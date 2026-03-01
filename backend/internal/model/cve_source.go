package model

import (
	"time"

	"github.com/uptrace/bun"
)

// CVESource represents a source of CVE data.
type CVESource struct {
	bun.BaseModel `bun:"table:cve_sources"`

	ID          string    `bun:",pk,notnull,nullzero,default:gen_random_uuid()" json:"id"`
	Name        string    `bun:",notnull,unique" json:"name"`
	Type        string    `bun:",notnull" json:"type"`
	URL         string    `bun:",nullzero" json:"url"`
	Description string    `bun:",nullzero" json:"description"`
	Enabled     bool      `bun:",default:true" json:"enabled"`
	CreatedAt   time.Time `bun:",default:current_timestamp" json:"created_at"`
}
