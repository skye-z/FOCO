package handler

import (
	"context"
	"encoding/json"
	"net/http"

	authpkg "foco/backend/api/internal/auth"
	"foco/backend/api/internal/domain/account"
	"foco/backend/api/internal/http/middleware"
)

type AccountService interface {
	BootstrapLearner(ctx context.Context, input account.BootstrapInput) ([]string, error)
	ListExams(ctx context.Context) ([]account.ExamSummary, error)
	ListAdminUsers(ctx context.Context) ([]account.AdminUserSummary, error)
	GetActiveEnrollment(ctx context.Context, userID string) (*account.ActiveEnrollment, error)
	EnsureEnrollment(ctx context.Context, input account.CreateEnrollmentInput) (*account.ActiveEnrollment, error)
}

type AuditWriter interface{}

type LearnerIdentityHandler struct {
	accountService AccountService
}

func NewLearnerIdentityHandler(accountService AccountService) *LearnerIdentityHandler {
	return &LearnerIdentityHandler{accountService: accountService}
}

func (h *LearnerIdentityHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var activeEnrollment *account.ActiveEnrollment
	var err error
	if h.accountService != nil {
		activeEnrollment, err = h.accountService.GetActiveEnrollment(r.Context(), claims.UserID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"user":                   toUserPayload(claims),
			"active_exam_enrollment": activeEnrollment,
		},
		"meta":  map[string]any{},
		"error": nil,
	})
}

func (h *LearnerIdentityHandler) Bootstrap(w http.ResponseWriter, r *http.Request) {
	if h.accountService == nil {
		http.Error(w, "account service unavailable", http.StatusInternalServerError)
		return
	}
	claims := middleware.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var body struct {
		DisplayName string `json:"display_name"`
	}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}

	roles, err := h.accountService.BootstrapLearner(r.Context(), account.BootstrapInput{
		UserID:      claims.UserID,
		DisplayName: body.DisplayName,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	claims.Roles = roles

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"user": toUserPayload(claims),
		},
		"meta":  map[string]any{},
		"error": nil,
	})
}

func (h *LearnerIdentityHandler) ListExams(w http.ResponseWriter, r *http.Request) {
	if h.accountService == nil {
		http.Error(w, "account service unavailable", http.StatusInternalServerError)
		return
	}
	items, err := h.accountService.ListExams(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data":  items,
		"meta":  map[string]any{},
		"error": nil,
	})
}

func (h *LearnerIdentityHandler) CreateEnrollment(w http.ResponseWriter, r *http.Request) {
	if h.accountService == nil {
		http.Error(w, "account service unavailable", http.StatusInternalServerError)
		return
	}
	claims := middleware.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var body struct {
		ExamID string `json:"exam_id"`
	}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}

	enrollment, err := h.accountService.EnsureEnrollment(r.Context(), account.CreateEnrollmentInput{
		UserID: claims.UserID,
		ExamID: body.ExamID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  enrollment,
		"meta":  map[string]any{},
		"error": nil,
	})
}

func toUserPayload(claims *authpkg.Claims) map[string]any {
	return map[string]any{
		"id":    claims.UserID,
		"email": claims.Email,
		"roles": claims.Roles,
	}
}

func writeJSON(w http.ResponseWriter, status int, payload map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
