package source

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

var getVendorCmd = &cobra.Command{
	Use:   "get-vendor",
	Short: "Get a vendor by ID",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		conn, err := grpc.NewClient(cfg.GRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to connect to gRPC server: %v\n", err)
			os.Exit(1)
		}
		defer conn.Close()

		client := secra_v1.NewVendorServiceClient(conn)
		id, _ := cmd.Flags().GetString("id")
		req := &secra_v1.GetVendorRequest{Id: id}
		res, err := client.GetVendor(context.Background(), req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error getting vendor: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Vendor: ID=%s Name=%s\n", res.Id, res.Name)
	},
}

func init() {
	getVendorCmd.Flags().String("id", "", "Vendor ID")
	getVendorCmd.MarkFlagRequired("id")
}
