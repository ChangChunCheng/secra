package importer

import (
	"context"

	"github.com/uptrace/bun"
	"gitlab.com/jacky850509/secra/internal/model"
)

func ImportWeaknesses(db *bun.DB, weaknesses []model.CVEWeakness, _ map[string]string) error {
	ctx := context.Background()
	for _, w := range weaknesses {
		// Use natural key (cve_id, weakness) for Upsert to support --force
		_, err := db.NewInsert().Model(&w).
			On("CONFLICT (cve_id, weakness) DO UPDATE SET created_at = EXCLUDED.created_at").
			Exec(ctx)
		if err != nil { return err }
	}
	return nil
}
