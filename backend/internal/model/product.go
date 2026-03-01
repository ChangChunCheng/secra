package model

import "github.com/uptrace/bun"

type Product struct {
	bun.BaseModel `bun:"table:products"`

	ID       string `bun:",pk,notnull" json:"id"`
	VendorID string `bun:",notnull" json:"vendor_id"`
	Name     string `bun:",notnull" json:"name"`

	Vendor *Vendor `bun:"rel:has-one,join:vendor_id=id" json:"vendor,omitempty"`
}
