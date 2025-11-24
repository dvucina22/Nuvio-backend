package models

type ProductVariant struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	SKU      string  `json:"sku"`
	Price    float64 `json:"price"`
	IsActive bool    `json:"isActive"`

	Images     []ProductImage     `json:"images"`
	Attributes []ProductAttribute `json:"attributes"`
	Inventory  Inventory          `json:"inventory"`
}
