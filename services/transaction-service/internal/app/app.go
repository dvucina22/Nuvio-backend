package app

import (
	"log"
	"time"

	"github.com/transaction-service/internal/api/rest"
	"github.com/transaction-service/internal/config"
	postgres "github.com/transaction-service/internal/db"
	"github.com/transaction-service/internal/repository"
	"github.com/transaction-service/internal/service"
	"github.com/transaction-service/pkg/utils"
)

func Run() {
	cfg := config.Load()

	db := postgres.ConnectPostgres(cfg.DatabaseDSN)

	cardRepo := repository.NewBankCardRepository(db)

	jwtManager := utils.NewJWTManager(cfg.JWTSecret, time.Duration(cfg.JWTExpiry)*time.Minute)

	cardService := service.NewCardService(cardRepo, cfg.EncryptionKey)

	server := rest.NewServer(cfg.Port, cardService, jwtManager)

	log.Printf("Transaction Service running on port %s", cfg.Port)
	log.Fatal(server.Run())
}
