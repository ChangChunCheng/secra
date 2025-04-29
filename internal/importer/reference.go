package importer

import (
	"context"
	"log"

	"github.com/uptrace/bun"
	"gitlab.com/jacky850509/secra/internal/model"
)

// ImportReferences 匯入 References，需要額外傳入 cveMap 進行 source_uid → UUID 映射
func ImportReferences(db *bun.DB, refs []model.CVEReference, cveMap map[string]string) error {
	ctx := context.Background()

	for _, ref := range refs {
		cveID, ok := cveMap[ref.CVEID]
		if !ok {
			log.Printf("⚠️ CVE not found for reference: %s (URL=%s)", ref.CVEID, ref.URL)
			continue
		}

		ref.CVEID = cveID

		_, err := db.NewInsert().
			Model(&ref).
			On("CONFLICT (cve_id, url) DO NOTHING").
			Exec(ctx)
		if err != nil {
			log.Printf("❌ Failed to insert reference (CVE=%s URL=%s): %v", ref.CVEID, ref.URL, err)
		}
	}

	return nil
}
