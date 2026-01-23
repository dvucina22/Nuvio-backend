package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/iso-com-service/internal/service"
	api "github.com/iso-com-service/pkg/models"
)

type AuthorizeHandler struct {
	svc *service.ISOService
}

func NewAuthorizeHandler(svc *service.ISOService) *AuthorizeHandler {
	return &AuthorizeHandler{svc: svc}
}

func (h *AuthorizeHandler) AuthorizeSale(w http.ResponseWriter, r *http.Request) {
	var req api.AuthorizeSaleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	jsonData, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal request to JSON: %v", err)
	} else {
		log.Printf("Iso com client received request:\n%s", jsonData)
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
