package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

type UserClaims struct {
	UserID string   `json:"user_id"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

type Roles []string

func NewJWTManager(secretKey string, duration time.Duration) *JWTManager {
	return &JWTManager{secretKey, duration}
}

func (m *JWTManager) Generate(userID uuid.UUID, email string, roles []string) (string, error) {
	claims := UserClaims{
		UserID: userID.String(),
		Email:  email,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "account-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secretKey))
}

func (m *JWTManager) Verify(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.secretKey), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, err
	}
	return claims, nil
}

func (r Roles) IsAdmin() bool {
	for _, role := range r {
		if role == "admin" {
			return true
		}
	}
	return false
}

func (r Roles) IsAdminOrSeller() bool {
	for _, role := range r {
		if role == "admin" || role == "seller" {
			return true
		}
	}
	return false
}
