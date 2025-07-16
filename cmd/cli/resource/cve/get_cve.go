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

var getCveCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieve a CVE record by ID",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		conn, err := grpc.NewClient(cfg.GRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to connect to gRPC server: %v\n", err)
			os.Exit(1)
		}
		defer conn.Close()

		client := secra_v1.NewCVEServiceClient(conn)
		id, _ := cmd.Flags().GetString("id")
		req := &secra_v1.GetCVERequest{Id: id}
		res, err := client.GetCVE(context.Background(), req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error retrieving CVE: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("ID=%s SourceID=%s SourceUID=%s Title=%s Description=%s Severity=%s CVSS=%.2f Status=%s PublishedAt=%s UpdatedAt=%s\n",
			res.Id, res.SourceId, res.SourceUid, res.Title, res.Description, res.Severity, res.CvssScore, res.Status, res.PublishedAt, res.UpdatedAt)
	},
}

func init() {
	getCveCmd.Flags().String("id", "", "CVE UUID")
	getCveCmd.MarkFlagRequired("id")
}
