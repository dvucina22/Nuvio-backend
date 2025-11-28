package service

import (
	"github.com/catalog-service/internal/repository"
	"github.com/catalog-service/pkg/models/products"
)

type CategoryService struct {
	repo repository.CategoryRepository
}

func NewCategoryService(repo repository.CategoryRepository) *CategoryService {
	return &CategoryService{
		repo: repo,
	}
}

func (s *CategoryService) GetAllCategories() ([]products.Category, error) {
	categories, err := s.repo.GetCategories()

	if err != nil {
		return []products.Category{}, err
	}

	return categories, nil
}
