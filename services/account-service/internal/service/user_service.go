package service

import (
	"context"

	"github.com/account-service/internal/repository"
	"github.com/account-service/pkg/models"
)

type UserService struct {
	user_repo    repository.UserRepository
	account_repo repository.AccountRepository
}

func NewUserService(ur repository.UserRepository, ar repository.AccountRepository) *UserService {
	return &UserService{
		user_repo:    ur,
		account_repo: ar,
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
