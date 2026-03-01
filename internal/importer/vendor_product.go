package importer

import (
	"context"

	"github.com/uptrace/bun"
	"gitlab.com/jacky850509/secra/internal/model"
	nvd_v1_parser "gitlab.com/jacky850509/secra/internal/parser/nvd/v1"
	nvd_v2_parser "gitlab.com/jacky850509/secra/internal/parser/nvd/v2"
	"gitlab.com/jacky850509/secra/internal/util"
)

// ImportVendorsAndProductsFromv1 simplified for v1
func ImportVendorsAndProductsFromv1(db *bun.DB, vendors []model.Vendor, products []model.Product, relations []nvd_v1_parser.CVEProductRelation, _, _ map[string]string) error {
	return importVendorsProductsRelations(db, vendors, products, relations)
}

// ImportVendorsAndProductsFromv2 simplified for v2
func ImportVendorsAndProductsFromv2(db *bun.DB, vendors []model.Vendor, products []model.Product, relations []nvd_v2_parser.CVEProductRelation, _, _ map[string]string) error {
	return importVendorsProductsRelations(db, vendors, products, relations)
}

func importVendorsProductsRelations[T any](db *bun.DB, vendors []model.Vendor, products []model.Product, relations []T) error {
	ctx := context.Background()

	for _, v := range vendors {
		// Fix: Use name as conflict target for vendors
		db.NewInsert().Model(&v).On("CONFLICT (name) DO UPDATE SET name = EXCLUDED.name").Exec(ctx)
	}

	for _, p := range products {
		// Fix: Use vendor_id and name as conflict target for products
		db.NewInsert().Model(&p).On("CONFLICT (vendor_id, name) DO UPDATE SET name = EXCLUDED.name").Exec(ctx)
	}

	for _, r := range relations {
		var cveUID, vName, pName string
		switch v := any(r).(type) {
		case nvd_v1_parser.CVEProductRelation:
			cveUID, vName, pName = v.CveSourceUID, v.VendorName, v.ProductName
		case nvd_v2_parser.CVEProductRelation:
			cveUID, vName, pName = v.CveSourceUID, v.VendorName, v.ProductName
		default:
			continue
		}

		cid := util.CVEID(cveUID)
		pid := util.ProductID(vName, pName)

		link := &model.CVEProduct{CVEID: cid, ProductID: pid}
		db.NewInsert().Model(link).On("CONFLICT DO NOTHING").Exec(ctx)
	}
	return nil
}

func BuildVendorMap(db *bun.DB) (map[string]string, error) {
	var results []model.Vendor
	err := db.NewSelect().Model(&results).Scan(context.Background())
	m := make(map[string]string)
	for _, r := range results { m[r.Name] = r.ID }
	return m, err
}

func BuildProductMap(db *bun.DB) (map[string]string, error) {
	type row struct { ID, Name, VendorName string `bun:"vendor_name"` }
	var results []row
	err := db.NewSelect().Table("products").Column("products.id", "products.name").
		ColumnExpr("v.name AS vendor_name").Join("JOIN vendors v ON v.id = products.vendor_id").Scan(context.Background(), &results)
	m := make(map[string]string)
	for _, r := range results { m[r.VendorName+":"+r.Name] = r.ID }
	return m, err
}
