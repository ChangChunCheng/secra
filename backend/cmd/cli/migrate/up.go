package migrate

import (
	"github.com/spf13/cobra"
	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/db"
	"gitlab.com/jacky850509/secra/internal/storage"
)

// UpCmd applies all pending migrations
var UpCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply all pending migrations",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		dbWrapper := storage.NewDB(cfg.PostgresDSN, false)
		defer dbWrapper.Close()

		db.RunUp(dbWrapper.DB)
	},
}
