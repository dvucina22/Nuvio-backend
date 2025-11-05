package config

import (
	"log"
	"os"
)

type Config struct {
	Port        string
	DatabaseDSN string
	JWTSecret   string
	JWTExpiry   int
}

func Load() *Config {
	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseDSN: getEnv("DATABASE_DSN", "postgres://postgres:ngL420Idk@localhost:5432/nuvio-local?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "81c6ff699c3d6c1321607dcd20b4d329dfb6ae33d967e97fa7e6244be9587b1f"),
		JWTExpiry:   60,
	}
	if cfg.DatabaseDSN == "" {
		log.Fatal("DATABASE_DSN not set")
	}
	return cfg
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
