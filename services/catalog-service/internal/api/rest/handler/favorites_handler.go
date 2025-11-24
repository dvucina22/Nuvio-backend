package handler

import (
	"encoding/json"
	"net/http"

	"github.com/catalog-service/internal/api/rest/middleware"
	"github.com/catalog-service/internal/service"
	"github.com/catalog-service/pkg/models/favorites"
	"github.com/catalog-service/pkg/response"
)

type FavoritesHandler struct {
	service *service.FavoritesService
}

func NewFavoritesHandler(s *service.FavoritesService) *FavoritesHandler {
	return &FavoritesHandler{service: s}
}

func (h *FavoritesHandler) AddToFavorites(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req favorites.FavoriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	err := h.service.AddToFavorites(r.Context(), claims.UserID, req.ProductID)
	if err != nil {
		switch err {
		case service.ErrInvalidProductID:
			response.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		case service.ErrAlreadyFavorited:
			response.JSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		default:
			response.JSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}

	response.JSON(w, http.StatusCreated, map[string]string{"message": "added to favorites"})
}

func (h *FavoritesHandler) RemoveFromFavorites(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req favorites.FavoriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	err := h.service.RemoveFromFavorites(r.Context(), claims.UserID, req.ProductID)
	if err != nil {
		switch err {
		case service.ErrInvalidProductID:
			response.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		case service.ErrNotFavorited:
			response.JSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		default:
			response.JSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "removed from favorites"})
}
