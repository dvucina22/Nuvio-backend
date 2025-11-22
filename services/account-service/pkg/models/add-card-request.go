package models

type AddCardRequest struct {
	CardNumber      string  `json:"cardNumber"`
	ExpirationMonth int     `json:"expirationMonth"`
	ExpirationYear  int     `json:"expirationYear"`
	FullnameOnCard  string  `json:"fullnameOnCard"`
	CardName        *string `json:"cardName,omitempty"`
	IsPrimary       bool    `json:"isPrimary"`
}
