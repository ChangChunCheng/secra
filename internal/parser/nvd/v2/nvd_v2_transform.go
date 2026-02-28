package v2

import (
	"strings"

	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/util"
)

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
			} else if len(c.Metrics.CvssMetricV30) > 0 {
				cvss := c.Metrics.CvssMetricV30[0].CvssData
				score = &cvss.BaseScore
				severity = strPtr(cvss.BaseSeverity)
			} else if len(c.Metrics.CvssMetricV2) > 0 {
				cvss := c.Metrics.CvssMetricV2[0].CvssData
				score = &cvss.BaseScore
				severity = strPtr(cvss.Severity)
			}
		}
		results = append(results, model.CVE{
			ID:          util.CVEID(c.ID),
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

func ExtractAllFromV2(feed *Nvdv2CveFeed) ([]model.Vendor, []model.Product, []CVEProductRelation, []model.CVEReference, []model.CVEWeakness) {
	vendorMap := map[string]string{}
	productMap := map[string]string{}
	var vendors []model.Vendor
	var products []model.Product
	var relations []CVEProductRelation
	var references []model.CVEReference
	var weaknesses []model.CVEWeakness

	for _, item := range feed.Vulnerabilities {
		sourceUID := item.Cve.ID
		for _, cfg := range item.Cve.Configurations {
			nodes := flattenNodes(cfg.Nodes)
			for _, node := range nodes {
				for _, cpe := range node.CpeMatch {
					vendor, product := parseCpe23Uri(cpe.Criteria)
					if vendor == "" || product == "" { continue }
					
					vendor = strings.ToLower(vendor)
					product = strings.ToLower(product)

					vID := util.VendorID(vendor)
					pID := util.ProductID(vendor, product)

					if _, exists := vendorMap[vendor]; !exists {
						vendorMap[vendor] = vID
						vendors = append(vendors, model.Vendor{ID: vID, Name: vendor})
					}
					
					if _, exists := productMap[pID]; !exists {
						productMap[pID] = vID // Temp storage
						products = append(products, model.Product{ID: pID, VendorID: vID, Name: product})
					}
					
					if cpe.Vulnerable {
						relations = append(relations, CVEProductRelation{
							CveSourceUID: sourceUID, VendorName: vendor, ProductName: product,
						})
					}
				}
			}
		}
		// References
		for _, ref := range item.Cve.References {
			references = append(references, model.CVEReference{
				ID:     util.NewUUIDv5("ref:" + sourceUID + ":" + ref.URL),
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
						ID:       util.NewUUIDv5("weakness:" + sourceUID + ":" + desc.Value),
						CVEID:    sourceUID,
						Weakness: desc.Value,
					})
				}
			}
		}
	}

	return vendors, products, relations, references, weaknesses
}

func flattenNodes(nodes []Nvdv2ConfigNode) []Nvdv2ConfigNode {
	var all []Nvdv2ConfigNode
	for _, node := range nodes {
		all = append(all, node)
		if len(node.Children) > 0 { all = append(all, flattenNodes(node.Children)...) }
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
		if d.Lang == "en" { return d.Value }
	}
	return ""
}

func strPtr(s string) *string { return &s }

func shortTitle(desc string) string {
	if len(desc) > 100 { return desc[:100] + "..." }
	return desc
}
