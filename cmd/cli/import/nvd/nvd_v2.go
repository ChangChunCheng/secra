// cmd/cli/import/nvd/nvd_v2.go
package nvd

import (
	"encoding/json"
	"log"
	"time"

	"github.com/spf13/cobra"

	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/fetcher"
	"gitlab.com/jacky850509/secra/internal/importer"
	nvd_v2_parser "gitlab.com/jacky850509/secra/internal/parser/nvd/v2"
	"gitlab.com/jacky850509/secra/internal/storage"
)

var (
	startDate string
	endDate   string
	apiKey    string
)

func init() {
	v2Nvd.Flags().StringVar(&startDate, "start", "", "Start date (YYYY-MM-DD) [required]")
	v2Nvd.Flags().StringVar(&endDate, "end", "", "End date (YYYY-MM-DD)")
	v2Nvd.Flags().StringVar(&apiKey, "apikey", "", "Optional NVD API key")
	v2Nvd.MarkFlagRequired("start")
}

var v2Nvd = &cobra.Command{
	Use:   "v2",
	Short: "NVD v2 feed",
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
			log.Fatalf("❌ --end date must be after --start date")
		}
		if end.Sub(start) > 30*24*time.Hour {
			log.Fatalf("❌ Date range cannot exceed 30 days")
		}
		ImportNvdv2(start, end)
	},
}

func ImportNvdv2(start, end time.Time) {
	cfg = config.Load()
	db = storage.NewDB(cfg.PostgresDSN, false)
	defer db.Close()

	// Step 1: 下載 feed 資料
	data, err := fetcher.FetchNvdv2Feed(cfg.NvdURLv2, fetcher.NvdV2QueryParams{
		PubStartDate:   start,
		PubEndDate:     end,
		StartIndex:     0,
		ResultsPerPage: 2000,
		ApiKey:         apiKey,
	})
	if err != nil {
		log.Fatalf("❌ Failed to fetch NVD v2: %v", err)
	}

	var feed nvd_v2_parser.Nvdv2CveFeed
	if err := json.Unmarshal(data, &feed); err != nil {
		log.Fatalf("❌ Failed to parse NVD v2 JSON: %v", err)
	}
	log.Printf("✅ Feed parsed with %d items.", len(feed.Vulnerabilities))

	// Step 2: 轉換 CVEs
	cves, err := nvd_v2_parser.ConvertToCVEsFromV2(&feed)
	if err != nil {
		log.Fatalf("❌ Failed to convert CVEs: %v", err)
	}

	// Step 3: 確保資料來源
	source, err := ensureCveSource(db.DB, "nvd-v2", "", cfg.NvdURLv2)
	if err != nil {
		log.Fatalf("❌ Failed to ensure source: %v", err)
	}

	log.Printf("📦 Importing %d CVEs...", len(cves))
	if err := importer.ImportCVEs(db.DB, source.ID, cves); err != nil {
		log.Fatalf("❌ CVE import failed: %v", err)
	}

	// Step 4: 擷取 vendor/product 關聯
	vendors, products, relations := nvd_v2_parser.ExtractVendorsAndProductsFromv2(&feed)

	// Step 5: 先插入 vendors
	if err := importer.ImportVendorsAndProductsFromv2(db.DB, vendors, nil, nil, nil, nil); err != nil {
		log.Fatalf("❌ Vendor insert failed: %v", err)
	}

	// Step 6: 補上 products 的 vendor UUID
	vendorMap, err := importer.BuildVendorMap(db.DB)
	if err != nil {
		log.Fatalf("❌ Failed to build vendor map: %v", err)
	}
	for i := range products {
		if realID, ok := vendorMap[products[i].VendorID]; ok {
			products[i].VendorID = realID
		} else {
			log.Printf("⚠️ Vendor not found for product: %s (vendor=%s)", products[i].Name, products[i].VendorID)
		}
	}

	// Step 7: 插入 products
	if err := importer.ImportVendorsAndProductsFromv2(db.DB, nil, products, nil, nil, nil); err != nil {
		log.Fatalf("❌ Product insert failed: %v", err)
	}

	// Step 8: 構建對照表
	productMap, err := importer.BuildProductMap(db.DB)
	if err != nil {
		log.Fatalf("❌ Failed to build product map: %v", err)
	}

	// Step 9: 使用 relations 裡的 UID 建立 cveMap
	cveUIDs := make([]string, 0, len(relations))
	for _, r := range relations {
		cveUIDs = append(cveUIDs, r.CveSourceUID)
	}
	cveMap, err := importer.BuildCveMap(db.DB, cveUIDs)
	if err != nil {
		log.Fatalf("❌ Failed to build CVE map: %v", err)
	}

	// Step 10: 建立 CVE-product 關聯
	log.Printf("🔗 Linking %d CVEs to products...", len(relations))
	if err := importer.ImportVendorsAndProductsFromv2(db.DB, nil, nil, relations, cveMap, productMap); err != nil {
		log.Fatalf("❌ Relation import failed: %v", err)
	}

	log.Println("✅ NVD v2 Import complete.")
}
