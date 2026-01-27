package router

import (
	"github.com/gorilla/mux"
	"github.com/transaction-service/internal/api/rest/handler"
	"github.com/transaction-service/internal/api/rest/middleware"
	"github.com/transaction-service/internal/config"
	"github.com/transaction-service/internal/service"
	"github.com/transaction-service/pkg/utils"
)

func NewRouter(
	jwtManager *utils.JWTManager,
	cardService *service.CardService,
	transactionService *service.TransactionService,
	cfg *config.Config,
) *mux.Router {
	r := mux.NewRouter()

	corsMiddleware := middleware.NewCORSMiddleware(cfg.AllowedOrigins)
	r.Use(corsMiddleware.Handle)

	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	cardHandler := handler.NewCardHandler(cardService)
	transactionHandler := handler.NewTransactionHandler(transactionService)

	transactionsAPI := r.PathPrefix("/api/transactions").Subrouter()

	protected := transactionsAPI.PathPrefix("").Subrouter()
	protected.Use(authMiddleware.RequireAuth)

	protected.HandleFunc("/cards", cardHandler.GetCards).Methods("GET", "OPTIONS")
	protected.HandleFunc("/cards/{card_id}", cardHandler.GetCard).Methods("GET", "OPTIONS")
	protected.HandleFunc("/cards", cardHandler.AddCard).Methods("POST", "OPTIONS")
	protected.HandleFunc("/cards/{card_id}", cardHandler.DeleteCard).Methods("DELETE", "OPTIONS")
	protected.HandleFunc("/cards/{card_id}/primary", cardHandler.SetPrimaryCard).Methods("PUT", "OPTIONS")

	protected.HandleFunc("/statistics", transactionHandler.GetStatistics).Methods("GET")

	protected.HandleFunc("/history", transactionHandler.GetFilteredTransactions).Methods("POST", "OPTIONS")
	protected.HandleFunc("/history/{transaction_id}", transactionHandler.GetTransactionDetail).Methods("GET", "OPTIONS")

	protected.HandleFunc("/sale", transactionHandler.CreateSale).Methods("POST", "OPTIONS")
	protected.HandleFunc("/sale/{transaction_id}/void", transactionHandler.VoidSale).Methods("POST", "OPTIONS")

	return r
}
