package models

type ProductImage struct {
	ID        int64  `json:"id"`
	URL       string `json:"url"`
	IsPrimary bool   `json:"isPrimary"`
	VariantID *int64 `json:"variantId,omitempty"`
}
