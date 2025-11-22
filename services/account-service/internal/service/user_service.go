package service

import (
	"context"

	"github.com/account-service/internal/repository"
	"github.com/account-service/pkg/models"
	"github.com/account-service/pkg/utils"
)

type UserService struct {
	user_repo       repository.UserRepository
	account_repo    repository.AccountRepository
	password_helper *utils.PasswordHelper
}

func NewUserService(ur repository.UserRepository, ar repository.AccountRepository, ph *utils.PasswordHelper) *UserService {
	return &UserService{
		user_repo:       ur,
		account_repo:    ar,
		password_helper: ph,
	}
}

func (s *UserService) GetUserInfo(userID string) (*models.UserMinimal, error) {
	return s.user_repo.GetUserInfo(userID)
}

func (s *UserService) UpdateUserInfo(ctx context.Context, userID string, user *models.UpdateUser) error {
	if user == nil {
		return nil
	}

	if user.Email != nil {
		existing, err := s.account_repo.FindByEmail(ctx, *user.Email)

		if err != nil {
			return err
		}

		if existing != nil && existing.ID.String() != userID {
			return models.ErrEmailAlreadyExists
		}
	}

	return s.user_repo.UpdateUserInfo(userID, user)
}

func (s *UserService) UpdateUserPassword(ctx context.Context, userID string, password *models.UpdatePassword) error {
	if (password.NewPassword == "") || (password.OldPassword == "") {
		return models.ErrMissingFields
	}

	user, err := s.account_repo.FindById(ctx, userID)

	if err != nil {
		return err
	}

	if user == nil {
		return models.ErrUserNotFound
	}

	if err := s.password_helper.ComparePassword(user.PasswordHash, password.OldPassword); err != nil {
		return models.ErrInvalidPassword
	}

	if err := s.password_helper.ValidatePassword(password.NewPassword); err != nil {
		return models.ErrPasswordWeak
	}

	hashedPassword, err := s.password_helper.HashPassword(password.NewPassword)
	if err != nil {
		return err
	}

	return s.user_repo.UpdateUserPassword(userID, hashedPassword)
}

func (s *UserService) UpdateUserProfilePicture(userID string, profilePictureURL *string) error {
	return s.user_repo.UpdateUserProfilePicture(userID, profilePictureURL)
}
