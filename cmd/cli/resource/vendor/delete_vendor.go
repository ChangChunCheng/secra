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

var deleteVendorCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a vendor",
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

		_, err = client.DeleteVendor(context.Background(), &secra_v1.DeleteVendorRequest{Id: id})
		if err != nil {
			fmt.Fprintf(os.Stderr, "error deleting vendor: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Deleted Vendor: ID=%s\n", id)
	},
}

func init() {
	deleteVendorCmd.Flags().String("id", "", "Vendor UUID")
	deleteVendorCmd.MarkFlagRequired("id")
}
