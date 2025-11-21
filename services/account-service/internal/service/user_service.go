package service

import (
	"github.com/account-service/internal/repository"
	"github.com/account-service/pkg/models"
)

type UserService struct {
	user_repo repository.UserRepository
}

func NewUserService(ur repository.UserRepository) *UserService {
	return &UserService{
		user_repo: ur,
	}
}

func (s *UserService) GetUserInfo(userID string) (*models.UserMinimal, error) {
	return s.user_repo.GetUserInfo(userID)
}
