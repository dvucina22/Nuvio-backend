package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/account-service/internal/api/rest/middleware"
	"github.com/account-service/internal/service"
	"github.com/account-service/pkg/models"
	"github.com/account-service/pkg/response"
	"github.com/account-service/pkg/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type UserHandler struct {
	service           *service.UserService
	password_helper   *utils.PasswordHelper
	cloudinaryService *service.CloudinaryService
}

func NewUserHandler(s *service.UserService, ph *utils.PasswordHelper, cs *service.CloudinaryService) *UserHandler {
	return &UserHandler{
		service:           s,
		password_helper:   ph,
		cloudinaryService: cs,
	}
}

func (h *UserHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	user, err := h.service.GetUserInfo(claims.UserID)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			response.JSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
			return
		}
		if errors.Is(err, models.ErrMissingFields) {
			response.JSON(w, http.StatusBadRequest, map[string]string{"error": "missing user ID"})
			return
		}
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

func (h *UserHandler) GetUploadSignature(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	publicID := fmt.Sprintf("user_%s", claims.UserID)

	signature, err := h.cloudinaryService.GenerateUploadSignature(publicID)
	if err != nil {
		http.Error(w, "Failed to generate signature", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(signature)
}

func (h *UserHandler) UpdateProfilePicture(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req struct {
		ProfilePictureURL string `json:"profilePictureUrl"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err := h.service.UpdateUserProfilePicture(claims.UserID, &req.ProfilePictureURL)
	if err != nil {
		http.Error(w, "Failed to update profile picture", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Profile picture updated successfully",
	})
}

func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())

	if claims == nil || !Roles(claims.Roles).isAdmin() {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	users, err := h.service.GetAllUsers(r.Context())
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	response.JSON(w, http.StatusOK, users)
}

func (h *UserHandler) DeactivateUser(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())

	if claims == nil || !Roles(claims.Roles).isAdmin() {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	userID := mux.Vars(r)["id"]
	if _, err := uuid.Parse(userID); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id format"})
		return
	}
	err := h.service.DeactivateUser(userID)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "deactivated user"})
}

func (h *UserHandler) FilterUsersByName(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())

	if claims == nil || !Roles(claims.Roles).isAdmin() {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	nameSubstring := r.URL.Query().Get("name")
	users, err := h.service.FilterUsersByName(nameSubstring)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	response.JSON(w, http.StatusOK, users)
}
