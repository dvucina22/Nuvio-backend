package router

import (
	"github.com/catalog-service/internal/api/rest/middleware"
	"github.com/catalog-service/pkg/utils"
	"github.com/gorilla/mux"
)

func NewRouter(
	jwtManager *utils.JWTManager,
) *mux.Router {
	r := mux.NewRouter()

	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	accountsAPI := r.PathPrefix("/api/catalog").Subrouter()

	protected := accountsAPI.PathPrefix("").Subrouter()
	protected.Use(authMiddleware.RequireAuth)
	return r
}
