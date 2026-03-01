package product

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

var listProductCmd = &cobra.Command{
	Use:   "list",
	Short: "List products",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		conn, err := grpc.NewClient(cfg.GRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to connect to gRPC server: %v\n", err)
			os.Exit(1)
		}
		defer conn.Close()

		client := secra_v1.NewProductServiceClient(conn)
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		req := &secra_v1.ListProductRequest{Limit: int32(limit), Offset: int32(offset)}
		res, err := client.ListProduct(context.Background(), req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error listing products: %v\n", err)
			os.Exit(1)
		}
		for _, p := range res.Products {
			fmt.Printf("ID=%s Name=%s\n", p.Id, p.Name)
		}
	},
}

func init() {
	listProductCmd.Flags().Int("limit", 10, "Maximum number of products")
	listProductCmd.Flags().Int("offset", 0, "Offset for pagination")
}
