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

var createCveSourceCmd = &cobra.Command{
	Use:   "create-cve-source",
	Short: "Create a new CVE source",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		conn, err := grpc.NewClient(cfg.GRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to connect to gRPC server: %v\n", err)
			os.Exit(1)
		}
		defer conn.Close()

		client := secra_v1.NewCVESourceServiceClient(conn)

		name, _ := cmd.Flags().GetString("name")
		ctype, _ := cmd.Flags().GetString("type")
		url, _ := cmd.Flags().GetString("url")
		desc, _ := cmd.Flags().GetString("description")

		req := &secra_v1.CreateCVESourceRequest{
			Source: &secra_v1.CVESource{
				Name:        name,
				Type:        ctype,
				Url:         url,
				Description: desc,
			},
		}
		res, err := client.CreateCVESource(context.Background(), req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating CVE resource: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Created CVE resource: ID=%s Name=%s URL=%s\n", res.Id, res.Name, res.Url)
	},
}

func init() {
	createCveSourceCmd.Flags().String("name", "", "Resource name")
	createCveSourceCmd.Flags().String("type", "", "CVE source type")
	createCveSourceCmd.Flags().String("url", "", "Resource URL")
	createCveSourceCmd.Flags().String("description", "", "Description")
	createCveSourceCmd.MarkFlagRequired("name")
	createCveSourceCmd.MarkFlagRequired("type")
	createCveSourceCmd.MarkFlagRequired("url")
	createCveSourceCmd.MarkFlagRequired("description")
}
