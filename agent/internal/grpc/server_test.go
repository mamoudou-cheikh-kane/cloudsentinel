package grpc

import (
	"context"
	"testing"

	pb "github.com/mamoudou-cheikh-kane/cloudsentinel/agent/internal/pb"
)

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

func TestInjectFault_Success(t *testing.T) {
	s := NewServer("n", "v")
	resp, _ := s.InjectFault(context.Background(), &pb.InjectFaultRequest{
		Type:            pb.FaultType_FAULT_TYPE_CPU_STRESS,
		DurationSeconds: 30,
	})
	if !resp.Accepted {
		t.Errorf("valid fault should be accepted: %s", resp.Message)
	}
	if resp.FaultId == "" {
		t.Errorf("fault_id should not be empty")
	}
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
	injectResp, _ := s.InjectFault(context.Background(), &pb.InjectFaultRequest{
		Type:            pb.FaultType_FAULT_TYPE_NETWORK_LATENCY,
		DurationSeconds: 60,
	})

	rollResp, _ := s.Rollback(context.Background(), &pb.RollbackRequest{
		FaultId: injectResp.FaultId,
	})
	if !rollResp.Success {
		t.Errorf("rollback should succeed: %s", rollResp.Message)
	}
}

func TestGetStatus(t *testing.T) {
	s := NewServer("n", "v")
	s.InjectFault(context.Background(), &pb.InjectFaultRequest{
		Type:            pb.FaultType_FAULT_TYPE_MEMORY_PRESSURE,
		DurationSeconds: 120,
	})
	s.InjectFault(context.Background(), &pb.InjectFaultRequest{
		Type:            pb.FaultType_FAULT_TYPE_DISK_FILL,
		DurationSeconds: 60,
	})

	resp, _ := s.GetStatus(context.Background(), &pb.StatusRequest{})
	if len(resp.ActiveFaults) != 2 {
		t.Errorf("expected 2 active faults, got %d", len(resp.ActiveFaults))
	}
}
