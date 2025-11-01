package handler

import (
	"github.com/account-service/internal/account"

	"github.com/go-chi/chi/v5"
)

func Routes(svc account.Service) chi.Router {
	r := chi.NewRouter()
	r.Post("/register", RegisterHandler(svc))
	r.Post("/login", LoginHandler(svc))
	return r
}
