package router

import (
	"github.com/gorilla/mux"
	"github.com/transaction-service/internal/api/rest/handler"
	"github.com/transaction-service/internal/api/rest/middleware"
	"github.com/transaction-service/internal/service"
	"github.com/transaction-service/pkg/utils"
)

func NewRouter(
	jwtManager *utils.JWTManager,
	cardService *service.CardService,
	transactionService *service.TransactionService,
) *mux.Router {
	r := mux.NewRouter()

	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	cardHandler := handler.NewCardHandler(cardService)
	transactionHandler := handler.NewTransactionHandler(transactionService)

	transactionsAPI := r.PathPrefix("/api/transactions").Subrouter()

	protected := transactionsAPI.PathPrefix("").Subrouter()
	protected.Use(authMiddleware.RequireAuth)

	protected.HandleFunc("/cards", cardHandler.GetCards).Methods("GET")
	protected.HandleFunc("/cards/{card_id}", cardHandler.GetCard).Methods("GET")
	protected.HandleFunc("/cards", cardHandler.AddCard).Methods("POST")
	protected.HandleFunc("/cards/{card_id}", cardHandler.DeleteCard).Methods("DELETE")
	protected.HandleFunc("/cards/{card_id}/primary", cardHandler.SetPrimaryCard).Methods("PUT")

	protected.HandleFunc("/sale", transactionHandler.CreateSale).Methods("POST")
	return r
}
