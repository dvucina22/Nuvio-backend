package account

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type bcryptHasher struct{ cost int }

func NewBCryptHasher(cost int) Hasher { return &bcryptHasher{cost: cost} }

func (b *bcryptHasher) Hash(pw string) (string, error) {
	hs, err := bcrypt.GenerateFromPassword([]byte(pw), b.cost)
	return string(hs), err
}
func (b *bcryptHasher) Compare(hashed, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
}

type jwtProvider struct {
	secret []byte
	ttl    time.Duration
}

func NewJWT(secret string, ttl time.Duration) TokenProvider {
	return &jwtProvider{secret: []byte(secret), ttl: ttl}
}

func (j *jwtProvider) Issue(email string) (TokenPair, error) {
	claims := jwt.MapClaims{
		"sub":   email,
		"exp":   time.Now().Add(j.ttl).Unix(),
		"iat":   time.Now().Unix(),
		"scope": "account:basic",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(j.secret)
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{AccessToken: signed}, nil
}
