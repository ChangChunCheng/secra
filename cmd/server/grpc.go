package main

import (
	"fmt"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"gitlab.com/jacky850509/secra/cmd/server/grpc_server"
	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/storage"
)

func main() {
	cfg := config.Load()

	db := storage.NewDB(cfg.PostgresDSN, false)
	defer db.Close()

	listener, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to listen on %s: %v\n", cfg.GRPCPort, err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()

	// Register Health Check Service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	grpc_server.RegisterServices(grpcServer, db.DB)

	fmt.Printf("gRPC server listening on %s\n", cfg.GRPCPort)
	if err := grpcServer.Serve(listener); err != nil {
		fmt.Fprintf(os.Stderr, "failed to serve gRPC server: %v\n", err)
		os.Exit(1)
	}
}
