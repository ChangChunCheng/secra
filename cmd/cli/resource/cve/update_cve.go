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

var updateCveCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing CVE record",
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
		sourceID, _ := cmd.Flags().GetString("source-id")
		sourceUID, _ := cmd.Flags().GetString("source-uid")
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")
		severity, _ := cmd.Flags().GetString("severity")
		cvss, _ := cmd.Flags().GetFloat32("cvss-score")
		status, _ := cmd.Flags().GetString("status")

		req := &secra_v1.UpdateCVERequest{
			Cve: &secra_v1.CVE{
				Id:          id,
				SourceId:    sourceID,
				SourceUid:   sourceUID,
				Title:       title,
				Description: description,
				Severity:    severity,
				CvssScore:   cvss,
				Status:      status,
			},
		}
		res, err := client.UpdateCVE(context.Background(), req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error updating CVE: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Updated CVE: ID=%s SourceID=%s SourceUID=%s\n", res.Id, res.SourceId, res.SourceUid)
	},
}

func init() {
	updateCveCmd.Flags().String("id", "", "CVE UUID")
	updateCveCmd.Flags().String("source-id", "", "Resource ID")
	updateCveCmd.Flags().String("source-uid", "", "Original CVE identifier")
	updateCveCmd.Flags().String("title", "", "CVE title")
	updateCveCmd.Flags().String("description", "", "CVE description")
	updateCveCmd.Flags().String("severity", "", "CVE severity")
	updateCveCmd.Flags().Float32("cvss-score", 0, "CVSS score")
	updateCveCmd.Flags().String("status", "", "CVE status")
	updateCveCmd.MarkFlagRequired("id")
	updateCveCmd.MarkFlagRequired("source-id")
	updateCveCmd.MarkFlagRequired("source-uid")
	updateCveCmd.MarkFlagRequired("title")
	updateCveCmd.MarkFlagRequired("description")
	updateCveCmd.MarkFlagRequired("severity")
	updateCveCmd.MarkFlagRequired("cvss-score")
	updateCveCmd.MarkFlagRequired("status")
}
