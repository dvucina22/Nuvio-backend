package config

import "os"

type Config struct {
	Port        string
	DatabaseDSN string
	JWTSecret   string
}

func Load() *Config {
	return &Config{
		Port:        getenv("PORT", "8080"),
		DatabaseDSN: getenv("DATABASE_DSN", "postgres://user:pass@localhost:5432/accountdb?sslmode=disable"),
		JWTSecret:   getenv("JWT_SECRET", "change-me"),
	}
}

func getenv(k, def string) string {
	if v, ok := os.LookupEnv(k); ok {
		return v
	}
	return def
}
