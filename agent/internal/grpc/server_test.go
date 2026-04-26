package grpc

import (
	"context"
	"testing"

	pb "github.com/mamoudou-cheikh-kane/cloudsentinel/agent/internal/pb"
)

// lightCPUParams returns parameters for a CPU stress fault that puts
// almost no real load on the test machine: 1 worker at 5% intensity.
// Tests should always use these to keep the suite fast and CI-friendly.
func lightCPUParams() map[string]string {
	return map[string]string{"intensity": "5", "workers": "1"}
}

func TestHealth(t *testing.T) {
	s := NewServer("test-node", "0.1.0")
	resp, err := s.Health(context.Background(), &pb.HealthRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Node != "test-node" {
		t.Errorf("expected node=test-node, got %q", resp.Node)
	}
	if resp.Version != "0.1.0" {
		t.Errorf("expected version=0.1.0, got %q", resp.Version)
	}
	if resp.UptimeSeconds < 0 {
		t.Errorf("uptime must be >= 0")
	}
}

func TestInjectFault_Unspecified(t *testing.T) {
	s := NewServer("n", "v")
	resp, _ := s.InjectFault(context.Background(), &pb.InjectFaultRequest{
		Type: pb.FaultType_FAULT_TYPE_UNSPECIFIED,
	})
	if resp.Accepted {
		t.Errorf("unspecified fault should not be accepted")
	}
}

func TestInjectFault_ZeroDuration(t *testing.T) {
	s := NewServer("n", "v")
	resp, _ := s.InjectFault(context.Background(), &pb.InjectFaultRequest{
		Type:            pb.FaultType_FAULT_TYPE_CPU_STRESS,
		DurationSeconds: 0,
		Parameters:      lightCPUParams(),
	})
	if resp.Accepted {
		t.Errorf("zero duration should not be accepted")
	}
}

func TestInjectFault_NotYetImplemented(t *testing.T) {
	s := NewServer("n", "v")
	tests := []pb.FaultType{
		pb.FaultType_FAULT_TYPE_NETWORK_LATENCY,
		pb.FaultType_FAULT_TYPE_MEMORY_PRESSURE,
		pb.FaultType_FAULT_TYPE_DISK_FILL,
	}
	for _, ft := range tests {
		t.Run(ft.String(), func(t *testing.T) {
			resp, _ := s.InjectFault(context.Background(), &pb.InjectFaultRequest{
				Type:            ft,
				DurationSeconds: 5,
			})
			if resp.Accepted {
				t.Errorf("%s is not implemented yet, should be rejected", ft)
			}
		})
	}
}

func TestInjectFault_CPUStress_Success(t *testing.T) {
	s := NewServer("n", "v")
	ctx := context.Background()

	resp, _ := s.InjectFault(ctx, &pb.InjectFaultRequest{
		Type:            pb.FaultType_FAULT_TYPE_CPU_STRESS,
		DurationSeconds: 5,
		Parameters:      lightCPUParams(),
	})
	if !resp.Accepted {
		t.Fatalf("CPU stress should be accepted: %s", resp.Message)
	}
	if resp.FaultId == "" {
		t.Errorf("fault_id should not be empty")
	}

	// Cleanup so the worker goroutine does not outlive the test.
	t.Cleanup(func() {
		_, _ = s.Rollback(ctx, &pb.RollbackRequest{FaultId: resp.FaultId})
	})
}

func TestRollback_NotFound(t *testing.T) {
	s := NewServer("n", "v")
	resp, _ := s.Rollback(context.Background(), &pb.RollbackRequest{
		FaultId: "inexistant",
	})
	if resp.Success {
		t.Errorf("rollback of unknown fault should fail")
	}
}

func TestRollback_Success(t *testing.T) {
	s := NewServer("n", "v")
	ctx := context.Background()

	injectResp, _ := s.InjectFault(ctx, &pb.InjectFaultRequest{
		Type:            pb.FaultType_FAULT_TYPE_CPU_STRESS,
		DurationSeconds: 60,
		Parameters:      lightCPUParams(),
	})
	if !injectResp.Accepted {
		t.Fatalf("inject must succeed before rollback test: %s", injectResp.Message)
	}

	rollResp, _ := s.Rollback(ctx, &pb.RollbackRequest{
		FaultId: injectResp.FaultId,
	})
	if !rollResp.Success {
		t.Errorf("rollback should succeed: %s", rollResp.Message)
	}
}

func TestGetStatus(t *testing.T) {
	s := NewServer("n", "v")
	ctx := context.Background()

	r1, _ := s.InjectFault(ctx, &pb.InjectFaultRequest{
		Type:            pb.FaultType_FAULT_TYPE_CPU_STRESS,
		DurationSeconds: 60,
		Parameters:      lightCPUParams(),
	})
	r2, _ := s.InjectFault(ctx, &pb.InjectFaultRequest{
		Type:            pb.FaultType_FAULT_TYPE_CPU_STRESS,
		DurationSeconds: 60,
		Parameters:      lightCPUParams(),
	})

	t.Cleanup(func() {
		_, _ = s.Rollback(ctx, &pb.RollbackRequest{FaultId: r1.FaultId})
		_, _ = s.Rollback(ctx, &pb.RollbackRequest{FaultId: r2.FaultId})
	})

	resp, _ := s.GetStatus(ctx, &pb.StatusRequest{})
	if len(resp.ActiveFaults) != 2 {
		t.Errorf("expected 2 active faults, got %d", len(resp.ActiveFaults))
	}

	// Each active fault should report the right type and a positive
	// remaining time.
	for _, af := range resp.ActiveFaults {
		if af.Type != pb.FaultType_FAULT_TYPE_CPU_STRESS {
			t.Errorf("expected CPU_STRESS, got %s", af.Type)
		}
		if af.RemainingSeconds <= 0 {
			t.Errorf("remaining_seconds should be > 0, got %d", af.RemainingSeconds)
		}
	}
}

func TestShutdown_StopsAllFaults(t *testing.T) {
	s := NewServer("n", "v")
	ctx := context.Background()

	_, _ = s.InjectFault(ctx, &pb.InjectFaultRequest{
		Type:            pb.FaultType_FAULT_TYPE_CPU_STRESS,
		DurationSeconds: 60,
		Parameters:      lightCPUParams(),
	})
	_, _ = s.InjectFault(ctx, &pb.InjectFaultRequest{
		Type:            pb.FaultType_FAULT_TYPE_CPU_STRESS,
		DurationSeconds: 60,
		Parameters:      lightCPUParams(),
	})

	s.Shutdown(ctx)

	resp, _ := s.GetStatus(ctx, &pb.StatusRequest{})
	if len(resp.ActiveFaults) != 0 {
		t.Errorf("after Shutdown, expected 0 faults, got %d", len(resp.ActiveFaults))
	}
}
