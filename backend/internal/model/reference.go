package model

import (
	"time"

	"github.com/uptrace/bun"
)

type CVEReference struct {
	bun.BaseModel `bun:"table:cve_references"`

	ID        string    `bun:",pk,notnull"`
	CVEID     string    `bun:"cve_id,pk"`           // UUID
	URL       string    `bun:"url,pk"`              // URL 也是 PK
	Source    string    `bun:"source,nullzero"`     // 可以為空
	Tags      []string  `bun:"tags,array,nullzero"` // 允許空陣列
	CreatedAt time.Time `bun:"created_at,default:current_timestamp"`
}
