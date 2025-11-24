package models

import "github.com/google/uuid"

type UserMinimal struct {
	ID                uuid.UUID `json:"id"`
	FirstName         *string   `json:"firstName,omitempty"`
	LastName          *string   `json:"lastName,omitempty"`
	Email             string    `json:"email,omitempty"`
	Gender            *string   `json:"gender,omitempty"`
	ProfilePictureURL *string   `json:"profilePictureUrl,omitempty"`
	PhoneNumber       *string   `json:"phoneNumber,omitempty"`
}
