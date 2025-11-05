package service

import (
	"context"
	"errors"
	"time"

	"github.com/account-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type LoginService struct {
	repo repository.AccountRepository
}

func NewLoginService(r repository.AccountRepository) *LoginService {
	return &LoginService{repo: r}
}

func (s *LoginService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("email and password required")
	}

	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("error retrieving user")
	}

	if user == nil {
		return nil, errors.New("invalid email")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid password")
	}

	now := time.Now()
	if err := s.repo.UpdateLastLogin(ctx, user.ID, now); err != nil {
		return nil, err
	}

	token := "dummy-token-" + user.ID.String()

	return &LoginResponse{Token: token}, nil
}
