package rest

import (
	"fmt"
	"net/http"

	"github.com/account-service/internal/api/rest/router"
	"github.com/account-service/internal/service"
	"github.com/account-service/pkg/utils"
)

type Server struct {
	port            string
	registerService *service.RegisterService
	loginService    *service.LoginService
	oauthService    *service.OAuthService
	userService     *service.UserService
	jwtManager      *utils.JWTManager
}

func NewServer(port string, registerService *service.RegisterService, loginService *service.LoginService,
	oauthService *service.OAuthService, userService *service.UserService, jwtManager *utils.JWTManager) *Server {
	return &Server{
		port:            port,
		registerService: registerService,
		loginService:    loginService,
		oauthService:    oauthService,
		userService:     userService,
		jwtManager:      jwtManager,
	}
}

func (s *Server) Run() error {
	r := router.NewRouter(s.registerService, s.loginService, s.oauthService, s.userService, s.jwtManager)
	return http.ListenAndServe(fmt.Sprintf(":%s", s.port), r)
}
