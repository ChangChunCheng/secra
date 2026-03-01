package nvd

import (
	"log"
	"time"

	"github.com/spf13/cobra"

	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/service"
	"gitlab.com/jacky850509/secra/internal/storage"
)

var (
	startDate string
	endDate   string
	force     bool
)

func init() {
	v2Nvd.Flags().StringVar(&startDate, "start", "", "Start date (YYYY-MM-DD) [required]")
	v2Nvd.Flags().StringVar(&endDate, "end", "", "End date (YYYY-MM-DD)")
	v2Nvd.Flags().BoolVarP(&force, "force", "f", false, "Force re-import even if data exists for the date")
	v2Nvd.MarkFlagRequired("start")
}

var v2Nvd = &cobra.Command{
	Use:   "v2",
	Short: "Import CVEs from NVD API v2 with smart interval merging",
	Run: func(cmd *cobra.Command, args []string) {
		start, err := time.Parse(time.DateOnly, startDate)
		if err != nil {
			log.Fatalf("❌ Invalid start date format: %v", err)
		}

		end := time.Now().UTC()
		if endDate != "" {
			end, err = time.Parse(time.DateOnly, endDate)
			if err != nil {
				log.Fatalf("❌ Invalid end date format: %v", err)
			}
		}

		cfg = config.Load()
		db = storage.NewDB(cfg.PostgresDSN, false)
		defer db.Close()

		// Use shared NVD import service (reused by scheduler)
		importService := service.NewNVDImportService(db.DB, cfg)
		count, err := importService.ImportDateRange(cmd.Context(), start, end, force)

		if err != nil {
			log.Fatalf("❌ Import failed: %v", err)
		}

		if count == 0 {
			log.Println("✅ No missing intervals found. Data is up to date.")
		} else {
			log.Printf("✅ Import completed successfully. Total: %d CVEs", count)
		}
	},
}
