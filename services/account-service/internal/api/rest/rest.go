package rest

import (
	"fmt"
	"net/http"

	"github.com/account-service/internal/api/rest/router"
	"github.com/account-service/internal/service"
)

type Server struct {
	port            string
	registerService *service.RegisterService
	loginService    *service.LoginService
}

func NewServer(port string, registerService *service.RegisterService, loginService *service.LoginService) *Server {
	return &Server{
		port:            port,
		registerService: registerService,
		loginService:    loginService,
	}
}

func (s *Server) Run() error {
	r := router.NewRouter(s.registerService, s.loginService)
	return http.ListenAndServe(fmt.Sprintf(":%s", s.port), r)
}
