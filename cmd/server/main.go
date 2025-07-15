package main

import (
	"fmt"
	"net"
	"os"

	"google.golang.org/grpc"

	"gitlab.com/jacky850509/secra/cmd/server/grpc_server"
	"gitlab.com/jacky850509/secra/internal/config"
)

func main() {
	cfg := config.Load()
	listener, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to listen on %s: %v\n", cfg.GRPCPort, err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	grpc_server.RegisterServices(grpcServer)

	fmt.Printf("gRPC server listening on %s\n", cfg.GRPCPort)
	if err := grpcServer.Serve(listener); err != nil {
		fmt.Fprintf(os.Stderr, "failed to serve gRPC server: %v\n", err)
		os.Exit(1)
	}
}
