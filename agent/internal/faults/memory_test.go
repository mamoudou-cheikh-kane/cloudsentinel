package faults

import (
	"context"
	"runtime"
	"testing"
	"time"
)

func TestNewMemoryPressure_Validation(t *testing.T) {
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
			params:   Params{Raw: map[string]string{"size_mb": "20000"}},
			wantErr:  true,
		},
		{
			name:     "valid custom size",
			duration: 1 * time.Second,
			params:   Params{Raw: map[string]string{"size_mb": "50"}},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := NewMemoryPressure(tt.duration, tt.params)
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
			if f.Type() != TypeMemoryPressure {
				t.Errorf("Type() = %q, want %q", f.Type(), TypeMemoryPressure)
			}
			if f.Duration() != tt.duration {
				t.Errorf("Duration() = %v, want %v", f.Duration(), tt.duration)
			}
		})
	}
}

func TestMemoryPressure_StartStop(t *testing.T) {
	f, err := NewMemoryPressure(5*time.Second, Params{
		Raw: map[string]string{"size_mb": "20"},
	})
	if err != nil {
		t.Fatalf("NewMemoryPressure: %v", err)
	}

	ctx := context.Background()
	if err := f.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	if err := f.Stop(ctx); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
}

// TestMemoryPressure_ActuallyAllocates proves the fault actually
// allocates physical memory by sampling runtime.MemStats before and
// after Start.
func TestMemoryPressure_ActuallyAllocates(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping memory allocation check in -short mode")
	}

	const sizeMB = 100
	const sizeBytes = sizeMB * 1024 * 1024

	// Read memory stats before allocation.
	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	f, err := NewMemoryPressure(5*time.Second, Params{
		Raw: map[string]string{"size_mb": "100"},
	})
	if err != nil {
		t.Fatalf("NewMemoryPressure: %v", err)
	}

	ctx := context.Background()
	if err := f.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Give the touch loop time to run at least once.
	time.Sleep(300 * time.Millisecond)

	// Read memory stats while the fault is active.
	var during runtime.MemStats
	runtime.ReadMemStats(&during)

	allocated := during.Alloc - before.Alloc

	t.Logf("Before: Alloc=%d MB", before.Alloc/(1024*1024))
	t.Logf("During: Alloc=%d MB", during.Alloc/(1024*1024))
	t.Logf("Delta:  %d MB (expected ~%d MB)", allocated/(1024*1024), sizeMB)

	// The buffer is at least sizeBytes; we allow 80% as a safety
	// margin against minor GC fluctuations.
	minExpected := uint64(float64(sizeBytes) * 0.8)
	if allocated < minExpected {
		t.Errorf(
			"allocated %d bytes, expected at least %d bytes",
			allocated, minExpected,
		)
	}

	if err := f.Stop(ctx); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	// After Stop and a GC cycle, the memory should be reclaimed.
	runtime.GC()
	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	t.Logf("After Stop+GC: Alloc=%d MB", after.Alloc/(1024*1024))
}

func TestMemoryPressure_StopIsIdempotent(t *testing.T) {
	f, err := NewMemoryPressure(10*time.Second, Params{
		Raw: map[string]string{"size_mb": "10"},
	})
	if err != nil {
		t.Fatalf("NewMemoryPressure: %v", err)
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
