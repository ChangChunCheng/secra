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

var createProductCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new product",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		conn, err := grpc.NewClient(cfg.GRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to connect to gRPC server: %v\n", err)
			os.Exit(1)
		}
		defer conn.Close()

		client := secra_v1.NewProductServiceClient(conn)
		name, _ := cmd.Flags().GetString("name")

		req := &secra_v1.CreateProductRequest{
			Product: &secra_v1.Product{
				Name: name,
			},
		}
		res, err := client.CreateProduct(context.Background(), req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating product: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Created Product: ID=%s Name=%s\n", res.Id, res.Name)
	},
}

func init() {
	createProductCmd.Flags().String("name", "", "Product name")
	createProductCmd.MarkFlagRequired("name")
}
