package app

import (
	"log"
	"time"

	"github.com/account-service/internal/api/rest"
	"github.com/account-service/internal/config"
	"github.com/account-service/internal/repository"
	"github.com/account-service/internal/service"
	"github.com/account-service/pkg/utils"
)

func Run() {
	cfg := config.Load()

	db := repository.ConnectPostgres(cfg.DatabaseDSN)

	accountRepo := repository.NewAccountRepository(db)
	roleRepo := repository.NewRoleRepository(db)

	jwtManager := utils.NewJWTManager(cfg.JWTSecret, time.Duration(cfg.JWTExpiry)*time.Minute)

	registerService := service.NewRegisterService(accountRepo, roleRepo)
	loginService := service.NewLoginService(accountRepo, jwtManager)

	server := rest.NewServer(cfg.Port, registerService, loginService)

	log.Printf("Account Service running on port %s", cfg.Port)
	log.Fatal(server.Run())
}
