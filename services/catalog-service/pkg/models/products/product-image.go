package products

type ProductImage struct {
	ID        int64  `json:"id"`
	URL       string `json:"url"`
	IsPrimary bool   `json:"isPrimary"`
}
