package user

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"gitlab.com/jacky850509/secra/internal/auth"
	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/storage"
)

var (
	resetUsername string
	resetPassword string
)

var resetPasswordCmd = &cobra.Command{
	Use:   "reset-password",
	Short: "Reset user password",
	Long:  "Reset a user's password by username. The password will be hashed using bcrypt.",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		db := storage.NewDB(cfg.PostgresDSN, false)
		defer db.Close()

		ctx := context.Background()

		// Check if user exists
		var user model.User
		err := db.DB.NewSelect().Model(&user).Where("username = ?", resetUsername).Scan(ctx)
		if err != nil {
			log.Fatalf("❌ User '%s' not found: %v", resetUsername, err)
		}

		// Hash the new password
		hashedPassword, err := auth.HashPassword(resetPassword)
		if err != nil {
			log.Fatalf("❌ Failed to hash password: %v", err)
		}

		// Update password
		_, err = db.DB.NewUpdate().
			Model(&user).
			Set("password_hash = ?", hashedPassword).
			Set("must_change_password = false").
			Where("username = ?", resetUsername).
			Exec(ctx)

		if err != nil {
			log.Fatalf("❌ Failed to update password: %v", err)
		}

		fmt.Printf("✅ Password reset successfully for user: %s\n", resetUsername)
		fmt.Printf("🔑 New password: %s\n", resetPassword)
	},
}

func init() {
	resetPasswordCmd.Flags().StringVarP(&resetUsername, "username", "u", "", "Username [required]")
	resetPasswordCmd.Flags().StringVarP(&resetPassword, "password", "p", "", "New password [required]")
	resetPasswordCmd.MarkFlagRequired("username")
	resetPasswordCmd.MarkFlagRequired("password")

	Cmd.AddCommand(resetPasswordCmd)
}
