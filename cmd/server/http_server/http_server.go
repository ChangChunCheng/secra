package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/bunrouter/extra/reqlog"

	"gitlab.com/jacky850509/secra/internal/config"
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

	// 載入設定
	cfg := config.Load()

	// 初始化資料庫連線
	db := storage.NewDB(cfg.PostgresDSN, true)
	defer db.Close()

	// 建立 BunRouter 實例，並加入日誌中介軟體
	router := bunrouter.New(
		bunrouter.Use(reqlog.NewMiddleware()),
	)

	// 註冊路由
	router.GET("/", func(w http.ResponseWriter, req bunrouter.Request) error {
		_, err := w.Write([]byte("Secra RESTful API is running."))
		return err
	})

	// 建立 HTTP 伺服器
	server := &http.Server{
		Addr:    cfg.HTTPPort,
		Handler: router,
	}

	// 啟動伺服器
	go func() {
		log.Printf("HTTP server listening on %s", cfg.HTTPPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// 等待中斷訊號以優雅地關閉伺服器
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down HTTP server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP server Shutdown: %v", err)
	}
}
