package handler

import (
	"encoding/json"
	"net/http"

	"github.com/account-service/internal/account"
)

func LoginHandler(svc account.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req account.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		tok, err := svc.Login(r.Context(), req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(tok)
	}
}
