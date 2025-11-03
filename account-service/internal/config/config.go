package config

import "os"

type Config struct {
	Port        string
	DatabaseDSN string
}

func Load() *Config {
	return &Config{
		Port:        getenv("PORT", "8080"),
		DatabaseDSN: getenv("DATABASE_DSN", "postgres://postgres:ngL420Idk@localhost:5432/nuvio-local?sslmode=disable"),
	}
}

func getenv(k, def string) string {
	if v, ok := os.LookupEnv(k); ok {
		return v
	}
	return def
}
