// Package faults defines the Fault interface and the in-memory registry
// that tracks active fault injections on the agent.
package faults

import (
	"context"
	"fmt"
	"time"
)

// Fault is the interface every fault implementation must satisfy.
//
// A Fault is a self-contained unit that knows how to apply a perturbation
// to the local node (Start) and how to revert it (Stop).
//
// Implementations must be safe to use across goroutines: Start and Stop
// can be called from different goroutines, and Stop must be idempotent
// (calling it twice is safe and is a no-op the second time).
type Fault interface {
	// ID returns a unique identifier for this fault instance.
	ID() string

	// Type returns the fault type as a stable string (e.g. "cpu_stress").
	Type() string

	// Start applies the perturbation. It must return quickly: any
	// long-running work (e.g. busy loop) belongs in a goroutine spawned
	// by Start.
	Start(ctx context.Context) error

	// Stop reverts the perturbation. It must be idempotent.
	Stop(ctx context.Context) error

	// StartedAt returns the time the fault was started.
	StartedAt() time.Time

	// Duration returns the configured duration of this fault.
	Duration() time.Duration
}

// Params carries the typed parameters for a fault, parsed from the
// gRPC `parameters` map.
type Params struct {
	// Raw is the untouched key/value map from the gRPC request.
	Raw map[string]string
}

// IntOr returns Raw[key] parsed as int, or fallback when the key is
// missing or unparseable.
func (p Params) IntOr(key string, fallback int) int {
	v, ok := p.Raw[key]
	if !ok {
		return fallback
	}
	var n int
	if _, err := fmt.Sscanf(v, "%d", &n); err != nil {
		return fallback
	}
	return n
}

// StringOr returns Raw[key], or fallback when the key is missing.
func (p Params) StringOr(key, fallback string) string {
	if v, ok := p.Raw[key]; ok && v != "" {
		return v
	}
	return fallback
}
