package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type RoleRepository interface {
	AddUserRole(ctx context.Context, userID uuid.UUID, roleID int) error
	RemoveUserRole(ctx context.Context, userID uuid.UUID, roleID int) error
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
