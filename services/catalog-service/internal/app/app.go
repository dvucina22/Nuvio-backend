package app

import (
	"log"
	"time"

	"github.com/catalog-service/internal/api/rest"
	"github.com/catalog-service/internal/config"
	postgres "github.com/catalog-service/internal/db"
	"github.com/catalog-service/internal/repository"
	"github.com/catalog-service/internal/service"
	"github.com/catalog-service/pkg/utils"
)

func Run() {
	cfg := config.Load()

	db := postgres.ConnectPostgres(cfg.DatabaseDSN)

	productRepo := repository.NewProductRepository(db)
	favoritesRepo := repository.NewFavoritesRepository(db)
	cartRepo := repository.NewCartRepository(db)

	productService := service.NewProductService(productRepo)
	favoritesService := service.NewFavoritesService(favoritesRepo, productRepo)
	cartService := service.NewCartService(cartRepo, productRepo)

	jwtManager := utils.NewJWTManager(cfg.JWTSecret, time.Duration(cfg.JWTExpiry)*time.Minute)

	server := rest.NewServer(cfg.Port, jwtManager, productService, favoritesService, cartService)

	log.Printf("Catalog Service running on port %s", cfg.Port)
	log.Fatal(server.Run())
}
