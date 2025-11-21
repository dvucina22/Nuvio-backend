package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/account-service/internal/api/rest/middleware"
	"github.com/account-service/internal/service"
	"github.com/account-service/pkg/models"
	"github.com/account-service/pkg/response"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(s *service.UserService) *UserHandler {
	return &UserHandler{service: s}
}

func (h *UserHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {

	claims := middleware.GetUserClaims(r.Context())

	if claims == nil {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	user, err := h.service.GetUserInfo(claims.UserID)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	if user == nil {
		response.JSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	response.JSON(w, http.StatusOK, user)
}

func (h *UserHandler) UpdateUserInfo(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())

	if claims == nil {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var updateUser *models.UpdateUser
	if err := json.NewDecoder(r.Body).Decode(&updateUser); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	err := h.service.UpdateUserInfo(r.Context(), claims.UserID, updateUser)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrEmailAlreadyExists):
			response.JSON(w, http.StatusConflict, map[string]string{"error": "email already in use"})
			return
		case errors.Is(err, models.ErrUserNotFound):
			response.JSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
			return
		default:
			response.JSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "user info updated successfully"})
}
