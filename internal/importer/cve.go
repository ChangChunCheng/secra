package importer

import (
	"context"
	"log"

	"github.com/uptrace/bun"
	"gitlab.com/jacky850509/secra/internal/model"
)

func ImportCVEs(db *bun.DB, sourceID string, cves []model.CVE) error {
	ctx := context.Background()

	for _, cve := range cves {
		cve.SourceID = sourceID

		_, err := db.NewInsert().
			Model(&cve).
			On("CONFLICT (source_uid) DO UPDATE").
			Set("title = EXCLUDED.title").
			Set("description = EXCLUDED.description").
			Set("cvss_score = EXCLUDED.cvss_score").
			Set("severity = EXCLUDED.severity").
			Set("updated_at = EXCLUDED.updated_at").
			Exec(ctx)

		if err != nil {
			log.Printf("❌ Failed to upsert CVE %s: %v", cve.SourceUID, err)
		}
	}

	return nil
}

// 查詢所有指定 source_uid 對應的 UUID → 建立 cveMap
func BuildCveMap(db *bun.DB, sourceUIDs []string) (map[string]string, error) {
	ctx := context.Background()

	type row struct {
		ID        string `bun:"id"`
		SourceUID string `bun:"source_uid"`
	}

	var results []row

	err := db.NewSelect().
		Table("cves").
		Column("id", "source_uid").
		Where("source_uid IN (?)", bun.In(sourceUIDs)).
		Scan(ctx, &results)
	if err != nil {
		return nil, err
	}

	m := make(map[string]string)
	for _, r := range results {
		m[r.SourceUID] = r.ID
	}
	return m, nil
}
