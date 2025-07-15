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

var getProfileCmd = &cobra.Command{
	Use:   "get-profile",
	Short: "取得使用者個人資料",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		db := storage.NewDB(cfg.PostgresDSN, false)
		userRepo := repo.NewUserRepository(db.DB)
		svc := service.NewUserService(userRepo)

		token, _ := cmd.Flags().GetString("token")
		profile, err := svc.GetProfile(context.Background(), token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "get-profile failed: %v\n", err)
			os.Exit(1)
		}
		// 輸出個人資料
		fmt.Printf("ID: %s\nUsername: %s\nEmail: %s\n", profile.ID.String(), profile.Username, profile.Email)
	},
}

func init() {
	getProfileCmd.Flags().String("token", "", "JWT token")
	getProfileCmd.MarkFlagRequired("token")
	Cmd.AddCommand(getProfileCmd)
}
