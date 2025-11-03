package router

import (
	"github.com/account-service/internal/api/rest/handler"
	"github.com/account-service/internal/service"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r *chi.Mux, svc *service.RegisterService) {
	h := handler.NewRegisterHandler(svc)
	r.Post("/api/v1/register", h.Register)
}
