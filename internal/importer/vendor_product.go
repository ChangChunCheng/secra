package importer

import (
	"context"
	"fmt"
	"log"

	"github.com/uptrace/bun"
	"gitlab.com/jacky850509/secra/internal/model"
	nvd_v1_parser "gitlab.com/jacky850509/secra/internal/parser/nvd/v1"
	nvd_v2_parser "gitlab.com/jacky850509/secra/internal/parser/nvd/v2"
)

// ImportVendorsAndProducts 實作 Vendor/Product/CVE 關聯匯入
func ImportVendorsAndProductsFromv1(
	db *bun.DB,
	vendors []model.Vendor,
	products []model.Product,
	relations []nvd_v1_parser.CVEProductRelation,
	cveMap map[string]string,
	productMap map[string]string,
) error {
	ctx := context.Background()

	// 插入 vendors
	for _, v := range vendors {
		_, err := db.NewInsert().
			Model(&v).
			On("CONFLICT (name) DO NOTHING").
			Exec(ctx)
		if err != nil {
			log.Printf("❌ Failed to insert vendor %s: %v", v.Name, err)
		}
	}

	// 插入 products
	for _, p := range products {
		_, err := db.NewInsert().
			Model(&p).
			On("CONFLICT (vendor_id, name) DO NOTHING").
			Exec(ctx)
		if err != nil {
			log.Printf("❌ Failed to insert product %s (vendor_id=%s): %v", p.Name, p.VendorID, err)
		}
	}

	// 建立 CVE ↔ Product 關聯
	for _, r := range relations {
		pkey := r.VendorName + ":" + r.ProductName

		pid, ok := productMap[pkey]
		if !ok {
			log.Printf("❌ Product not mapped: %s", pkey)
			continue
		}

		cid, ok := cveMap[r.CveSourceUID]
		if !ok {
			log.Printf("❌ CVE not mapped: %s", r.CveSourceUID)
			continue
		}

		_, err := db.NewInsert().
			Model(&model.CVEProduct{
				CVEID:     cid,
				ProductID: pid,
			}).
			On("CONFLICT DO NOTHING").
			Exec(ctx)
		if err != nil {
			log.Printf("❌ Failed to link CVE %s and Product %s: %v", r.CveSourceUID, pkey, err)
		}
	}

	return nil
}

func ImportVendorsAndProductsFromv2(
	db *bun.DB,
	vendors []model.Vendor,
	products []model.Product,
	relations []nvd_v2_parser.CVEProductRelation,
	cveMap map[string]string,
	productMap map[string]string,
) error {
	ctx := context.Background()

	// 插入 vendors
	for _, v := range vendors {
		_, err := db.NewInsert().
			Model(&v).
			On("CONFLICT (name) DO NOTHING").
			Exec(ctx)
		if err != nil {
			log.Printf("❌ Failed to insert vendor %s: %v", v.Name, err)
		}
	}

	// 插入 products
	for _, p := range products {
		_, err := db.NewInsert().
			Model(&p).
			On("CONFLICT (vendor_id, name) DO NOTHING").
			Exec(ctx)
		if err != nil {
			log.Printf("❌ Failed to insert product %s (vendor_id=%s): %v", p.Name, p.VendorID, err)
		}
	}

	// 建立 CVE ↔ Product 關聯
	for _, r := range relations {
		pkey := r.VendorName + ":" + r.ProductName

		pid, ok := productMap[pkey]
		if !ok {
			log.Printf("❌ Product not mapped: %s", pkey)
			continue
		}

		cid, ok := cveMap[r.CveSourceUID]
		if !ok {
			// log.Printf("❌ CVE not mapped: %s", r.CveSourceUID)
			continue
		}

		_, err := db.NewInsert().
			Model(&model.CVEProduct{
				CVEID:     cid,
				ProductID: pid,
			}).
			On("CONFLICT DO NOTHING").
			Exec(ctx)
		if err != nil {
			log.Printf("❌ Failed to link CVE %s and Product %s: %v", r.CveSourceUID, pkey, err)
		}
	}

	return nil
}

// 建立 vendor id:product name → product id 的對照表
func BuildProductMap(db *bun.DB) (map[string]string, error) {
	ctx := context.Background()

	type ProductRow struct {
		ProductID   string `bun:"id"`
		ProductName string `bun:"name"`
		VendorName  string `bun:"vendor_name"`
	}

	var results []ProductRow

	err := db.NewSelect().
		Table("products").
		Column("products.id", "products.name").
		ColumnExpr("vendors.name AS vendor_name").
		Join("JOIN vendors ON vendors.id = products.vendor_id").
		Scan(ctx, &results)
	if err != nil {
		return nil, err
	}

	m := make(map[string]string)
	for _, row := range results {
		key := fmt.Sprintf("%s:%s", row.VendorName, row.ProductName)
		m[key] = row.ProductID
	}
	return m, nil
}

func BuildVendorMap(db *bun.DB) (map[string]string, error) {
	ctx := context.Background()

	type VendorRow struct {
		ID   string `bun:"id"`
		Name string `bun:"name"`
	}

	var results []VendorRow

	err := db.NewSelect().
		Table("vendors").
		Column("id", "name").
		Scan(ctx, &results)
	if err != nil {
		return nil, fmt.Errorf("BuildVendorMap failed: %w", err)
	}

	m := make(map[string]string)
	for _, row := range results {
		m[row.Name] = row.ID
	}
	return m, nil
}
