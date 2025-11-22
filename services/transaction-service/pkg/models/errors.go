package models

import "errors"

var (
	ErrInvalidCard   = errors.New("invalid card data")
	ErrCardNotFound  = errors.New("card not found")
	ErrMissingFields = errors.New("missing required fields")
)
