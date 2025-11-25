package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/transaction-service/internal/api/rest/middleware"
	"github.com/transaction-service/internal/service"
	"github.com/transaction-service/pkg/models"
	"github.com/transaction-service/pkg/response"
)

type CardHandler struct {
	svc *service.CardService
}

func NewCardHandler(s *service.CardService) *CardHandler {
	return &CardHandler{svc: s}
}

func (h *CardHandler) AddCard(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	userID := claims.UserID

	var req models.AddCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	card, err := h.svc.AddCard(r.Context(), userID, &req)
	if err != nil {
		statusCode := h.getStatusCode(err)
		response.JSON(w, statusCode, map[string]string{"error": err.Error()})
		return
	}

	response.JSON(w, http.StatusCreated, map[string]any{"data": card})
}

func (h *CardHandler) GetCards(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	userID := claims.UserID

	cards, err := h.svc.GetCards(r.Context(), userID)
	if err != nil {
		statusCode := h.getStatusCode(err)
		response.JSON(w, statusCode, map[string]string{"error": err.Error()})
		return
	}

	if len(cards) == 0 {
		response.JSON(w, http.StatusOK, map[string]any{"data": []any{}})
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{"data": cards})
}

func (h *CardHandler) GetCard(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	userID := claims.UserID

	idStr := mux.Vars(r)["card_id"]
	cardID, err := strconv.Atoi(idStr)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid card id"})
		return
	}

	card, err := h.svc.GetCard(r.Context(), userID, cardID)
	if err != nil {
		statusCode := h.getStatusCode(err)
		response.JSON(w, statusCode, map[string]string{"error": err.Error()})
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{"data": card})
}

func (h *CardHandler) SetPrimaryCard(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	userID := claims.UserID

	idStr := mux.Vars(r)["card_id"]
	cardID, err := strconv.Atoi(idStr)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid card id"})
		return
	}

	if err := h.svc.SetPrimaryCard(r.Context(), userID, cardID); err != nil {
		statusCode := h.getStatusCode(err)
		response.JSON(w, statusCode, map[string]string{"error": err.Error()})
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "card set as primary"})
}

func (h *CardHandler) DeleteCard(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	userID := claims.UserID

	idStr := mux.Vars(r)["card_id"]
	cardID, err := strconv.Atoi(idStr)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid card id"})
		return
	}

	if err := h.svc.DeleteCard(r.Context(), userID, cardID); err != nil {
		statusCode := h.getStatusCode(err)
		response.JSON(w, statusCode, map[string]string{"error": err.Error()})
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "card deleted successfully"})
}

func (h *CardHandler) getStatusCode(err error) int {
	switch {
	case errors.Is(err, models.ErrCardNotFound):
		return http.StatusNotFound

	case errors.Is(err, models.ErrInvalidCard),
		errors.Is(err, models.ErrInvalidCardNumber),
		errors.Is(err, models.ErrCardExpired),
		errors.Is(err, models.ErrMissingFields),
		errors.Is(err, models.ErrInvalidUserId):
		return http.StatusBadRequest

	case errors.Is(err, models.ErrEncryptionFailed),
		errors.Is(err, models.ErrDatabaseOperation):
		return http.StatusInternalServerError

	default:
		return http.StatusInternalServerError
	}
}
