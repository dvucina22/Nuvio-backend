package handler

import (
	"net/http"

	"github.com/catalog-service/internal/service"

	"github.com/catalog-service/pkg/response"
)

type BrandHandler struct {
	service *service.BrandService
}

func NewBrandHandler(service *service.BrandService) *BrandHandler {
	return &BrandHandler{
		service: service,
	}
}

func (h *BrandHandler) GetAllBrands(w http.ResponseWriter, r *http.Request) {
	brands, err := h.service.GetAllBrands()

	if err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	response.JSON(w, http.StatusOK, brands)
}
