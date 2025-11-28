package service

import (
	"github.com/catalog-service/internal/repository"
	"github.com/catalog-service/pkg/models/products"
)

type BrandService struct {
	repo repository.BrandRepository
}

func NewBrandService(repo repository.BrandRepository) *BrandService {
	return &BrandService{
		repo: repo,
	}
}

func (s *BrandService) GetAllBrands() ([]products.Brand, error) {
	brand, err := s.repo.GetBrands()

	if err != nil {
		return []products.Brand{}, err
	}

	return brand, nil
}
