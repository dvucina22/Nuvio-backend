package models

import "time"

type BankCard struct {
	ID              int       `json:"id"`
	Token           string    `json:"token"`
	LastFourDigits  string    `json:"lastFourDigits"`
	CardBrand       string    `json:"cardBrand"`
	ExpirationMonth int       `json:"expirationMonth"`
	ExpirationYear  int       `json:"expirationYear"`
	FullnameOnCard  string    `json:"fullnameOnCard"`
	CardName        *string   `json:"cardName,omitempty"`
	IsPrimary       bool      `json:"isPrimary"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`

	PANEncrypted []byte `json:"-"`
	IV           []byte `json:"-"`
}
