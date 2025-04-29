package importer

import (
	"context"
	"log"

	"github.com/uptrace/bun"
	"gitlab.com/jacky850509/secra/internal/model"
)

// ImportWeaknesses 匯入 Weaknesses，需要額外傳入 cveMap 進行 source_uid → UUID 映射
func ImportWeaknesses(db *bun.DB, weaknesses []model.CVEWeakness, cveMap map[string]string) error {
	ctx := context.Background()

	for _, w := range weaknesses {
		cveID, ok := cveMap[w.CVEID]
		if !ok {
			log.Printf("⚠️ CVE not found for weakness: %s (Weakness=%s)", w.CVEID, w.Weakness)
			continue
		}

		w.CVEID = cveID

		_, err := db.NewInsert().
			Model(&w).
			On("CONFLICT (cve_id, weakness) DO NOTHING").
			Exec(ctx)
		if err != nil {
			log.Printf("❌ Failed to insert weakness (CVE=%s Weakness=%s): %v", w.CVEID, w.Weakness, err)
		}
	}

	return nil
}
