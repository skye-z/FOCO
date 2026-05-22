package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"foco/backend/api/internal/domain/diagnostic"
	"foco/backend/api/internal/http/middleware"
)

type DiagnosticService interface {
	GetCurrent(ctx context.Context, userID, examID string, now time.Time) (*diagnostic.CurrentPayload, error)
	Restart(ctx context.Context, userID, examID, triggerType string, now time.Time) (*diagnostic.CurrentPayload, error)
	Submit(ctx context.Context, input diagnostic.SubmitInput, now time.Time) (*diagnostic.ProfileSummary, error)
}

type LearnerDiagnosticHandler struct {
	service DiagnosticService
}

func NewLearnerDiagnosticHandler(service DiagnosticService) *LearnerDiagnosticHandler {
	return &LearnerDiagnosticHandler{service: service}
}

func (h *LearnerDiagnosticHandler) GetCurrent(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		http.Error(w, "diagnostic service unavailable", http.StatusInternalServerError)
		return
	}
	claims := middleware.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	payload, err := h.service.GetCurrent(r.Context(), claims.UserID, r.URL.Query().Get("exam_id"), time.Now().UTC())
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

func (h *LearnerDiagnosticHandler) Restart(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		http.Error(w, "diagnostic service unavailable", http.StatusInternalServerError)
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

	payload, err := h.service.Restart(r.Context(), claims.UserID, body.ExamID, diagnostic.TriggerTypeManualRestart, time.Now().UTC())
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

func (h *LearnerDiagnosticHandler) Submit(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		http.Error(w, "diagnostic service unavailable", http.StatusInternalServerError)
		return
	}
	claims := middleware.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var body struct {
		Answers map[string][]string `json:"answers"`
	}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}

	payload, err := h.service.Submit(r.Context(), diagnostic.SubmitInput{
		UserID:    claims.UserID,
		AttemptID: r.PathValue("attemptId"),
		Answers:   body.Answers,
	}, time.Now().UTC())
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
