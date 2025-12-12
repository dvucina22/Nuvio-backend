package models

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID                    int64
	UserID                uuid.UUID
	BankCardID            *int64
	Type                  string
	Status                string
	PANMasked             string
	CardFirstDigit        string
	CardExpirationYY      string
	CardExpirationMM      string
	ProcessingCode        string
	Amount                int64
	CurrencyCode          string
	STAN                  string
	TransactionTime       string
	TransactionDate       string
	RRN                   string
	TerminalTID           string
	MerchantMID           string
	HostType              string
	OriginalTransactionID *int64
	RequestPayload        []byte
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type TransactionProduct struct {
	ID            int64
	TransactionID int64
	ProductID     int64
	UnitPrice     int64
	Quantity      int
	ProductName   *string
	ProductSKU    *string
}
