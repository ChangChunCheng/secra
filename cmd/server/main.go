package main

import (
	"context"
	"fmt"
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

	// Initialize default admin user if not exists
	initializeDefaultAdmin(db, cfg)

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
	log.Println("✅ Shutdown complete")
}

func initializeDefaultAdmin(db *storage.DBWrapper, cfg *config.AppConfig) {
	ctx := context.Background()

	// Check if any user exists
	var count int
	err := db.DB.NewSelect().Table("users").ColumnExpr("COUNT(*)").Scan(ctx, &count)
	if err != nil {
		log.Printf("⚠️  Failed to check existing users: %v", err)
		return
	}

	if count > 0 {
		return // Users already exist, skip initialization
	}

	// Create default admin user
	log.Printf("📝 Creating default admin user: %s", cfg.DefaultAdminUsername)

	email := cfg.DefaultAdminUsername + "@secra.local"

	// Use Bun's raw query for SQL functions
	query := fmt.Sprintf(`
		INSERT INTO users (id, username, email, password_hash, role, status, must_change_password, created_at, updated_at)
		VALUES (gen_random_uuid(), '%s', '%s', crypt('%s', gen_salt('bf')), 'admin', 'active', false, now(), now())
	`, cfg.DefaultAdminUsername, email, cfg.DefaultAdminPassword)

	_, err = db.DB.ExecContext(ctx, query)

	if err != nil {
		log.Printf("⚠️  Failed to create default admin user: %v", err)
		log.Printf("💡 You can create an admin user manually via /register endpoint")
	} else {
		log.Printf("✅ Default admin user '%s' created successfully", cfg.DefaultAdminUsername)
		log.Printf("🔑 Username: %s, Password: %s", cfg.DefaultAdminUsername, cfg.DefaultAdminPassword)
	}
}
