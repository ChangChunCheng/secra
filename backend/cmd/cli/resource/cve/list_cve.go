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

var listCveCmd = &cobra.Command{
	Use:   "list",
	Short: "List CVE records with pagination",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		conn, err := grpc.Dial(cfg.GRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to connect to gRPC server: %v\n", err)
			os.Exit(1)
		}
		defer conn.Close()

		client := secra_v1.NewCVEServiceClient(conn)
		limit, _ := cmd.Flags().GetInt32("limit")
		offset, _ := cmd.Flags().GetInt32("offset")
		req := &secra_v1.ListCVERequest{Limit: limit, Offset: offset}
		res, err := client.ListCVE(context.Background(), req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error listing CVEs: %v\n", err)
			os.Exit(1)
		}

		for _, cve := range res.Cves {
			fmt.Printf("ID=%s SourceID=%s SourceUID=%s Title=%s Severity=%s CVSS=%.2f Status=%s PublishedAt=%s UpdatedAt=%s\n",
				cve.Id, cve.SourceId, cve.SourceUid, cve.Title, cve.Severity, cve.CvssScore, cve.Status, cve.PublishedAt, cve.UpdatedAt)
		}
	},
}

func init() {
	listCveCmd.Flags().Int32("limit", 10, "Maximum number of CVEs to list")
	listCveCmd.Flags().Int32("offset", 0, "Offset for pagination")
}
