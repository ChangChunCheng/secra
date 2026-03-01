package health

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/storage"
	"gitlab.com/jacky850509/secra/internal/service"
)

var (
	checkType string
	testEmail string
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check system health or test services",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()

		switch checkType {
		case "db":
			db := storage.NewDB(cfg.PostgresDSN, false)
			defer db.Close()
			if err := db.DB.Ping(); err != nil {
				log.Fatalf("❌ Database connection failed: %v", err)
			}
			fmt.Println("✅ Database is healthy.")

		case "grpc":
			conn, err := grpc.Dial(cfg.GRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Fatalf("❌ gRPC connection failed: %v", err)
			}
			defer conn.Close()
			fmt.Println("✅ gRPC server is reachable.")

		case "email":
			if testEmail == "" {
				log.Fatal("❌ Please provide a recipient email using --to flag")
			}
			db := storage.NewDB(cfg.PostgresDSN, false)
			defer db.Close()
			
			notifier := service.NewNotificationService(cfg.SMTPConfig, db.DB)
			log.Printf("📧 Sending test email to %s...", testEmail)
			err := notifier.SendEmail(context.Background(), testEmail, "SECRA SMTP Test", "Congratulations! Your SECRA notification system is working correctly.")
			if err != nil {
				log.Fatalf("❌ Email failed: %v", err)
			}
			fmt.Println("✅ Test email sent successfully!")

		default:
			fmt.Println("Usage: secra health check --type [db|grpc|email] [--to recipient@example.com]")
		}
	},
}

func init() {
	checkCmd.Flags().StringVar(&checkType, "type", "db", "Type of health check (db, grpc, email)")
	checkCmd.Flags().StringVar(&testEmail, "to", "", "Recipient email for testing notification")
}
