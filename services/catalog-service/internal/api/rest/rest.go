package rest

import (
	"fmt"
	"net/http"

	"github.com/catalog-service/internal/api/rest/router"
	"github.com/catalog-service/internal/service"
	"github.com/catalog-service/pkg/utils"
)

type Server struct {
	port              string
	jwtManager        *utils.JWTManager
	productService    *service.ProductService
	favoritesService  *service.FavoritesService
	cartService       *service.CartService
	brandService      *service.BrandService
	categoryService   *service.CategoryService
	attributesService *service.AttributesService
}

func NewServer(port string, jwtManager *utils.JWTManager, productService *service.ProductService, favoritesService *service.FavoritesService,
	cartService *service.CartService, brandService *service.BrandService, categoryService *service.CategoryService,
	attributesService *service.AttributesService) *Server {
	return &Server{
		port:              port,
		jwtManager:        jwtManager,
		productService:    productService,
		favoritesService:  favoritesService,
		cartService:       cartService,
		brandService:      brandService,
		categoryService:   categoryService,
		attributesService: attributesService,
	}
}

func (s *Server) Run() error {
	r := router.NewRouter(s.jwtManager, s.productService, s.favoritesService, s.cartService, s.brandService, s.categoryService, s.attributesService)
	return http.ListenAndServe(fmt.Sprintf(":%s", s.port), r)
}
