package model

import (
	"github.com/uptrace/bun"
)

// TargetType represents a row in the target_types table.
type TargetType struct {
	bun.BaseModel `bun:"table:target_types,alias:target_type"`

	ID   int    `bun:"id,pk,autoincrement"`
	Name string `bun:"name,notnull,unique"`
}
