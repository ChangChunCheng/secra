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

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new user",
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
		email, _ := cmd.Flags().GetString("email")
		password, _ := cmd.Flags().GetString("password")

		req := &secra_v1.RegisterRequest{
			Username:        username,
			Email:           email,
			Password:        password,
			ConfirmPassword: password,
		}
		res, err := client.Register(context.Background(), req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error registering user: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Registered user: %s (message=%s)\n", email, res.Message)
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
