package service

import (
	"context"
	"errors"
	"time"

	"github.com/account-service/internal/repository"
	"github.com/account-service/pkg/utils"
	"golang.org/x/crypto/bcrypt"
)

type LoginService struct {
	repo       repository.AccountRepository
	jwtManager *utils.JWTManager
}

func NewLoginService(r repository.AccountRepository, jwtManager *utils.JWTManager) *LoginService {
	return &LoginService{
		repo:       r,
		jwtManager: jwtManager,
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Email     string `json:"email,omitempty"`
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
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	now := time.Now()
	if err := s.repo.UpdateLastLogin(ctx, user.ID, now); err != nil {
		return nil, err
	}

	token, err := s.jwtManager.Generate(user.ID, user.Email)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	var firstName, lastName string
	if user.FirstName != nil {
		firstName = *user.FirstName
	}
	if user.LastName != nil {
		lastName = *user.LastName
	}

	return &LoginResponse{
		Token:     token,
		FirstName: firstName,
		LastName:  lastName,
		Email:     user.Email,
	}, nil
}
