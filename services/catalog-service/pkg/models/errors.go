package models

import "errors"

var (
	ErrInvalidData      = errors.New("invalid data")
	ErrNotFound         = errors.New("resource not found")
	ErrProductNotFound  = errors.New("product not found")
	ErrInvalidFilter    = errors.New("invalid filter")
	ErrInvalidProductID = errors.New("invalid product ID")
	ErrInvalidQuantity  = errors.New("invalid quantity")
	ErrProductNotInCart = errors.New("product not in cart")
	ErrCartNotFound     = errors.New("cart not found")
	ErrBrandNotFound    = errors.New("brand not found")
	ErrCategoryNotFound = errors.New("category not found")
)
