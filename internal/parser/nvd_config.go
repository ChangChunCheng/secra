package parser

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

// 解析 vendor/product 關聯
func ExtractVendorsAndProducts(feed *NvdCveFeed) ([]model.Vendor, []model.Product, []CVEProductRelation) {
	vendorSet := map[string]string{}  // name → uuid
	productSet := map[string]string{} // vendor:product → uuid
	var relations []CVEProductRelation

	for _, item := range feed.Items {
		sourceUID := item.CVE.DataMeta.ID
		nodes := flattenNodes(item.Configurations.Nodes)

		for _, node := range nodes {
			for _, cpe := range node.CpeMatch {
				if !cpe.Vulnerable {
					continue
				}
				vendor, product := parseCpe23Uri(cpe.Cpe23Uri)
				if vendor == "" || product == "" {
					continue
				}

				// 記錄 vendor
				if _, exists := vendorSet[vendor]; !exists {
					vendorSet[vendor] = uuid.NewString()
				}

				// 記錄 product
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

	// 建立 vendor struct
	var vendors []model.Vendor
	for name, id := range vendorSet {
		vendors = append(vendors, model.Vendor{
			ID:   id,
			Name: name,
		})
	}

	// 建立 product struct（VendorID 暫用 vendor name）
	var products []model.Product
	for key, id := range productSet {
		parts := strings.SplitN(key, ":", 2)
		vendorName := parts[0]
		productName := parts[1]

		products = append(products, model.Product{
			ID:       id,
			VendorID: vendorName, // ⬅ 暫時使用 vendor name，後續 CLI 會補成真 UUID
			Name:     productName,
		})
	}

	return vendors, products, relations
}

func flattenNodes(nodes []ConfigNode) []ConfigNode {
	var all []ConfigNode
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
