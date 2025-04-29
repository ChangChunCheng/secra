package v2

import (
	"strings"

	"github.com/google/uuid"
	"gitlab.com/jacky850509/secra/internal/model"
)

// ConvertToCVEsFromV2 把 NVD v2 Feed 轉成內部 model.CVE
func ConvertToCVEsFromV2(feed *Nvdv2CveFeed) ([]model.CVE, error) {
	var results []model.CVE

	for _, item := range feed.Vulnerabilities {
		c := item.Cve
		desc := extractEnglishDescription(c.Descriptions)

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
			ID:          uuid.NewString(),
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

// 抽取全部資料 (Vendor, Product, Relation, Reference, Weakness)
func ExtractAllFromV2(feed *Nvdv2CveFeed) (
	[]model.Vendor, []model.Product, []CVEProductRelation,
	[]model.CVEReference, []model.CVEWeakness,
) {
	vendorSet := map[string]string{}
	productSet := map[string]string{}

	var (
		vendors    []model.Vendor
		products   []model.Product
		relations  []CVEProductRelation
		references []model.CVEReference
		weaknesses []model.CVEWeakness
	)

	for _, item := range feed.Vulnerabilities {
		sourceUID := item.Cve.ID

		// Configurations → Vendor/Product
		for _, cfg := range item.Cve.Configurations {
			nodes := flattenNodes(cfg.Nodes)
			for _, node := range nodes {
				for _, cpe := range node.CpeMatch {
					if !cpe.Vulnerable {
						continue
					}
					vendor, product := parseCpe23Uri(cpe.Criteria)
					if vendor == "" || product == "" {
						continue
					}
					if _, exists := vendorSet[vendor]; !exists {
						vendorSet[vendor] = uuid.NewString()
					}
					pkey := vendor + ":" + product
					if _, exists := productSet[pkey]; !exists {
						productSet[pkey] = uuid.NewString()
					}
					relations = append(relations, CVEProductRelation{
						CveSourceUID: sourceUID,
						VendorName:   vendor,
						ProductName:  product,
					})
				}
			}
		}

		// References
		for _, ref := range item.Cve.References {
			references = append(references, model.CVEReference{
				ID:     uuid.NewString(),
				CVEID:  sourceUID,
				URL:    ref.URL,
				Source: ref.Source,
				Tags:   ref.Tags,
			})
		}

		// Weaknesses
		for _, w := range item.Cve.Weaknesses {
			for _, desc := range w.Description {
				if desc.Lang == "en" {
					weaknesses = append(weaknesses, model.CVEWeakness{
						ID:       uuid.NewString(),
						CVEID:    sourceUID,
						Weakness: desc.Value,
					})
				}
			}
		}
	}

	// Build vendors
	for name, id := range vendorSet {
		vendors = append(vendors, model.Vendor{ID: id, Name: name})
	}

	// Build products
	for key, id := range productSet {
		parts := strings.SplitN(key, ":", 2)
		vendorName := parts[0]
		productName := parts[1]
		products = append(products, model.Product{
			ID:       id,
			VendorID: vendorName, // 後面會用 vendorMap 轉正確 uuid
			Name:     productName,
		})
	}

	return vendors, products, relations, references, weaknesses
}

func flattenNodes(nodes []Nvdv2ConfigNode) []Nvdv2ConfigNode {
	var all []Nvdv2ConfigNode
	for _, node := range nodes {
		all = append(all, node)
		if len(node.Children) > 0 {
			all = append(all, flattenNodes(node.Children)...)
		}
	}
	return all
}

func parseCpe23Uri(uri string) (vendor string, product string) {
	parts := strings.Split(uri, ":")
	if len(parts) >= 5 {
		return parts[3], parts[4]
	}
	return "", ""
}

func extractEnglishDescription(descs []LangValue) string {
	for _, d := range descs {
		if d.Lang == "en" {
			return d.Value
		}
	}
	return ""
}

func strPtr(s string) *string {
	return &s
}

func shortTitle(desc string) string {
	if len(desc) > 80 {
		return desc[:80] + "..."
	}
	return desc
}
