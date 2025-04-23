package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	secra_v1 "gitlab.com/jacky850509/secra/api/gen/v1"
	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/service"
	"gitlab.com/jacky850509/secra/internal/storage"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

func main() {
	fmt.Println("🚀 CLI Running")
	fmt.Printf("Version: %s\nCommit: %s\nBuildDate: %s\n", Version, Commit, BuildDate)

	// 1. 載入設定
	cfg := config.Load()

	// 2. 初始化資料庫連線
	db := storage.NewDB(cfg.PostgresDSN, true)
	defer db.Close()

	// 3. 啟動 gRPC Server
	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	// 4. 註冊服務實作
	handler := service.NewSecraHandler(db.DB)
	secra_v1.RegisterSecraServiceServer(grpcServer, handler)

	// 5. graceful shutdown
	go func() {
		log.Printf("gRPC server listening on %s\n", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	// 6. 等待中斷訊號
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down gRPC server...")
	grpcServer.GracefulStop()
}
