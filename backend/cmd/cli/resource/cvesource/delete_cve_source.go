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

var deleteCveSourceCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a CVE source by ID",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		conn, err := grpc.Dial(cfg.GRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to connect to gRPC server: %v\n", err)
			os.Exit(1)
		}
		defer conn.Close()

		client := secra_v1.NewCVESourceServiceClient(conn)
		id, _ := cmd.Flags().GetString("id")
		req := &secra_v1.DeleteCVESourceRequest{Id: id}
		_, err = client.DeleteCVESource(context.Background(), req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error deleting CVE source: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Deleted CVE source: ID=%s\n", id)
	},
}

func init() {
	deleteCveSourceCmd.Flags().String("id", "", "Resource ID")
	deleteCveSourceCmd.MarkFlagRequired("id")
}
