package nvd

import (
	"encoding/json"
	"log"
	"time"

	"github.com/spf13/cobra"

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
)

func init() {
	v2Nvd.Flags().StringVar(&startDate, "start", "", "Start date (YYYY-MM-DD) [required]")
	v2Nvd.Flags().StringVar(&endDate, "end", "", "End date (YYYY-MM-DD)")
	v2Nvd.Flags().StringVar(&apiKey, "apikey", "", "Optional NVD API key")
	v2Nvd.MarkFlagRequired("start")
}

var v2Nvd = &cobra.Command{
	Use:   "v2",
	Short: "Import CVEs from NVD API v2",
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

		// Chunked import for large date ranges
		// NVD API v2 allows max 120 days, but we use 30 days to keep memory usage low
		chunkSize := 30 * 24 * time.Hour
		for currentStart := start; currentStart.Before(end); {
			currentEnd := currentStart.Add(chunkSize)
			if currentEnd.After(end) {
				currentEnd = end
			}

			log.Printf("📅 Processing chunk: %s to %s", currentStart.Format(time.DateOnly), currentEnd.Format(time.DateOnly))
			ImportNvdv2Chunk(currentStart, currentEnd)

			currentStart = currentEnd
			// Add a delay between chunks to respect Rate Limit
			// NVD recommends 6 seconds without API key, or less with key
			if currentStart.Before(end) {
				delay := 6 * time.Second
				if apiKey != "" {
					delay = 1 * time.Second
				}
				log.Printf("Waiting %v before next chunk...", delay)
				time.Sleep(delay)
			}
		}
	},
}

func ImportNvdv2Chunk(start, end time.Time) {
	cfg = config.Load()
	db = storage.NewDB(cfg.PostgresDSN, false)
	defer db.Close()

	pageSize := 2000
	startIndex := 0

	source, err := ensureCveSource(db.DB, "nvd-v2", "", cfg.NvdURLv2)
	if err != nil {
		log.Fatalf("❌ Failed to ensure source: %v", err)
	}

	// Step 0: 準備暫存累積
	var (
		allCVEs       []model.CVE
		allRelations  []nvd_v2_parser.CVEProductRelation
		allVendors    []model.Vendor
		allProducts   []model.Product
		allReferences []model.CVEReference
		allWeaknesses []model.CVEWeakness
		allSourceUIDs = make(map[string]struct{})
	)

	for {
		log.Printf("📥 Fetching NVD v2 feed: startIndex=%d...", startIndex)

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
			log.Println("✅ No more vulnerabilities to process.")
			break
		}

		log.Printf("✅ Parsed %d vulnerabilities, totalResults=%d.", len(feed.Vulnerabilities), feed.TotalResults)

		// 轉換 CVEs
		cves, err := nvd_v2_parser.ConvertToCVEsFromV2(&feed)
		if err != nil {
			log.Fatalf("❌ Failed to convert CVEs: %v", err)
		}
		allCVEs = append(allCVEs, cves...)

		// 擷取 vendor/product/relation/reference/weakness
		vendors, products, relations, references, weaknesses := nvd_v2_parser.ExtractAllFromV2(&feed)

		allVendors = append(allVendors, vendors...)
		allProducts = append(allProducts, products...)
		allRelations = append(allRelations, relations...)
		allReferences = append(allReferences, references...)
		allWeaknesses = append(allWeaknesses, weaknesses...)

		// 收集所有 source_uid
		for _, cve := range cves {
			allSourceUIDs[cve.SourceUID] = struct{}{}
		}
		for _, ref := range references {
			allSourceUIDs[ref.CVEID] = struct{}{}
		}
		for _, w := range weaknesses {
			allSourceUIDs[w.CVEID] = struct{}{}
		}

		// 檢查是否還有下一頁
		startIndex += len(feed.Vulnerabilities)
		if startIndex >= feed.TotalResults {
			log.Println("✅ All pages fetched.")
			break
		}
		// Delay between pages if no API key
		if apiKey == "" {
			time.Sleep(1 * time.Second)
		}
	}

	// Step 1: Insert CVEs
	log.Printf("📦 Importing %d CVEs...", len(allCVEs))
	if err := importer.ImportCVEs(db.DB, source.ID, allCVEs); err != nil {
		log.Fatalf("❌ CVE import failed: %v", err)
	}

	// Step 2: Insert vendors
	log.Printf("📦 Importing %d vendors...", len(allVendors))
	if err := importer.ImportVendorsAndProductsFromv2(db.DB, allVendors, nil, nil, nil, nil); err != nil {
		log.Fatalf("❌ Vendor insert failed: %v", err)
	}

	// Step 3: Resolve vendor ID for products
	vendorMap, err := importer.BuildVendorMap(db.DB)
	if err != nil {
		log.Fatalf("❌ Failed to build vendor map: %v", err)
	}
	for i := range allProducts {
		if realID, ok := vendorMap[allProducts[i].VendorID]; ok {
			allProducts[i].VendorID = realID
		} else {
			log.Printf("⚠️ Vendor not found for product: %s (vendor=%s)", allProducts[i].Name, allProducts[i].VendorID)
		}
	}

	// Step 4: Insert products
	log.Printf("📦 Importing %d products...", len(allProducts))
	if err := importer.ImportVendorsAndProductsFromv2(db.DB, nil, allProducts, nil, nil, nil); err != nil {
		log.Fatalf("❌ Product insert failed: %v", err)
	}

	// Step 5: Insert CVE ↔ Product relations
	productMap, err := importer.BuildProductMap(db.DB)
	if err != nil {
		log.Fatalf("❌ Failed to build product map: %v", err)
	}

	cveUIDList := make([]string, 0, len(allSourceUIDs))
	for uid := range allSourceUIDs {
		cveUIDList = append(cveUIDList, uid)
	}

	cveMap, err := importer.BuildCveMap(db.DB, cveUIDList)
	if err != nil {
		log.Fatalf("❌ Failed to build CVE map: %v", err)
	}

	log.Printf("🔗 Linking %d CVEs to products...", len(allRelations))
	if err := importer.ImportVendorsAndProductsFromv2(db.DB, nil, nil, allRelations, cveMap, productMap); err != nil {
		log.Fatalf("❌ Relation insert failed: %v", err)
	}

	// Step 6: Insert references
	log.Printf("🔗 Importing %d references...", len(allReferences))
	if err := importer.ImportReferences(db.DB, allReferences, cveMap); err != nil {
		log.Fatalf("❌ Failed to import references: %v", err)
	}

	// Step 7: Insert weaknesses
	log.Printf("🔗 Importing %d weaknesses...", len(allWeaknesses))
	if err := importer.ImportWeaknesses(db.DB, allWeaknesses, cveMap); err != nil {
		log.Fatalf("❌ Failed to import weaknesses: %v", err)
	}

	log.Println("✅ NVD v2 Import fully complete for this chunk.")
}
