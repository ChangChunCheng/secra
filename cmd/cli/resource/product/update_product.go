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

var updateProductCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a product",
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
		name, _ := cmd.Flags().GetString("name")

		req := &secra_v1.UpdateProductRequest{
			Product: &secra_v1.Product{Id: id, Name: name},
		}
		res, err := client.UpdateProduct(context.Background(), req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error updating product: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Updated Product: ID=%s Name=%s\n", res.Id, res.Name)
	},
}

func init() {
	updateProductCmd.Flags().String("id", "", "Product UUID")
	updateProductCmd.Flags().String("name", "", "New product name")
	updateProductCmd.MarkFlagRequired("id")
	updateProductCmd.MarkFlagRequired("name")
}
