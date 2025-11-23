package app

import (
	"log"
	"time"

	"github.com/catalog-service/internal/api/rest"
	"github.com/catalog-service/internal/config"
	"github.com/catalog-service/pkg/utils"
)

func Run() {
	cfg := config.Load()

	jwtManager := utils.NewJWTManager(cfg.JWTSecret, time.Duration(cfg.JWTExpiry)*time.Minute)

	server := rest.NewServer(cfg.Port, jwtManager)

	log.Printf("Catalog Service running on port %s", cfg.Port)
	log.Fatal(server.Run())
}
