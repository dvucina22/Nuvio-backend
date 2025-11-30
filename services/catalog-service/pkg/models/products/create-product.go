package products

type CreateProduct struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description,omitempty"`
	ModelNumber string  `json:"modelNumber,omitempty"`
	SKU         string  `json:"sku,omitempty"`
	BasePrice   float64 `json:"basePrice" binding:"required"`

	BrandID    int64 `json:"brandId" binding:"required"`
	CategoryID int64 `json:"categoryId" binding:"required"`

	Quantity int64 `json:"quantity,omitempty"`
}
