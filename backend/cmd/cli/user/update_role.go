package user

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/storage"
)

var (
	updateUsername string
	newRole        string
)

var updateRoleCmd = &cobra.Command{
	Use:   "update-role",
	Short: "Update user role",
	Long:  "Update a user's role to either 'user' or 'admin'.",
	Run: func(cmd *cobra.Command, args []string) {
		if newRole != "user" && newRole != "admin" {
			log.Fatalf("❌ Invalid role '%s'. Must be 'user' or 'admin'", newRole)
		}

		cfg := config.Load()
		db := storage.NewDB(cfg.PostgresDSN, false)
		defer db.Close()

		ctx := context.Background()

		// Check if user exists
		var user model.User
		err := db.DB.NewSelect().Model(&user).Where("username = ?", updateUsername).Scan(ctx)
		if err != nil {
			log.Fatalf("❌ User '%s' not found: %v", updateUsername, err)
		}

		oldRole := user.Role

		// Update role
		_, err = db.DB.NewUpdate().
			Model(&user).
			Set("role = ?", newRole).
			Where("username = ?", updateUsername).
			Exec(ctx)

		if err != nil {
			log.Fatalf("❌ Failed to update role: %v", err)
		}

		fmt.Printf("✅ Role updated successfully for user: %s\n", updateUsername)
		fmt.Printf("   Previous role: %s → New role: %s\n", oldRole, newRole)
	},
}

func init() {
	updateRoleCmd.Flags().StringVarP(&updateUsername, "username", "u", "", "Username [required]")
	updateRoleCmd.Flags().StringVarP(&newRole, "role", "r", "", "New role: user or admin [required]")
	updateRoleCmd.MarkFlagRequired("username")
	updateRoleCmd.MarkFlagRequired("role")

	Cmd.AddCommand(updateRoleCmd)
}
