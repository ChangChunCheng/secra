package resource

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	secra_v1 "gitlab.com/jacky850509/secra/api/gen/v1"
	"gitlab.com/jacky850509/secra/internal/config"
	"google.golang.org/grpc"
)

var createVendorCmd = &cobra.Command{
	Use:   "create-vendor",
	Short: "Create a new vendor",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		conn, err := grpc.Dial(cfg.GRPCPort, grpc.WithInsecure())
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to connect to gRPC server: %v\n", err)
			os.Exit(1)
		}
		defer conn.Close()

		client := secra_v1.NewVendorServiceClient(conn)

		name, _ := cmd.Flags().GetString("name")

		req := &secra_v1.CreateVendorRequest{
			Vendor: &secra_v1.Vendor{
				Name: name,
			},
		}

		res, err := client.CreateVendor(context.Background(), req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating vendor: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Created Vendor: ID=%s Name=%s\n", res.Id, res.Name)
	},
}

func init() {
	createVendorCmd.Flags().String("name", "", "Vendor name")
	createVendorCmd.MarkFlagRequired("name")
}
