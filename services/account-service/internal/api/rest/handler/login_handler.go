package handler

import (
	"encoding/json"
	"net/http"

	"github.com/account-service/internal/service"
	"github.com/account-service/pkg/models"
	"github.com/account-service/pkg/response"
)

type LoginHandler struct {
	service *service.LoginService
}

func NewLoginHandler(s *service.LoginService) *LoginHandler {
	return &LoginHandler{service: s}
}

func (h *LoginHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	res, err := h.service.Login(r.Context(), req)
	if err != nil {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}

	response.JSON(w, http.StatusOK, res)
}
