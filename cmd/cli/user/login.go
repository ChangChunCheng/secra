package user

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/repo"
	"gitlab.com/jacky850509/secra/internal/service"
	"gitlab.com/jacky850509/secra/internal/storage"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate user and return JWT token",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		db := storage.NewDB(cfg.PostgresDSN, false)
		userRepo := repo.NewUserRepository(db.DB)
		svc := service.NewUserService(userRepo)

		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")

		token, expireAt, err := svc.Login(context.Background(), username, password)
		if err != nil {
			fmt.Fprintf(os.Stderr, "login failed: %v\n", err)
			os.Exit(1)
		}
		t := time.Unix(expireAt, 0)
		fmt.Println(token, t.Format("2006-01-02T15:04:05"))
	},
}

func init() {
	loginCmd.Flags().String("username", "", "Username")
	loginCmd.Flags().String("password", "", "Password")
	loginCmd.MarkFlagRequired("username")
	loginCmd.MarkFlagRequired("password")
}
