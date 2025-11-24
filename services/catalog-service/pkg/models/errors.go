package models

import "errors"

var (
	ErrInvalidData     = errors.New("invalid data")
	ErrNotFound        = errors.New("resource not found")
	ErrProductNotFound = errors.New("product not found")
	ErrInvalidFilter   = errors.New("invalid filter")
)
