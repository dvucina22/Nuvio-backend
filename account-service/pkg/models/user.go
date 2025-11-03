package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	Email        string     `json:"email" db:"email"`
	PhoneNumber  *string    `json:"phoneNumber,omitempty" db:"phone_number"`
	PasswordHash string     `json:"-" db:"password_hash"`
	FirstName    *string    `json:"firstName,omitempty" db:"first_name"`
	LastName     *string    `json:"lastName,omitempty" db:"last_name"`
	IsActive     bool       `json:"isActive" db:"is_active"`
	CreatedAt    time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time  `json:"updatedAt" db:"updated_at"`
	LastLoginAt  *time.Time `json:"lastLoginAt,omitempty" db:"last_login_at"`
}
