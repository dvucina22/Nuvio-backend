package handler

import (
	"encoding/json"
	"net/http"

	"github.com/account-service/internal/account"
)

func RegisterHandler(svc account.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req account.RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		acc, err := svc.Register(r.Context(), req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(acc)
	}
}
