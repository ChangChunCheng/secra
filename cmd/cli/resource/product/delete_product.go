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

var deleteProductCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a product",
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

		_, err = client.DeleteProduct(context.Background(), &secra_v1.DeleteProductRequest{Id: id})
		if err != nil {
			fmt.Fprintf(os.Stderr, "error deleting product: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Deleted Product: ID=%s\n", id)
	},
}

func init() {
	deleteProductCmd.Flags().String("id", "", "Product UUID")
	deleteProductCmd.MarkFlagRequired("id")
}
