package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/catalog-service/internal/service"
	"github.com/catalog-service/pkg/models"
	"github.com/catalog-service/pkg/response"
)

type ProductHandler struct {
	service *service.ProductService
}

func NewProductHandler(s *service.ProductService) *ProductHandler {
	return &ProductHandler{service: s}
}

func (h *ProductHandler) GetFilteredProducts(w http.ResponseWriter, r *http.Request) {
	var filter models.ProductFilter

	if err := json.NewDecoder(r.Body).Decode(&filter); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	products, err := h.service.GetFilteredProducts(r.Context(), &filter)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrInvalidFilter):
			response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid filter"})
			return

		case errors.Is(err, models.ErrProductNotFound):
			response.JSON(w, http.StatusNotFound, map[string]string{"error": "no products found"})
			return

		default:
			response.JSON(w, http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
			return
		}
	}

	response.JSON(w, http.StatusOK, products)
}
