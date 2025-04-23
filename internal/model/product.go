package model

import "github.com/uptrace/bun"

type Product struct {
	bun.BaseModel `bun:"table:products"`

	ID       string `bun:",pk,notnull"`
	VendorID string `bun:",notnull"`
	Name     string `bun:",notnull"`

	Vendor *Vendor `bun:"rel:has-one,join:vendor_id=id"`
}
