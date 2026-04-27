package faults

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// TypeNetworkLatency is the stable string identifier for the network
// latency fault.
const TypeNetworkLatency = "network_latency"

// defaultIface is the interface NetworkLatency targets when no
// "interface" parameter is provided.
const defaultIface = "eth0"

// TCExecutor abstracts the call to `tc` (traffic control). The
// production implementation in network_tc.go runs the real binary;
// tests use a mock that records the calls.
//
// This indirection lets unit tests verify the right command is
// issued without needing CAP_NET_ADMIN or even tc on the host.
type TCExecutor interface {
	// AddDelay installs a netem qdisc on iface that adds the given
	// fixed delay (and optional jitter) to outbound packets.
	AddDelay(ctx context.Context, iface string, delay, jitter time.Duration) error

	// RemoveDelay removes the root qdisc on iface (best-effort).
	RemoveDelay(ctx context.Context, iface string) error
}

// NetworkLatency is a Fault that adds a configurable latency to the
// outbound traffic of a network interface using Linux tc/netem.
//
// The fault delegates the actual tc invocations to a TCExecutor so
// it stays unit-testable on machines that have neither tc nor
// CAP_NET_ADMIN.
type NetworkLatency struct {
	id        string
	iface     string
	delay     time.Duration
	jitter    time.Duration
	duration  time.Duration
	startedAt time.Time

	exec TCExecutor

	mu      sync.Mutex
	stopped bool
	applied bool // true once AddDelay returned successfully
}

// NewNetworkLatency builds a NetworkLatency fault with the production
// TCExecutor. Use NewNetworkLatencyWithExecutor in tests.
//
// Recognized parameters:
//   - delay_ms  (int, default 100) — added latency in milliseconds
//   - jitter_ms (int, default 0)   — random jitter around delay
//   - interface (string, default "eth0") — network interface to target
func NewNetworkLatency(duration time.Duration, params Params) (*NetworkLatency, error) {
	return NewNetworkLatencyWithExecutor(duration, params, &realTCExecutor{})
}

// NewNetworkLatencyWithExecutor is the test-friendly constructor that
// lets the caller inject a custom TCExecutor.
func NewNetworkLatencyWithExecutor(
	duration time.Duration,
	params Params,
	exec TCExecutor,
) (*NetworkLatency, error) {
	if duration <= 0 {
		return nil, fmt.Errorf("network_latency: duration must be > 0")
	}

	delayMs := params.IntOr("delay_ms", 100)
	if delayMs < 1 {
		return nil, fmt.Errorf("network_latency: delay_ms must be >= 1, got %d", delayMs)
	}
	if delayMs > 60000 {
		return nil, fmt.Errorf("network_latency: delay_ms must be <= 60000 (60 s), got %d", delayMs)
	}

	jitterMs := params.IntOr("jitter_ms", 0)
	if jitterMs < 0 {
		return nil, fmt.Errorf("network_latency: jitter_ms must be >= 0, got %d", jitterMs)
	}
	if jitterMs > delayMs {
		return nil, fmt.Errorf(
			"network_latency: jitter_ms (%d) cannot exceed delay_ms (%d)",
			jitterMs, delayMs,
		)
	}

	iface := params.StringOr("interface", defaultIface)

	if exec == nil {
		return nil, fmt.Errorf("network_latency: TCExecutor must not be nil")
	}

	return &NetworkLatency{
		id:        uuid.NewString(),
		iface:     iface,
		delay:     time.Duration(delayMs) * time.Millisecond,
		jitter:    time.Duration(jitterMs) * time.Millisecond,
		duration:  duration,
		startedAt: time.Now(),
		exec:      exec,
	}, nil
}

// ID returns the unique fault identifier.
func (n *NetworkLatency) ID() string { return n.id }

// Type returns the stable type string.
func (n *NetworkLatency) Type() string { return TypeNetworkLatency }

// StartedAt returns the time the fault was started.
func (n *NetworkLatency) StartedAt() time.Time { return n.startedAt }

// Duration returns the configured fault duration.
func (n *NetworkLatency) Duration() time.Duration { return n.duration }

// Iface returns the network interface this fault is targeting.
func (n *NetworkLatency) Iface() string { return n.iface }

// Delay returns the configured delay.
func (n *NetworkLatency) Delay() time.Duration { return n.delay }

// Jitter returns the configured jitter.
func (n *NetworkLatency) Jitter() time.Duration { return n.jitter }

// Start applies the netem qdisc on the configured interface.
func (n *NetworkLatency) Start(ctx context.Context) error {
	if err := n.exec.AddDelay(ctx, n.iface, n.delay, n.jitter); err != nil {
		return fmt.Errorf("network_latency: AddDelay failed: %w", err)
	}
	n.mu.Lock()
	n.applied = true
	n.mu.Unlock()
	return nil
}

// Stop removes the qdisc. Safe to call multiple times. If Start never
// succeeded, Stop is a no-op.
func (n *NetworkLatency) Stop(ctx context.Context) error {
	n.mu.Lock()
	if n.stopped {
		n.mu.Unlock()
		return nil
	}
	n.stopped = true
	applied := n.applied
	n.mu.Unlock()

	if !applied {
		return nil
	}
	if err := n.exec.RemoveDelay(ctx, n.iface); err != nil {
		return fmt.Errorf("network_latency: RemoveDelay failed: %w", err)
	}
	return nil
}
