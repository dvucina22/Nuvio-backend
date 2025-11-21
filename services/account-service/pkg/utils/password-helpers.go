package utils

import (
	"errors"
	"unicode"

	"github.com/account-service/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

type PasswordHelper struct{}

func NewPasswordHelper() *PasswordHelper {
	return &PasswordHelper{}
}

func (p *PasswordHelper) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", models.ErrPasswordHashFailed
	}
	return string(bytes), nil
}

func (p *PasswordHelper) ValidatePassword(password string) error {
	if len(password) < 8 {
		return models.ErrPasswordInvalid
	}

	var hasUpper bool
	var hasDigit bool

	for _, r := range password {
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
		if hasUpper && hasDigit {
			break
		}
	}

	if !hasUpper || !hasDigit {
		return models.ErrPasswordInvalid
	}
	return nil
}

func (p *PasswordHelper) ComparePassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return bcrypt.ErrMismatchedHashAndPassword
		}
		return models.ErrPasswordCompareFailed
	}
	return nil
}

func (p *PasswordHelper) ValidateAndHash(password string) (string, error) {
	if err := p.ValidatePassword(password); err != nil {
		return "", err
	}
	return p.HashPassword(password)
}
