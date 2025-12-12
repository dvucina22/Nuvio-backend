package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/transaction-service/internal/api/rest/middleware"
	"github.com/transaction-service/internal/service"
	"github.com/transaction-service/pkg/models"
	"github.com/transaction-service/pkg/response"
	"github.com/transaction-service/pkg/utils"
)

type TransactionHandler struct {
	svc *service.TransactionService
}

func NewTransactionHandler(s *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{svc: s}
}

func (h *TransactionHandler) CreateSale(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	userID := claims.UserID
	var req models.SaleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}

	req.UserID = parsedUserID

	res, err := h.svc.CreateSale(r.Context(), userID, &req)
	if err != nil {
		statusCode := h.getStatusCode(err)
		response.JSON(w, statusCode, map[string]string{"error": err.Error()})
		return
	}

	response.JSON(w, http.StatusCreated, res)
}

func (h *TransactionHandler) VoidSale(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	if !isAdmin(claims) {
		response.JSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	vars := mux.Vars(r)
	idStr := vars["transaction_id"]
	txID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid transaction id"})
		return
	}

	req := &models.VoidRequest{
		TransactionID: txID,
	}

	voidTx, err := h.svc.VoidSale(r.Context(), req)
	if err != nil {
		statusCode := h.getStatusCode(err)
		response.JSON(w, statusCode, map[string]string{"error": err.Error()})
		return
	}

	response.JSON(w, http.StatusCreated, map[string]any{
		"data": voidTx,
	})
}

func (h *TransactionHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	userID := claims.UserID

	vars := mux.Vars(r)
	idStr := vars["transaction_id"]
	txID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid transaction id"})
		return
	}

	tx, products, err := h.svc.GetTransactionDetail(r.Context(), userID, txID)
	if err != nil {
		statusCode := h.getStatusCode(err)
		response.JSON(w, statusCode, map[string]string{"error": err.Error()})
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"transaction": tx,
			"products":    products,
		},
	})
}

func (h *TransactionHandler) getStatusCode(err error) int {
	switch {
	case errors.Is(err, models.ErrCardNotFound):
		return http.StatusNotFound

	case errors.Is(err, models.ErrInvalidCard),
		errors.Is(err, models.ErrInvalidCardNumber),
		errors.Is(err, models.ErrCardExpired),
		errors.Is(err, models.ErrMissingFields),
		errors.Is(err, models.ErrInvalidUserId),
		errors.Is(err, models.ErrInvalidCurrencyCode),
		errors.Is(err, models.ErrInvalidProducts),
		errors.Is(err, models.ErrInvalidAmount),
		errors.Is(err, models.ErrInvalidTransactionType),
		errors.Is(err, models.ErrInvalidTransactionState):
		return http.StatusBadRequest

	case errors.Is(err, models.ErrTransactionNotFound):
		return http.StatusNotFound

	case errors.Is(err, models.ErrVoidAlreadyExists):
		return http.StatusConflict

	case errors.Is(err, models.ErrTerminalCredentialsNotFound):
		return http.StatusUnprocessableEntity

	case errors.Is(err, models.ErrEncryptionFailed),
		errors.Is(err, models.ErrDatabaseOperation):
		return http.StatusInternalServerError

	default:
		return http.StatusInternalServerError
	}
}

func isAdmin(claims *utils.UserClaims) bool {
	if claims == nil {
		return false
	}

	for _, r := range claims.Roles {
		if r == "ADMIN" || r == "admin" {
			return true
		}
	}

	return false
}
