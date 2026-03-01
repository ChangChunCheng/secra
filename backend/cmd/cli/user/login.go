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

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate user and return JWT token",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		conn, err := grpc.NewClient(cfg.GRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to connect to gRPC server: %v\n", err)
			os.Exit(1)
		}
		defer conn.Close()

		client := secra_v1.NewUserServiceClient(conn)

		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")

		req := &secra_v1.LoginRequest{
			Username: username,
			Password: password,
		}
		res, err := client.Login(context.Background(), req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "login failed: %v\n", err)
			os.Exit(1)
		}
		// parse ExpireAt string to int64
		fmt.Println(res.Token, res.ExpireAt)
	},
}

func init() {
	loginCmd.Flags().String("username", "", "Username")
	loginCmd.Flags().String("password", "", "Password")
	loginCmd.MarkFlagRequired("username")
	loginCmd.MarkFlagRequired("password")
}
