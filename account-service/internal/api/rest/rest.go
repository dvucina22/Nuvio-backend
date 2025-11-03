package rest

import (
	"net/http"

	"github.com/account-service/internal/api/rest/router"
	"github.com/account-service/internal/service"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	port string
	svc  *service.RegisterService
}

func NewServer(port string, svc *service.RegisterService) *Server {
	return &Server{port: port, svc: svc}
}

func (s *Server) Run() error {
	r := chi.NewRouter()
	router.RegisterRoutes(r, s.svc)
	return http.ListenAndServe(":"+s.port, r)
}
