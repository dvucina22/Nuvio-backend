package config

import (
	"bufio"
	"encoding/base64"
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
	Port          string
	DatabaseDSN   string
	JWTSecret     string
	JWTExpiry     int
	EncryptionKey []byte
}

func Load() *Config {
	loadEnvFile(".env")

	cfg := &Config{
		Port:          getEnv("TRANSACTION_PORT", ""),
		DatabaseDSN:   getEnv("DATABASE_DSN", ""),
		JWTSecret:     getEnv("JWT_SECRET", ""),
		JWTExpiry:     604800,
		EncryptionKey: loadEncryptionKey(),
	}

	if cfg.Port == "" {
		log.Fatal("TRANSACTION_PORT not set in .env")
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

func loadEncryptionKey() []byte {
	keyB64 := getEnv("ENCRYPTION_KEY", "")
	rawKey, err := base64.StdEncoding.DecodeString(keyB64)
	if err != nil {
		log.Fatal("Invalid ENCRYPTION_KEY (must be base64 encoded 32 bytes)")
	}
	if len(rawKey) != 32 {
		log.Fatal("ENCRYPTION_KEY must decode to exactly 32 bytes")
	}

	return rawKey
}
