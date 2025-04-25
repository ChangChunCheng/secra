// parser/nvd_v2_transform.go
package v2

import (
	"github.com/google/uuid"
	"gitlab.com/jacky850509/secra/internal/model"
)

// ConvertToCVEsFromV2 轉換 NVD v2 feed 結構為內部 CVE 模型列表
func ConvertToCVEsFromV2(feed *Nvdv2CveFeed) ([]model.CVE, error) {
	var results []model.CVE

	for _, item := range feed.Vulnerabilities {
		c := item.Cve
		desc := extractEnglishDescriptionFromV2(c.Descriptions)

		var (
			score    *float64
			severity *string
		)

		if c.Metrics != nil {
			if len(c.Metrics.CvssMetricV31) > 0 {
				cvss := c.Metrics.CvssMetricV31[0].CvssData
				score = &cvss.BaseScore
				severity = strPtr(cvss.BaseSeverity)
			} else if len(c.Metrics.CvssMetricV2) > 0 {
				cvss := c.Metrics.CvssMetricV2[0].CvssData
				score = &cvss.BaseScore
				severity = strPtr(cvss.Severity)
			}
		}

		results = append(results, model.CVE{
			ID:          uuid.NewString(), // ✅ 修正為自動生成 UUID
			SourceID:    "",               // 由 CLI 注入
			SourceUID:   c.ID,
			Title:       shortTitle(desc),
			Description: desc,
			Severity:    severity,
			CVSSScore:   score,
			Status:      "active",
			PublishedAt: c.Published.Time,
			UpdatedAt:   c.LastModified.Time,
		})
	}

	return results, nil
}

// 提取英文描述
func extractEnglishDescriptionFromV2(descs []LangValue) string {
	for _, d := range descs {
		if d.Lang == "en" {
			return d.Value
		}
	}
	return ""
}

// 建立短標題
func shortTitle(desc string) string {
	if len(desc) > 80 {
		return desc[:80] + "..."
	}
	return desc
}

// 字串指標工具
func strPtr(s string) *string {
	return &s
}
