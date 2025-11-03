package app

import (
	"log"

	"github.com/account-service/internal/api/rest"
	"github.com/account-service/internal/config"
	"github.com/account-service/internal/repository"
	"github.com/account-service/internal/service"
)

func Run() {
	cfg := config.Load()

	db := repository.ConnectPostgres(cfg.DatabaseDSN)
	repo := repository.NewAccountRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	registerService := service.NewRegisterService(repo, roleRepo)

	server := rest.NewServer(cfg.Port, registerService)
	log.Fatal(server.Run())
}
