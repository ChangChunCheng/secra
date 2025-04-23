package model

import (
	"time"

	"github.com/uptrace/bun"
)

type CVE struct {
	bun.BaseModel `bun:"table:cves"`

	ID          string    `bun:",pk,notnull"`
	SourceID    string    `bun:",notnull"`
	SourceUID   string    `bun:",notnull"`
	Title       string    `bun:",notnull"`
	Description string    `bun:",notnull"`
	Severity    *string   `bun:",nullzero"`
	CVSSScore   *float64  `bun:",nullzero"`
	Status      string    `bun:",default:'active'"`
	PublishedAt time.Time `bun:",nullzero"`
	UpdatedAt   time.Time `bun:",nullzero"`
}
