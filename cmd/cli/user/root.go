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

var cfg *config.AppConfig
var db *storage.DBWrapper

var Cmd = &cobra.Command{
	Use:   "user",
	Short: "User management commands",
}

func init() {
	Cmd.AddCommand(registerCmd)
}

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new user",
	Run: func(cmd *cobra.Command, args []string) {
		// initialize config and db
		// load config and initialize DB
		cfg = config.Load()
		db = storage.NewDB(cfg.PostgresDSN, false)
		userRepo := repo.NewUserRepository(db.DB)
		svc := service.NewUserService(userRepo)

		username, _ := cmd.Flags().GetString("username")
		email, _ := cmd.Flags().GetString("email")
		password, _ := cmd.Flags().GetString("password")
		// here password is plain; hash or adjust per service signature

		user, err := svc.Register(context.Background(), username, email, password)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Registered user: %s (id=%d)\n", user.Email, user.ID)
	},
}

func init() {
	registerCmd.Flags().String("username", "", "Username for login")
	registerCmd.Flags().String("email", "", "User email")
	registerCmd.Flags().String("password", "", "User password")
	registerCmd.MarkFlagRequired("username")
	registerCmd.MarkFlagRequired("email")
	registerCmd.MarkFlagRequired("password")
}
