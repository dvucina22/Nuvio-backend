package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/account-service/pkg/models"
	"github.com/google/uuid"
)

type OAuthRepository interface {
	FindByProviderID(ctx context.Context, provider, providerUserID string) (*models.User, error)
	LinkAccount(ctx context.Context, userID uuid.UUID, provider, providerUserID, accessToken string, refreshToken *string) error
}

type oauthRepo struct {
	db *sql.DB
}

func NewOAuthRepo(db *sql.DB) OAuthRepository {
	return &oauthRepo{db: db}
}

func (r *oauthRepo) FindByProviderID(ctx context.Context, provider, providerUserID string) (*models.User, error) {
	const q = `
        SELECT 
            u.id,
            u.email,
            u.phone_number,
            u.first_name,
            u.last_name,
            u.is_active,
            u.created_at,
            u.updated_at,
            u.last_login_at,
            COALESCE(
                json_agg(
                    json_build_object('id', r2.id, 'name', r2.name)
                ) FILTER (WHERE r2.id IS NOT NULL),
                '[]'
            ) AS roles
        FROM account.oauth_accounts oa
        JOIN account.users u ON u.id = oa.user_id
        LEFT JOIN account.user_roles_map m ON m.user_id = u.id
        LEFT JOIN account.user_roles r2 ON r2.id = m.role_id
        WHERE oa.provider = $1 AND oa.provider_user_id = $2
        GROUP BY u.id;
    `

	var u models.User
	var rolesJSON []byte

	err := r.db.QueryRowContext(ctx, q, provider, providerUserID).Scan(
		&u.ID,
		&u.Email,
		&u.PhoneNumber,
		&u.FirstName,
		&u.LastName,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.LastLoginAt,
		&rolesJSON,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(rolesJSON, &u.Roles); err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *oauthRepo) LinkAccount(ctx context.Context, userID uuid.UUID, provider, providerUserID, accessToken string, refreshToken *string) error {
	const q = `INSERT INTO account.oauth_accounts (user_id, provider, provider_user_id, access_token, refresh_token)
				VALUES ($1,$2,$3,$4,$5)
				ON CONFLICT (provider, provider_user_id) DO UPDATE
				SET user_id = EXCLUDED.user_id,
					access_token = EXCLUDED.access_token,
					refresh_token = EXCLUDED.refresh_token`
	_, err := r.db.ExecContext(ctx, q, userID, provider, providerUserID, accessToken, refreshToken)
	return err
}
