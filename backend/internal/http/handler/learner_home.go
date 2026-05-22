package handler

import (
	"context"
	"net/http"
	"time"

	"foco/backend/api/internal/domain/home"
	"foco/backend/api/internal/http/middleware"
)

type HomeService interface {
	BuildHome(ctx context.Context, userID, examID string, now time.Time) (*home.Payload, error)
	BuildRecommendations(ctx context.Context, userID, examID string, now time.Time) ([]home.Recommendation, error)
}

type LearnerHomeHandler struct {
	service HomeService
}

func NewLearnerHomeHandler(service HomeService) *LearnerHomeHandler {
	return &LearnerHomeHandler{service: service}
}

func (h *LearnerHomeHandler) GetHome(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		http.Error(w, "home service unavailable", http.StatusInternalServerError)
		return
	}
	claims := middleware.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	examID := r.URL.Query().Get("exam_id")
	payload, err := h.service.BuildHome(r.Context(), claims.UserID, examID, time.Now().UTC())
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

func (h *LearnerHomeHandler) GetRecommendations(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		http.Error(w, "home service unavailable", http.StatusInternalServerError)
		return
	}
	claims := middleware.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	examID := r.URL.Query().Get("exam_id")
	payload, err := h.service.BuildRecommendations(r.Context(), claims.UserID, examID, time.Now().UTC())
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
