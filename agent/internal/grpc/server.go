// Package grpc provides the CloudSentinel agent gRPC server.
package grpc

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/mamoudou-cheikh-kane/cloudsentinel/agent/internal/faults"
	pb "github.com/mamoudou-cheikh-kane/cloudsentinel/agent/internal/pb"
)

// Server implements pb.AgentServiceServer.
//
// The server is intentionally thin: it only knows how to translate
// protobuf requests into Fault objects and delegate execution to the
// Registry. All the real work (busy loops, memory allocation, tc rules)
// lives in the faults package.
type Server struct {
	pb.UnimplementedAgentServiceServer

	nodeName  string
	version   string
	startedAt time.Time
	registry  *faults.Registry
}

// NewServer creates a new gRPC Server for the agent.
func NewServer(nodeName, version string) *Server {
	return &Server{
		nodeName:  nodeName,
		version:   version,
		startedAt: time.Now(),
		registry:  faults.NewRegistry(),
	}
}

// Health returns the agent's health status.
func (s *Server) Health(_ context.Context, _ *pb.HealthRequest) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{
		Node:          s.nodeName,
		Version:       s.version,
		UptimeSeconds: int64(time.Since(s.startedAt).Seconds()),
	}, nil
}

// InjectFault constructs the right Fault implementation from the
// request and registers it with the Registry, which actually runs it.
func (s *Server) InjectFault(
	ctx context.Context,
	req *pb.InjectFaultRequest,
) (*pb.InjectFaultResponse, error) {
	if req.Type == pb.FaultType_FAULT_TYPE_UNSPECIFIED {
		return &pb.InjectFaultResponse{
			Accepted: false,
			Message:  "fault type must be specified",
		}, nil
	}
	if req.DurationSeconds <= 0 {
		return &pb.InjectFaultResponse{
			Accepted: false,
			Message:  "duration_seconds must be > 0",
		}, nil
	}

	duration := time.Duration(req.DurationSeconds) * time.Second
	params := faults.Params{Raw: req.Parameters}

	fault, err := buildFault(req.Type, duration, params)
	if err != nil {
		slog.Warn("InjectFault: bad request", "err", err)
		return &pb.InjectFaultResponse{
			Accepted: false,
			Message:  err.Error(),
		}, nil
	}

	if err := s.registry.Add(ctx, fault); err != nil {
		slog.Error("InjectFault: registry.Add failed", "err", err)
		return &pb.InjectFaultResponse{
			Accepted: false,
			Message:  "failed to start fault: " + err.Error(),
		}, nil
	}

	slog.Info("fault injected",
		"fault_id", fault.ID(),
		"type", fault.Type(),
		"duration_seconds", req.DurationSeconds,
	)
	return &pb.InjectFaultResponse{
		FaultId:  fault.ID(),
		Accepted: true,
		Message:  "fault started",
	}, nil
}

// Rollback stops a running fault.
func (s *Server) Rollback(
	ctx context.Context,
	req *pb.RollbackRequest,
) (*pb.RollbackResponse, error) {
	err := s.registry.Stop(ctx, req.FaultId)
	if errors.Is(err, faults.ErrFaultNotFound) {
		return &pb.RollbackResponse{
			Success: false,
			Message: "fault_id not found",
		}, nil
	}
	if err != nil {
		return &pb.RollbackResponse{
			Success: false,
			Message: "failed to stop fault: " + err.Error(),
		}, nil
	}
	return &pb.RollbackResponse{
		Success: true,
		Message: "fault rolled back",
	}, nil
}

// GetStatus returns the list of currently active faults.
func (s *Server) GetStatus(
	_ context.Context,
	_ *pb.StatusRequest,
) (*pb.StatusResponse, error) {
	active := s.registry.List()
	out := make([]*pb.ActiveFault, 0, len(active))
	for _, f := range active {
		remaining := int32(f.Duration().Seconds() - time.Since(f.StartedAt()).Seconds())
		if remaining < 0 {
			remaining = 0
		}
		out = append(out, &pb.ActiveFault{
			FaultId:          f.ID(),
			Type:             pbTypeFromString(f.Type()),
			StartedAtUnix:    f.StartedAt().Unix(),
			RemainingSeconds: remaining,
		})
	}
	return &pb.StatusResponse{ActiveFaults: out}, nil
}

// Shutdown gracefully stops all active faults. Call this when the
// agent process is shutting down.
func (s *Server) Shutdown(ctx context.Context) {
	s.registry.StopAll(ctx)
}

// buildFault is the dispatcher: it picks the right Fault constructor
// based on the protobuf FaultType. New fault types are added here.
func buildFault(t pb.FaultType, duration time.Duration, params faults.Params) (faults.Fault, error) {
	switch t {
	case pb.FaultType_FAULT_TYPE_CPU_STRESS:
		return faults.NewCPUStress(duration, params)
	case pb.FaultType_FAULT_TYPE_MEMORY_PRESSURE:
		return faults.NewMemoryPressure(duration, params)
	case pb.FaultType_FAULT_TYPE_DISK_FILL:
		return faults.NewDiskFill(duration, params)
	case pb.FaultType_FAULT_TYPE_NETWORK_LATENCY:
		return faults.NewNetworkLatency(duration, params)
	default:
		return nil, errors.New("unknown fault type")
	}
}

// pbTypeFromString translates the stable Type() string back to the
// protobuf enum value.
func pbTypeFromString(s string) pb.FaultType {
	switch s {
	case faults.TypeCPUStress:
		return pb.FaultType_FAULT_TYPE_CPU_STRESS
	case faults.TypeMemoryPressure:
		return pb.FaultType_FAULT_TYPE_MEMORY_PRESSURE
	case faults.TypeDiskFill:
		return pb.FaultType_FAULT_TYPE_DISK_FILL
	case faults.TypeNetworkLatency:
		return pb.FaultType_FAULT_TYPE_NETWORK_LATENCY
	default:
		return pb.FaultType_FAULT_TYPE_UNSPECIFIED
	}
}
