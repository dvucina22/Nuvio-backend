package service

import (
	"context"
	"errors"
	"time"

	"github.com/account-service/internal/repository"
	"github.com/account-service/pkg/models"
	"github.com/account-service/pkg/utils"
)

type RegisterService struct {
	repo            repository.AccountRepository
	role_repo       repository.RoleRepository
	password_helper *utils.PasswordHelper
}

func NewRegisterService(r repository.AccountRepository, rr repository.RoleRepository, ph *utils.PasswordHelper) *RegisterService {
	return &RegisterService{
		repo:            r,
		role_repo:       rr,
		password_helper: ph,
	}
}

func (s *RegisterService) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("email and password required")
	}

	if err := s.password_helper.ValidatePassword(req.Password); err != nil {
		return nil, err
	}

	existing, err := s.repo.FindByEmail(ctx, req.Email)

	if err != nil {
		return nil, err
	}

	if existing != nil {
		return nil, errors.New("email already exists")
	}

	hashedPassword, err := s.password_helper.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:        req.Email,
		PhoneNumber:  req.PhoneNumber,
		PasswordHash: hashedPassword,
		Gender:       req.Gender,
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
