package router

import (
	"github.com/account-service/internal/api/rest/handler"
	"github.com/account-service/internal/api/rest/middleware"
	"github.com/account-service/internal/service"
	"github.com/account-service/pkg/utils"
	"github.com/gorilla/mux"
)

func NewRouter(
	registerService *service.RegisterService,
	loginService *service.LoginService,
	oauthService *service.OAuthService,
	userService *service.UserService,
	jwtManager *utils.JWTManager,
	passwordHelper *utils.PasswordHelper,
) *mux.Router {
	r := mux.NewRouter()

	registerHandler := handler.NewRegisterHandler(registerService)
	loginHandler := handler.NewLoginHandler(loginService)
	oauthHandler := handler.NewOAuthHandler(oauthService)
	userHandler := handler.NewUserHandler(userService, passwordHelper)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	accountsAPI := r.PathPrefix("/api/accounts").Subrouter()

	accountsAPI.HandleFunc("/register", registerHandler.Register).Methods("POST")
	accountsAPI.HandleFunc("/login", loginHandler.Login).Methods("POST")

	accountsAPI.HandleFunc("/oauth/{provider}/verify", oauthHandler.VerifyToken).Methods("POST")

	protected := accountsAPI.PathPrefix("").Subrouter()
	protected.Use(authMiddleware.RequireAuth)

	protected.HandleFunc("/logged-user", userHandler.GetUserInfo).Methods("GET")
	protected.HandleFunc("/logged-user", userHandler.UpdateUserInfo).Methods("PUT")
	protected.HandleFunc("/update-password", userHandler.UpdateUserPassword).Methods("POST")

	return r
}
