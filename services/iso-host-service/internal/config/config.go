package config

import (
	"bufio"
	"log"
	"os"
	"strings"
)

type Config struct {
	Addr string
}

func Load() *Config {
	loadEnvFile(".env")

	addr := strings.TrimSpace(getEnv("ISO_HOST_ADDR", ""))
	port := strings.TrimSpace(getEnv("ISO_HOST_PORT", "8005"))

	if addr == "" {
		addr = "0.0.0.0:" + port
	} else if strings.HasPrefix(addr, ":") {
		addr = "0.0.0.0" + addr
	}

	cfg := &Config{Addr: addr}

	if cfg.Addr == "" {
		log.Fatal("ISO_HOST_ADDR/ISO_HOST_PORT not set")
	}

	isoHostAddr := strings.TrimSpace(getEnv("ISO_HOST_ADDR", ""))
	if isoHostAddr == "" {
		log.Fatal("ISO_HOST_ADDR not set")
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
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
