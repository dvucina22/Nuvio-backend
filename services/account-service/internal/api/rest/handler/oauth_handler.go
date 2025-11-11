package handler

import (
	"encoding/json"
	"net/http"

	"github.com/account-service/internal/service"
	"github.com/account-service/pkg/response"
	"github.com/account-service/pkg/types"
	"github.com/gorilla/mux"
)

type OAuthHandler struct {
	svc *service.OAuthService
}

func NewOAuthHandler(s *service.OAuthService) *OAuthHandler {
	return &OAuthHandler{svc: s}
}

func (h *OAuthHandler) VerifyToken(w http.ResponseWriter, r *http.Request) {
	providerStr := mux.Vars(r)["provider"]

	provider, err := types.ParseProvider(providerStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req struct {
		IDToken string `json:"id_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.IDToken == "" {
		http.Error(w, "id_token is required", http.StatusBadRequest)
		return
	}

	resp, err := h.svc.VerifyIDToken(r.Context(), provider, req.IDToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}
