package usercmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"

	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/repo"
	"gitlab.com/jacky850509/secra/internal/storage"
)

var (
	username string
	email    string
	password string
	role     string
)

var registerLocalCmd = &cobra.Command{
	Use:   "register-local",
	Short: "Register a local user",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		dbw := storage.NewDB(cfg.PostgresDSN, false)
		defer dbw.Close()

		r := repo.NewUserRepo(dbw.DB)
		if role == "" {
			role = "user"
		}
		err := r.CreateLocalUser(context.Background(), username, email, password, role)
		if err != nil {
			log.Fatalf("Failed to create user: %v", err)
		}
		log.Println("✅ User created")
	},
}

func init() {
	registerLocalCmd.Flags().StringVar(&username, "username", "", "Username")
	registerLocalCmd.Flags().StringVar(&email, "email", "", "Email")
	registerLocalCmd.Flags().StringVar(&password, "password", "", "Password")
	registerLocalCmd.Flags().StringVar(&role, "role", "user", "Role (user/admin)")
	registerLocalCmd.MarkFlagRequired("username")
	registerLocalCmd.MarkFlagRequired("email")
	registerLocalCmd.MarkFlagRequired("password")
}
