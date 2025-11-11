package router

import (
	"github.com/account-service/internal/api/rest/handler"
	"github.com/account-service/internal/service"
	"github.com/gorilla/mux"
)

func NewRouter(
	registerService *service.RegisterService,
	loginService *service.LoginService,
	oauthService *service.OAuthService,
) *mux.Router {
	r := mux.NewRouter()

	registerHandler := handler.NewRegisterHandler(registerService)
	loginHandler := handler.NewLoginHandler(loginService)
	oauthHandler := handler.NewOAuthHandler(oauthService)

	accountsAPI := r.PathPrefix("/api/accounts").Subrouter()

	accountsAPI.HandleFunc("/register", registerHandler.Register).Methods("POST")
	accountsAPI.HandleFunc("/login", loginHandler.Login).Methods("POST")

	accountsAPI.HandleFunc("/oauth/{provider}/verify", oauthHandler.VerifyToken).Methods("POST")
	return r
}
