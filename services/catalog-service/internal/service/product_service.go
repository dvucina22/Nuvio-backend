package service

import (
	"context"

	"github.com/catalog-service/internal/repository"
	"github.com/catalog-service/pkg/models"
)

type ProductService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) *ProductService {
	return &ProductService{
		repo: repo,
	}
}

func (s *ProductService) GetFilteredProducts(ctx context.Context, filter *models.ProductFilter) ([]models.Product, error) {
	if filter == nil {
		return nil, models.ErrInvalidFilter
	}

	if filter.PriceMin != nil && filter.PriceMax != nil && *filter.PriceMin > *filter.PriceMax {
		return nil, models.ErrInvalidFilter
	}

	if filter.Limit == 0 {
		filter.Limit = 20
	}

	products, err := s.repo.GetFilteredProducts(filter)
	if err != nil {
		return nil, err
	}

	if len(products) == 0 {
		return nil, models.ErrProductNotFound
	}

	return products, nil
}
