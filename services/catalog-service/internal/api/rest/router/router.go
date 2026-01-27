package router

import (
	"github.com/catalog-service/internal/api/rest/handler"
	"github.com/catalog-service/internal/api/rest/middleware"
	"github.com/catalog-service/internal/config"
	"github.com/catalog-service/internal/service"
	"github.com/catalog-service/pkg/utils"
	"github.com/gorilla/mux"
)

func NewRouter(
	jwtManager *utils.JWTManager,
	productService *service.ProductService,
	favoritesService *service.FavoritesService,
	cartService *service.CartService,
	brandService *service.BrandService,
	categoryService *service.CategoryService,
	attributesService *service.AttributesService,
	cfg *config.Config,
) *mux.Router {
	r := mux.NewRouter()

	corsMiddleware := middleware.NewCORSMiddleware(cfg.AllowedOrigins)
	r.Use(corsMiddleware.Handle)

	productHandler := handler.NewProductHandler(productService)
	favoritesHanlder := handler.NewFavoritesHandler(favoritesService)
	cartHandler := handler.NewCartHandler(cartService)
	brandHandler := handler.NewBrandHandler(brandService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	attributesHandler := handler.NewAttributesHandler(attributesService)

	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	catalogAPI := r.PathPrefix("/api/catalog").Subrouter()

	protected := catalogAPI.PathPrefix("").Subrouter()
	protected.Use(authMiddleware.RequireAuth)
	optionalProtected := catalogAPI.PathPrefix("").Subrouter()
	optionalProtected.Use(authMiddleware.OptionalAuth)

	optionalProtected.HandleFunc("/products", productHandler.CreateProduct).Methods("POST", "OPTIONS")
	optionalProtected.HandleFunc("/products/filter", productHandler.GetFilteredProducts).Methods("POST", "OPTIONS")
	optionalProtected.HandleFunc("/products/images", productHandler.GetPrimaryImages).Methods("POST", "OPTIONS")
	optionalProtected.HandleFunc("/products/{id}", productHandler.GetProductByID).Methods("GET", "OPTIONS")
	optionalProtected.HandleFunc("/products/{id}", productHandler.UpdateProductByID).Methods("PUT", "OPTIONS")
	optionalProtected.HandleFunc("/products/{id}", productHandler.DeleteProductByID).Methods("DELETE", "OPTIONS")

	optionalProtected.HandleFunc("/brands", brandHandler.GetAllBrands).Methods("GET", "OPTIONS")
	optionalProtected.HandleFunc("/categories", categoryHandler.GetAllCategories).Methods("GET", "OPTIONS")
	optionalProtected.HandleFunc("/attributes", attributesHandler.GetAttributes).Methods("GET", "OPTIONS")

	protected.HandleFunc("/products/favorite", favoritesHanlder.AddToFavorites).Methods("POST", "OPTIONS")
	protected.HandleFunc("/products/favorite", favoritesHanlder.RemoveFromFavorites).Methods("DELETE", "OPTIONS")

	protected.HandleFunc("/products/cart/{id}", cartHandler.AddProductToCart).Methods("POST", "OPTIONS")
	protected.HandleFunc("/products/cart/{id}", cartHandler.RemoveProductFromCart).Methods("DELETE", "OPTIONS")
	protected.HandleFunc("/products/cart", cartHandler.GetCartContents).Methods("GET", "OPTIONS")
	protected.HandleFunc("/products/cart/empty", cartHandler.EmptyCart).Methods("GET", "OPTIONS")
	return r
}
