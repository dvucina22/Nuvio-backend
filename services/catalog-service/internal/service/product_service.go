package service

import (
	"context"

	"github.com/catalog-service/internal/api/rest/middleware"
	"github.com/catalog-service/internal/repository"
	"github.com/catalog-service/pkg/models"
	"github.com/catalog-service/pkg/models/products"
	"github.com/google/uuid"
)

type ProductService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) *ProductService {
	return &ProductService{
		repo: repo,
	}
}

func (s *ProductService) GetFilteredProducts(ctx context.Context, filter *products.ProductFilter) ([]products.ProductMinimal, error) {
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

	return products, nil
}

func (s *ProductService) GetProductByID(ctx context.Context, productID int) (*products.Product, error) {
	if productID <= 0 {
		return nil, models.ErrInvalidProductID
	}

	var userId *uuid.UUID

	claims := middleware.GetUserClaims(ctx)
	if claims != nil && claims.UserID != "" {
		parsed, err := uuid.Parse(claims.UserID)
		if err == nil {
			userId = &parsed
		}
	}

	product, err := s.repo.GetProductByID(userId, productID)
	if err != nil {
		return nil, err
	}

	if product == nil {
		return nil, models.ErrProductNotFound
	}

	return product, nil
}

func (s *ProductService) DeleteProductByID(ctx context.Context, productID int) error {
	if productID <= 0 {
		return models.ErrInvalidProductID
	}

	exists, err := s.repo.ExistsByID(ctx, productID)
	if err != nil {
		return err
	}

	if !exists {
		return models.ErrInvalidProductID
	}

	return s.repo.DeleteProductByID(productID)
}

func (s *ProductService) UpdateProductByID(ctx context.Context, productID int, product *products.UpdateProduct) error {
	if productID <= 0 {
		return models.ErrInvalidProductID
	}
	exists, err := s.repo.ExistsByID(ctx, productID)
	if err != nil {
		return err
	}
	if !exists {
		return models.ErrInvalidProductID
	}
	if product == nil {
		return models.ErrInvalidData
	}

	return s.repo.UpdateProductByID(productID, product)
}

func (s *ProductService) CreateProduct(product *products.CreateProduct) error {
	if product == nil {
		return models.ErrInvalidData
	}

	return s.repo.CreateProduct(product)
}

func (s *ProductService) GetPrimaryImages(ctx context.Context, productIds []int64) ([]products.ProductImageResponse, error) {
	if len(productIds) == 0 {
		return []products.ProductImageResponse{}, nil
	}
	
	if len(productIds) > 100 {
		return nil, models.ErrInvalidFilter
	}
	
	imagesMap, err := s.repo.GetPrimaryImages(ctx, productIds)
	if err != nil {
		return nil, err
	}
	
	result := make([]products.ProductImageResponse, 0, len(productIds))
	
	for _, productID := range productIds {
		imageURL, exists := imagesMap[productID]
		if exists {
			result = append(result, products.ProductImageResponse{
				ProductID: productID,
				ImageURL:  imageURL,
			})
		}
	}
	
	return result, nil
}