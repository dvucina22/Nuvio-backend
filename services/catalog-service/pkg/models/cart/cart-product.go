package cart

import "github.com/catalog-service/pkg/models/products"

type CartProduct struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	BasePrice float64 `json:"basePrice"`

	BrandName    string                      `json:"brand"`
	CategoryName string                      `json:"category"`
	Attributes   []products.ProductAttribute `json:"attributes"`

	ImageURL string `json:"imageUrl"`

	Quantity int `json:"quantity"`

	IsFavorite bool `json:"isFavorite"`
}
