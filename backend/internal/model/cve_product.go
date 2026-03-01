package model

type CVEProduct struct {
	CVEID     string `bun:"cve_id,pk"`
	ProductID string `bun:"product_id,pk"`
}
