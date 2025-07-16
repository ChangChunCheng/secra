package cvesource

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

var updateCveSourceCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing CVE source",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		conn, err := grpc.NewClient(cfg.GRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to connect to gRPC server: %v\n", err)
			os.Exit(1)
		}
		defer conn.Close()

		client := secra_v1.NewCVESourceServiceClient(conn)
		id, _ := cmd.Flags().GetString("id")
		name, _ := cmd.Flags().GetString("name")
		ctype, _ := cmd.Flags().GetString("type")
		url, _ := cmd.Flags().GetString("url")
		desc, _ := cmd.Flags().GetString("description")

		req := &secra_v1.UpdateCVESourceRequest{
			Source: &secra_v1.CVESource{
				Id:          id,
				Name:        name,
				Type:        ctype,
				Url:         url,
				Description: desc,
				Enabled:     true,
			},
		}
		res, err := client.UpdateCVESource(context.Background(), req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error updating CVE source: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Updated CVE source: ID=%s Name=%s URL=%s\n", res.Id, res.Name, res.Url)
	},
}

func init() {
	updateCveSourceCmd.Flags().String("id", "", "Resource ID")
	updateCveSourceCmd.Flags().String("name", "", "Resource name")
	updateCveSourceCmd.Flags().String("type", "", "CVE resource type")
	updateCveSourceCmd.Flags().String("url", "", "Resource URL")
	updateCveSourceCmd.Flags().String("description", "", "Description")
	updateCveSourceCmd.MarkFlagRequired("id")
	updateCveSourceCmd.MarkFlagRequired("name")
	updateCveSourceCmd.MarkFlagRequired("type")
	updateCveSourceCmd.MarkFlagRequired("url")
	updateCveSourceCmd.MarkFlagRequired("description")
}
