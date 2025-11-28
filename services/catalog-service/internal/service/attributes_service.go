package service

import (
	"github.com/catalog-service/internal/repository"
	"github.com/catalog-service/pkg/models/products"
)

type AttributesService struct {
	attributesRepo repository.AttributesRepository
}

func NewAttributesService(attributesRepo repository.AttributesRepository) *AttributesService {
	return &AttributesService{attributesRepo: attributesRepo}
}

func (s *AttributesService) GetAttributes() ([]products.AttributeValues, error) {
	return s.attributesRepo.GetAttributes()
}
