package rest

import (
	"fmt"
	"net/http"

	"github.com/catalog-service/internal/api/rest/router"
	"github.com/catalog-service/internal/service"
	"github.com/catalog-service/pkg/utils"
)

type Server struct {
	port             string
	jwtManager       *utils.JWTManager
	productService   *service.ProductService
	favoritesService *service.FavoritesService
}

func NewServer(port string, jwtManager *utils.JWTManager, productService *service.ProductService, favoritesService *service.FavoritesService) *Server {
	return &Server{
		port:             port,
		jwtManager:       jwtManager,
		productService:   productService,
		favoritesService: favoritesService,
	}
}

func (s *Server) Run() error {
	r := router.NewRouter(s.jwtManager, s.productService, s.favoritesService)
	return http.ListenAndServe(fmt.Sprintf(":%s", s.port), r)
}
