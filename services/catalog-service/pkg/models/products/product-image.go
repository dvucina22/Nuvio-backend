package products

type ProductImage struct {
	ID        int64  `json:"id"`
	URL       string `json:"url"`
	IsPrimary bool   `json:"isPrimary"`
}

type ProductImageRequest struct {
	ProductIds []int64 `json:"productIds"`
}

type ProductImageResponse struct {
	ProductID int64  `json:"productId"`
	ImageURL  string `json:"imageUrl"`
}
