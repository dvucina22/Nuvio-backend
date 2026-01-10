package models

type SaleItemDTO struct {
	Code      string `json:"code"`
	UnitPrice int64  `json:"unitPrice"`
	Quantity  int    `json:"quantity"`
}

type AuthorizeSaleReqDTO struct {
	UserID       string        `json:"userId"`
	CardNumber   string        `json:"cardNumber"`
	ExpiryMonth  int           `json:"expiryMonth"`
	ExpiryYear   int           `json:"expiryYear"`
	CurrencyCode string        `json:"currencyCode"`
	Amount       int64         `json:"amount"`
	Items        []SaleItemDTO `json:"items"`
}

type AuthorizeSaleRespDTO struct {
	HostType string `json:"hostType"`
	TID      string `json:"tid"`
	MID      string `json:"mid"`
	STAN     string `json:"stan"`
	RRN      string `json:"rrn"`

	Status       string `json:"status"`
	ResponseCode string `json:"responseCode"`
	AuthCode     string `json:"authCode,omitempty"`

	RawRequestHex  string `json:"rawRequestHex"`
	RawResponseHex string `json:"rawResponseHex"`
}

type AuthorizeVoidReqDTO struct {
	UserID             string `json:"userId"`
	OriginalRequestHex string `json:"originalRequestHex"`
}

type AuthorizeVoidRespDTO struct {
	HostType string `json:"hostType"`
	TID      string `json:"tid"`
	MID      string `json:"mid"`
	STAN     string `json:"stan"`

	Status       string `json:"status"`
	ResponseCode string `json:"responseCode"`

	RawRequestHex  string `json:"rawRequestHex"`
	RawResponseHex string `json:"rawResponseHex"`
}
