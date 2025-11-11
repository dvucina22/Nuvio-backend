package service

import (
	"context"
	"errors"
	"time"

	"github.com/account-service/internal/repository"
	"github.com/account-service/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Email       string  `json:"email"`
	Password    string  `json:"password"`
	PhoneNumber *string `json:"phoneNumber,omitempty"`
	FirstName   *string `json:"firstName,omitempty"`
	LastName    *string `json:"lastName,omitempty"`
}

type RegisterService struct {
	repo      repository.AccountRepository
	role_repo repository.RoleRepository
}

func NewRegisterService(r repository.AccountRepository, rr repository.RoleRepository) *RegisterService {
	return &RegisterService{
		repo:      r,
		role_repo: rr,
	}
}

func (s *RegisterService) Register(ctx context.Context, req RegisterRequest) (*models.User, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("email and password required")
	}

	if err := validatePassword(req.Password); err != nil {
		return nil, err
	}

	existing, err := s.repo.FindByEmail(ctx, req.Email)

	if err != nil {
		return nil, err
	}

	if existing != nil {
		return nil, errors.New("email already exists")
	}

	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:        req.Email,
		PhoneNumber:  req.PhoneNumber,
		PasswordHash: hashedPassword,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	if err := s.role_repo.AddUserRole(ctx, user.ID, 2); err != nil {
		return nil, err
	}
	user.PasswordHash = ""
	return user, nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func validatePassword(pasword string) error {
	if len(pasword) < 8 {
		return errors.New("password must be at least 8 characters long, contain one uppercase letter and one number")
	}

	var hasUpper bool
	var hasDigit bool

	for _, r := range pasword {
		if r >= 'A' && r <= 'Z' {
			hasUpper = true
		}
		if r >= '0' && r <= '9' {
			hasDigit = true
		}
		if hasUpper && hasDigit {
			break
		}
	}

	if !hasUpper || !hasDigit {
		return errors.New("password must be at least 8 characters long, contain one uppercase letter and one number")
	}
	return nil
}
