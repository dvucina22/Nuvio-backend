package cart

type AddToCartRequest struct {
	ProductID int `json:"productId"`
	Quantity  int `json:"quantity"`
}
