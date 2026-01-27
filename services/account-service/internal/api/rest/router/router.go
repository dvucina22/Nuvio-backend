package router

import (
	"github.com/account-service/internal/api/rest/handler"
	"github.com/account-service/internal/api/rest/middleware"
	"github.com/account-service/internal/config"
	"github.com/account-service/internal/service"
	"github.com/account-service/pkg/utils"
	"github.com/gorilla/mux"
)

func NewRouter(
	registerService *service.RegisterService,
	loginService *service.LoginService,
	oauthService *service.OAuthService,
	userService *service.UserService,
	jwtManager *utils.JWTManager,
	passwordHelper *utils.PasswordHelper,
	clousdinaryService *service.CloudinaryService,
	roleService *service.RoleService,
	cfg *config.Config,
) *mux.Router {
	r := mux.NewRouter()

	corsMiddleware := middleware.NewCORSMiddleware(cfg.AllowedOrigins)
	r.Use(corsMiddleware.Handle)

	registerHandler := handler.NewRegisterHandler(registerService)
	loginHandler := handler.NewLoginHandler(loginService)
	oauthHandler := handler.NewOAuthHandler(oauthService)
	userHandler := handler.NewUserHandler(userService, passwordHelper, clousdinaryService)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)
	roleHandler := handler.NewRoleHandler(roleService)

	accountsAPI := r.PathPrefix("/api/accounts").Subrouter()

	accountsAPI.HandleFunc("/register", registerHandler.Register).Methods("POST", "OPTIONS")
	accountsAPI.HandleFunc("/login", loginHandler.Login).Methods("POST", "OPTIONS")

	accountsAPI.HandleFunc("/oauth/{provider}/verify", oauthHandler.VerifyToken).Methods("POST", "OPTIONS")

	protected := accountsAPI.PathPrefix("").Subrouter()
	protected.Use(authMiddleware.RequireAuth)

	protected.HandleFunc("/logged-user", userHandler.GetUserInfo).Methods("GET", "OPTIONS")
	protected.HandleFunc("/logged-user", userHandler.UpdateUserInfo).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/update-password", userHandler.UpdateUserPassword).Methods("POST", "OPTIONS")

	protected.HandleFunc("/profile-picture/update", userHandler.UpdateProfilePicture).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/profile-picture/upload-signature", userHandler.GetUploadSignature).Methods("GET", "OPTIONS")

	protected.HandleFunc("/roles", roleHandler.GetAllRoles).Methods("GET", "OPTIONS")
	protected.HandleFunc("/roles/{role_id}/user/{user_id}", roleHandler.AddUserRole).Methods("POST", "OPTIONS")
	protected.HandleFunc("/roles/{role_id}/user/{user_id}", roleHandler.RemoveUserRole).Methods("DELETE", "OPTIONS")

	protected.HandleFunc("/users", userHandler.GetAllUsers).Methods("GET", "OPTIONS")
	protected.HandleFunc("/users/{id}", userHandler.DeactivateUser).Methods("DELETE", "OPTIONS")
	protected.HandleFunc("/users/filter", userHandler.FilterUsersByName).Methods("GET", "OPTIONS")

	return r
}
