package models

import "time"

type SaleProductResponse struct {
	ProductID int64   `json:"productId"`
	Quantity  int32   `json:"quantity"`
	UnitPrice int64   `json:"unitPrice"`
	Name      *string `json:"name,omitempty"`
	SKU       *string `json:"sku,omitempty"`
}

type SaleResponse struct {
	ID           int64                 `json:"id"`
	Status       string                `json:"status"`
	Amount       int64                 `json:"amount"`
	CurrencyCode string                `json:"currencyCode"`
	CreatedAt    time.Time             `json:"createdAt"`
	Products     []SaleProductResponse `json:"products"`
}
