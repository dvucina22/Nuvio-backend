package products

type UpdateProduct struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	ModelNumber *string  `json:"modelNumber,omitempty"`
	SKU         *string  `json:"sku,omitempty"`
	BasePrice   *float64 `json:"basePrice,omitempty"`
	IsActive    *bool    `json:"isActive,omitempty"`

	BrandID    *int64 `json:"brandId,omitempty"`
	CategoryID *int64 `json:"categoryId,omitempty"`

	Quantity *int64 `json:"quantity,omitempty"`
}
