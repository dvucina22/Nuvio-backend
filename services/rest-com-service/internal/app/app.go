package app

import (
	"log"
	"time"

	"github.com/rest-com-service/internal/api/rest"
	"github.com/rest-com-service/internal/api/rest/handler"
	"github.com/rest-com-service/internal/client/resttcp"
	"github.com/rest-com-service/internal/config"
	postgres "github.com/rest-com-service/internal/db"
	"github.com/rest-com-service/internal/repository"
	"github.com/rest-com-service/internal/service"
	"github.com/rest-com-service/pkg/utils"
)

func Run() {
	cfg := config.Load()

	db := postgres.ConnectPostgres(cfg.DatabaseDSN)

	terminalCredRepo := repository.NewTerminalCredentialRepository(db)
	linkStateRepo := repository.NewH2HLinkStateRepository(db)

	tcpClient := resttcp.New(cfg.RestHostAddr)

	restService := service.NewRESTService(terminalCredRepo, linkStateRepo, tcpClient)
	terminalService := service.NewTerminalService(terminalCredRepo)

	jwtManager := utils.NewJWTManager(cfg.JWTSecret, time.Duration(cfg.JWTExpiry)*time.Minute)

	authHandler := handler.NewAuthorizeHandler(restService)
	terminalHandler := handler.NewTerminalHandler(terminalService)

	server := rest.NewServer(cfg.Port, jwtManager, authHandler, terminalHandler)

	log.Printf("Rest Communication Service running on port %s", cfg.Port)
	log.Fatal(server.Run())
}
