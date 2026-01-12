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
	ResponsePayload       []byte
	RequestPayload        []byte
	CreatedAt             time.Time
	UpdatedAt             time.Time
	ResponseCode          *string
	AuthCode              *string
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

type TransactionListItem struct {
    ID           int64     `json:"id"`
    Status       string    `json:"status"`
    Amount       int64     `json:"amount"`
    CurrencyCode string    `json:"currencyCode"`
    PANMasked    string    `json:"panMasked"`
    ProductCount int       `json:"productCount"`
    CreatedAt    time.Time `json:"createdAt"`
}

type TransactionListResponse struct {
    Transactions []*TransactionListItem `json:"transactions"`
    Total        int64                  `json:"total"`
    Page         int                    `json:"page"`
    PageSize     int                    `json:"pageSize"`
}

type TransactionProductDetail struct {
    ID          int64   `json:"id"`
    ProductID   int64   `json:"productId"`
    UnitPrice   int64   `json:"unitPrice"`
    Quantity    int     `json:"quantity"`
    LineTotal   int64   `json:"lineTotal"`
    ProductName *string `json:"productName,omitempty"`
    ProductSKU  *string `json:"productSku,omitempty"`
}

type TransactionDetail struct {
    ID               int64                      `json:"id"`
    Status           string                     `json:"status"`
    Amount           int64                      `json:"amount"`
    CurrencyCode     string                     `json:"currencyCode"`
    PANMasked        string                     `json:"panMasked"`
    CardExpirationYY string                     `json:"cardExpirationYy"`
    CardExpirationMM string                     `json:"cardExpirationMm"`
    TransactionDate  string                     `json:"transactionDate"`
    TransactionTime  string                     `json:"transactionTime"`
    CreatedAt        time.Time                  `json:"createdAt"`
    Products         []TransactionProductDetail `json:"products"`
}

type AdminTransactionListItem struct {
    ID                    int64      `json:"id"`
    UserID                uuid.UUID  `json:"userId"`
    Type                  string     `json:"type"`
    Status                string     `json:"status"`
    Amount                int64      `json:"amount"`
    CurrencyCode          string     `json:"currencyCode"`
    PANMasked             string     `json:"panMasked"`
    ResponseCode          *string    `json:"responseCode,omitempty"`
    AuthCode              *string    `json:"authCode,omitempty"`
    OriginalTransactionID *int64     `json:"originalTransactionId,omitempty"`
    ProductCount          int        `json:"productCount"`
    CreatedAt             time.Time  `json:"createdAt"`
}

type AdminTransactionListResponse struct {
    Transactions []*AdminTransactionListItem `json:"transactions"`
    Total        int64                       `json:"total"`
    Page         int                         `json:"page"`
    PageSize     int                         `json:"pageSize"`
}

type AdminTransactionDetail struct {
    ID                    int64                      `json:"id"`
    UserID                uuid.UUID                  `json:"userId"`
    BankCardID            *int64                     `json:"bankCardId,omitempty"`
    Type                  string                     `json:"type"`
    Status                string                     `json:"status"`
    PANMasked             string                     `json:"panMasked"`
    CardFirstDigit        string                     `json:"cardFirstDigit"`
    CardExpirationYY      string                     `json:"cardExpirationYy"`
    CardExpirationMM      string                     `json:"cardExpirationMm"`
    ProcessingCode        string                     `json:"processingCode"`
    Amount                int64                      `json:"amount"`
    CurrencyCode          string                     `json:"currencyCode"`
    STAN                  string                     `json:"stan"`
    TransactionTime       string                     `json:"transactionTime"`
    TransactionDate       string                     `json:"transactionDate"`
    RRN                   string                     `json:"rrn"`
    TerminalTID           string                     `json:"terminalTid"`
    MerchantMID           string                     `json:"merchantMid"`
    HostType              string                     `json:"hostType"`
    OriginalTransactionID *int64                     `json:"originalTransactionId,omitempty"`
    ResponseCode          *string                    `json:"responseCode,omitempty"`
    AuthCode              *string                    `json:"authCode,omitempty"`
    CreatedAt             time.Time                  `json:"createdAt"`
    UpdatedAt             time.Time                  `json:"updatedAt"`
    Products              []TransactionProductDetail `json:"products,omitempty"`
}
