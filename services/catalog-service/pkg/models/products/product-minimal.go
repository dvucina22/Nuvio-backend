package products

type ProductMinimal struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	ModelNumber string  `json:"modelNumber,omitempty"`
	SKU         string  `json:"sku,omitempty"`
	BasePrice   float64 `json:"basePrice"`
	IsActive    bool    `json:"isActive"`

	BrandName    string `json:"brand"`
	CategoryName string `json:"category"`

	ImageURL string `json:"imageUrl"`

	Attributes []ProductAttribute `json:"attributes"`

	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}
