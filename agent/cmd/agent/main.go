// Package main is the entry point for the CloudSentinel agent.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	csgrpc "github.com/mamoudou-cheikh-kane/cloudsentinel/agent/internal/grpc"
	"github.com/mamoudou-cheikh-kane/cloudsentinel/agent/internal/metrics"
	pb "github.com/mamoudou-cheikh-kane/cloudsentinel/agent/internal/pb"
)

const (
	metricsAddr     = ":9100"
	metricsPath     = "/metrics"
	grpcAddr        = ":50051"
	shutdownTimeout = 10 * time.Second
	agentVersion    = "0.1.0"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		hostname, _ := os.Hostname()
		nodeName = hostname
	}

	slog.Info("starting cloudsentinel-agent",
		"version", agentVersion,
		"node", nodeName,
		"metrics_addr", metricsAddr,
		"grpc_addr", grpcAddr,
	)

	// ---------- Prometheus metrics server ----------
	registry := prometheus.NewRegistry()
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		metrics.NewSystemCollector(nodeName),
	)

	mux := http.NewServeMux()
	mux.Handle(metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		Registry: registry,
	}))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	httpSrv := &http.Server{
		Addr:              metricsAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// ---------- gRPC server ----------
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		slog.Error("failed to listen", "addr", grpcAddr, "error", err)
		os.Exit(1)
	}
	grpcSrv := grpc.NewServer()
	csServer := csgrpc.NewServer(nodeName, agentVersion)
	pb.RegisterAgentServiceServer(grpcSrv, csServer)
	reflection.Register(grpcSrv)

	// ---------- Run both servers ----------
	errCh := make(chan error, 2)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("metrics server listening", "addr", metricsAddr)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("grpc server listening", "addr", grpcAddr)
		if err := grpcSrv.Serve(lis); err != nil {
			errCh <- err
		}
	}()

	// ---------- Wait for signal or fatal error ----------
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		slog.Error("server failed", "error", err)
	case sig := <-sigCh:
		slog.Info("shutdown signal received", "signal", sig.String())
	}

	// ---------- Graceful shutdown ----------
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Stop every active fault before stopping the gRPC server, so
	// we get a chance to clean tc rules, remove filler files, and
	// free the memory buffers. This is best-effort: we keep going
	// even if a single fault fails to stop.
	slog.Info("stopping active faults")
	csServer.Shutdown(ctx)

	grpcSrv.GracefulStop()
	if err := httpSrv.Shutdown(ctx); err != nil {
		slog.Error("graceful http shutdown failed", "error", err)
	}

	wg.Wait()
	slog.Info("agent stopped cleanly")
}
