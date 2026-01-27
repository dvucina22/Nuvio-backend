package rest

import (
	"fmt"
	"net/http"

	"github.com/rest-com-service/internal/api/rest/handler"
	"github.com/rest-com-service/internal/api/rest/router"
	"github.com/rest-com-service/pkg/utils"
)

type Server struct {
	port            string
	jwtManager      *utils.JWTManager
	authHandler     *handler.AuthorizeHandler
	terminalHandler *handler.TerminalHandler
}

func NewServer(port string, jwtManager *utils.JWTManager, authHandler *handler.AuthorizeHandler, terminalHandler *handler.TerminalHandler) *Server {
	return &Server{
		port:            port,
		jwtManager:      jwtManager,
		authHandler:     authHandler,
		terminalHandler: terminalHandler,
	}
}

func (s *Server) Run() error {
	r := router.NewRouter(s.jwtManager, s.authHandler, s.terminalHandler)
	return http.ListenAndServe(fmt.Sprintf(":%s", s.port), r)
}
