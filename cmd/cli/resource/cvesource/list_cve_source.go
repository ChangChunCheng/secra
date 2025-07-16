package cvesource

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	secra_v1 "gitlab.com/jacky850509/secra/api/gen/v1"
	"gitlab.com/jacky850509/secra/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var listCveSourceCmd = &cobra.Command{
	Use:   "list",
	Short: "List CVE sources with pagination",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		conn, err := grpc.NewClient(cfg.GRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to connect to gRPC server: %v\n", err)
			os.Exit(1)
		}
		defer conn.Close()

		client := secra_v1.NewCVESourceServiceClient(conn)
		limit, _ := cmd.Flags().GetInt32("limit")
		offset, _ := cmd.Flags().GetInt32("offset")
		req := &secra_v1.ListCVESourceRequest{Limit: limit, Offset: offset}
		res, err := client.ListCVESource(context.Background(), req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error listing CVE sources: %v\n", err)
			os.Exit(1)
		}
		for _, src := range res.Sources {
			fmt.Printf("ID=%s Name=%s URL=%s Enabled=%v\n", src.Id, src.Name, src.Url, src.Enabled)
		}
	},
}

func init() {
	listCveSourceCmd.Flags().Int32("limit", 10, "Limit number of results")
	listCveSourceCmd.Flags().Int32("offset", 0, "Offset for results")
}
