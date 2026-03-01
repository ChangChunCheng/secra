package subscription

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

var deleteSubscriptionCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a subscription",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		conn, err := grpc.NewClient(cfg.GRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to connect to gRPC server: %v\n", err)
			os.Exit(1)
		}
		defer conn.Close()

		client := secra_v1.NewSubscriptionServiceClient(conn)

		subID, _ := cmd.Flags().GetString("subscription-id")

		req := &secra_v1.DeleteSubscriptionRequest{
			Id: subID,
		}
		_, err = client.DeleteSubscription(context.Background(), req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error deleting subscription: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Deleted subscription %s\n", subID)
	},
}

func init() {
	deleteSubscriptionCmd.Flags().String("subscription-id", "", "Subscription UUID to delete")
	deleteSubscriptionCmd.MarkFlagRequired("subscription-id")
}
