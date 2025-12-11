package handler

import (
	"net/http"
	"strconv"

	"github.com/account-service/internal/api/rest/middleware"
	"github.com/account-service/internal/service"
	"github.com/account-service/pkg/models"
	"github.com/account-service/pkg/response"
	"github.com/account-service/pkg/utils"
	"github.com/gorilla/mux"
)

type RoleHandler struct {
	roleService *service.RoleService
}

type Roles []string

func NewRoleHandler(rs *service.RoleService) *RoleHandler {
	return &RoleHandler{
		roleService: rs,
	}
}

func (h *RoleHandler) GetAllRoles(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())

	if claims == nil || !utils.Roles(claims.Roles).IsAdmin() {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	roles, err := h.roleService.GetAllRoles(r.Context())
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	response.JSON(w, http.StatusOK, roles)
}

func (h *RoleHandler) AddUserRole(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())

	if claims == nil || !utils.Roles(claims.Roles).IsAdmin() {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	userID := mux.Vars(r)["user_id"]
	roleIDStr := mux.Vars(r)["role_id"]
	roleID, err := strconv.Atoi(roleIDStr)

	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid role_id"})
		return
	}

	if err := h.roleService.AddUserRole(r.Context(), userID, roleID); err != nil {
		if err == models.ErrCannotAssignRole {
			response.JSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
			return
		}
		if err == models.ErrUserAlreadyHasRole {
			response.JSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
		response.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "role added successfully"})
}

func (h *RoleHandler) RemoveUserRole(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())

	if claims == nil || !utils.Roles(claims.Roles).IsAdmin() {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	userID := mux.Vars(r)["user_id"]
	roleIDStr := mux.Vars(r)["role_id"]
	roleID, err := strconv.Atoi(roleIDStr)

	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid role_id"})
		return
	}

	if err := h.roleService.RemoveUserRole(r.Context(), userID, roleID); err != nil {
		if err == models.ErrCannotRemoveRole {
			response.JSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
			return
		}
		if err == models.ErrUserDoesNotHaveRole {
			response.JSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		response.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "role removed successfully"})
}
