package models

type Product struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	ModelNumber string  `json:"modelNumber,omitempty"`
	SKU         string  `json:"sku,omitempty"`
	BasePrice   float64 `json:"basePrice"`
	IsActive    bool    `json:"isActive"`

	Brand    Brand    `json:"brand"`
	Category Category `json:"category"`

	Images     []ProductImage     `json:"images"`
	Attributes []ProductAttribute `json:"attributes"`

	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}
