package app

import (
	"log"
	"time"

	"github.com/account-service/internal/api/rest"
	"github.com/account-service/internal/config"
	postgres "github.com/account-service/internal/db"
	"github.com/account-service/internal/repository"
	"github.com/account-service/internal/service"
	"github.com/account-service/pkg/utils"
)

func Run() {
	cfg := config.Load()

	db := postgres.ConnectPostgres(cfg.DatabaseDSN)

	accountRepo := repository.NewAccountRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	oauthRepo := repository.NewOAuthRepo(db)
	userRepo := repository.NewUserRepo(db)

	jwtManager := utils.NewJWTManager(cfg.JWTSecret, time.Duration(cfg.JWTExpiry)*time.Minute)

	oauth2Configs := cfg.BuildOAuth2Configs()

	registerService := service.NewRegisterService(accountRepo, roleRepo)
	loginService := service.NewLoginService(accountRepo, jwtManager)
	oauthService := service.NewOAuthService(accountRepo, oauthRepo, jwtManager, oauth2Configs)
	userService := service.NewUserService(userRepo)

	server := rest.NewServer(cfg.Port, registerService, loginService, oauthService, userService, jwtManager)

	log.Printf("Account Service running on port %s", cfg.Port)
	log.Fatal(server.Run())
}
