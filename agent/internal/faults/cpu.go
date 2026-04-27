package faults

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
)

// TypeCPUStress is the stable string identifier for the CPU stress fault.
const TypeCPUStress = "cpu_stress"

// CPUStress is a Fault that loads the CPU to a configurable percentage
// for a configurable duration.
//
// It works by spawning N worker goroutines (one per logical CPU by
// default). Each worker alternates between a busy loop and a sleep,
// where the duty cycle is set so that the long-run CPU usage matches
// the configured intensity.
//
// Example: intensity=80 means each worker spends 80 ms computing and
// 20 ms sleeping per 100 ms slice.
type CPUStress struct {
	id        string
	intensity int           // 1..100
	workers   int           // number of busy goroutines
	duration  time.Duration // how long the fault runs
	startedAt time.Time

	cancel  context.CancelFunc
	wg      sync.WaitGroup
	mu      sync.Mutex
	stopped bool
}

// NewCPUStress builds a CPUStress fault from typed parameters.
//
// Recognized parameters:
//   - intensity (int, 1..100, default 80) — target CPU percentage per core
//   - workers   (int, default = runtime.NumCPU()) — number of stress goroutines
func NewCPUStress(duration time.Duration, params Params) (*CPUStress, error) {
	if duration <= 0 {
		return nil, fmt.Errorf("cpu_stress: duration must be > 0")
	}

	intensity := params.IntOr("intensity", 80)
	if intensity < 1 || intensity > 100 {
		return nil, fmt.Errorf("cpu_stress: intensity must be 1..100, got %d", intensity)
	}

	workers := params.IntOr("workers", runtime.NumCPU())
	if workers < 1 {
		workers = 1
	}

	return &CPUStress{
		id:        uuid.NewString(),
		intensity: intensity,
		workers:   workers,
		duration:  duration,
		startedAt: time.Now(),
	}, nil
}

// ID returns the unique fault identifier.
func (c *CPUStress) ID() string { return c.id }

// Type returns the stable type string.
func (c *CPUStress) Type() string { return TypeCPUStress }

// StartedAt returns the time the fault was started.
func (c *CPUStress) StartedAt() time.Time { return c.startedAt }

// Duration returns the configured fault duration.
func (c *CPUStress) Duration() time.Duration { return c.duration }

// Start spawns the worker goroutines and returns immediately.
func (c *CPUStress) Start(ctx context.Context) error {
	// Create a derived cancellable context so Stop can short-circuit
	// the workers even before the parent context is done.
	workerCtx, cancel := context.WithCancel(ctx)
	c.mu.Lock()
	c.cancel = cancel
	c.mu.Unlock()

	// 1 ms of "work then sleep" smoothes the CPU usage over time without
	// adding too much scheduling overhead.
	const slice = 100 * time.Millisecond
	workTime := time.Duration(c.intensity) * slice / 100
	sleepTime := slice - workTime

	for i := 0; i < c.workers; i++ {
		c.wg.Add(1)
		go c.run(workerCtx, workTime, sleepTime)
	}
	return nil
}

// Stop cancels the worker goroutines and waits for them to exit.
// Safe to call multiple times.
func (c *CPUStress) Stop(_ context.Context) error {
	c.mu.Lock()
	if c.stopped {
		c.mu.Unlock()
		return nil
	}
	c.stopped = true
	cancel := c.cancel
	c.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	c.wg.Wait()
	return nil
}

// run is the worker goroutine: alternate burning CPU and sleeping.
func (c *CPUStress) run(ctx context.Context, workTime, sleepTime time.Duration) {
	defer c.wg.Done()

	// Use a local accumulator so the compiler cannot optimize the loop
	// away. The exact computation is irrelevant; only the side effect
	// of keeping the CPU busy matters.
	var acc uint64

	for {
		// Busy-loop for workTime.
		end := time.Now().Add(workTime)
		for time.Now().Before(end) {
			acc++
			if ctx.Err() != nil {
				_ = acc // keep acc live so the compiler does not elide it
				return
			}
		}

		// Sleep for sleepTime, but wake early if the context is done.
		if sleepTime > 0 {
			select {
			case <-ctx.Done():
				return
			case <-time.After(sleepTime):
			}
		} else if ctx.Err() != nil {
			return
		}
	}
}
