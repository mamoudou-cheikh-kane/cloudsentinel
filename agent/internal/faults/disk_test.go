package faults

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewDiskFill_Validation(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		params   Params
		wantErr  bool
	}{
		{
			name:     "valid default",
			duration: 1 * time.Second,
			params:   Params{Raw: nil},
			wantErr:  false,
		},
		{
			name:     "zero duration rejected",
			duration: 0,
			params:   Params{Raw: nil},
			wantErr:  true,
		},
		{
			name:     "negative duration rejected",
			duration: -1 * time.Second,
			params:   Params{Raw: nil},
			wantErr:  true,
		},
		{
			name:     "size_mb 0 rejected",
			duration: 1 * time.Second,
			params:   Params{Raw: map[string]string{"size_mb": "0"}},
			wantErr:  true,
		},
		{
			name:     "size_mb negative rejected",
			duration: 1 * time.Second,
			params:   Params{Raw: map[string]string{"size_mb": "-100"}},
			wantErr:  true,
		},
		{
			name:     "size_mb too large rejected",
			duration: 1 * time.Second,
			params:   Params{Raw: map[string]string{"size_mb": "2000000"}},
			wantErr:  true,
		},
		{
			name:     "valid custom size and directory",
			duration: 1 * time.Second,
			params: Params{Raw: map[string]string{
				"size_mb":   "10",
				"directory": "/tmp/test-cs",
			}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := NewDiskFill(tt.duration, tt.params)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if f.ID() == "" {
				t.Errorf("ID should not be empty")
			}
			if f.Type() != TypeDiskFill {
				t.Errorf("Type() = %q, want %q", f.Type(), TypeDiskFill)
			}
			if f.Duration() != tt.duration {
				t.Errorf("Duration() = %v, want %v", f.Duration(), tt.duration)
			}
		})
	}
}

func TestDiskFill_StartStop(t *testing.T) {
	dir := t.TempDir()
	f, err := NewDiskFill(5*time.Second, Params{
		Raw: map[string]string{"size_mb": "5", "directory": dir},
	})
	if err != nil {
		t.Fatalf("NewDiskFill: %v", err)
	}

	ctx := context.Background()
	if err := f.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// File path is set after Start.
	if f.Path() == "" {
		t.Errorf("Path() should not be empty after Start")
	}

	time.Sleep(50 * time.Millisecond)

	if err := f.Stop(ctx); err != nil {
		t.Fatalf("Stop: %v", err)
	}
}

// TestDiskFill_ActuallyFillsFile proves the fault really writes the
// requested amount of bytes to disk.
func TestDiskFill_ActuallyFillsFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping disk fill check in -short mode")
	}

	dir := t.TempDir()
	const sizeMB = 10
	const sizeBytes = sizeMB * 1024 * 1024

	f, err := NewDiskFill(30*time.Second, Params{
		Raw: map[string]string{"size_mb": "10", "directory": dir},
	})
	if err != nil {
		t.Fatalf("NewDiskFill: %v", err)
	}

	ctx := context.Background()
	if err := f.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Wait long enough for the fill loop to complete.
	// 10 MiB at ~5ms per MiB plus filesystem sync ~ a few hundred ms.
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		stat, err := os.Stat(f.Path())
		if err == nil && stat.Size() >= int64(sizeBytes) {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	stat, err := os.Stat(f.Path())
	if err != nil {
		t.Fatalf("file should exist on disk: %v", err)
	}

	t.Logf("File size: %d bytes (expected %d bytes)", stat.Size(), sizeBytes)

	if stat.Size() < int64(sizeBytes) {
		t.Errorf(
			"file is only %d bytes, expected at least %d",
			stat.Size(), sizeBytes,
		)
	}

	if err := f.Stop(ctx); err != nil {
		t.Fatalf("Stop: %v", err)
	}
}

func TestDiskFill_StopRemovesFile(t *testing.T) {
	dir := t.TempDir()
	f, err := NewDiskFill(30*time.Second, Params{
		Raw: map[string]string{"size_mb": "2", "directory": dir},
	})
	if err != nil {
		t.Fatalf("NewDiskFill: %v", err)
	}

	ctx := context.Background()
	if err := f.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Wait for the file to exist.
	deadline := time.Now().Add(2 * time.Second)
	path := f.Path()
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file should exist before Stop: %v", err)
	}

	if err := f.Stop(ctx); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	// After Stop, the file must be gone.
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("file should be removed after Stop, got err=%v", err)
	}
}

func TestDiskFill_StopIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	f, err := NewDiskFill(10*time.Second, Params{
		Raw: map[string]string{"size_mb": "1", "directory": dir},
	})
	if err != nil {
		t.Fatalf("NewDiskFill: %v", err)
	}

	ctx := context.Background()
	if err := f.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}

	if err := f.Stop(ctx); err != nil {
		t.Fatalf("first Stop: %v", err)
	}
	if err := f.Stop(ctx); err != nil {
		t.Fatalf("second Stop: %v", err)
	}
	if err := f.Stop(ctx); err != nil {
		t.Fatalf("third Stop: %v", err)
	}
}

func TestDiskFill_DefaultDirectoryCreated(t *testing.T) {
	// Use a custom dir so we don't pollute the real /tmp.
	dir := filepath.Join(t.TempDir(), "nested", "subdir")

	f, err := NewDiskFill(5*time.Second, Params{
		Raw: map[string]string{"size_mb": "1", "directory": dir},
	})
	if err != nil {
		t.Fatalf("NewDiskFill: %v", err)
	}

	ctx := context.Background()
	if err := f.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer f.Stop(ctx)

	if _, err := os.Stat(dir); err != nil {
		t.Errorf("Start should have created the directory %q: %v", dir, err)
	}
}
