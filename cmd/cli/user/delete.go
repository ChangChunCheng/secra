package user

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/storage"
)

var (
	deleteUsername string
	forceDelete    bool
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a user",
	Long:  "Delete a user from the system. This action cannot be undone.",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		db := storage.NewDB(cfg.PostgresDSN, false)
		defer db.Close()

		ctx := context.Background()

		// Check if user exists
		var user model.User
		err := db.DB.NewSelect().Model(&user).Where("username = ?", deleteUsername).Scan(ctx)
		if err != nil {
			log.Fatalf("❌ User '%s' not found: %v", deleteUsername, err)
		}

		// Confirm deletion unless --force is used
		if !forceDelete {
			fmt.Printf("⚠️  Are you sure you want to delete user '%s' (%s, %s)? [y/N]: ", user.Username, user.Email, user.Role)
			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				log.Fatalf("❌ Failed to read input: %v", err)
			}
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				fmt.Println("❌ Deletion cancelled")
				return
			}
		}

		// Delete user
		_, err = db.DB.NewDelete().Model(&user).Where("username = ?", deleteUsername).Exec(ctx)
		if err != nil {
			log.Fatalf("❌ Failed to delete user: %v", err)
		}

		fmt.Printf("✅ User '%s' deleted successfully\n", deleteUsername)
	},
}

func init() {
	deleteCmd.Flags().StringVarP(&deleteUsername, "username", "u", "", "Username [required]")
	deleteCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "Skip confirmation prompt")
	deleteCmd.MarkFlagRequired("username")

	Cmd.AddCommand(deleteCmd)
}
