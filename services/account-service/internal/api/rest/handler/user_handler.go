package handler

import (
	"net/http"

	"github.com/account-service/internal/api/rest/middleware"
	"github.com/account-service/internal/service"
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
