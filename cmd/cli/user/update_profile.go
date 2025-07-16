package user

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	secra_v1 "gitlab.com/jacky850509/secra/api/gen/v1"
	"gitlab.com/jacky850509/secra/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var updateProfileCmd = &cobra.Command{
	Use:   "update-profile",
	Short: "更新使用者個人資料 (目前僅支援更新 Email)",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		conn, err := grpc.NewClient(cfg.GRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to connect to gRPC server: %v\n", err)
			os.Exit(1)
		}
		defer conn.Close()

		client := secra_v1.NewUserServiceClient(conn)

		token, _ := cmd.Flags().GetString("token")
		email, _ := cmd.Flags().GetString("email")
		password, _ := cmd.Flags().GetString("password")

		req := &secra_v1.UpdateProfileRequest{
			Token:           token,
			Email:           email,
			Password:        password,
			ConfirmPassword: password,
		}
		res, err := client.UpdateProfile(context.Background(), req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "update-profile failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Updated Profile:\nID: %s\nUsername: %s\nEmail: %s\n", res.Id, res.Username, res.Email)
	},
}

func init() {
	updateProfileCmd.Flags().String("token", "", "JWT token")
	updateProfileCmd.Flags().String("email", "", "New email address")
	updateProfileCmd.Flags().String("password", "", "New password (leave empty to keep unchanged)")
	updateProfileCmd.MarkFlagRequired("token")
	updateProfileCmd.MarkFlagRequired("email")
}
