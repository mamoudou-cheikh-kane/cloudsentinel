package faults

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// mockTCExecutor records the AddDelay/RemoveDelay calls so tests can
// assert on them without invoking real tc commands.
type mockTCExecutor struct {
	mu sync.Mutex

	addCalls    []addCall
	removeCalls []string

	// errors to inject
	addErr    error
	removeErr error
}

type addCall struct {
	iface  string
	delay  time.Duration
	jitter time.Duration
}

func (m *mockTCExecutor) AddDelay(_ context.Context, iface string, delay, jitter time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.addErr != nil {
		return m.addErr
	}
	m.addCalls = append(m.addCalls, addCall{iface: iface, delay: delay, jitter: jitter})
	return nil
}

func (m *mockTCExecutor) RemoveDelay(_ context.Context, iface string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.removeErr != nil {
		return m.removeErr
	}
	m.removeCalls = append(m.removeCalls, iface)
	return nil
}

func (m *mockTCExecutor) addCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.addCalls)
}

func (m *mockTCExecutor) removeCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.removeCalls)
}

func TestNewNetworkLatency_Validation(t *testing.T) {
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
			name:     "delay_ms 0 rejected",
			duration: 1 * time.Second,
			params:   Params{Raw: map[string]string{"delay_ms": "0"}},
			wantErr:  true,
		},
		{
			name:     "delay_ms too large rejected",
			duration: 1 * time.Second,
			params:   Params{Raw: map[string]string{"delay_ms": "70000"}},
			wantErr:  true,
		},
		{
			name:     "jitter_ms negative rejected",
			duration: 1 * time.Second,
			params:   Params{Raw: map[string]string{"jitter_ms": "-10"}},
			wantErr:  true,
		},
		{
			name:     "jitter_ms larger than delay rejected",
			duration: 1 * time.Second,
			params:   Params{Raw: map[string]string{"delay_ms": "100", "jitter_ms": "200"}},
			wantErr:  true,
		},
		{
			name:     "valid custom params",
			duration: 1 * time.Second,
			params: Params{Raw: map[string]string{
				"delay_ms":  "200",
				"jitter_ms": "20",
				"interface": "lo",
			}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := NewNetworkLatencyWithExecutor(tt.duration, tt.params, &mockTCExecutor{})
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
			if f.Type() != TypeNetworkLatency {
				t.Errorf("Type() = %q, want %q", f.Type(), TypeNetworkLatency)
			}
			if f.Duration() != tt.duration {
				t.Errorf("Duration() = %v, want %v", f.Duration(), tt.duration)
			}
		})
	}
}

func TestNewNetworkLatency_NilExecutorRejected(t *testing.T) {
	_, err := NewNetworkLatencyWithExecutor(1*time.Second, Params{}, nil)
	if err == nil {
		t.Fatal("nil executor should be rejected")
	}
}

func TestNetworkLatency_StartCallsTCWithRightArgs(t *testing.T) {
	mock := &mockTCExecutor{}
	f, err := NewNetworkLatencyWithExecutor(5*time.Second, Params{
		Raw: map[string]string{
			"delay_ms":  "150",
			"jitter_ms": "20",
			"interface": "lo",
		},
	}, mock)
	if err != nil {
		t.Fatalf("NewNetworkLatencyWithExecutor: %v", err)
	}

	if err := f.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}

	if mock.addCallCount() != 1 {
		t.Fatalf("expected 1 AddDelay call, got %d", mock.addCallCount())
	}
	call := mock.addCalls[0]
	if call.iface != "lo" {
		t.Errorf("iface = %q, want %q", call.iface, "lo")
	}
	if call.delay != 150*time.Millisecond {
		t.Errorf("delay = %v, want 150ms", call.delay)
	}
	if call.jitter != 20*time.Millisecond {
		t.Errorf("jitter = %v, want 20ms", call.jitter)
	}
}

func TestNetworkLatency_StopCallsRemove(t *testing.T) {
	mock := &mockTCExecutor{}
	f, _ := NewNetworkLatencyWithExecutor(5*time.Second, Params{
		Raw: map[string]string{"interface": "lo"},
	}, mock)

	ctx := context.Background()
	if err := f.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := f.Stop(ctx); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	if mock.removeCallCount() != 1 {
		t.Errorf("expected 1 RemoveDelay call, got %d", mock.removeCallCount())
	}
	if mock.removeCalls[0] != "lo" {
		t.Errorf("Remove was called with iface=%q, want %q", mock.removeCalls[0], "lo")
	}
}

func TestNetworkLatency_StopWithoutStartIsNoop(t *testing.T) {
	mock := &mockTCExecutor{}
	f, _ := NewNetworkLatencyWithExecutor(5*time.Second, Params{}, mock)

	if err := f.Stop(context.Background()); err != nil {
		t.Errorf("Stop without Start should not error: %v", err)
	}
	if mock.removeCallCount() != 0 {
		t.Errorf("Stop without Start should not call RemoveDelay, got %d calls", mock.removeCallCount())
	}
}

func TestNetworkLatency_StopIsIdempotent(t *testing.T) {
	mock := &mockTCExecutor{}
	f, _ := NewNetworkLatencyWithExecutor(5*time.Second, Params{}, mock)

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

	// Even after 3 Stop() calls, RemoveDelay must have been called only once.
	if mock.removeCallCount() != 1 {
		t.Errorf("RemoveDelay should be called exactly once, got %d", mock.removeCallCount())
	}
}

func TestNetworkLatency_StartFailureLeavesNoArtifact(t *testing.T) {
	mock := &mockTCExecutor{addErr: errors.New("simulated tc failure")}
	f, _ := NewNetworkLatencyWithExecutor(5*time.Second, Params{}, mock)

	if err := f.Start(context.Background()); err == nil {
		t.Fatal("Start should propagate AddDelay error")
	}

	// If Start failed, Stop must NOT call RemoveDelay (nothing to clean up).
	if err := f.Stop(context.Background()); err != nil {
		t.Errorf("Stop after failed Start should not error: %v", err)
	}
	if mock.removeCallCount() != 0 {
		t.Errorf("RemoveDelay should not be called when Start failed, got %d", mock.removeCallCount())
	}
}
