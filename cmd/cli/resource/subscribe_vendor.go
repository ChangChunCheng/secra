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

var subscribeVendorCmd = &cobra.Command{
	Use:   "subscribe-vendor",
	Short: "Subscribe to a vendor",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		db := storage.NewDB(cfg.PostgresDSN, false)
		subRepo := repo.NewSubscriptionRepository(db.DB)
		svc := service.NewSubscriptionService(subRepo)

		userID, _ := cmd.Flags().GetString("user-id")
		vendorID, _ := cmd.Flags().GetString("vendor-id")
		severity, _ := cmd.Flags().GetString("severity")
		// 將 severity 轉大寫以便映射
		severity = strings.ToUpper(severity)

		target := model.SubscriptionTarget{
			TargetTypeID: 1, // vendor type
			TargetID:     uuid.MustParse(vendorID),
		}
		sub, err := svc.CreateSubscription(context.Background(), userID, []model.SubscriptionTarget{target}, severity)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Subscription created: User=%s Vendors=[%s] Severity=%s\n", sub.UserID, strings.Join([]string{vendorID}, ","), severity)
	},
}

func init() {
	Cmd.AddCommand(subscribeVendorCmd)
	subscribeVendorCmd.Flags().String("user-id", "", "User UUID")
	subscribeVendorCmd.Flags().String("vendor-id", "", "Vendor UUID to subscribe")
	subscribeVendorCmd.Flags().String("severity", "low", "Severity threshold")
	subscribeVendorCmd.MarkFlagRequired("user-id")
	subscribeVendorCmd.MarkFlagRequired("vendor-id")
}
