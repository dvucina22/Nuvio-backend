package rest

import (
	"fmt"
	"net/http"

	"github.com/transaction-service/internal/api/rest/router"
	"github.com/transaction-service/internal/config"
	"github.com/transaction-service/internal/service"
	"github.com/transaction-service/pkg/utils"
)

type Server struct {
	port               string
	cardService        *service.CardService
	jwtManager         *utils.JWTManager
	transactionService *service.TransactionService
	cfg                *config.Config
}

func NewServer(port string, cardService *service.CardService, jwtManager *utils.JWTManager, transactionService *service.TransactionService, cfg *config.Config) *Server {
	return &Server{
		port:               port,
		cardService:        cardService,
		jwtManager:         jwtManager,
		transactionService: transactionService,
		cfg:                cfg,
	}
}

func (s *Server) Run() error {
	r := router.NewRouter(s.jwtManager, s.cardService, s.transactionService, s.cfg)
	return http.ListenAndServe(fmt.Sprintf(":%s", s.port), r)
}
