package faults

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"
)

// ErrFaultNotFound is returned when a fault_id does not match any
// active fault in the registry.
var ErrFaultNotFound = errors.New("fault not found")

// Registry keeps track of currently active faults on the agent.
//
// The registry is responsible for:
//   - Thread-safe insertion and removal of faults
//   - Auto-stopping faults when their duration elapses
//   - Stopping all faults gracefully on shutdown
type Registry struct {
	mu     sync.Mutex
	faults map[string]Fault
	timers map[string]*time.Timer
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		faults: make(map[string]Fault),
		timers: make(map[string]*time.Timer),
	}
}

// Add registers and starts a fault. It schedules an automatic Stop after
// the fault's Duration. If Start fails, the fault is not registered.
func (r *Registry) Add(ctx context.Context, f Fault) error {
	if err := f.Start(ctx); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.faults[f.ID()] = f

	// Schedule auto-stop after the fault's duration.
	timer := time.AfterFunc(f.Duration(), func() {
		r.autoStop(f.ID())
	})
	r.timers[f.ID()] = timer

	slog.Info("fault registered",
		"fault_id", f.ID(),
		"type", f.Type(),
		"duration", f.Duration(),
	)
	return nil
}

// Stop stops the fault identified by id. Returns ErrFaultNotFound if no
// such fault is active.
func (r *Registry) Stop(ctx context.Context, id string) error {
	r.mu.Lock()
	f, ok := r.faults[id]
	if !ok {
		r.mu.Unlock()
		return ErrFaultNotFound
	}
	if t, hasTimer := r.timers[id]; hasTimer {
		t.Stop()
		delete(r.timers, id)
	}
	delete(r.faults, id)
	r.mu.Unlock()

	if err := f.Stop(ctx); err != nil {
		slog.Error("fault stop failed", "fault_id", id, "err", err)
		return err
	}
	slog.Info("fault stopped", "fault_id", id, "type", f.Type())
	return nil
}

// List returns a snapshot of all currently active faults.
func (r *Registry) List() []Fault {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := make([]Fault, 0, len(r.faults))
	for _, f := range r.faults {
		out = append(out, f)
	}
	return out
}

// StopAll stops every active fault. Used on shutdown.
func (r *Registry) StopAll(ctx context.Context) {
	r.mu.Lock()
	ids := make([]string, 0, len(r.faults))
	for id := range r.faults {
		ids = append(ids, id)
	}
	r.mu.Unlock()

	for _, id := range ids {
		_ = r.Stop(ctx, id)
	}
}

// autoStop is called by the timer when a fault's duration elapses.
func (r *Registry) autoStop(id string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := r.Stop(ctx, id); err != nil && !errors.Is(err, ErrFaultNotFound) {
		slog.Error("auto-stop failed", "fault_id", id, "err", err)
	}
}
