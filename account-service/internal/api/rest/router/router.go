package router

import (
	"github.com/account-service/internal/api/rest/handler"
	"github.com/account-service/internal/service"
	"github.com/gorilla/mux"
)

func NewRouter(registerService *service.RegisterService, loginService *service.LoginService) *mux.Router {
	r := mux.NewRouter()

	registerHandler := handler.NewRegisterHandler(registerService)
	loginHandler := handler.NewLoginHandler(loginService)

	r.HandleFunc("/api/register", registerHandler.Register).Methods("POST")
	r.HandleFunc("/api/login", loginHandler.Login).Methods("POST")

	return r
}
