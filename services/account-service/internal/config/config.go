package config

import (
	"bufio"
	"log"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type OAuthProviderConfig struct {
	Enabled      bool
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

type OAuthConfig struct {
	Google OAuthProviderConfig
}

type Config struct {
	Port        string
	DatabaseDSN string
	JWTSecret   string
	JWTExpiry   int

	OAuth OAuthConfig
}

func Load() *Config {
	loadEnvFile(".env")

	cfg := &Config{
		Port:        getEnv("ACCOUNT_PORT", ""),
		DatabaseDSN: getEnv("DATABASE_DSN", ""),
		JWTSecret:   getEnv("JWT_SECRET", ""),
		JWTExpiry:   60,
		OAuth: OAuthConfig{
			Google: OAuthProviderConfig{
				ClientID:     getEnv("OAUTH_GOOGLE_CLIENT_ID", ""),
				ClientSecret: getEnv("OAUTH_GOOGLE_CLIENT_SECRET", ""),
				RedirectURL:  getEnv("OAUTH_GOOGLE_REDIRECT_URL", ""),
				Scopes:       []string{"openid", "email", "profile"},
			},
		},
	}

	if cfg.OAuth.Google.ClientID != "" && cfg.OAuth.Google.ClientSecret != "" && cfg.OAuth.Google.RedirectURL != "" {
		cfg.OAuth.Google.Enabled = true
	}

	if cfg.Port == "" {
		log.Fatal("ACCOUNT_PORT not set in .env")
	}
	if cfg.DatabaseDSN == "" {
		log.Fatal("DATABASE_DSN not set in .env")
	}
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET not set in .env")
	}

	return cfg
}

func loadEnvFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading .env file: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

type OAuth2Configs struct {
	Google *oauth2.Config
}

func (c *Config) BuildOAuth2Configs() *OAuth2Configs {
	out := &OAuth2Configs{}

	if c.OAuth.Google.Enabled {
		out.Google = &oauth2.Config{
			ClientID:     c.OAuth.Google.ClientID,
			ClientSecret: c.OAuth.Google.ClientSecret,
			Endpoint:     google.Endpoint,
			RedirectURL:  c.OAuth.Google.RedirectURL,
			Scopes:       c.OAuth.Google.Scopes,
		}
	}

	return out
}
