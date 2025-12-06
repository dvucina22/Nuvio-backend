package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/account-service/pkg/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type AccountRepository interface {
	Create(ctx context.Context, u *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateLastLogin(ctx context.Context, id uuid.UUID, t time.Time) error
	FindById(ctx context.Context, id string) (*models.User, error)
}

type accountRepo struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) AccountRepository {
	return &accountRepo{db: db}
}

func (r *accountRepo) Create(ctx context.Context, u *models.User) error {
	q := `INSERT INTO account.users (
			email, phone_number, password_hash, first_name, last_name, is_active, gender
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRowContext(
		ctx,
		q,
		u.Email,
		u.PhoneNumber,
		u.PasswordHash,
		u.FirstName,
		u.LastName,
		u.IsActive,
		u.Gender,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

func (r *accountRepo) FindByEmail(ctx context.Context, id string) (*models.User, error) {
	var (
		u         models.User
		roleIDs   []sql.NullInt64
		roleNames []sql.NullString
	)

	q := `
        SELECT 
            u.id,
            u.email,
            u.phone_number,
            u.password_hash,
            u.gender,
            u.profile_picture_url,
            u.first_name,
            u.last_name,
            u.is_active,
            u.created_at,
            u.updated_at,
            u.last_login_at,
            COALESCE(array_agg(r.id)   FILTER (WHERE r.id   IS NOT NULL), '{}') AS role_ids,
            COALESCE(array_agg(r.name) FILTER (WHERE r.name IS NOT NULL), '{}') AS role_names
        FROM account.users u
        LEFT JOIN account.user_roles_map m ON m.user_id = u.id
        LEFT JOIN account.user_roles r ON r.id = m.role_id
        WHERE u.email = $1
        GROUP BY u.id
        LIMIT 1
    `

	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&u.ID,
		&u.Email,
		&u.PhoneNumber,
		&u.PasswordHash,
		&u.Gender,
		&u.ProfilePictureURL,
		&u.FirstName,
		&u.LastName,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.LastLoginAt,
		pq.Array(&roleIDs),
		pq.Array(&roleNames),
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if len(roleIDs) > 0 && len(roleIDs) == len(roleNames) {
		roles := make([]models.Role, 0, len(roleIDs))
		for i := range roleIDs {
			if !roleIDs[i].Valid && !roleNames[i].Valid {
				continue
			}

			role := models.Role{}
			if roleIDs[i].Valid {
				role.ID = int(roleIDs[i].Int64)
			}
			if roleNames[i].Valid {
				role.Name = roleNames[i].String
			}

			roles = append(roles, role)
		}

		u.Roles = roles
	}

	return &u, nil
}

func (r *accountRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID, t time.Time) error {
	q := `UPDATE account.users SET last_login_at = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, q, t, id)
	return err
}

func (r *accountRepo) FindById(ctx context.Context, id string) (*models.User, error) {
	var (
		u         models.User
		roleIDs   []sql.NullInt64
		roleNames []sql.NullString
	)

	q := `
        SELECT 
            u.id,
            u.email,
            u.phone_number,
            u.password_hash,
            u.gender,
            u.profile_picture_url,
            u.first_name,
            u.last_name,
            u.is_active,
            u.created_at,
            u.updated_at,
            u.last_login_at,
            COALESCE(array_agg(r.id)   FILTER (WHERE r.id   IS NOT NULL), '{}') AS role_ids,
            COALESCE(array_agg(r.name) FILTER (WHERE r.name IS NOT NULL), '{}') AS role_names
        FROM account.users u
        LEFT JOIN account.user_roles_map m ON m.user_id = u.id
        LEFT JOIN account.user_roles r ON r.id = m.role_id
        WHERE u.id = $1
        GROUP BY u.id
        LIMIT 1
    `

	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&u.ID,
		&u.Email,
		&u.PhoneNumber,
		&u.PasswordHash,
		&u.Gender,
		&u.ProfilePictureURL,
		&u.FirstName,
		&u.LastName,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.LastLoginAt,
		pq.Array(&roleIDs),
		pq.Array(&roleNames),
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if len(roleIDs) > 0 && len(roleIDs) == len(roleNames) {
		roles := make([]models.Role, 0, len(roleIDs))
		for i := range roleIDs {
			if !roleIDs[i].Valid && !roleNames[i].Valid {
				continue
			}

			role := models.Role{}
			if roleIDs[i].Valid {
				role.ID = int(roleIDs[i].Int64)
			}
			if roleNames[i].Valid {
				role.Name = roleNames[i].String
			}

			roles = append(roles, role)
		}

		u.Roles = roles
	}

	return &u, nil
}
