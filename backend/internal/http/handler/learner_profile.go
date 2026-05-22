package handler

import (
	"context"
	"net/http"
	"time"

	"foco/backend/api/internal/domain/profile"
	"foco/backend/api/internal/http/middleware"
)

type ProfileService interface {
	BuildProfile(ctx context.Context, userID, examID string, now time.Time) (*profile.Payload, error)
}

type LearnerProfileHandler struct {
	service ProfileService
}

func NewLearnerProfileHandler(service ProfileService) *LearnerProfileHandler {
	return &LearnerProfileHandler{service: service}
}

func (h *LearnerProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		http.Error(w, "profile service unavailable", http.StatusInternalServerError)
		return
	}
	claims := middleware.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	examID := r.URL.Query().Get("exam_id")
	payload, err := h.service.BuildProfile(r.Context(), claims.UserID, examID, time.Now().UTC())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  payload,
		"meta":  map[string]any{},
		"error": nil,
	})
}
