package service

import (
	"context"
	"errors"

	"github.com/catalog-service/internal/repository"
)

var (
	ErrInvalidProductID = errors.New("invalid product ID")
	ErrAlreadyFavorited = errors.New("product already in favorites")
	ErrNotFavorited     = errors.New("product not in favorites")
)

type FavoritesService struct {
	repo        repository.FavoritesRepository
	productRepo repository.ProductRepository
}

func NewFavoritesService(repo repository.FavoritesRepository, productRepo repository.ProductRepository) *FavoritesService {
	return &FavoritesService{
		repo:        repo,
		productRepo: productRepo,
	}
}

func (s *FavoritesService) AddToFavorites(ctx context.Context, userID string, productID int) error {
	if productID <= 0 {
		return ErrInvalidProductID
	}

	productExists, err := s.productRepo.ExistsByID(ctx, productID)
	if err != nil {
		return err
	}

	if !productExists {
		return ErrInvalidProductID
	}

	exists, err := s.repo.IsFavorited(userID, productID)
	if err != nil {
		return err
	}
	if exists {
		return ErrAlreadyFavorited
	}

	return s.repo.AddToFavorites(userID, productID)
}

func (s *FavoritesService) RemoveFromFavorites(ctx context.Context, userID string, productID int) error {
	if productID <= 0 {
		return ErrInvalidProductID
	}

	productExists, err := s.productRepo.ExistsByID(ctx, productID)
	if err != nil {
		return err
	}

	if !productExists {
		return ErrInvalidProductID
	}

	exists, err := s.repo.IsFavorited(userID, productID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotFavorited
	}

	return s.repo.RemoveFromFavorites(userID, productID)
}
