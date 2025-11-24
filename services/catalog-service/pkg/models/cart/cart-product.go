package cart

type CartProduct struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	BasePrice float64 `json:"basePrice"`

	BrandName string `json:"brand"`

	ImageURL string `json:"imageUrl"`

	Quantity int `json:"quantity"`

	IsFavorite bool `json:"isFavorite"`
}
