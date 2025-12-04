package service

import (
	"context"

	"github.com/account-service/internal/repository"
	"github.com/account-service/pkg/models"
	"github.com/google/uuid"
)

type RoleService struct {
	roleRepo repository.RoleRepository
}

func NewRoleService(rr repository.RoleRepository) *RoleService {
	return &RoleService{
		roleRepo: rr,
	}
}

func (s *RoleService) AddUserRole(ctx context.Context, userID string, roleID int) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	if roleID == 1 {
		return models.ErrCannotAssignRole
	}

	if hasRole, err := s.roleRepo.UserHasRole(ctx, uid, roleID); err != nil {
		return err
	} else if hasRole {
		return models.ErrUserAlreadyHasRole
	}

	return s.roleRepo.AddUserRole(ctx, uid, roleID)
}

func (s *RoleService) RemoveUserRole(ctx context.Context, userID string, roleID int) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	if roleID == 1 {
		return models.ErrCannotRemoveRole
	}

	if hasRole, err := s.roleRepo.UserHasRole(ctx, uid, roleID); err != nil {
		return err
	} else if !hasRole {
		return models.ErrUserDoesNotHaveRole
	}

	return s.roleRepo.RemoveUserRole(ctx, uid, roleID)
}

func (s *RoleService) GetAllRoles(ctx context.Context) ([]models.Role, error) {
	return s.roleRepo.GetAllRoles(ctx)
}
