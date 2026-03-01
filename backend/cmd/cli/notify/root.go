package notify

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/uptrace/bun"
	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/service"
	"gitlab.com/jacky850509/secra/internal/storage"
)

var Cmd = &cobra.Command{
	Use:   "notify",
	Short: "Notification dispatch commands",
}

var digestCmd = &cobra.Command{
	Use:   "digest",
	Short: "Send scheduled digests to users",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		db := storage.NewDB(cfg.PostgresDSN, false)
		defer db.Close()

		notifier := service.NewNotificationService(cfg.SMTPConfig, db.DB)
		ctx := context.Background()

		// 1. Fetch all users with scheduled notifications
		var users []model.User
		db.DB.NewSelect().Model(&users).Where("notification_frequency != 'immediate'").Scan(ctx)

		for _, u := range users {
			if shouldNotifyUser(u) {
				log.Printf("📬 Preparing digest for %s (%s)", u.Username, u.Email)
				sendUserDigest(ctx, db.DB, notifier, u)
			}
		}
	},
}

func shouldNotifyUser(u model.User) bool {
	loc, err := time.LoadLocation(u.Timezone)
	if err != nil { loc = time.UTC }

	nowLocal := time.Now().In(loc)
	
	if nowLocal.Hour() != 8 { return false }

	if u.LastNotifiedAt.In(loc).Format("2006-01-02") == nowLocal.Format("2006-01-02") {
		return false
	}

	if u.NotificationFrequency == "weekly" && nowLocal.Weekday() != time.Monday {
		return false
	}

	return true
}

func sendUserDigest(ctx context.Context, db *bun.DB, notifier service.NotificationService, u model.User) {
	type cveInfo struct {
		SourceUID string `bun:"source_uid"`
		Title     string `bun:"title"`
		Severity  string `bun:"severity"`
	}
	var matches []cveInfo

	err := db.NewSelect().
		TableExpr("cves AS c").
		ColumnExpr("c.source_uid, c.title, c.severity").
		Join("JOIN cve_products cp ON cp.cve_id = c.id").
		Join("JOIN products p ON p.id = cp.product_id").
		Join("JOIN subscription_targets st ON (st.target_type_id = 2 AND st.target_id = p.vendor_id) OR (st.target_type_id = 3 AND st.target_id = p.id)").
		Join("JOIN subscriptions s ON s.id = st.subscription_id").
		Where("s.user_id = ? AND c.published_at > ?", u.ID, u.LastNotifiedAt).
		Group("c.source_uid", "c.title", "c.severity").
		Scan(ctx, &matches)

	if err != nil || len(matches) == 0 {
		log.Printf("ℹ️ No new CVEs for %s since %v", u.Username, u.LastNotifiedAt)
		return
	}

	body := fmt.Sprintf("Hello %s,\n\nHere is your vulnerability summary since %s:\n\n", u.Username, u.LastNotifiedAt.Format("2006-01-02"))
	for _, m := range matches {
		body += fmt.Sprintf("- [%s] %s (Severity: %s)\n", m.SourceUID, m.Title, m.Severity)
	}
	body += "\nCheck details at: http://localhost:8081\n"

	err = notifier.SendEmail(ctx, u.Email, "SECRA: Your Vulnerability Digest", body)
	if err == nil {
		db.NewUpdate().Model(&u).Set("last_notified_at = ?", time.Now()).WherePK().Exec(ctx)
		log.Printf("✅ Digest sent to %s", u.Email)
	} else {
		log.Printf("❌ Failed to send digest: %v", err)
	}
}

func init() {
	Cmd.AddCommand(digestCmd)
}
