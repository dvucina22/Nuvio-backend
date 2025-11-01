package account

import (
	"context"
	"errors"
	"time"
)

type Account struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"createdAt"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenPair struct {
	AccessToken string `json:"accessToken"`
}

type Repository interface {
	Create(ctx context.Context, a *Account) error
	FindByEmail(ctx context.Context, email string) (*Account, error)
}

type Hasher interface {
	Hash(pw string) (string, error)
	Compare(hashed, plain string) error
}

type TokenProvider interface {
	Issue(email string) (TokenPair, error)
}

type Service interface {
	Register(ctx context.Context, req RegisterRequest) (*Account, error)
	Login(ctx context.Context, req LoginRequest) (*TokenPair, error)
}

type service struct {
	repo   Repository
	hasher Hasher
	token  TokenProvider
}

func NewService(r Repository, h Hasher, t TokenProvider) Service {
	return &service{repo: r, hasher: h, token: t}
}

func (s *service) Register(ctx context.Context, req RegisterRequest) (*Account, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("email and password required")
	}
	if _, err := s.repo.FindByEmail(ctx, req.Email); err == nil {
		return nil, errors.New("email already exists")
	}

	hash, err := s.hasher.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	acc := &Account{
		Email:     req.Email,
		Password:  hash,
		CreatedAt: time.Now(),
	}
	if err := s.repo.Create(ctx, acc); err != nil {
		return nil, err
	}
	acc.Password = ""
	return acc, nil
}

func (s *service) Login(ctx context.Context, req LoginRequest) (*TokenPair, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("email and password required")
	}
	acc, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}
	if err := s.hasher.Compare(acc.Password, req.Password); err != nil {
		return nil, errors.New("invalid email or password")
	}
	tok, err := s.token.Issue(acc.Email)
	if err != nil {
		return nil, err
	}
	return &tok, nil
}
