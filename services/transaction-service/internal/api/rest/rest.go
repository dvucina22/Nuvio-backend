package rest

import (
	"fmt"
	"net/http"

	"github.com/transaction-service/internal/api/rest/router"
	"github.com/transaction-service/internal/service"
	"github.com/transaction-service/pkg/utils"
)

type Server struct {
	port        string
	cardService *service.CardService
	jwtManager  *utils.JWTManager
}

func NewServer(port string, cardService *service.CardService, jwtManager *utils.JWTManager) *Server {
	return &Server{
		port:        port,
		cardService: cardService,
		jwtManager:  jwtManager,
	}
}

func (s *Server) Run() error {
	r := router.NewRouter(s.jwtManager, s.cardService)
	return http.ListenAndServe(fmt.Sprintf(":%s", s.port), r)
}
