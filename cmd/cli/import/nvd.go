package importcmd

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
	"gitlab.com/jacky850509/secra/internal/parser"
	"gitlab.com/jacky850509/secra/internal/storage"
)

var nvdCmd = &cobra.Command{
	Use:   "nvd",
	Short: "Import CVEs from NVD feed",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		db := storage.NewDB(cfg.PostgresDSN, false)
		defer db.Close()

		log.Printf("📥 Downloading NVD, url=%s,  feed for year %d...", cfg.NvdURLv1, year)
		data, err := fetcher.DownloadNvdv1Feed(year, cfg.NvdURLv1)
		if err != nil {
			log.Fatalf("❌ Failed to fetch feed: %v", err)
		}

		var feed parser.Nvdv1CveFeed
		if err := json.Unmarshal(data, &feed); err != nil {
			log.Fatalf("❌ Failed to parse feed JSON: %v", err)
		}
		log.Printf("✅ Feed parsed with %d items.", len(feed.Items))

		// Step 1: 轉換 CVEs
		cves, err := parser.ConvertToCVEs(&feed)
		if err != nil {
			log.Fatalf("❌ Failed to convert CVEs: %v", err)
		}

		// Step 2: 確保來源
		source, err := ensureCveSource(db.DB, sourceName, cfg.NvdURLv1, "")
		if err != nil {
			log.Fatalf("❌ Failed to ensure source: %v", err)
		}

		// Step 3: 匯入 CVEs
		log.Printf("📦 Importing %d CVEs...", len(cves))
		if err := importer.ImportCVEs(db.DB, source.ID, cves); err != nil {
			log.Fatalf("❌ CVE import failed: %v", err)
		}

		// Step 4: 萃取 vendor/product 關聯
		log.Println("🔍 Extracting vendors/products from configurations...")
		vendors, products, relations := parser.ExtractVendorsAndProducts(&feed)

		// Step 5: 寫入 vendors
		log.Printf("📦 Inserting %d vendors...", len(vendors))
		if err := importer.ImportVendorsAndProducts(db.DB, vendors, nil, nil, nil, nil); err != nil {
			log.Fatalf("❌ Vendor insert failed: %v", err)
		}

		// Step 6: 查出 vendorMap 以補 products 的 VendorID
		vendorMap, err := importer.BuildVendorMap(db.DB)
		if err != nil {
			log.Fatalf("❌ Failed to build vendor map: %v", err)
		}

		for i := range products {
			name := products[i].VendorID // 此時仍為 vendor name
			if realID, ok := vendorMap[name]; ok {
				products[i].VendorID = realID
			} else {
				log.Printf("❌ Vendor not found before inserting product: %s", name)
			}
		}

		// Step 7: 寫入 products
		log.Printf("📦 Inserting %d products...", len(products))
		if err := importer.ImportVendorsAndProducts(db.DB, nil, products, nil, nil, nil); err != nil {
			log.Fatalf("❌ Product insert failed: %v", err)
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
		}

		productMap, err := importer.BuildProductMap(db.DB)
		if err != nil {
			log.Fatalf("❌ Failed to build product map: %v", err)
		}

		log.Printf("🔗 Linking %d CVEs to products...", len(relations))
		if err := importer.ImportVendorsAndProducts(db.DB, nil, nil, relations, cveMap, productMap); err != nil {
			log.Fatalf("❌ CVE-product relation insert failed: %v", err)
		}

		log.Println("✅ Import complete.")
	},
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
