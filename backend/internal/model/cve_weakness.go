package model

import (
	"time"

	"github.com/uptrace/bun"
)

type CVEWeakness struct {
	bun.BaseModel `bun:"table:cve_weaknesses"`

	ID        string    `bun:",pk,notnull"`
	CVEID     string    `bun:"cve_id,pk"`   // UUID
	Weakness  string    `bun:"weakness,pk"` // Weakness 文字 (例如 CWE-79)
	CreatedAt time.Time `bun:"created_at,default:current_timestamp"`
}
