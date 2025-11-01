package account

import (
	"context"
	"database/sql"
	"errors"
)

type sqlRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) Repository {
	return &sqlRepository{db: db}
}

func (r *sqlRepository) Create(ctx context.Context, a *Account) error {
	q := `
		INSERT INTO accounts (email, password, created_at)
		VALUES ($1, $2, NOW())
		RETURNING id, created_at;
	`
	return r.db.QueryRowContext(ctx, q, a.Email, a.Password).Scan(&a.ID, &a.CreatedAt)
}

func (r *sqlRepository) FindByEmail(ctx context.Context, email string) (*Account, error) {
	var a Account
	q := `SELECT id, email, password, created_at FROM accounts WHERE email=$1 LIMIT 1;`
	err := r.db.QueryRowContext(ctx, q, email).Scan(&a.ID, &a.Email, &a.Password, &a.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("not found")
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}
