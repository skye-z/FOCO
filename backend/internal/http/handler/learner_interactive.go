package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"foco/backend/api/internal/http/middleware"
)

type LearnerInteractiveHandler struct {
	interactiveService InteractiveService
}

func NewLearnerInteractiveHandler(s InteractiveService) *LearnerInteractiveHandler {
	return &LearnerInteractiveHandler{interactiveService: s}
}

func (h *LearnerInteractiveHandler) ListUnits(w http.ResponseWriter, r *http.Request) {
	if h.interactiveService == nil {
		http.Error(w, "interactive service unavailable", http.StatusInternalServerError)
		return
	}
	units, err := h.interactiveService.ListUnits(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": units, "meta": map[string]any{}, "error": nil})
}

func (h *LearnerInteractiveHandler) GetUnit(w http.ResponseWriter, r *http.Request) {
	if h.interactiveService == nil {
		http.Error(w, "interactive service unavailable", http.StatusInternalServerError)
		return
	}
	unitVersionID := r.PathValue("unitVersionId")
	unit, err := h.interactiveService.GetUnit(r.Context(), unitVersionID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	if unit == nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "unit not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": unit, "meta": map[string]any{}, "error": nil})
}

func (h *LearnerInteractiveHandler) StartAttempt(w http.ResponseWriter, r *http.Request) {
	if h.interactiveService == nil {
		http.Error(w, "interactive service unavailable", http.StatusInternalServerError)
		return
	}
	unitVersionID := r.PathValue("unitVersionId")
	claims := middleware.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	attempt, err := h.interactiveService.StartAttempt(r.Context(), unitVersionID, claims.UserID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": attempt})
}

func (h *LearnerInteractiveHandler) SubmitStepAction(w http.ResponseWriter, r *http.Request) {
	if h.interactiveService == nil {
		http.Error(w, "interactive service unavailable", http.StatusInternalServerError)
		return
	}
	attemptID := r.PathValue("attemptId")
	stepID := r.PathValue("stepId")
	var payload map[string]any
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
	}
	feedback, err := h.interactiveService.SubmitStepAction(r.Context(), attemptID, stepID, payload)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": feedback})
}

func (h *LearnerInteractiveHandler) CompleteAttempt(w http.ResponseWriter, r *http.Request) {
	if h.interactiveService == nil {
		http.Error(w, "interactive service unavailable", http.StatusInternalServerError)
		return
	}
	attemptID := r.PathValue("attemptId")
	summary, err := h.interactiveService.CompleteAttempt(r.Context(), attemptID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": summary})
}

func learnerInteractiveError(err error) int {
	if err == context.Canceled {
		return http.StatusRequestTimeout
	}
	return http.StatusInternalServerError
}
