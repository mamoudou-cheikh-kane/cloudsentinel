package faults

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// realTCExecutor implements TCExecutor by invoking the real `tc`
// binary. It is the default executor used in production.
//
// Requires:
//   - the tc binary in PATH (provided by iproute2)
//   - the kernel module `sch_netem` loaded (default on most distros)
//   - the CAP_NET_ADMIN capability for the calling process
type realTCExecutor struct{}

// AddDelay runs:
//
//	tc qdisc add dev <iface> root netem delay <delay>ms [<jitter>ms]
func (r *realTCExecutor) AddDelay(
	ctx context.Context,
	iface string,
	delay, jitter time.Duration,
) error {
	delayMs := fmt.Sprintf("%dms", delay.Milliseconds())
	args := []string{"qdisc", "add", "dev", iface, "root", "netem", "delay", delayMs}
	if jitter > 0 {
		args = append(args, fmt.Sprintf("%dms", jitter.Milliseconds()))
	}

	cmd := exec.CommandContext(ctx, "tc", args...) //nolint:gosec // args are sanitized
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tc add: %w (output: %s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// RemoveDelay runs:
//
//	tc qdisc del dev <iface> root
//
// If the qdisc does not exist (e.g. it was never added), tc returns
// a non-zero exit code; we swallow that case so Stop stays idempotent.
func (r *realTCExecutor) RemoveDelay(ctx context.Context, iface string) error {
	cmd := exec.CommandContext(ctx, "tc", "qdisc", "del", "dev", iface, "root") //nolint:gosec
	out, err := cmd.CombinedOutput()
	if err != nil {
		s := strings.TrimSpace(string(out))
		// "Cannot find device" or "Cannot delete qdisc with handle of zero"
		// mean we are already in the desired state.
		if strings.Contains(s, "Cannot delete") || strings.Contains(s, "No such") {
			return nil
		}
		return fmt.Errorf("tc del: %w (output: %s)", err, s)
	}
	return nil
}
