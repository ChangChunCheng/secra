package v2

import (
	"strings"

	"github.com/google/uuid"
	"gitlab.com/jacky850509/secra/internal/model"
)

type CVEProductRelation struct {
	CveSourceUID string
	VendorName   string
	ProductName  string
}

// ExtractVendorsAndProductsFromv2 擷取 vendor/product 關聯 from NVD v2 feed
func ExtractVendorsAndProductsFromv2(feed *Nvdv2CveFeed) ([]model.Vendor, []model.Product, []CVEProductRelation) {
	vendorSet := map[string]string{}  // name → uuid
	productSet := map[string]string{} // vendor:product → uuid
	var relations []CVEProductRelation

	for _, item := range feed.Vulnerabilities {
		sourceUID := item.Cve.ID
		for _, cfg := range item.Cve.Configurations { // ⚠️ configurations 是 slice
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
	}

	var vendors []model.Vendor
	for name, id := range vendorSet {
		vendors = append(vendors, model.Vendor{
			ID:   id,
			Name: name,
		})
	}

	var products []model.Product
	for key, id := range productSet {
		parts := strings.SplitN(key, ":", 2)
		vendorName := parts[0]
		productName := parts[1]

		products = append(products, model.Product{
			ID:       id,
			VendorID: vendorName,
			Name:     productName,
		})
	}

	return vendors, products, relations
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
