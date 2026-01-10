package models

import (
	"errors"
	"net/http"
)

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

	ErrTransactionNotFound     = errors.New("transaction not found")
	ErrInvalidTransactionType  = errors.New("invalid transaction type")
	ErrInvalidTransactionState = errors.New("invalid transaction state")
	ErrVoidAlreadyExists       = errors.New("void already exists for this transaction")
)

type ApiErr struct {
	Status  int
	Code    string
	Message string
}

func MapError(err error) ApiErr {
	switch {
	case errors.Is(err, ErrCardNotFound):
		return ApiErr{Status: http.StatusNotFound, Code: "CARD_NOT_FOUND", Message: err.Error()}

	case errors.Is(err, ErrTransactionNotFound):
		return ApiErr{Status: http.StatusNotFound, Code: "TRANSACTION_NOT_FOUND", Message: err.Error()}

	case errors.Is(err, ErrVoidAlreadyExists):
		return ApiErr{Status: http.StatusConflict, Code: "VOID_ALREADY_EXISTS", Message: err.Error()}

	case errors.Is(err, ErrTerminalCredentialsNotFound):
		return ApiErr{Status: http.StatusUnprocessableEntity, Code: "TERMINAL_CREDENTIALS_NOT_FOUND", Message: err.Error()}

	case errors.Is(err, ErrInvalidCard),
		errors.Is(err, ErrInvalidCardNumber),
		errors.Is(err, ErrCardExpired),
		errors.Is(err, ErrMissingFields),
		errors.Is(err, ErrInvalidUserId),
		errors.Is(err, ErrInvalidCurrencyCode),
		errors.Is(err, ErrInvalidProducts),
		errors.Is(err, ErrInvalidAmount),
		errors.Is(err, ErrInvalidTransactionType),
		errors.Is(err, ErrInvalidTransactionState):
		return ApiErr{Status: http.StatusBadRequest, Code: "VALIDATION_ERROR", Message: err.Error()}

	case errors.Is(err, ErrEncryptionFailed):
		return ApiErr{Status: http.StatusInternalServerError, Code: "ENCRYPTION_FAILED", Message: err.Error()}

	case errors.Is(err, ErrDatabaseOperation):
		return ApiErr{Status: http.StatusInternalServerError, Code: "DATABASE_OPERATION_FAILED", Message: err.Error()}

	default:
		return ApiErr{Status: http.StatusInternalServerError, Code: "INTERNAL_ERROR", Message: "internal error"}
	}
}
