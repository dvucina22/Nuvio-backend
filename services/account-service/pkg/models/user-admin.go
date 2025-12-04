package models

import (
	"time"

	"github.com/google/uuid"
)

type UserAdmin struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	Email             string     `json:"email" db:"email"`
	PhoneNumber       *string    `json:"phoneNumber,omitempty" db:"phone_number"`
	FirstName         *string    `json:"firstName,omitempty" db:"first_name"`
	LastName          *string    `json:"lastName,omitempty" db:"last_name"`
	IsActive          bool       `json:"isActive" db:"is_active"`
	CreatedAt         time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt         time.Time  `json:"updatedAt" db:"updated_at"`
	LastLoginAt       *time.Time `json:"lastLoginAt,omitempty" db:"last_login_at"`
	ProfilePictureURL *string    `json:"profilePictureUrl,omitempty" db:"profile_picture_url"`
	Gender            *string    `json:"gender,omitempty" db:"gender"`
	Roles             []Role     `json:"roles,omitempty"`
}
