package rest

import (
	"fmt"
	"net/http"

	"github.com/catalog-service/internal/api/rest/router"
	"github.com/catalog-service/pkg/utils"
)

type Server struct {
	port       string
	jwtManager *utils.JWTManager
}

func NewServer(port string, jwtManager *utils.JWTManager) *Server {
	return &Server{
		port:       port,
		jwtManager: jwtManager,
	}
}

func (s *Server) Run() error {
	r := router.NewRouter(s.jwtManager)
	return http.ListenAndServe(fmt.Sprintf(":%s", s.port), r)
}
