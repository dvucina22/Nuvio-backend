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
}

type terminalCredentialRepo struct {
	db *sql.DB
}

func NewTerminalCredentialRepository(db *sql.DB) TerminalCredentialRepository {
	return &terminalCredentialRepo{db: db}
}

func (r *terminalCredentialRepo) GetActiveByUserAndHost(
	ctx context.Context,
	userID uuid.UUID,
	hostType string,
) (*TerminalCredentials, error) {
	q := `
        SELECT tid, mid
        FROM transaction.user_terminal_credentials
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
