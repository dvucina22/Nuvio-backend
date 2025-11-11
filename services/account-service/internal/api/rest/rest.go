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
	oauthService    *service.OAuthService
}

func NewServer(port string, registerService *service.RegisterService, loginService *service.LoginService,
	oauthService *service.OAuthService) *Server {
	return &Server{
		port:            port,
		registerService: registerService,
		loginService:    loginService,
		oauthService:    oauthService,
	}
}

func (s *Server) Run() error {
	r := router.NewRouter(s.registerService, s.loginService, s.oauthService)
	return http.ListenAndServe(fmt.Sprintf(":%s", s.port), r)
}
