package nvd

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/uptrace/bun"

	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/fetcher"
	"gitlab.com/jacky850509/secra/internal/importer"
	"gitlab.com/jacky850509/secra/internal/model"
	nvd_v2_parser "gitlab.com/jacky850509/secra/internal/parser/nvd/v2"
	"gitlab.com/jacky850509/secra/internal/storage"
)

var (
	startDate string
	endDate   string
	apiKey    string
	force     bool
)

func init() {
	v2Nvd.Flags().StringVar(&startDate, "start", "", "Start date (YYYY-MM-DD) [required]")
	v2Nvd.Flags().StringVar(&endDate, "end", "", "End date (YYYY-MM-DD)")
	v2Nvd.Flags().StringVar(&apiKey, "apikey", "", "Optional NVD API key")
	v2Nvd.Flags().BoolVarP(&force, "force", "f", false, "Force re-import even if data exists for the date")
	v2Nvd.MarkFlagRequired("start")
}

var v2Nvd = &cobra.Command{
	Use:   "v2",
	Short: "Import CVEs from NVD API v2 with resume capability",
	Run: func(cmd *cobra.Command, args []string) {
		start, err := time.Parse(time.DateOnly, startDate)
		if err != nil {
			log.Fatalf("❌ Invalid --start date format (expected YYYY-MM-DD): %v", err)
		}
		end := time.Now().UTC()
		if endDate != "" {
			end, err = time.Parse(time.DateOnly, endDate)
			if err != nil {
				log.Fatalf("❌ Invalid --end date format (expected YYYY-MM-DD): %v", err)
			}
		}
		if end.Before(start) {
			log.Fatalf("❌ --end must be after --start")
		}

		cfg := config.Load()
		db := storage.NewDB(cfg.PostgresDSN, false)
		defer db.Close()

		// Process day by day to support granular resume
		for current := start; !current.After(end); current = current.AddDate(0, 0, 1) {
			dayStr := current.Format(time.DateOnly)
			
			if !force {
				// Check if we already have data for this specific day
				count, _ := db.DB.NewSelect().Model((*model.CVE)(nil)).
					Where("published_at::date = ?", dayStr).
					Count(cmd.Context())
				
				if count > 0 {
					log.Printf("⏩ Skipping %s (already has %d CVEs). Use -f to force re-import.", dayStr, count)
					continue
				}
			}

			dayEnd := current.AddDate(0, 0, 1) // Next day 00:00:00
			log.Printf("📅 Processing: %s", dayStr)
			
			ImportNvdv2Daily(cmd.Context(), db.DB, cfg, current, dayEnd)

			// Delay to respect Rate Limit
			delay := 6 * time.Second
			if apiKey != "" {
				delay = 1 * time.Second
			}
			if !current.Equal(end) {
				time.Sleep(delay)
			}
		}
	},
}

func ImportNvdv2Daily(ctx context.Context, db *bun.DB, cfg *config.AppConfig, start, end time.Time) {
	pageSize := 2000
	startIndex := 0

	source, err := ensureCveSource(db, "nvd-v2", "", cfg.NvdURLv2)
	if err != nil {
		log.Fatalf("❌ Failed to ensure source: %v", err)
	}

	var (
		allCVEs       []model.CVE
		allVendors    []model.Vendor
		allProducts   []model.Product
		allRelations  []nvd_v2_parser.CVEProductRelation
		allReferences []model.CVEReference
		allWeaknesses []model.CVEWeakness
		allSourceUIDs = make(map[string]struct{})
	)

	for {
		data, err := fetcher.FetchNvdv2Feed(cfg.NvdURLv2, fetcher.NvdV2QueryParams{
			PubStartDate:   start,
			PubEndDate:     end,
			StartIndex:     startIndex,
			ResultsPerPage: pageSize,
			ApiKey:         apiKey,
		})
		if err != nil {
			log.Fatalf("❌ Failed to fetch NVD v2 feed: %v", err)
		}

		var feed nvd_v2_parser.Nvdv2CveFeed
		if err := json.Unmarshal(data, &feed); err != nil {
			log.Fatalf("❌ Failed to parse NVD v2 feed JSON: %v", err)
		}

		if len(feed.Vulnerabilities) == 0 {
			break
		}

		cves, _ := nvd_v2_parser.ConvertToCVEsFromV2(&feed)
		allCVEs = append(allCVEs, cves...)

		vendors, products, relations, references, weaknesses := nvd_v2_parser.ExtractAllFromV2(&feed)
		allVendors = append(allVendors, vendors...)
		allProducts = append(allProducts, products...)
		allRelations = append(allRelations, relations...)
		allReferences = append(allReferences, references...)
		allWeaknesses = append(allWeaknesses, weaknesses...)

		for _, cve := range cves { allSourceUIDs[cve.SourceUID] = struct{}{} }

		startIndex += len(feed.Vulnerabilities)
		if startIndex >= feed.TotalResults {
			break
		}
		time.Sleep(1 * time.Second)
	}

	if len(allCVEs) == 0 {
		log.Printf("ℹ️ No data found for this date.")
		return
	}

	// Import Logic (using existing importer logic)
	importer.ImportCVEs(db, source.ID, allCVEs)
	importer.ImportVendorsAndProductsFromv2(db, allVendors, nil, nil, nil, nil)
	
	vendorMap, _ := importer.BuildVendorMap(db)
	for i := range allProducts {
		if realID, ok := vendorMap[allProducts[i].VendorID]; ok {
			allProducts[i].VendorID = realID
		}
	}
	importer.ImportVendorsAndProductsFromv2(db, nil, allProducts, nil, nil, nil)

	productMap, _ := importer.BuildProductMap(db)
	cveUIDList := make([]string, 0, len(allSourceUIDs))
	for uid := range allSourceUIDs { cveUIDList = append(cveUIDList, uid) }
	cveMap, _ := importer.BuildCveMap(db, cveUIDList)

	importer.ImportVendorsAndProductsFromv2(db, nil, nil, allRelations, cveMap, productMap)
	importer.ImportReferences(db, allReferences, cveMap)
	importer.ImportWeaknesses(db, allWeaknesses, cveMap)

	log.Printf("✅ Daily import complete: %d CVEs processed.", len(allCVEs))
}

func ensureCveSource(db *bun.DB, name, description, url string) (*model.CVESource, error) {
	ctx := context.Background()
	source := new(model.CVESource)
	err := db.NewSelect().Model(source).Where("name = ?", name).Scan(ctx)
	if err == nil {
		return source, nil
	}

	source = &model.CVESource{
		ID:          uuid.New().String(),
		Name:        name,
		Type:        "nvd",
		URL:         url,
		Description: description,
		Enabled:     true,
	}

	_, err = db.NewInsert().Model(source).Exec(ctx)
	return source, err
}
