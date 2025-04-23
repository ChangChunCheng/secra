package migrate

import (
	"github.com/spf13/cobra"
	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/db"
	"gitlab.com/jacky850509/secra/internal/storage"
)

// StatusCmd 顯示目前 migration 的執行狀態
var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		dbWrapper := storage.NewDB(cfg.PostgresDSN, true)
		defer dbWrapper.Close()

		db.RunStatus(dbWrapper.DB)
	},
}
