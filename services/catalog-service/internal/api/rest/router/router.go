package router

import (
	"github.com/catalog-service/internal/api/rest/handler"
	"github.com/catalog-service/internal/api/rest/middleware"
	"github.com/catalog-service/internal/service"
	"github.com/catalog-service/pkg/utils"
	"github.com/gorilla/mux"
)

func NewRouter(
	jwtManager *utils.JWTManager,
	productService *service.ProductService,

) *mux.Router {
	r := mux.NewRouter()

	productHandler := handler.NewProductHandler(productService)

	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	catalogAPI := r.PathPrefix("/api/catalog").Subrouter()

	protected := catalogAPI.PathPrefix("").Subrouter()
	protected.Use(authMiddleware.RequireAuth)

	catalogAPI.HandleFunc("/products/filter", productHandler.GetFilteredProducts).Methods("POST")
	return r
}
