package models

type RegisterRequest struct {
	Email       string  `json:"email"`
	Password    string  `json:"password"`
	PhoneNumber *string `json:"phoneNumber,omitempty"`
	FirstName   *string `json:"firstName,omitempty"`
	LastName    *string `json:"lastName,omitempty"`
}
