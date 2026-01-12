package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/transaction-service/internal/api/rest/middleware"
	"github.com/transaction-service/internal/client/host"
	"github.com/transaction-service/internal/service"
	"github.com/transaction-service/pkg/models"
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
	if claims == nil {
		models.Fail(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
		return
	}

	userID := claims.UserID

	var req models.SaleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		models.Fail(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body", map[string]any{
			"reason": err.Error(),
		})
		return
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		models.Fail(w, http.StatusBadRequest, "INVALID_USER_ID", "invalid user id", nil)
		return
	}
	req.UserID = parsedUserID

	authHeader := r.Header.Get("Authorization")
	ctx := host.WithAuthHeader(r.Context(), authHeader)

	res, err := h.svc.CreateSale(ctx, userID, &req)
	if err != nil {
		apiErr := models.MapError(err)
		models.Fail(w, apiErr.Status, apiErr.Code, apiErr.Message, nil)
		return
	}

	if rc, ok := getResponseCode(res); ok && rc == "30" {
		models.Fail(w, http.StatusUnprocessableEntity, "HOST_DECLINED", "transaction declined by host", map[string]any{
			"responseCode": rc,
			"status":       getStatus(res),
			"id":           getID(res),
		})
		return
	}

	models.Ok(w, http.StatusCreated, res, nil)
}

func (h *TransactionHandler) VoidSale(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		models.Fail(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
		return
	}

	if !isAdmin(claims) {
		models.Fail(w, http.StatusForbidden, "FORBIDDEN", "forbidden", nil)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["transaction_id"]
	txID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		models.Fail(w, http.StatusBadRequest, "INVALID_TRANSACTION_ID", "invalid transaction id", nil)
		return
	}

	req := &models.VoidRequest{
		TransactionID: txID,
	}

	authHeader := r.Header.Get("Authorization")
	ctx := host.WithAuthHeader(r.Context(), authHeader)

	res, err := h.svc.VoidSale(ctx, req)
	if err != nil {
		apiErr := models.MapError(err)
		models.Fail(w, apiErr.Status, apiErr.Code, apiErr.Message, nil)
		return
	}

	if rc, ok := getResponseCode(res); ok && rc == "30" {
		models.Fail(w, http.StatusUnprocessableEntity, "HOST_DECLINED", "void declined by host", map[string]any{
			"responseCode": rc,
			"status":       getStatus(res),
			"id":           getID(res),
		})
		return
	}

	models.Ok(w, http.StatusCreated, res, nil)
}

func (h *TransactionHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		models.Fail(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
		return
	}

	userID := claims.UserID

	vars := mux.Vars(r)
	idStr := vars["transaction_id"]
	txID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		models.Fail(w, http.StatusBadRequest, "INVALID_TRANSACTION_ID", "invalid transaction id", nil)
		return
	}

	tx, products, err := h.svc.GetTransactionDetail(r.Context(), userID, txID)
	if err != nil {
		apiErr := models.MapError(err)
		models.Fail(w, apiErr.Status, apiErr.Code, apiErr.Message, nil)
		return
	}

	models.Ok(w, http.StatusOK, map[string]any{
		"transaction": tx,
		"products":    products,
	}, nil)
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

func getResponseCode(v any) (string, bool) {
	switch x := v.(type) {
	case *models.SaleCreateResponse:
		return strings.TrimSpace(x.ResponseCode), true

	case *models.VoidCreateResponse:
		return strings.TrimSpace(x.ResponseCode), true

	case *models.Transaction:
		if x.ResponseCode == nil {
			return "", false
		}
		return strings.TrimSpace(*x.ResponseCode), true
	}

	return "", false
}

func getStatus(v any) any {
	switch x := v.(type) {
	case *models.SaleCreateResponse:
		return x.Status
	case *models.VoidCreateResponse:
		return x.Status
	case *models.Transaction:
		return x.Status
	}
	return nil
}

func getID(v any) any {
	switch x := v.(type) {
	case *models.SaleCreateResponse:
		return x.ID
	case *models.VoidCreateResponse:
		return x.ID
	case *models.Transaction:
		return x.ID
	}
	return nil
}
