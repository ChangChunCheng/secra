package importer

import (
	"context"

	"github.com/uptrace/bun"
	"gitlab.com/jacky850509/secra/internal/model"
)

func ImportReferences(db *bun.DB, refs []model.CVEReference, _ map[string]string) error {
	ctx := context.Background()
	for _, r := range refs {
		// Use natural key (cve_id, url) for Upsert to support --force
		_, err := db.NewInsert().Model(&r).
			On("CONFLICT (cve_id, url) DO UPDATE SET source = EXCLUDED.source, tags = EXCLUDED.tags").
			Exec(ctx)
		if err != nil { return err }
	}
	return nil
}
