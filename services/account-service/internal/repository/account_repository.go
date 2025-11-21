package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/account-service/pkg/models"
	"github.com/google/uuid"
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
			email, phone_number, password_hash, first_name, last_name, is_active
		)
		VALUES ($1, $2, $3, $4, $5, $6)
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
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

func (r *accountRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User

	q := `SELECT 
			id,
			email,
			phone_number,
			password_hash,
			first_name,
			last_name,
			is_active,
			created_at,
			updated_at,
			last_login_at
		FROM account.users
		WHERE email = $1
		LIMIT 1`

	err := r.db.QueryRowContext(ctx, q, email).Scan(
		&u.ID,
		&u.Email,
		&u.PhoneNumber,
		&u.PasswordHash,
		&u.FirstName,
		&u.LastName,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.LastLoginAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *accountRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID, t time.Time) error {
	q := `UPDATE account.users SET last_login_at = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, q, t, id)
	return err
}

func (r *accountRepo) FindById(ctx context.Context, id string) (*models.User, error) {
	var u models.User
	q := `SELECT id, email, phone_number, password_hash, first_name, last_name, is_active, created_at, updated_at, last_login_at
		FROM account.users
		WHERE id = $1
		LIMIT 1`

	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&u.ID,
		&u.Email,
		&u.PhoneNumber,
		&u.PasswordHash,
		&u.FirstName,
		&u.LastName,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.LastLoginAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
