package router

import (
	"github.com/gorilla/mux"
	"github.com/iso-com-service/internal/api/rest/handler"
	"github.com/iso-com-service/internal/api/rest/middleware"
	"github.com/iso-com-service/pkg/utils"
)

func NewRouter(
	jwtManager *utils.JWTManager,
	authorizeHandler *handler.AuthorizeHandler,
	terminalHandler *handler.TerminalHandler,
) *mux.Router {
	r := mux.NewRouter()

	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	api := r.PathPrefix("/api/bank-comm").Subrouter()

	protected := api.PathPrefix("").Subrouter()
	protected.Use(authMiddleware.RequireAuth)

	protected.HandleFunc("/authorize/sale", authorizeHandler.AuthorizeSale).Methods("POST")
	protected.HandleFunc("/authorize/void", authorizeHandler.AuthorizeVoid).Methods("POST")

	admin := api.PathPrefix("").Subrouter()
	admin.Use(authMiddleware.RequireAuth)
	admin.Use(authMiddleware.RequireAdmin)

	admin.HandleFunc("/terminal-credentials", terminalHandler.CreateTerminalCredentials).Methods("POST")

	return r
}
