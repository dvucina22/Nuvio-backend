package app

import (
	"log"
	"time"

	"github.com/iso-com-service/internal/api/rest"
	"github.com/iso-com-service/internal/api/rest/handler"
	"github.com/iso-com-service/internal/client/isotcp"
	"github.com/iso-com-service/internal/config"
	postgres "github.com/iso-com-service/internal/db"
	"github.com/iso-com-service/internal/repository"
	"github.com/iso-com-service/internal/service"
	"github.com/iso-com-service/pkg/utils"
)

func Run() {
	cfg := config.Load()

	db := postgres.ConnectPostgres(cfg.DatabaseDSN)

	terminalCredRepo := repository.NewTerminalCredentialRepository(db)
	linkStateRepo := repository.NewH2HLinkStateRepository(db)

	tcpClient := isotcp.New(cfg.IsoHostAddr)

	isoService := service.NewISOService(terminalCredRepo, linkStateRepo, tcpClient)
	terminalService := service.NewTerminalService(terminalCredRepo)

	jwtManager := utils.NewJWTManager(cfg.JWTSecret, time.Duration(cfg.JWTExpiry)*time.Minute)

	authHandler := handler.NewAuthorizeHandler(isoService)
	terminalHandler := handler.NewTerminalHandler(terminalService)

	server := rest.NewServer(cfg.Port, jwtManager, authHandler, terminalHandler)

	log.Printf("Iso Communication Service running on port %s", cfg.Port)
	log.Fatal(server.Run())
}
