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
	favoritesService *service.FavoritesService,
) *mux.Router {
	r := mux.NewRouter()

	productHandler := handler.NewProductHandler(productService)
	favoritesHanlder := handler.NewFavoritesHandler(favoritesService)

	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	catalogAPI := r.PathPrefix("/api/catalog").Subrouter()

	protected := catalogAPI.PathPrefix("").Subrouter()
	protected.Use(authMiddleware.RequireAuth)

	protected.HandleFunc("/products/filter", productHandler.GetFilteredProducts).Methods("POST")

	protected.HandleFunc("/products/favorite", favoritesHanlder.AddToFavorites).Methods("POST")
	protected.HandleFunc("/products/favorite", favoritesHanlder.RemoveFromFavorites).Methods("DELETE")
	return r
}
