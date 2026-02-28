package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	secra_v1 "gitlab.com/jacky850509/secra/api/gen/v1"
	"gitlab.com/jacky850509/secra/cmd/server/http_server"
	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/storage"
)

func main() {
	cfg := config.Load()

	// Connect to Database
	db := storage.NewDB(cfg.PostgresDSN, false)
	defer db.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 1. Setup gRPC-Gateway
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	
	grpcAddr := "localhost" + cfg.GRPCPort
	// Register all services to the gateway
	err := secra_v1.RegisterCVEServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts)
	if err != nil {
		log.Fatalf("failed to register CVE gateway: %v", err)
	}
	_ = secra_v1.RegisterProductServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts)
	_ = secra_v1.RegisterVendorServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts)
	_ = secra_v1.RegisterUserServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts)
	_ = secra_v1.RegisterSubscriptionServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts)

	// 2. Setup Web UI Server (net/http based)
	webServer := http_server.NewServer(db.DB)

	// 3. Combined Handler
	mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// API requests go to grpc-gateway
		if strings.HasPrefix(r.URL.Path, "/v1/") {
			mux.ServeHTTP(w, r)
			return
		}
		// Everything else goes to Web UI
		webServer.ServeHTTP(w, r)
	})

	port := cfg.HTTPPort
	if port == "" || port == ":8080" {
		port = ":8081"
	}

	addr := "127.0.0.1" + port
	log.Printf("🚀 Combined HTTP Server (API + Web UI) starting on %s...", addr)
	if err := http.ListenAndServe(addr, mainHandler); err != nil {
		log.Fatalf("❌ Failed to start HTTP server: %v", err)
	}
}
