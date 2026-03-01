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
	"google.golang.org/grpc/metadata"
)

var getProfileCmd = &cobra.Command{
	Use:   "get-profile",
	Short: "取得使用者個人資料",
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
		// Include JWT in metadata for authentication
		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", token))

		req := &secra_v1.TokenRequest{Token: token}
		res, err := client.GetProfile(ctx, req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "get-profile failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("ID: %s\nUsername: %s\nEmail: %s\n", res.Id, res.Username, res.Email)
	},
}

func init() {
	getProfileCmd.Flags().String("token", "", "Token")
	getProfileCmd.MarkFlagRequired("token")
}
