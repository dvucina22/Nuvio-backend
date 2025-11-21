package handler

import (
	"encoding/json"
	"net/http"

	"github.com/account-service/internal/service"
	"github.com/account-service/pkg/models"
	"github.com/account-service/pkg/response"
)

type RegisterHandler struct {
	svc *service.RegisterService
}

func NewRegisterHandler(svc *service.RegisterService) *RegisterHandler {
	return &RegisterHandler{svc: svc}
}

func (h *RegisterHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req *models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	acc, err := h.svc.Register(r.Context(), req)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	response.JSON(w, http.StatusCreated, acc)
}
