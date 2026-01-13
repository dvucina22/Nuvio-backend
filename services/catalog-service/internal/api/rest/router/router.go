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
	brandService *service.BrandService,
	categoryService *service.CategoryService,
	attributesService *service.AttributesService,
) *mux.Router {
	r := mux.NewRouter()

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

	optionalProtected.HandleFunc("/products", productHandler.CreateProduct).Methods("POST")
	optionalProtected.HandleFunc("/products/filter", productHandler.GetFilteredProducts).Methods("POST")
	optionalProtected.HandleFunc("/products/images", productHandler.GetPrimaryImages).Methods("POST")
	optionalProtected.HandleFunc("/products/{id}", productHandler.GetProductByID).Methods("GET")
	optionalProtected.HandleFunc("/products/{id}", productHandler.UpdateProductByID).Methods("PUT")
	optionalProtected.HandleFunc("/products/{id}", productHandler.DeleteProductByID).Methods("DELETE")

	optionalProtected.HandleFunc("/brands", brandHandler.GetAllBrands).Methods("GET")
	optionalProtected.HandleFunc("/categories", categoryHandler.GetAllCategories).Methods("GET")
	optionalProtected.HandleFunc("/attributes", attributesHandler.GetAttributes).Methods("GET")

	protected.HandleFunc("/products/favorite", favoritesHanlder.AddToFavorites).Methods("POST")
	protected.HandleFunc("/products/favorite", favoritesHanlder.RemoveFromFavorites).Methods("DELETE")

	protected.HandleFunc("/products/cart/{id}", cartHandler.AddProductToCart).Methods("POST")
	protected.HandleFunc("/products/cart/{id}", cartHandler.RemoveProductFromCart).Methods("DELETE")
	protected.HandleFunc("/products/cart", cartHandler.GetCartContents).Methods("GET")
	protected.HandleFunc("/products/cart/empty", cartHandler.EmptyCart).Methods("GET")
	return r
}
