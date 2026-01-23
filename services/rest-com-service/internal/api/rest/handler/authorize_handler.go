package handler

import (
	"encoding/json"
	"net/http"

	"github.com/rest-com-service/internal/service"
	api "github.com/rest-com-service/pkg/models"
)

type AuthorizeHandler struct {
	svc *service.RESTService
}

func NewAuthorizeHandler(svc *service.RESTService) *AuthorizeHandler {
	return &AuthorizeHandler{svc: svc}
}

func (h *AuthorizeHandler) AuthorizeSale(w http.ResponseWriter, r *http.Request) {
	var req api.AuthorizeSaleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	resp, err := h.svc.AuthorizeSale(r.Context(), &req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *AuthorizeHandler) AuthorizeVoid(w http.ResponseWriter, r *http.Request) {
	var req api.AuthorizeVoidRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	resp, err := h.svc.AuthorizeVoid(r.Context(), &req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
