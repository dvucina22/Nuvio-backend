package service

import (
	"context"
	"errors"
	"time"

	"github.com/account-service/internal/repository"
	"github.com/account-service/pkg/models"
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

func (s *LoginService) Login(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error) {
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

	roleStrings := make([]string, len(user.Roles))
	for i, r := range user.Roles {
		roleStrings[i] = string(r.Name)
	}
	token, err := s.jwtManager.Generate(user.ID, user.Email, roleStrings)
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

	return &models.LoginResponse{
		Token:             token,
		FirstName:         firstName,
		LastName:          lastName,
		Email:             user.Email,
		Gender:            user.Gender,
		ProfilePictureURL: user.ProfilePictureURL,
		PhoneNumber:       user.PhoneNumber,
		Roles:             &user.Roles,
	}, nil
}
