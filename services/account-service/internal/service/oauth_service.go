package service

import (
	"context"
	"strings"

	"errors"
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/auth/credentials/idtoken"
	"github.com/account-service/internal/config"
	"github.com/account-service/internal/repository"
	"github.com/account-service/pkg/models"
	"github.com/account-service/pkg/types"
	"github.com/account-service/pkg/utils"
)

type OAuthService struct {
	accounts   repository.AccountRepository
	oauthRepo  repository.OAuthRepository
	jwtManager *utils.JWTManager
	cfgs       *config.OAuth2Configs
	httpClient *http.Client
}

func NewOAuthService(a repository.AccountRepository, o repository.OAuthRepository, jwt *utils.JWTManager,
	cfgs *config.OAuth2Configs) *OAuthService {
	return &OAuthService{
		accounts:   a,
		oauthRepo:  o,
		jwtManager: jwt,
		cfgs:       cfgs,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type OAuthLoginResponse struct {
	Token string `json:"token"`
}

func (s *OAuthService) VerifyIDToken(ctx context.Context, p types.Provider, idToken string) (*OAuthLoginResponse, error) {
	switch p {
	case types.ProviderGoogle:
		payload, err := idtoken.Validate(ctx, idToken, s.cfgs.Google.ClientID)
		if err != nil {
			return nil, fmt.Errorf("invalid google token: %w", err)
		}

		email, _ := payload.Claims["email"].(string)
		name, _ := payload.Claims["name"].(string)
		sub := payload.Subject

		if email == "" {
			return nil, errors.New("email not found in token")
		}

		user, err := s.oauthRepo.FindByProviderID(ctx, p.String(), sub)
		if err != nil {
			return nil, fmt.Errorf("database error: %w", err)
		}

		if user == nil {
			user, err = s.accounts.FindByEmail(ctx, email)
			if err != nil {
				return nil, fmt.Errorf("database error: %w", err)
			}

			if user == nil {
				now := time.Now()
				firstName, lastName := splitName(name)
				user = &models.User{
					Email:     email,
					FirstName: &firstName,
					LastName:  &lastName,
					IsActive:  true,
					CreatedAt: now,
					UpdatedAt: now,
				}
				if err := s.accounts.Create(ctx, user); err != nil {
					return nil, fmt.Errorf("failed to create user: %w", err)
				}
			}

			if err := s.oauthRepo.LinkAccount(ctx, user.ID, p.String(), sub, "", nil); err != nil {
				return nil, fmt.Errorf("failed to link account: %w", err)
			}
		}

		if err := s.accounts.UpdateLastLogin(ctx, user.ID, time.Now()); err != nil {
			return nil, fmt.Errorf("failed to update last login: %w", err)
		}

		jwt, err := s.jwtManager.Generate(user.ID, user.Email)
		if err != nil {
			return nil, errors.New("failed to generate authentication token")
		}

		return &OAuthLoginResponse{Token: jwt}, nil

	default:
		return nil, fmt.Errorf("unsupported provider: %s", p)
	}
}

func splitName(full string) (string, string) {
	full = strings.TrimSpace(full)
	if full == "" {
		return "", ""
	}
	parts := strings.Fields(full)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.Join(parts[1:], " ")
}
