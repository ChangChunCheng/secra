package nvd

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/fetcher"
	"gitlab.com/jacky850509/secra/internal/importer"
	nvd_v1_parser "gitlab.com/jacky850509/secra/internal/parser/nvd/v1"
	"gitlab.com/jacky850509/secra/internal/storage"
)

var (
	recent   bool
	modified bool
	year     uint16
)

var v1Nvd = &cobra.Command{
	Use:   "v1",
	Short: "NVD v1 feed",
	Run: func(cmd *cobra.Command, args []string) {
		if !(recent || modified || year >= 2002) {
			log.Printf("❌ No feed type specified. Use --recent=%t, --modified=%t, or --year=%d.", recent, modified, year)
			return
		}
		ImportNvdv1(recent, modified, year)
	},
}

func init() {
	v1Nvd.Flags().BoolVar(&recent, "recent", false, "Recent import even if data exists")
	v1Nvd.Flags().BoolVar(&modified, "modified", false, "Modified import even if data exists")
	v1Nvd.Flags().Uint16Var(&year, "year", 0, "Year of feed (NVD only)")
}

func ImportNvdv1(recent, modified bool, year uint16) {
	cfg = config.Load()
	db = storage.NewDB(cfg.PostgresDSN, false)
	defer db.Close()

	var data []byte
	var err error
	processData := make(map[string][]byte)

	if recent {
		log.Printf("📥 Downloading NVD, url=%s, recent feed...", cfg.NvdURLv1)
		data, err = fetcher.DownloadNvdv1FeedRecent(cfg.NvdURLv1)
		if err != nil {
			log.Fatalf("❌ Failed to fetch feed: %v", err)
		} else {
			processData["recent"] = data
		}
	}
	if modified {
		log.Printf("📥 Downloading NVD, url=%s, modified feed...", cfg.NvdURLv1)
		data, err = fetcher.DownloadNvdv1FeedModified(cfg.NvdURLv1)
		if err != nil {
			log.Fatalf("❌ Failed to fetch feed: %v", err)
		} else {
			processData["modified"] = data
		}
	}
	if year >= 2002 {
		log.Printf("📥 Downloading NVD, url=%s, feed for year %d...", cfg.NvdURLv1, year)
		data, err = fetcher.DownloadNvdv1FeedYear(year, cfg.NvdURLv1)
		if err != nil {
			log.Fatalf("❌ Failed to fetch feed: %v", err)
		} else {
			processData[fmt.Sprintf("year-%d", year)] = data
		}
	}
	if len(processData) == 0 {
		log.Printf("❌ No feed type specified. Use --recent, --modified, or --year.")
		return
	}
	log.Printf("📥 Downloaded %d feeds.", len(processData))

	log.Printf("📦 Processing %d feeds...", len(processData))
	for feedName, data := range processData {
		log.Printf("📦 Processing %s feed...", feedName)
		if err = ProcessImportNvdv1(db, data, feedName); err != nil {
			log.Fatalf("❌ Failed to process feed %s: %v", feedName, err)
		}
		log.Printf("✅ %s feed processed successfully.", feedName)
	}
	log.Println("✅ All feeds processed.")
}

func ProcessImportNvdv1(db *storage.DBWrapper, data []byte, sourceName string) error {
	var feed nvd_v1_parser.Nvdv1CveFeed
	if err := json.Unmarshal(data, &feed); err != nil {
		log.Fatalf("❌ Failed to parse feed JSON: %v", err)
		return err
	}
	log.Printf("✅ Feed parsed with %d items.", len(feed.Items))

	cves, err := nvd_v1_parser.ConvertToCVEs(&feed)
	if err != nil {
		log.Fatalf("❌ Failed to convert CVEs: %v", err)
		return err
	}

	source, err := ensureCveSource(db.DB, sourceName, "NVD v1 data feed", cfg.NvdURLv1)
	if err != nil {
		log.Fatalf("❌ Failed to ensure source: %v", err)
		return err
	}

	log.Printf("📦 Importing %d CVEs...", len(cves))
	if err := importer.ImportCVEs(db.DB, source.ID, cves); err != nil {
		log.Fatalf("❌ CVE import failed: %v", err)
		return err
	}

	log.Println("🔍 Extracting vendors/products from configurations...")
	vendors, products, relations := nvd_v1_parser.ExtractVendorsAndProductsFromv1(&feed)

	log.Printf("📦 Inserting %d vendors...", len(vendors))
	if err := importer.ImportVendorsAndProductsFromv1(db.DB, vendors, nil, nil, nil, nil); err != nil {
		log.Fatalf("❌ Vendor insert failed: %v", err)
		return err
	}

	vendorMap, err := importer.BuildVendorMap(db.DB)
	if err != nil {
		log.Fatalf("❌ Failed to build vendor map: %v", err)
		return err
	}

	for i := range products {
		name := products[i].VendorID
		if realID, ok := vendorMap[name]; ok {
			products[i].VendorID = realID
		} else {
			log.Printf("❌ Vendor not found before inserting product: %s", name)
			return fmt.Errorf("vendor not found: %s", name)
		}
	}

	log.Printf("📦 Inserting %d products...", len(products))
	if err := importer.ImportVendorsAndProductsFromv1(db.DB, nil, products, nil, nil, nil); err != nil {
		log.Fatalf("❌ Product insert failed: %v", err)
		return err
	}

	uids := make([]string, 0, len(cves))
	for _, cve := range cves {
		uids = append(uids, cve.SourceUID)
	}
	cveMap, err := importer.BuildCveMap(db.DB, uids)
	if err != nil {
		log.Fatalf("❌ Failed to build CVE map: %v", err)
		return err
	}

	productMap, err := importer.BuildProductMap(db.DB)
	if err != nil {
		log.Fatalf("❌ Failed to build product map: %v", err)
		return err
	}

	log.Printf("🔗 Linking %d CVEs to products...", len(relations))
	if err := importer.ImportVendorsAndProductsFromv1(db.DB, nil, nil, relations, cveMap, productMap); err != nil {
		log.Fatalf("❌ CVE-product relation insert failed: %v", err)
		return err
	}

	log.Println("✅ Import complete.")
	return nil
}
