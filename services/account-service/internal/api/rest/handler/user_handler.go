package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/account-service/internal/api/rest/middleware"
	"github.com/account-service/internal/service"
	"github.com/account-service/pkg/models"
	"github.com/account-service/pkg/response"
	"github.com/account-service/pkg/utils"
)

type UserHandler struct {
	service         *service.UserService
	password_helper *utils.PasswordHelper
}

func NewUserHandler(s *service.UserService, ph *utils.PasswordHelper) *UserHandler {
	return &UserHandler{service: s, password_helper: ph}
}

func (h *UserHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	user, err := h.service.GetUserInfo(claims.UserID)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get user info"})
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

	var body models.UpdateUser
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	err := h.service.UpdateUserInfo(r.Context(), claims.UserID, &body)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrEmailAlreadyExists):
			response.JSON(w, http.StatusConflict, map[string]string{"error": "email already in use"})
			return
		case errors.Is(err, models.ErrUserNotFound):
			response.JSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
			return
		default:
			response.JSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update user info"})
			return
		}
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "user info updated successfully"})
}

func (h *UserHandler) UpdateUserPassword(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var body models.UpdatePassword
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	err := h.service.UpdateUserPassword(r.Context(), claims.UserID, &body)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrMissingFields):
			response.JSON(w, http.StatusBadRequest, map[string]string{"error": "oldPassword and newPassword required"})
			return
		case errors.Is(err, models.ErrUserNotFound):
			response.JSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
			return
		case errors.Is(err, models.ErrInvalidPassword):
			response.JSON(w, http.StatusForbidden, map[string]string{"error": "old password incorrect"})
			return
		case errors.Is(err, models.ErrPasswordWeak):
			response.JSON(w, http.StatusBadRequest, map[string]string{"error": "new password is too weak"})
			return
		default:
			response.JSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update password"})
			return
		}
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "password updated successfully"})
}

func (h *UserHandler) UpdateUserProfilePicture(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var body struct {
		ProfilePictureURL *string `json:"profilePictureUrl"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	err := h.service.UpdateUserProfilePicture(claims.UserID, body.ProfilePictureURL)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrUserNotFound):
			response.JSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
			return
		default:
			response.JSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update profile picture"})
			return
		}
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "profile picture updated successfully"})
}
