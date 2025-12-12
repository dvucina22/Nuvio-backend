package models

import "github.com/google/uuid"

type SaleProduct struct {
	ProductID int64   `json:"productId"`
	Quantity  int     `json:"quantity"`
	UnitPrice int64   `json:"unitPrice"`
	Name      *string `json:"name,omitempty"`
	SKU       *string `json:"sku,omitempty"`
}

type SaleRequest struct {
	UserID       uuid.UUID `json:"userId"`
	CardNumber   string    `json:"cardNumber"`
	ExpiryMonth  int       `json:"expiryMonth"`
	ExpiryYear   int       `json:"expiryYear"`
	CurrencyCode string    `json:"currencyCode"`

	Products []SaleProduct `json:"products"`

	TotalAmount int64 `json:"totalAmount"`
}
