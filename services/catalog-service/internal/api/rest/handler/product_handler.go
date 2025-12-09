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
	"github.com/catalog-service/pkg/utils"
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

func (h *ProductHandler) UpdateProductByID(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())

	if claims == nil || !utils.Roles(claims.Roles).IsAdmin() {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	vars := mux.Vars(r)["id"]
	productID, err := strconv.Atoi(vars)

	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid product ID"})
		return
	}

	var body products.UpdateProduct
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	err = h.service.UpdateProductByID(r.Context(), productID, &body)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrProductNotFound):
			response.JSON(w, http.StatusNotFound, map[string]string{"error": "product not found"})
			return
		default:
			response.JSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update product"})
			return
		}
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "product info updated successfully"})
}

func (h *ProductHandler) DeleteProductByID(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())

	if claims == nil || !utils.Roles(claims.Roles).IsAdmin() {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	vars := mux.Vars(r)["id"]
	productID, err := strconv.Atoi(vars)

	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid product ID"})
		return
	}

	err = h.service.DeleteProductByID(r.Context(), productID)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrProductNotFound):
			response.JSON(w, http.StatusNotFound, map[string]string{"error": "product not found"})
			return
		default:
			response.JSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete product"})
			return
		}
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "deleted product"})
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())

	if claims == nil || !utils.Roles(claims.Roles).IsAdmin() {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var body products.CreateProduct
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.service.CreateProduct(&body); err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	response.JSON(w, http.StatusCreated, map[string]string{"message": "created new product"})
}
