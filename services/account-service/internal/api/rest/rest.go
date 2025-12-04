package rest

import (
	"fmt"
	"net/http"

	"github.com/account-service/internal/api/rest/router"
	"github.com/account-service/internal/service"
	"github.com/account-service/pkg/utils"
)

type Server struct {
	port              string
	registerService   *service.RegisterService
	loginService      *service.LoginService
	oauthService      *service.OAuthService
	userService       *service.UserService
	jwtManager        *utils.JWTManager
	passwordHelper    *utils.PasswordHelper
	cloudinaryService *service.CloudinaryService
	roleService       *service.RoleService
}

func NewServer(port string, registerService *service.RegisterService, loginService *service.LoginService,
	oauthService *service.OAuthService, userService *service.UserService, jwtManager *utils.JWTManager,
	passwordHelper *utils.PasswordHelper, cloudinaryService *service.CloudinaryService, roleService *service.RoleService) *Server {
	return &Server{
		port:              port,
		registerService:   registerService,
		loginService:      loginService,
		oauthService:      oauthService,
		userService:       userService,
		jwtManager:        jwtManager,
		passwordHelper:    passwordHelper,
		cloudinaryService: cloudinaryService,
		roleService:       roleService,
	}
}

func (s *Server) Run() error {
	r := router.NewRouter(s.registerService, s.loginService, s.oauthService, s.userService,
		s.jwtManager, s.passwordHelper, s.cloudinaryService,
		s.roleService)
	return http.ListenAndServe(fmt.Sprintf(":%s", s.port), r)
}
