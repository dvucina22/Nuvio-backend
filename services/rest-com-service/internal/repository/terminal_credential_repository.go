package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type TerminalCredentials struct {
	TID string
	MID string
}

type TerminalCredentialRepository interface {
	GetActiveByUserAndHost(ctx context.Context, userID uuid.UUID, hostType string) (*TerminalCredentials, error)
	Create(ctx context.Context, userID uuid.UUID, hostType, tid, mid string, active bool) error
}

type terminalCredentialRepo struct {
	db *sql.DB
}

func NewTerminalCredentialRepository(db *sql.DB) TerminalCredentialRepository {
	return &terminalCredentialRepo{db: db}
}

func (r *terminalCredentialRepo) GetActiveByUserAndHost(ctx context.Context, userID uuid.UUID, hostType string) (*TerminalCredentials, error) {
	q := `
        SELECT tid, mid
        FROM rest_bank_comm.user_terminal_credentials
        WHERE user_id = $1
          AND host_type = $2
          AND active = true
        ORDER BY created_at DESC
        LIMIT 1
    `

	var creds TerminalCredentials

	err := r.db.QueryRowContext(ctx, q, userID, hostType).Scan(
		&creds.TID,
		&creds.MID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get terminal credentials: %w", err)
	}

	return &creds, nil
}

func (r *terminalCredentialRepo) Create(ctx context.Context, userID uuid.UUID, hostType, tid, mid string, active bool) error {
	q := `
		INSERT INTO rest_bank_comm.user_terminal_credentials (user_id, host_type, tid, mid, active)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.ExecContext(ctx, q, userID, hostType, tid, mid, active)
	if err != nil {
		return fmt.Errorf("failed to create terminal credentials: %w", err)
	}

	return nil
}
