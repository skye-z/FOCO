package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	authpkg "foco/backend/api/internal/auth"
	"foco/backend/api/internal/domain/account"
	"foco/backend/api/internal/http/middleware"
)

type AdminUserService interface {
	ListAdminUsers(ctx context.Context) ([]account.AdminUserSummary, error)
	GrantRole(ctx context.Context, userID, role, grantedBy string) error
	DisableUser(ctx context.Context, userID string) error
	ResetPassword(ctx context.Context, userID, newPassword string) error
}

type AdminUsersHandler struct {
	userService AdminUserService
}

func NewAdminUsersHandler(userService AdminUserService) *AdminUsersHandler {
	return &AdminUsersHandler{userService: userService}
}

func (h *AdminUsersHandler) List(w http.ResponseWriter, r *http.Request) {
	if !requireAdminRole(w, r) {
		return
	}

	if h.userService == nil {
		http.Error(w, "user service unavailable", http.StatusInternalServerError)
		return
	}

	users, err := h.userService.ListAdminUsers(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  users,
		"meta":  map[string]any{},
		"error": nil,
	})
}

func (h *AdminUsersHandler) GrantRole(w http.ResponseWriter, r *http.Request) {
	claims := requireAdminClaims(w, r)
	if claims == nil {
		return
	}
	if h.userService == nil {
		http.Error(w, "user service unavailable", http.StatusInternalServerError)
		return
	}

	var body struct {
		Role string `json:"role"`
	}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	if body.Role == "" {
		http.Error(w, "role is required", http.StatusBadRequest)
		return
	}

	if err := h.userService.GrantRole(r.Context(), r.PathValue("userId"), strings.TrimSpace(body.Role), claims.UserID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"ok": true}, "meta": map[string]any{}, "error": nil})
}

func (h *AdminUsersHandler) DisableUser(w http.ResponseWriter, r *http.Request) {
	claims := requireAdminClaims(w, r)
	if claims == nil {
		return
	}
	if h.userService == nil {
		http.Error(w, "user service unavailable", http.StatusInternalServerError)
		return
	}
	if claims.UserID == r.PathValue("userId") {
		http.Error(w, "cannot disable current admin account", http.StatusBadRequest)
		return
	}

	if err := h.userService.DisableUser(r.Context(), r.PathValue("userId")); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"ok": true}, "meta": map[string]any{}, "error": nil})
}

func (h *AdminUsersHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if !requireAdminRole(w, r) {
		return
	}
	if h.userService == nil {
		http.Error(w, "user service unavailable", http.StatusInternalServerError)
		return
	}

	var body struct {
		NewPassword string `json:"new_password"`
	}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	if len(strings.TrimSpace(body.NewPassword)) < 8 {
		http.Error(w, "new_password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	if err := h.userService.ResetPassword(r.Context(), r.PathValue("userId"), body.NewPassword); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"ok": true}, "meta": map[string]any{}, "error": nil})
}

func requireAdminRole(w http.ResponseWriter, r *http.Request) bool {
	return requireAdminClaims(w, r) != nil
}

func requireAdminClaims(w http.ResponseWriter, r *http.Request) *authpkg.Claims {
	claims := middleware.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return nil
	}
	for _, role := range claims.Roles {
		if role == "admin" {
			return claims
		}
	}
	http.Error(w, "forbidden", http.StatusForbidden)
	return nil
}
