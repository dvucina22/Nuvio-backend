package favorites

type FavoriteRequest struct {
	ProductID int `json:"productId" validate:"required"`
}
