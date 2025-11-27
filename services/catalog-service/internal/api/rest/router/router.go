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
	cartService *service.CartService,
) *mux.Router {
	r := mux.NewRouter()

	productHandler := handler.NewProductHandler(productService)
	favoritesHanlder := handler.NewFavoritesHandler(favoritesService)
	cartHandler := handler.NewCartHandler(cartService)

	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	catalogAPI := r.PathPrefix("/api/catalog").Subrouter()

	protected := catalogAPI.PathPrefix("").Subrouter()
	protected.Use(authMiddleware.RequireAuth)
	optionalProtected := catalogAPI.PathPrefix("").Subrouter()
	optionalProtected.Use(authMiddleware.OptionalAuth)

	optionalProtected.HandleFunc("/products/filter", productHandler.GetFilteredProducts).Methods("POST")
	optionalProtected.HandleFunc("/products/{id}", productHandler.GetProductByID).Methods("GET")

	protected.HandleFunc("/products/favorite", favoritesHanlder.AddToFavorites).Methods("POST")
	protected.HandleFunc("/products/favorite", favoritesHanlder.RemoveFromFavorites).Methods("DELETE")

	protected.HandleFunc("/products/cart", cartHandler.AddProductToCart).Methods("POST")
	protected.HandleFunc("/products/cart", cartHandler.RemoveProductFromCart).Methods("DELETE")
	protected.HandleFunc("/products/cart", cartHandler.GetCartContents).Methods("GET")
	protected.HandleFunc("/products/cart/empty", cartHandler.EmptyCart).Methods("DELETE")
	return r
}
