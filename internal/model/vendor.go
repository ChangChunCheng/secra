package model

import "github.com/uptrace/bun"

type Vendor struct {
	bun.BaseModel `bun:"table:vendors"`

	ID   string `bun:",pk,notnull"`
	Name string `bun:",notnull,unique"`
}
