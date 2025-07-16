package resource

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

var listVendorCmd = &cobra.Command{
	Use:   "list-vendor",
	Short: "List vendors",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		conn, err := grpc.NewClient(cfg.GRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "connect error: %v\n", err)
			os.Exit(1)
		}
		defer conn.Close()

		client := secra_v1.NewVendorServiceClient(conn)
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		req := &secra_v1.ListVendorRequest{Limit: int32(limit), Offset: int32(offset)}
		res, err := client.ListVendor(context.Background(), req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error listing vendors: %v\n", err)
			os.Exit(1)
		}

		for _, v := range res.Vendors {
			fmt.Printf("ID=%s Name=%s\n", v.Id, v.Name)
		}
	},
}

func init() {
	listVendorCmd.Flags().Int("limit", 10, "Maximum number of vendors")
	listVendorCmd.Flags().Int("offset", 0, "Offset for pagination")
}
