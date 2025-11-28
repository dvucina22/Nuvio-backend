package service

import (
	"context"

	"github.com/catalog-service/internal/repository"
	"github.com/catalog-service/pkg/models"
	"github.com/catalog-service/pkg/models/cart"
)

type CartService struct {
	repo        repository.CartRepository
	productRepo repository.ProductRepository
}

func NewCartService(repo repository.CartRepository, productRepo repository.ProductRepository) *CartService {
	return &CartService{
		repo:        repo,
		productRepo: productRepo,
	}
}

func (s *CartService) AddProductToCart(ctx context.Context, userID string, productID int) error {
	if productID <= 0 {
		return models.ErrInvalidProductID
	}

	exists, err := s.productRepo.ExistsByID(ctx, productID)
	if err != nil {
		return err
	}

	if !exists {
		return models.ErrInvalidProductID
	}

	return s.repo.AddProductToCart(userID, productID)
}

func (s *CartService) RemoveProductFromCart(ctx context.Context, userID string, productID int) error {
	if productID <= 0 {
		return models.ErrInvalidProductID
	}

	exists, err := s.productRepo.ExistsByID(ctx, productID)
	if err != nil {
		return err
	}

	if !exists {
		return models.ErrInvalidProductID
	}

	cartContents, err := s.repo.GetCartContents(userID)
	if err != nil {
		return err
	}

	if _, ok := cartContents[productID]; !ok {
		return models.ErrProductNotInCart
	}

	return s.repo.RemoveProductFromCart(userID, productID)
}

func (s *CartService) GetCartContents(ctx context.Context, userID string) ([]cart.CartProduct, error) {
	cartExists, err := s.repo.CartExists(userID)
	if err != nil {
		return nil, err
	}

	if !cartExists {
		return nil, nil
	}

	return s.repo.GetCartProducts(userID)
}

func (s *CartService) EmptyCart(ctx context.Context, userID string) error {
	cartExists, err := s.repo.CartExists(userID)

	if err != nil {
		return err
	}

	if !cartExists {
		return models.ErrCartNotFound
	}

	err = s.repo.EmptyCart(userID)
	if err != nil {
		return err
	}

	return nil
}
