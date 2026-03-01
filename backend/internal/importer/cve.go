package importer

import (
	"context"
	"log"

	"github.com/uptrace/bun"
	"gitlab.com/jacky850509/secra/internal/model"
)

func ImportCVEs(db *bun.DB, sourceID string, cves []model.CVE) error {
	ctx := context.Background()

	for i := range cves {
		cves[i].SourceID = sourceID
		// CRITICAL: Change conflict target to source_uid to enable true Upsert by CVE-ID
		_, err := db.NewInsert().Model(&cves[i]).
			On("CONFLICT (source_uid) DO UPDATE SET title = EXCLUDED.title, description = EXCLUDED.description, severity = EXCLUDED.severity, cvss_score = EXCLUDED.cvss_score, status = EXCLUDED.status, updated_at = EXCLUDED.updated_at, source_id = EXCLUDED.source_id").
			Exec(ctx)

		if err != nil {
			log.Printf("❌ Failed to import CVE %s: %v", cves[i].SourceUID, err)
			continue
		}
	}

	// Update daily stats after batch
	_, _ = db.NewRaw(`INSERT INTO daily_cve_counts (day, count)
		SELECT published_at::date as day, count(*) FROM cves GROUP BY day
		ON CONFLICT (day) DO UPDATE SET count = EXCLUDED.count`).Exec(ctx)

	return nil
}
