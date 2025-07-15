package resource

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/repo"
	"gitlab.com/jacky850509/secra/internal/service"
	"gitlab.com/jacky850509/secra/internal/storage"
)

var subscribeCveResourceCmd = &cobra.Command{
	Use:   "subscribe-cve-resource",
	Short: "Subscribe to a CVE resource",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		db := storage.NewDB(cfg.PostgresDSN, false)
		subRepo := repo.NewSubscriptionRepository(db.DB)
		svc := service.NewSubscriptionService(subRepo)

		userID, _ := cmd.Flags().GetString("user-id")
		resourceID, _ := cmd.Flags().GetString("resource-id")
		severity, _ := cmd.Flags().GetString("severity")
		// 將 severity 轉大寫以便映射
		severity = strings.ToUpper(severity)

		target := model.SubscriptionTarget{
			TargetTypeID: 3, // cve_resource type
			TargetID:     uuid.MustParse(resourceID),
		}
		sub, err := svc.CreateSubscription(context.Background(), userID, []model.SubscriptionTarget{target}, severity)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Subscription created: User=%s CVEResources=[%s] Severity=%s\n", sub.UserID, strings.Join([]string{resourceID}, ","), severity)
	},
}

func init() {
	subscribeCveResourceCmd.Flags().String("user-id", "", "User UUID")
	subscribeCveResourceCmd.Flags().String("resource-id", "", "CVE Resource UUID to subscribe")
	subscribeCveResourceCmd.Flags().String("severity", "low", "Severity threshold")
	subscribeCveResourceCmd.MarkFlagRequired("user-id")
	subscribeCveResourceCmd.MarkFlagRequired("resource-id")
}
