package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"gitlab.com/jacky850509/secra/cmd/server/grpc_server"
	"gitlab.com/jacky850509/secra/cmd/server/http_server"
	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/storage"
)

func main() {
	cfg := config.Load()

	// 1. Unified Database Connection
	db := storage.NewDB(cfg.PostgresDSN, true)
	defer db.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 2. Prepare gRPC - Ensure it listens on all interfaces (0.0.0.0)
	grpcAddr := cfg.GRPCPort
	if strings.Contains(grpcAddr, "127.0.0.1") {
		grpcAddr = ":" + strings.Split(grpcAddr, ":")[1]
	}
	
	grpcListener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("❌ Failed to listen on gRPC port %s: %v", grpcAddr, err)
	}
	grpcSrv := grpc.NewServer()
	
	healthSrv := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcSrv, healthSrv)
	healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	grpc_server.RegisterServices(grpcSrv, db.DB)

	// 3. Prepare HTTP
	httpSrv := http_server.NewServer(db.DB)
	webSrv := &http.Server{
		Addr:    cfg.HTTPPort,
		Handler: httpSrv,
	}

	// 4. Start Servers
	go func() {
		log.Printf("🚀 gRPC Server starting on %s", grpcAddr)
		if err := grpcSrv.Serve(grpcListener); err != nil && err != grpc.ErrServerStopped {
			log.Fatalf("❌ gRPC server error: %v", err)
		}
	}()

	go func() {
		log.Printf("🚀 HTTP Server (Web + REST) starting on %s", cfg.HTTPPort)
		if err := webSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ HTTP server error: %v", err)
		}
	}()

	// 5. Graceful Shutdown
	<-ctx.Done()
	log.Println("🛑 Shutting down servers...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := webSrv.Shutdown(shutdownCtx); err != nil {
		log.Printf("⚠️ HTTP shutdown error: %v", err)
	}
	grpcSrv.GracefulStop()

	log.Println("✅ Servers exited gracefully")
}
