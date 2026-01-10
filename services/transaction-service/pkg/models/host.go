package models

type HostSaleRequest struct {
	CardNumber   string
	ExpiryMonth  int
	ExpiryYear   int
	CurrencyCode string
	Amount       int64

	Products []*TransactionProduct
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

	ResponseCode *string
	AuthCode     *string
}
