package model

import (
	"time"

	"github.com/uptrace/bun"
)

type CVE struct {
	bun.BaseModel `bun:"table:cves"`

	ID          string    `bun:",pk,notnull,nullzero,default:gen_random_uuid()" json:"id"`
	SourceID    string    `bun:",notnull" json:"source_id"`
	SourceUID   string    `bun:",notnull" json:"source_uid"`
	Title       string    `bun:",notnull" json:"title"`
	Description string    `bun:",notnull" json:"description"`
	Severity    *string   `bun:",nullzero" json:"severity"`
	CVSSScore   *float64  `bun:",nullzero" json:"cvss_score"`
	Status      string    `bun:",default:'active'" json:"status"`
	PublishedAt time.Time `bun:",nullzero" json:"published_at"`
	UpdatedAt   time.Time `bun:",nullzero" json:"updated_at"`
}
