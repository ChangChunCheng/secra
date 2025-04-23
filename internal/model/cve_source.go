package model

import (
	"time"

	"github.com/uptrace/bun"
)

type CVESource struct {
	bun.BaseModel `bun:"table:cve_sources"`

	ID          string    `bun:",pk,notnull"`
	Name        string    `bun:",notnull,unique"`
	Type        string    `bun:",notnull"`
	URL         string    `bun:",nullzero"`
	Description string    `bun:",nullzero"`
	Enabled     bool      `bun:",default:true"`
	CreatedAt   time.Time `bun:",default:current_timestamp"`
}
