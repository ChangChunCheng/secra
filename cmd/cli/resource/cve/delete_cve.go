package cve

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

var deleteCveCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a CVE record",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		conn, err := grpc.NewClient(cfg.GRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "connect error: %v\n", err)
			os.Exit(1)
		}
		defer conn.Close()

		client := secra_v1.NewCVEServiceClient(conn)
		id, _ := cmd.Flags().GetString("id")

		_, err = client.DeleteCVE(context.Background(), &secra_v1.DeleteCVERequest{Id: id})
		if err != nil {
			fmt.Fprintf(os.Stderr, "error deleting CVE: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Deleted CVE: ID=%s\n", id)
	},
}

func init() {
	deleteCveCmd.Flags().String("id", "", "CVE UUID")
	deleteCveCmd.MarkFlagRequired("id")
}
