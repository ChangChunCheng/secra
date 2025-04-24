package nvd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/uptrace/bun"

	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/fetcher"
	"gitlab.com/jacky850509/secra/internal/importer"
	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/parser"
	"gitlab.com/jacky850509/secra/internal/storage"
)

var recent bool
var modified bool
var year uint16

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
	processDatas := make(map[string][]byte)

	if recent {
		log.Printf("📥 Downloading NVD, url=%s, recent feed...", cfg.NvdURLv1)
		data, err = fetcher.DownloadNvdv1FeedRecent(cfg.NvdURLv1)
		if err != nil {
			log.Fatalf("❌ Failed to fetch feed: %v", err)
		} else {
			processDatas["recent"] = data
		}
	}
	if modified {
		log.Printf("📥 Downloading NVD, url=%s, modified feed...", cfg.NvdURLv1)
		data, err = fetcher.DownloadNvdv1FeedModified(cfg.NvdURLv1)
		if err != nil {
			log.Fatalf("❌ Failed to fetch feed: %v", err)
		} else {
			processDatas["modified"] = data
		}
	}
	if year >= 2002 {
		log.Printf("📥 Downloading NVD, url=%s, feed for year %d...", cfg.NvdURLv1, year)
		data, err = fetcher.DownloadNvdv1FeedYear(year, cfg.NvdURLv1)
		if err != nil {
			log.Fatalf("❌ Failed to fetch feed: %v", err)
		} else {
			processDatas[fmt.Sprintf("year-%d", year)] = data
		}
	}
	if len(processDatas) == 0 {
		log.Printf("❌ No feed type specified. Use --recent, --modified, or --year.")
		return
	}
	log.Printf("📥 Downloaded %d feeds.", len(processDatas))

	log.Printf("📦 Processing %d feeds...", len(processDatas))
	// Step 1: 轉換 CVEs
	for feedName, data := range processDatas {
		log.Printf("📦 Processing %s feed...", feedName)
		if err = ProcessImportNvdv1(db, data, feedName); err != nil {
			log.Fatalf("❌ Failed to process feed %s: %v", feedName, err)
		}
		log.Printf("✅ %s feed processed successfully.", feedName)
	}
	log.Println("✅ All feeds processed.")
}

func ProcessImportNvdv1(db *storage.DBWrapper, data []byte, sourceName string) error {

	var feed parser.Nvdv1CveFeed
	if err := json.Unmarshal(data, &feed); err != nil {
		log.Fatalf("❌ Failed to parse feed JSON: %v", err)
		return err
	}
	log.Printf("✅ Feed parsed with %d items.", len(feed.Items))

	// Step 1: 轉換 CVEs
	cves, err := parser.ConvertToCVEs(&feed)
	if err != nil {
		log.Fatalf("❌ Failed to convert CVEs: %v", err)
		return err
	}

	// Step 2: 確保來源
	source, err := ensureCveSource(db.DB, sourceName, cfg.NvdURLv1, "")
	if err != nil {
		log.Fatalf("❌ Failed to ensure source: %v", err)
		return err
	}

	// Step 3: 匯入 CVEs
	log.Printf("📦 Importing %d CVEs...", len(cves))
	if err := importer.ImportCVEs(db.DB, source.ID, cves); err != nil {
		log.Fatalf("❌ CVE import failed: %v", err)
		return err
	}

	// Step 4: 萃取 vendor/product 關聯
	log.Println("🔍 Extracting vendors/products from configurations...")
	vendors, products, relations := parser.ExtractVendorsAndProducts(&feed)

	// Step 5: 寫入 vendors
	log.Printf("📦 Inserting %d vendors...", len(vendors))
	if err := importer.ImportVendorsAndProducts(db.DB, vendors, nil, nil, nil, nil); err != nil {
		log.Fatalf("❌ Vendor insert failed: %v", err)
		return err
	}

	// Step 6: 查出 vendorMap 以補 products 的 VendorID
	vendorMap, err := importer.BuildVendorMap(db.DB)
	if err != nil {
		log.Fatalf("❌ Failed to build vendor map: %v", err)
		return err
	}

	for i := range products {
		name := products[i].VendorID // 此時仍為 vendor name
		if realID, ok := vendorMap[name]; ok {
			products[i].VendorID = realID
		} else {
			log.Printf("❌ Vendor not found before inserting product: %s", name)
			return err
		}
	}

	// Step 7: 寫入 products
	log.Printf("📦 Inserting %d products...", len(products))
	if err := importer.ImportVendorsAndProducts(db.DB, nil, products, nil, nil, nil); err != nil {
		log.Fatalf("❌ Product insert failed: %v", err)
		return err
	}

	// Step 8: 準備對照表並關聯 CVE ↔ Product
	// 建立所有 source_uid 清單
	uids := make([]string, 0, len(cves))
	for _, cve := range cves {
		uids = append(uids, cve.SourceUID)
	}
	// 無論是新寫入或已有的，統一查詢 UUID
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
	if err := importer.ImportVendorsAndProducts(db.DB, nil, nil, relations, cveMap, productMap); err != nil {
		log.Fatalf("❌ CVE-product relation insert failed: %v", err)
		return err
	}

	log.Println("✅ Import complete.")

	return nil
}

func ensureCveSource(db *bun.DB, name, description, urlStr string) (*model.CVESource, error) {
	ctx := context.Background()

	var src model.CVESource
	err := db.NewSelect().Model(&src).Where("name = ?", name).Scan(ctx)
	if err == nil {
		return &src, nil // 已存在，直接回傳
	}

	// 只在來源尚未存在時，才使用提供的 description 和 url
	var urlPtr string
	if urlStr != "" {
		urlPtr = urlStr
	}

	src = model.CVESource{
		ID:          uuid.NewString(),
		Name:        name,
		Type:        "nvd",
		URL:         urlPtr,
		Description: description,
		Enabled:     true,
		CreatedAt:   time.Now(),
	}

	_, err = db.NewInsert().Model(&src).Exec(ctx)
	if err != nil {
		return nil, err
	}

	return &src, nil
}
