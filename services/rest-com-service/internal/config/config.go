package config

import (
	"bufio"
	"log"
	"os"
	"strings"
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

	RestHostAddr string
}

func Load() *Config {
	loadEnvFile(".env")

	cfg := &Config{
		Port:         getEnv("REST_COM_PORT", ""),
		DatabaseDSN:  getEnv("DATABASE_DSN", ""),
		JWTSecret:    getEnv("JWT_SECRET", ""),
		JWTExpiry:    604800,
		RestHostAddr: getEnv("REST_HOST_ADDR", ""),
	}

	if cfg.Port == "" {
		log.Fatal("REST_COM_PORT not set in .env")
	}
	if cfg.DatabaseDSN == "" {
		log.Fatal("DATABASE_DSN not set in .env")
	}
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET not set in .env")
	}

	if cfg.RestHostAddr == "" {
		log.Fatal("REST_HOST_ADDR not set in .env")
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
