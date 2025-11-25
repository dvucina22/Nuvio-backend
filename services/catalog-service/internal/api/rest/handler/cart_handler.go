package handler

import (
	"encoding/json"
	"net/http"

	"github.com/catalog-service/internal/api/rest/middleware"
	"github.com/catalog-service/internal/service"
	"github.com/catalog-service/pkg/models"
	"github.com/catalog-service/pkg/models/cart"
	"github.com/catalog-service/pkg/response"
)

type CartHandler struct {
	service *service.CartService
}

func NewCartHandler(s *service.CartService) *CartHandler {
	return &CartHandler{service: s}
}

func (h *CartHandler) AddProductToCart(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req cart.AddToCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	err := h.service.AddProductToCart(r.Context(), claims.UserID, req.ProductID, req.Quantity)
	if err != nil {
		switch err {
		case models.ErrInvalidProductID, models.ErrInvalidQuantity:
			response.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		default:
			response.JSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}

	response.JSON(w, http.StatusCreated, map[string]string{"message": "product added to cart"})
}

func (h *CartHandler) RemoveProductFromCart(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req cart.RemoveFromCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	err := h.service.RemoveProductFromCart(r.Context(), claims.UserID, req.ProductID)
	if err != nil {
		switch err {
		case models.ErrInvalidProductID:
			response.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		case models.ErrProductNotInCart:
			response.JSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		default:
			response.JSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "removed from cart"})
}

func (h *CartHandler) GetCartContents(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	items, err := h.service.GetCartContents(r.Context(), claims.UserID)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	response.JSON(w, http.StatusOK, items)
}

func (h *CartHandler) EmptyCart(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	err := h.service.EmptyCart(r.Context(), claims.UserID)
	if err == models.ErrCartNotFound {
		response.JSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	if err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "cart emptied"})
}
