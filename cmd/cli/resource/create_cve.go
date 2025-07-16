package resource

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

var createCveCmd = &cobra.Command{
	Use:   "create-cve",
	Short: "Create a new CVE record",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		conn, err := grpc.NewClient(cfg.GRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to connect to gRPC server: %v\n", err)
			os.Exit(1)
		}
		defer conn.Close()

		client := secra_v1.NewCVEServiceClient(conn)

		sourceID, _ := cmd.Flags().GetString("source-id")
		sourceUID, _ := cmd.Flags().GetString("source-uid")
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")

		req := &secra_v1.CreateCVERequest{
			Cve: &secra_v1.CVE{
				SourceId:    sourceID,
				SourceUid:   sourceUID,
				Title:       title,
				Description: description,
			},
		}
		res, err := client.CreateCVE(context.Background(), req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating CVE: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Created CVE: ID=%s SourceID=%s SourceUID=%s\n", res.Id, res.SourceId, res.SourceUid)
	},
}

func init() {
	createCveCmd.Flags().String("source-id", "", "Resource ID")
	createCveCmd.Flags().String("source-uid", "", "Original CVE identifier")
	createCveCmd.Flags().String("title", "", "CVE title")
	createCveCmd.Flags().String("description", "", "CVE description")
	createCveCmd.MarkFlagRequired("source-id")
	createCveCmd.MarkFlagRequired("source-uid")
	createCveCmd.MarkFlagRequired("title")
	createCveCmd.MarkFlagRequired("description")
}
