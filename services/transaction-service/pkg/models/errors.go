package models

import "errors"

var (
	ErrInvalidCard       = errors.New("invalid card data")
	ErrCardNotFound      = errors.New("card not found")
	ErrMissingFields     = errors.New("missing required fields")
	ErrInvalidUserId     = errors.New("invalid user ID")
	ErrDatabaseOperation = errors.New("database operation failed")
	ErrInvalidCardNumber = errors.New("invalid card number")
	ErrCardExpired       = errors.New("card has expired")
	ErrEncryptionFailed  = errors.New("failed to encrypt card data")

	ErrInvalidCurrencyCode         = errors.New("invalid currency code")
	ErrInvalidProducts             = errors.New("invalid products")
	ErrTerminalCredentialsNotFound = errors.New("terminal credentials not found")
	ErrInvalidAmount               = errors.New("invalid amount")
)
