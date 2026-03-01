package user

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/repo"
	"gitlab.com/jacky850509/secra/internal/service"
	"gitlab.com/jacky850509/secra/internal/storage"
)

var (
	newUsername string
	newEmail    string
	newPassword string
	isAdmin     bool
)

var Cmd = &cobra.Command{
	Use:   "user",
	Short: "User management commands",
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new user with specified role",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		db := storage.NewDB(cfg.PostgresDSN, false)
		defer db.Close()

		userRepo := repo.NewUserRepository(db.DB)
		svc := service.NewUserService(userRepo)

		u, err := svc.Register(context.Background(), newUsername, newEmail, newPassword, newPassword)
		if err != nil {
			log.Fatalf("❌ Failed to create user: %v", err)
		}

		if isAdmin {
			_, err = db.DB.NewUpdate().Table("users").Set("role = 'admin'").Where("id = ?", u.ID).Exec(context.Background())
			if err != nil {
				log.Fatalf("❌ Failed to elevate to admin: %v", err)
			}
			fmt.Printf("✅ Admin user created: %s (%s)\n", newUsername, newEmail)
		} else {
			fmt.Printf("✅ Regular user created: %s (%s)\n", newUsername, newEmail)
		}
	},
}

func init() {
	createCmd.Flags().StringVarP(&newUsername, "username", "u", "", "Username [required]")
	createCmd.Flags().StringVarP(&newEmail, "email", "e", "", "Email [required]")
	createCmd.Flags().StringVarP(&newPassword, "password", "p", "", "Password [required]")
	createCmd.Flags().BoolVar(&isAdmin, "admin", false, "Create as admin user")
	createCmd.MarkFlagRequired("username")
	createCmd.MarkFlagRequired("email")
	createCmd.MarkFlagRequired("password")

	Cmd.AddCommand(createCmd)
}
