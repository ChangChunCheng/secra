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

var getProductCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a product by ID",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		conn, err := grpc.NewClient(cfg.GRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to connect to gRPC server: %v\n", err)
			os.Exit(1)
		}
		defer conn.Close()

		client := secra_v1.NewProductServiceClient(conn)
		id, _ := cmd.Flags().GetString("id")

		req := &secra_v1.GetProductRequest{Id: id}
		res, err := client.GetProduct(context.Background(), req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error getting product: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Product: ID=%s Name=%s\n", res.Id, res.Name)
	},
}

func init() {
	getProductCmd.Flags().String("id", "", "Product ID")
	getProductCmd.MarkFlagRequired("id")
}
