package handler

import (
	"encoding/json"
	"net/http"

	"github.com/rest-com-service/internal/service"
)

type TerminalHandler struct {
	svc *service.TerminalService
}

func NewTerminalHandler(svc *service.TerminalService) *TerminalHandler {
	return &TerminalHandler{svc: svc}
}

type createTerminalReq struct {
	UserID   string `json:"userId"`
	HostType string `json:"hostType"`
	TID      string `json:"tid"`
	MID      string `json:"mid"`
	Active   bool   `json:"active"`
}

func (h *TerminalHandler) CreateTerminalCredentials(w http.ResponseWriter, r *http.Request) {
	var req createTerminalReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	if err := h.svc.CreateTerminalCredentials(r.Context(), req.UserID, req.HostType, req.TID, req.MID, req.Active); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]bool{"ok": true})
}
