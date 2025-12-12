package models

type PaymentCard struct {
	PAN             string
	ExpirationMonth int
	ExpirationYear  int
	Brand           string
	Last4           string
}
