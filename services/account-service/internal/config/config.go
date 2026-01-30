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
	Port           string
	DatabaseDSN    string
	JWTSecret      string
	JWTExpiry      int
	AllowedOrigins []string

	OAuth OAuthConfig
}

func Load() *Config {
	if _, err := os.Stat(".env"); err == nil {
		log.Println("Loading .env file...")
		loadEnvFile(".env")
	} else {
		log.Println("No .env file found, using environment variables from docker-compose")
	}

	cfg := &Config{
		Port:           getEnv("ACCOUNT_PORT", ""),
		DatabaseDSN:    getEnv("DATABASE_DSN", ""),
		JWTSecret:      getEnv("JWT_SECRET", ""),
		JWTExpiry:      604800,
		AllowedOrigins: loadAllowedOrigins(),
		OAuth: OAuthConfig{
			Google: OAuthProviderConfig{
				ClientID:     getEnv("OAUTH_GOOGLE_CLIENT_ID", ""),
				ClientSecret: getEnv("OAUTH_GOOGLE_CLIENT_SECRET", ""),
				RedirectURL:  getEnv("OAUTH_GOOGLE_REDIRECT_URL", ""),
				Scopes:       []string{"openid", "email", "profile"},
			},
		},
	}

	log.Println("=== OAUTH CONFIG DEBUG ===")
	log.Printf("Google ClientID exists: %v (length: %d)", cfg.OAuth.Google.ClientID != "", len(cfg.OAuth.Google.ClientID))
	log.Printf("Google ClientSecret exists: %v (length: %d)", cfg.OAuth.Google.ClientSecret != "", len(cfg.OAuth.Google.ClientSecret))
	log.Printf("Google RedirectURL: %s", cfg.OAuth.Google.RedirectURL)

	if cfg.OAuth.Google.ClientID != "" {
		log.Printf("Google ClientID (first 20 chars): %s...", cfg.OAuth.Google.ClientID[:min(20, len(cfg.OAuth.Google.ClientID))])
	}
	if cfg.OAuth.Google.ClientSecret != "" {
		log.Printf("Google ClientSecret (first 10 chars): %s...", cfg.OAuth.Google.ClientSecret[:min(10, len(cfg.OAuth.Google.ClientSecret))])
	}

	if cfg.OAuth.Google.ClientID != "" && cfg.OAuth.Google.ClientSecret != "" && cfg.OAuth.Google.RedirectURL != "" {
		cfg.OAuth.Google.Enabled = true
		log.Println("Google OAuth is ENABLED")
	} else {
		log.Println("Google OAuth is DISABLED - missing credentials:")
		log.Printf("  - ClientID present: %v", cfg.OAuth.Google.ClientID != "")
		log.Printf("  - ClientSecret present: %v", cfg.OAuth.Google.ClientSecret != "")
		log.Printf("  - RedirectURL present: %v", cfg.OAuth.Google.RedirectURL != "")
	}
	log.Println("==========================")

	if cfg.Port == "" {
		log.Fatal("FATAL: ACCOUNT_PORT not set in environment")
	}
	if cfg.DatabaseDSN == "" {
		log.Fatal("FATAL: DATABASE_DSN not set in environment")
	}
	if cfg.JWTSecret == "" {
		log.Fatal("FATAL: JWT_SECRET not set in environment")
	}

	log.Printf("Config loaded successfully - Port: %s", cfg.Port)
	return cfg
}

func loadEnvFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Printf("Could not open %s: %v", filename, err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	loadedVars := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			log.Printf("Warning: Invalid line %d in %s: %s", lineNum, filename, line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)

		if _, exists := os.LookupEnv(key); !exists {
			if err := os.Setenv(key, value); err != nil {
				log.Printf("Error setting env var %s: %v", key, err)
			} else {
				loadedVars++
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading %s: %v", filename, err)
	} else {
		log.Printf("Loaded %d variables from %s", loadedVars, filename)
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

	log.Println("=== BUILDING OAUTH2 CONFIGS ===")

	if c.OAuth.Google.Enabled {
		out.Google = &oauth2.Config{
			ClientID:     c.OAuth.Google.ClientID,
			ClientSecret: c.OAuth.Google.ClientSecret,
			Endpoint:     google.Endpoint,
			RedirectURL:  c.OAuth.Google.RedirectURL,
			Scopes:       c.OAuth.Google.Scopes,
		}
		log.Println("✓ Google OAuth2 config created successfully")
		log.Printf("  - ClientID: %s...", out.Google.ClientID[:min(20, len(out.Google.ClientID))])
		log.Printf("  - Endpoint: %s", out.Google.Endpoint.AuthURL)
		log.Printf("  - RedirectURL: %s", out.Google.RedirectURL)
		log.Printf("  - Scopes: %v", out.Google.Scopes)
	} else {
		log.Println("✗ Google OAuth2 config NOT created (disabled)")
	}

	log.Println("===============================")

	if out.Google == nil {
		log.Fatal("FATAL: Google OAuth2 config is nil! Check your OAUTH_GOOGLE_* environment variables in .env file")
	}

	return out
}

func loadAllowedOrigins() []string {
	originsStr := getEnv("ALLOWED_ORIGINS", "")
	if originsStr == "" {
		log.Println("Warning: ALLOWED_ORIGINS not set, using empty array")
		return []string{}
	}

	origins := strings.Split(originsStr, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	log.Printf("Allowed origins: %v", origins)
	return origins
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
