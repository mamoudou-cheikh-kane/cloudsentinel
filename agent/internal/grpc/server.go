// Package grpc provides the CloudSentinel agent gRPC server.
package grpc

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	pb "github.com/mamoudou-cheikh-kane/cloudsentinel/agent/internal/pb"
)

// Server implements pb.AgentServiceServer.
type Server struct {
	pb.UnimplementedAgentServiceServer

	nodeName     string
	version      string
	startedAt    time.Time
	mu           sync.Mutex
	activeFaults map[string]*activeFault
}

type activeFault struct {
	id        string
	faultType pb.FaultType
	startedAt time.Time
	duration  time.Duration
}

// NewServer creates a new gRPC Server for the agent.
func NewServer(nodeName, version string) *Server {
	return &Server{
		nodeName:     nodeName,
		version:      version,
		startedAt:    time.Now(),
		activeFaults: make(map[string]*activeFault),
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

// InjectFault records a new fault injection.
// The actual fault execution is a TODO — this is a skeleton.
func (s *Server) InjectFault(
	_ context.Context,
	req *pb.InjectFaultRequest,
) (*pb.InjectFaultResponse, error) {
	if req.Type == pb.FaultType_FAULT_TYPE_UNSPECIFIED {
		return &pb.InjectFaultResponse{
			Accepted: false,
			Message:  "fault type must be specified",
		}, nil
	}

	faultID := uuid.NewString()
	s.mu.Lock()
	s.activeFaults[faultID] = &activeFault{
		id:        faultID,
		faultType: req.Type,
		startedAt: time.Now(),
		duration:  time.Duration(req.DurationSeconds) * time.Second,
	}
	s.mu.Unlock()

	slog.Info("fault injected",
		"fault_id", faultID,
		"type", req.Type.String(),
		"duration_seconds", req.DurationSeconds,
	)

	return &pb.InjectFaultResponse{
		FaultId:  faultID,
		Accepted: true,
		Message:  "fault recorded (execution TODO)",
	}, nil
}

// Rollback removes an active fault injection.
func (s *Server) Rollback(
	_ context.Context,
	req *pb.RollbackRequest,
) (*pb.RollbackResponse, error) {
	s.mu.Lock()
	_, ok := s.activeFaults[req.FaultId]
	delete(s.activeFaults, req.FaultId)
	s.mu.Unlock()

	if !ok {
		return &pb.RollbackResponse{
			Success: false,
			Message: "fault_id not found",
		}, nil
	}

	slog.Info("fault rolled back", "fault_id", req.FaultId)
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
	s.mu.Lock()
	defer s.mu.Unlock()

	faults := make([]*pb.ActiveFault, 0, len(s.activeFaults))
	for _, f := range s.activeFaults {
		remaining := int32(f.duration.Seconds() - time.Since(f.startedAt).Seconds())
		if remaining < 0 {
			remaining = 0
		}
		faults = append(faults, &pb.ActiveFault{
			FaultId:          f.id,
			Type:             f.faultType,
			StartedAtUnix:    f.startedAt.Unix(),
			RemainingSeconds: remaining,
		})
	}

	return &pb.StatusResponse{ActiveFaults: faults}, nil
}
