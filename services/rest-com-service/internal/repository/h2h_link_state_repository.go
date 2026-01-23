package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type H2HLinkStateRepository interface {
	NextSTAN(ctx context.Context, hostType string) (string, error)
}

type h2hLinkStateRepo struct {
	db *sql.DB
}

func NewH2HLinkStateRepository(db *sql.DB) H2HLinkStateRepository {
	return &h2hLinkStateRepo{db: db}
}

func (r *h2hLinkStateRepo) NextSTAN(ctx context.Context, hostType string) (string, error) {
	q := `
        INSERT INTO rest_bank_comm.h2h_link_state (host_type, last_stan)
        VALUES ($1, 0)
        ON CONFLICT (host_type)
        DO UPDATE SET last_stan = (rest_bank_comm.h2h_link_state.last_stan + 1) % 1000000
        RETURNING last_stan
    `

	var stan int
	if err := r.db.QueryRowContext(ctx, q, hostType).Scan(&stan); err != nil {
		return "", fmt.Errorf("failed to get next stan: %w", err)
	}

	return fmt.Sprintf("%06d", stan), nil
}
