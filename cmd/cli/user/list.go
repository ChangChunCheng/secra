package user

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/storage"
)

var (
	filterRole   string
	filterStatus string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all users",
	Long:  "List all users in the system with optional filters for role and status.",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		db := storage.NewDB(cfg.PostgresDSN, false)
		defer db.Close()

		ctx := context.Background()

		query := db.DB.NewSelect().Model((*model.User)(nil))

		// Apply filters
		if filterRole != "" {
			query = query.Where("role = ?", filterRole)
		}
		if filterStatus != "" {
			query = query.Where("status = ?", filterStatus)
		}

		query = query.Order("created_at ASC")

		var users []model.User
		err := query.Scan(ctx, &users)
		if err != nil {
			log.Fatalf("❌ Failed to fetch users: %v", err)
		}

		if len(users) == 0 {
			fmt.Println("No users found")
			return
		}

		// Print header
		fmt.Printf("%-20s %-30s %-10s %-10s %-20s\n", "USERNAME", "EMAIL", "ROLE", "STATUS", "CREATED AT")
		fmt.Println("────────────────────────────────────────────────────────────────────────────────────────────")

		// Print users
		for _, user := range users {
			fmt.Printf("%-20s %-30s %-10s %-10s %-20s\n",
				user.Username,
				user.Email,
				user.Role,
				user.Status,
				user.CreatedAt.Format("2006-01-02 15:04:05"),
			)
		}

		fmt.Printf("\nTotal: %d users\n", len(users))
	},
}

func init() {
	listCmd.Flags().StringVarP(&filterRole, "role", "r", "", "Filter by role (user/admin)")
	listCmd.Flags().StringVarP(&filterStatus, "status", "s", "", "Filter by status (active/inactive)")

	Cmd.AddCommand(listCmd)
}
