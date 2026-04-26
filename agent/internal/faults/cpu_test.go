package faults

import (
	"context"
	"syscall"
	"testing"
	"time"
)

func TestNewCPUStress_Validation(t *testing.T) {
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
			name:     "intensity 0 rejected",
			duration: 1 * time.Second,
			params:   Params{Raw: map[string]string{"intensity": "0"}},
			wantErr:  true,
		},
		{
			name:     "intensity 101 rejected",
			duration: 1 * time.Second,
			params:   Params{Raw: map[string]string{"intensity": "101"}},
			wantErr:  true,
		},
		{
			name:     "valid custom params",
			duration: 1 * time.Second,
			params:   Params{Raw: map[string]string{"intensity": "50", "workers": "2"}},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := NewCPUStress(tt.duration, tt.params)
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
			if f.Type() != TypeCPUStress {
				t.Errorf("Type() = %q, want %q", f.Type(), TypeCPUStress)
			}
			if f.Duration() != tt.duration {
				t.Errorf("Duration() = %v, want %v", f.Duration(), tt.duration)
			}
		})
	}
}

func TestCPUStress_StartStop(t *testing.T) {
	f, err := NewCPUStress(5*time.Second, Params{
		Raw: map[string]string{"intensity": "50", "workers": "2"},
	})
	if err != nil {
		t.Fatalf("NewCPUStress: %v", err)
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

// TestCPUStress_ActuallyStressesCPU proves the fault actually consumes
// CPU cycles. Uses syscall.Getrusage for an accurate measurement on
// Linux (the target platform — agent runs in a Linux container).
func TestCPUStress_ActuallyStressesCPU(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CPU stress check in -short mode")
	}

	f, err := NewCPUStress(2*time.Second, Params{
		Raw: map[string]string{"intensity": "80", "workers": "2"},
	})
	if err != nil {
		t.Fatalf("NewCPUStress: %v", err)
	}

	startCPU := selfCPU(t)

	ctx := context.Background()
	if err := f.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}

	time.Sleep(1 * time.Second)

	if err := f.Stop(ctx); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	endCPU := selfCPU(t)
	delta := endCPU - startCPU

	const minExpected = 1.0
	if delta < minExpected {
		t.Errorf(
			"CPU stress consumed only %.2f CPU-seconds, expected at least %.2f",
			delta, minExpected,
		)
	}
	t.Logf("CPU stress consumed %.2f CPU-seconds in 1 wall second", delta)
}

func TestCPUStress_StopIsIdempotent(t *testing.T) {
	f, err := NewCPUStress(10*time.Second, Params{
		Raw: map[string]string{"workers": "1"},
	})
	if err != nil {
		t.Fatalf("NewCPUStress: %v", err)
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

// selfCPU returns the total CPU time consumed by this process so far
// in CPU-seconds, using POSIX getrusage. The agent always runs on
// Linux, so we do not need a portable fallback.
func selfCPU(t *testing.T) float64 {
	t.Helper()
	var ru syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &ru); err != nil {
		t.Fatalf("getrusage: %v", err)
	}
	user := float64(ru.Utime.Sec) + float64(ru.Utime.Usec)/1e6
	sys := float64(ru.Stime.Sec) + float64(ru.Stime.Usec)/1e6
	return user + sys
}
