package vendor

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

var updateVendorCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a vendor",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		conn, err := grpc.NewClient(cfg.GRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "connect error: %v\n", err)
			os.Exit(1)
		}
		defer conn.Close()

		client := secra_v1.NewVendorServiceClient(conn)
		id, _ := cmd.Flags().GetString("id")
		name, _ := cmd.Flags().GetString("name")

		req := &secra_v1.UpdateVendorRequest{
			Vendor: &secra_v1.Vendor{Id: id, Name: name},
		}
		res, err := client.UpdateVendor(context.Background(), req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error updating vendor: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Updated Vendor: ID=%s Name=%s\n", res.Id, res.Name)
	},
}

func init() {
	updateVendorCmd.Flags().String("id", "", "Vendor UUID")
	updateVendorCmd.Flags().String("name", "", "New vendor name")
	updateVendorCmd.MarkFlagRequired("id")
	updateVendorCmd.MarkFlagRequired("name")
}
