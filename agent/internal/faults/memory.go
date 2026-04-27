package faults

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// TypeMemoryPressure is the stable string identifier for the memory
// pressure fault.
const TypeMemoryPressure = "memory_pressure"

// pageSize is the granularity at which we touch the allocated buffer.
// 4 KiB matches the default Linux page size on x86_64 and arm64.
const pageSize = 4096

// MemoryPressure is a Fault that allocates a configurable amount of
// memory and keeps it resident by touching every page periodically.
//
// Without the periodic touch, the Linux kernel would happily swap the
// pages out to disk on a memory-constrained host, defeating the
// purpose of the fault. Touching every page forces the kernel to
// keep them in physical RAM.
type MemoryPressure struct {
	id        string
	sizeMB    int
	duration  time.Duration
	startedAt time.Time

	cancel  context.CancelFunc
	wg      sync.WaitGroup
	mu      sync.Mutex
	stopped bool

	// buffer holds the allocated memory. It is set to nil on Stop so
	// the GC can reclaim it.
	bufferMu sync.Mutex
	buffer   []byte
}

// NewMemoryPressure builds a MemoryPressure fault from typed parameters.
//
// Recognized parameters:
//   - size_mb (int, default 100) — number of megabytes to allocate
func NewMemoryPressure(duration time.Duration, params Params) (*MemoryPressure, error) {
	if duration <= 0 {
		return nil, fmt.Errorf("memory_pressure: duration must be > 0")
	}

	sizeMB := params.IntOr("size_mb", 100)
	if sizeMB < 1 {
		return nil, fmt.Errorf("memory_pressure: size_mb must be >= 1, got %d", sizeMB)
	}
	if sizeMB > 16384 {
		return nil, fmt.Errorf("memory_pressure: size_mb must be <= 16384 (16 GiB), got %d", sizeMB)
	}

	return &MemoryPressure{
		id:        uuid.NewString(),
		sizeMB:    sizeMB,
		duration:  duration,
		startedAt: time.Now(),
	}, nil
}

// ID returns the unique fault identifier.
func (m *MemoryPressure) ID() string { return m.id }

// Type returns the stable type string.
func (m *MemoryPressure) Type() string { return TypeMemoryPressure }

// StartedAt returns the time the fault was started.
func (m *MemoryPressure) StartedAt() time.Time { return m.startedAt }

// Duration returns the configured fault duration.
func (m *MemoryPressure) Duration() time.Duration { return m.duration }

// Start allocates the buffer and spawns the touch goroutine.
func (m *MemoryPressure) Start(ctx context.Context) error {
	bytes := m.sizeMB * 1024 * 1024
	buf := make([]byte, bytes)

	// Touch every page once at allocation time so the kernel actually
	// commits the physical memory (lazy allocation otherwise).
	for i := 0; i < bytes; i += pageSize {
		buf[i] = 1
	}

	m.bufferMu.Lock()
	m.buffer = buf
	m.bufferMu.Unlock()

	workerCtx, cancel := context.WithCancel(ctx)
	m.mu.Lock()
	m.cancel = cancel
	m.mu.Unlock()

	m.wg.Add(1)
	go m.touchLoop(workerCtx)
	return nil
}

// Stop cancels the touch loop and releases the buffer for the GC.
// Safe to call multiple times.
func (m *MemoryPressure) Stop(_ context.Context) error {
	m.mu.Lock()
	if m.stopped {
		m.mu.Unlock()
		return nil
	}
	m.stopped = true
	cancel := m.cancel
	m.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	m.wg.Wait()

	// Drop the reference so the GC can reclaim the buffer.
	m.bufferMu.Lock()
	m.buffer = nil
	m.bufferMu.Unlock()
	return nil
}

// touchLoop wakes up every 200 ms and writes one byte to every page
// of the buffer. This prevents the kernel from swapping the pages
// out under memory pressure.
func (m *MemoryPressure) touchLoop(ctx context.Context) {
	defer m.wg.Done()

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	// Use a rotating "tag" byte so each touch is observable in case
	// we ever read the buffer back during testing.
	tag := byte(1)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.bufferMu.Lock()
			buf := m.buffer
			if buf == nil {
				m.bufferMu.Unlock()
				return
			}
			for i := 0; i < len(buf); i += pageSize {
				buf[i] = tag
			}
			m.bufferMu.Unlock()
			tag++
		}
	}
}
