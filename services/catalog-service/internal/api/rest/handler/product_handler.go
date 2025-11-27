package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/catalog-service/internal/api/rest/middleware"
	"github.com/catalog-service/internal/service"
	"github.com/catalog-service/pkg/models"
	"github.com/catalog-service/pkg/models/products"
	"github.com/catalog-service/pkg/response"
	"github.com/gorilla/mux"
)

type ProductHandler struct {
	service *service.ProductService
}

func NewProductHandler(s *service.ProductService) *ProductHandler {
	return &ProductHandler{service: s}
}

func (h *ProductHandler) GetFilteredProducts(w http.ResponseWriter, r *http.Request) {
	var filter products.ProductFilter

	if err := json.NewDecoder(r.Body).Decode(&filter); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	if filter.IsFavorite != nil && *filter.IsFavorite {
		claims := middleware.GetUserClaims(r.Context())
		if claims == nil {
			response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}

		filter.IsFavoriteUserID = &claims.UserID
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

func (h *ProductHandler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)["id"]
	productID, err := strconv.Atoi(vars)

	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid product id"})
		return
	}

	product, err := h.service.GetProductByID(r.Context(), productID)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrProductNotFound):
			response.JSON(w, http.StatusNotFound, map[string]string{"error": "product not found"})
			return
		default:
			response.JSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get product"})
			return
		}
	}
	response.JSON(w, http.StatusOK, product)
}
