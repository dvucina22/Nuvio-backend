package models

type SaleItem struct {
	Code      string `json:"code"`
	UnitPrice int64  `json:"unitPrice"`
	Quantity  int    `json:"quantity"`
}

type AuthorizeSaleRequest struct {
	MessageType  string     `json:"messageType"`
	UserID       string     `json:"userId"`
	CardNumber   string     `json:"cardNumber"`
	ExpiryMonth  int        `json:"expiryMonth"`
	ExpiryYear   int        `json:"expiryYear"`
	CurrencyCode string     `json:"currencyCode"`
	Amount       int64      `json:"amount"`
	Items        []SaleItem `json:"items"`
}

type AuthorizeSaleResponse struct {
	Status       string `json:"status"`
	ResponseCode string `json:"responseCode"`
	AuthCode     string `json:"authCode,omitempty"`

	HostType string `json:"hostType"`
	TID      string `json:"tid"`
	MID      string `json:"mid"`
	STAN     string `json:"stan"`
	RRN      string `json:"rrn"`

	RawRequestHex  string `json:"rawRequestHex"`
	RawResponseHex string `json:"rawResponseHex"`
}

type AuthorizeVoidRequest struct {
	MessageType   string `json:"messageType"`
	UserID        string `json:"userId"`
	TransactionID int64  `json:"transactionId"`
}

type AuthorizeVoidResponse struct {
	Status       string `json:"status"`
	ResponseCode string `json:"responseCode"`

	HostType string `json:"hostType"`
	TID      string `json:"tid"`
	MID      string `json:"mid"`
	STAN     string `json:"stan"`

	RawRequestHex  string `json:"rawRequestHex"`
	RawResponseHex string `json:"rawResponseHex"`
}

type HostResponse struct {
	Status       string `json:"status"`
	ResponseCode string `json:"responseCode"`
	AuthCode     string `json:"authCode,omitempty"`
}
