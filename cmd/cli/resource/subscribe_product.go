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

var subscribeProductCmd = &cobra.Command{
	Use:   "subscribe-product",
	Short: "Subscribe to a product",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		db := storage.NewDB(cfg.PostgresDSN, false)
		subRepo := repo.NewSubscriptionRepository(db.DB)
		svc := service.NewSubscriptionService(subRepo)

		userID, _ := cmd.Flags().GetString("user-id")
		productID, _ := cmd.Flags().GetString("product-id")
		severity, _ := cmd.Flags().GetString("severity")
		// 將 severity 轉大寫以便映射
		severity = strings.ToUpper(severity)

		target := model.SubscriptionTarget{
			TargetTypeID: 2, // product type
			TargetID:     uuid.MustParse(productID),
		}
		sub, err := svc.CreateSubscription(context.Background(), userID, []model.SubscriptionTarget{target}, severity)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Subscription created: User=%s Products=[%s] Severity=%s\n", sub.UserID, strings.Join([]string{productID}, ","), severity)
	},
}

func init() {
	Cmd.AddCommand(subscribeProductCmd)
	subscribeProductCmd.Flags().String("user-id", "", "User UUID")
	subscribeProductCmd.Flags().String("product-id", "", "Product UUID to subscribe")
	subscribeProductCmd.Flags().String("severity", "low", "Severity threshold")
	subscribeProductCmd.MarkFlagRequired("user-id")
	subscribeProductCmd.MarkFlagRequired("product-id")
}
