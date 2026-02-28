package health

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"

	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/storage"
)

var (
	probeType string
	addr      string
	timeout   time.Duration
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Perform a health check (db, http, or grpc)",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		var err error
		switch probeType {
		case "db":
			err = checkDB(ctx)
		case "http":
			err = checkHTTP(ctx, addr)
		case "grpc":
			err = checkGRPC(ctx, addr)
		default:
			fmt.Printf("❌ Unknown probe type: %s\n", probeType)
			os.Exit(1)
		}

		if err != nil {
			fmt.Printf("❌ Health check failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✅ Health check passed")
	},
}

func init() {
	checkCmd.Flags().StringVar(&probeType, "type", "db", "Probe type: db, http, or grpc")
	checkCmd.Flags().StringVar(&addr, "addr", "localhost:8081", "Address to check (for http/grpc)")
	checkCmd.Flags().DurationVar(&timeout, "timeout", 2*time.Second, "Timeout for the check")
}

func checkDB(ctx context.Context) error {
	cfg := config.Load()
	dbWrapper := storage.NewDB(cfg.PostgresDSN, false)
	defer dbWrapper.Close()

	// dbWrapper.DB is *bun.DB, it has PingContext
	return dbWrapper.DB.PingContext(ctx)
}

func checkHTTP(ctx context.Context, target string) error {
	url := target
	if !strings.HasPrefix(target, "http") {
		url = "http://" + target + "/health"
	}
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status: %d", resp.StatusCode)
	}
	return nil
}

func checkGRPC(ctx context.Context, target string) error {
	conn, err := grpc.DialContext(ctx, target, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return err
	}
	defer conn.Close()
	client := grpc_health_v1.NewHealthClient(conn)
	resp, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: ""})
	if err != nil {
		return err
	}
	if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		return fmt.Errorf("status: %s", resp.Status)
	}
	return nil
}
