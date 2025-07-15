package user

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/repo"
	"gitlab.com/jacky850509/secra/internal/service"
	"gitlab.com/jacky850509/secra/internal/storage"
)

var updateProfileCmd = &cobra.Command{
	Use:   "update-profile",
	Short: "更新使用者個人資料 (目前僅支援更新 Email)",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		db := storage.NewDB(cfg.PostgresDSN, false)
		userRepo := repo.NewUserRepository(db.DB)
		svc := service.NewUserService(userRepo)

		token, _ := cmd.Flags().GetString("token")
		email, _ := cmd.Flags().GetString("email")
		// fullName flag exists but is ignored by service
		_, _ = cmd.Flags().GetString("fullName")

		profile, err := svc.UpdateProfile(context.Background(), token, email)
		if err != nil {
			fmt.Fprintf(os.Stderr, "update-profile failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Updated Profile:\nID: %s\nUsername: %s\nEmail: %s\n", profile.ID.String(), profile.Username, profile.Email)
	},
}

func init() {
	updateProfileCmd.Flags().String("token", "", "JWT token")
	updateProfileCmd.Flags().String("email", "", "New email address")
	updateProfileCmd.MarkFlagRequired("token")
	updateProfileCmd.MarkFlagRequired("email")
}
