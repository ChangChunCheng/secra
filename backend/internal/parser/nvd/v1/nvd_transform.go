package v1

import (
	"strings"

	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/util"
)

// ConvertToCVEs 轉換解析後的 feed → DB-ready CVE 結構
func ConvertToCVEs(feed *Nvdv1CveFeed) ([]model.CVE, error) {
	var results []model.CVE

	for _, item := range feed.Items {
		meta := item.CVE.DataMeta
		desc := extractEnglishDescription(item.CVE.Description)

		var (
			score    *float64
			severity *string
		)

		if item.Impact != nil {
			if item.Impact.CVSSv4 != nil {
				score = &item.Impact.CVSSv4.CvssData.BaseScore
				severity = strPtr(item.Impact.CVSSv4.CvssData.BaseSeverity)
			} else if item.Impact.CVSSv3 != nil {
				score = &item.Impact.CVSSv3.CvssData.BaseScore
				severity = strPtr(item.Impact.CVSSv3.CvssData.Severity)
			} else if item.Impact.CVSSv2 != nil {
				score = &item.Impact.CVSSv2.CvssData.BaseScore
				severity = strPtr(item.Impact.CVSSv2.CvssData.Severity)
			}
		}

		cve := model.CVE{
			ID:          util.CVEID(meta.ID), // Deterministic UUID v5
			SourceID:    "",
			SourceUID:   meta.ID,
			Title:       shortTitle(desc),
			Description: desc,
			Severity:    severity,
			CVSSScore:   score,
			Status:      "active",
			PublishedAt: item.PublishedDate.Time,
			UpdatedAt:   item.LastModifiedDate.Time,
		}

		results = append(results, cve)
	}

	return results, nil
}

func shortTitle(full string) string {
	if len(full) > 180 { full = full[:180] }
	if idx := strings.Index(full, "."); idx != -1 { return full[:idx+1] }
	return full
}

func extractEnglishDescription(desc Nvdv1CveDescription) string {
	for _, d := range desc.DescriptionData {
		if d.Lang == "en" { return d.Value }
	}
	return ""
}

func strPtr(s string) *string { return &s }
