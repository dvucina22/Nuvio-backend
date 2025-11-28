package handler

import (
	"net/http"

	"github.com/catalog-service/internal/service"

	"github.com/catalog-service/pkg/response"
)

type CategoryHandler struct {
	service *service.CategoryService
}

func NewCategoryHandler(service *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		service: service,
	}
}

func (h *CategoryHandler) GetAllCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.service.GetAllCategories()

	if err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	response.JSON(w, http.StatusOK, categories)
}
