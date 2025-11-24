package models

type UpdateUser struct {
	FirstName         *string `json:"firstName,omitempty"`
	LastName          *string `json:"lastName,omitempty"`
	PhoneNumber       *string `json:"phoneNumber,omitempty"`
	Gender            *string `json:"gender,omitempty"`
	ProfilePictureURL *string `json:"profilePictureUrl,omitempty"`
}
