package subscribe

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	secra_v1 "gitlab.com/jacky850509/secra/api/gen/v1"
	"gitlab.com/jacky850509/secra/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var subscribeCveSourceCmd = &cobra.Command{
	Use:   "cve-source",
	Short: "Subscribe to a CVE source",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		conn, err := grpc.NewClient(cfg.GRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to connect to gRPC server: %v\n", err)
			os.Exit(1)
		}
		defer conn.Close()

		client := secra_v1.NewSubscriptionServiceClient(conn)

		userID, _ := cmd.Flags().GetString("user-id")
		resourceID, _ := cmd.Flags().GetString("resource-id")
		severity, _ := cmd.Flags().GetString("severity")
		severity = strings.ToUpper(severity)

		req := &secra_v1.CreateSubscriptionRequest{
			UserId: userID,
			Targets: []*secra_v1.SubscriptionTarget{
				{
					TargetType: "cve_source",
					TargetId:   resourceID,
				}},
		}
		resp, err := client.CreateSubscription(context.Background(), req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating subscription: %v\n", err)
			os.Exit(1)
		}
		sub := resp.GetSubscription()
		fmt.Printf("Subscription ID=%s User=%s Targets=%v Severity=%s\n", sub.GetId(), sub.GetUserId(), sub.GetTargets(), sub.GetSeverityThreshold())
	},
}

func init() {
	subscribeCveSourceCmd.Flags().String("user-id", "", "User UUID")
	subscribeCveSourceCmd.Flags().String("resource-id", "", "CVE Resource UUID to subscribe")
	subscribeCveSourceCmd.Flags().String("severity", "low", "Severity threshold")
	subscribeCveSourceCmd.MarkFlagRequired("user-id")
	subscribeCveSourceCmd.MarkFlagRequired("resource-id")
}
