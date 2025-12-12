package models

import (
	"context"

	"github.com/google/uuid"
)

type HostSaleRequest struct {
	CardNumber   string
	ExpiryMonth  int
	ExpiryYear   int
	CurrencyCode string
	Amount       int64
}

type HostSaleResponse struct {
	HostType        string
	TerminalTID     string
	MerchantMID     string
	STAN            string
	TransactionTime string
	TransactionDate string
	RRN             string
	Status          string
	CardFirstDigit  string
	RawRequest      []byte
	RawResponse     []byte
}

type HostClient interface {
	AuthorizeSale(ctx context.Context, userID uuid.UUID, req *HostSaleRequest) (*HostSaleResponse, error)
}
