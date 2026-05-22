package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"foco/backend/api/internal/domain/interactive"
)

type InteractiveService interface {
	ListUnits(ctx context.Context) ([]interactive.UnitSummary, error)
	GetUnit(ctx context.Context, unitVersionID string) (*interactive.UnitView, error)
	StartAttempt(ctx context.Context, unitVersionID, userID string) (*interactive.UnitAttempt, error)
	SubmitStepAction(ctx context.Context, attemptID, stepID string, payload map[string]any) (*interactive.StepFeedback, error)
	CompleteAttempt(ctx context.Context, attemptID string) (*interactive.CompletionSummary, error)

	AdminListUnits(ctx context.Context, examID, subjectID string) ([]interactive.AdminUnitSummary, error)
	AdminCreateUnit(ctx context.Context, examID, subjectID, title string) (*interactive.AdminUnitSummary, error)
	AdminListVersions(ctx context.Context, unitID string) ([]interactive.AdminVersionSummary, error)
	AdminCreateVersion(ctx context.Context, unitID string) (*interactive.AdminVersionDetail, error)
	AdminGetVersionDetail(ctx context.Context, versionID string) (*interactive.AdminVersionDetail, error)
	AdminUpdateVersion(ctx context.Context, versionID, title string, steps []interactive.StepSchema) (*interactive.AdminVersionDetail, error)
	AdminPublishVersion(ctx context.Context, versionID string) (*interactive.AdminVersionDetail, error)
	AdminDeleteUnit(ctx context.Context, unitID string) error
}

type AdminInteractiveHandler struct {
	interactiveService InteractiveService
}

func NewAdminInteractiveHandler(s InteractiveService) *AdminInteractiveHandler {
	return &AdminInteractiveHandler{interactiveService: s}
}

func mapInteractiveError(err error) int {
	switch {
	case errors.Is(err, interactive.ErrVersionNotFound):
		return http.StatusNotFound
	case errors.Is(err, interactive.ErrVersionReadOnly):
		return http.StatusConflict
	case errors.Is(err, interactive.ErrVersionNotPublishable):
		return http.StatusBadRequest
	case errors.Is(err, interactive.ErrUnitNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

func (h *AdminInteractiveHandler) ListUnits(w http.ResponseWriter, r *http.Request) {
	if h.interactiveService == nil {
		http.Error(w, "interactive service unavailable", http.StatusInternalServerError)
		return
	}
	examID := r.URL.Query().Get("exam_id")
	subjectID := r.URL.Query().Get("subject_id")
	units, err := h.interactiveService.AdminListUnits(r.Context(), examID, subjectID)
	if err != nil {
		writeJSON(w, mapInteractiveError(err), map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": units, "meta": map[string]any{}, "error": nil})
}

func (h *AdminInteractiveHandler) CreateUnit(w http.ResponseWriter, r *http.Request) {
	if h.interactiveService == nil {
		http.Error(w, "interactive service unavailable", http.StatusInternalServerError)
		return
	}
	var body struct {
		ExamID    string `json:"exam_id"`
		SubjectID string `json:"subject_id"`
		Title     string `json:"title"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	unit, err := h.interactiveService.AdminCreateUnit(r.Context(), body.ExamID, body.SubjectID, body.Title)
	if err != nil {
		writeJSON(w, mapInteractiveError(err), map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": unit})
}

func (h *AdminInteractiveHandler) ListVersions(w http.ResponseWriter, r *http.Request) {
	if h.interactiveService == nil {
		http.Error(w, "interactive service unavailable", http.StatusInternalServerError)
		return
	}
	unitID := r.PathValue("unitId")
	versions, err := h.interactiveService.AdminListVersions(r.Context(), unitID)
	if err != nil {
		writeJSON(w, mapInteractiveError(err), map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": versions, "meta": map[string]any{}, "error": nil})
}

func (h *AdminInteractiveHandler) CreateVersion(w http.ResponseWriter, r *http.Request) {
	if h.interactiveService == nil {
		http.Error(w, "interactive service unavailable", http.StatusInternalServerError)
		return
	}
	unitID := r.PathValue("unitId")
	detail, err := h.interactiveService.AdminCreateVersion(r.Context(), unitID)
	if err != nil {
		writeJSON(w, mapInteractiveError(err), map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": detail})
}

func (h *AdminInteractiveHandler) GetVersionDetail(w http.ResponseWriter, r *http.Request) {
	if h.interactiveService == nil {
		http.Error(w, "interactive service unavailable", http.StatusInternalServerError)
		return
	}
	versionID := r.PathValue("versionId")
	detail, err := h.interactiveService.AdminGetVersionDetail(r.Context(), versionID)
	if err != nil {
		writeJSON(w, mapInteractiveError(err), map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": detail, "meta": map[string]any{}, "error": nil})
}

func (h *AdminInteractiveHandler) UpdateVersion(w http.ResponseWriter, r *http.Request) {
	if h.interactiveService == nil {
		http.Error(w, "interactive service unavailable", http.StatusInternalServerError)
		return
	}
	versionID := r.PathValue("versionId")
	var body struct {
		Title string                   `json:"title"`
		Steps []interactive.StepSchema `json:"steps"`
	}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	detail, err := h.interactiveService.AdminUpdateVersion(r.Context(), versionID, body.Title, body.Steps)
	if err != nil {
		writeJSON(w, mapInteractiveError(err), map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": detail, "meta": map[string]any{}, "error": nil})
}

func (h *AdminInteractiveHandler) PublishVersion(w http.ResponseWriter, r *http.Request) {
	if h.interactiveService == nil {
		http.Error(w, "interactive service unavailable", http.StatusInternalServerError)
		return
	}
	versionID := r.PathValue("versionId")
	detail, err := h.interactiveService.AdminPublishVersion(r.Context(), versionID)
	if err != nil {
		writeJSON(w, mapInteractiveError(err), map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": detail})
}

func (h *AdminInteractiveHandler) DeleteUnit(w http.ResponseWriter, r *http.Request) {
	if h.interactiveService == nil {
		http.Error(w, "interactive service unavailable", http.StatusInternalServerError)
		return
	}
	unitID := r.PathValue("unitId")
	if err := h.interactiveService.AdminDeleteUnit(r.Context(), unitID); err != nil {
		writeJSON(w, mapInteractiveError(err), map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": nil})
}
