package models

type EncryptedCardData struct {
	PANEncrypted    []byte `json:"-"`
	IV              []byte `json:"-"`
	LastFourDigits  string `json:"lastFourDigits"`
	CardBrand       string `json:"cardBrand"`
	ExpirationMonth int    `json:"expirationMonth"`
	ExpirationYear  int    `json:"expirationYear"`
	FullnameOnCard  string `json:"fullnameOnCard"`
}
