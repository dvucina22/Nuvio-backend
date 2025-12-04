package repository

import (
	"context"
	"database/sql"

	"github.com/account-service/pkg/models"
	"github.com/google/uuid"
)

type RoleRepository interface {
	AddUserRole(ctx context.Context, userID uuid.UUID, roleID int) error
	RemoveUserRole(ctx context.Context, userID uuid.UUID, roleID int) error
	GetAllRoles(ctx context.Context) ([]models.Role, error)
	UserHasRole(ctx context.Context, userID uuid.UUID, roleID int) (bool, error)
}

type roleRepo struct {
	db *sql.DB
}

func NewRoleRepository(db *sql.DB) RoleRepository {
	return &roleRepo{db: db}
}

func (r *roleRepo) AddUserRole(ctx context.Context, userID uuid.UUID, roleID int) error {
	q := `INSERT INTO account.user_roles_map (user_id, role_id) VALUES ($1, $2)`

	_, err := r.db.ExecContext(ctx, q, userID, roleID)
	return err
}

func (r *roleRepo) RemoveUserRole(ctx context.Context, userID uuid.UUID, roleID int) error {
	q := `DELETE FROM account.user_roles_map WHERE user_id = $1 AND role_id = $2`

	_, err := r.db.ExecContext(ctx, q, userID, roleID)
	return err
}

func (r *roleRepo) GetAllRoles(ctx context.Context) ([]models.Role, error) {
	q := `SELECT id, name FROM account.user_roles`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.Name); err != nil {
			return nil, err
		}
		if role.ID == 1 {
			continue
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (r *roleRepo) UserHasRole(ctx context.Context, userID uuid.UUID, roleID int) (bool, error) {
	q := `SELECT COUNT(1) FROM account.user_roles_map WHERE user_id = $1 AND role_id = $2`

	var count int

	err := r.db.QueryRowContext(ctx, q, userID, roleID).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
