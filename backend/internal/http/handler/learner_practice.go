package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"foco/backend/api/internal/domain/practice"
	"foco/backend/api/internal/http/middleware"
)

type PracticeService interface {
	CreateSession(ctx context.Context, input practice.CreateSessionInput) (*practice.PracticeSessionView, error)
	GetSession(ctx context.Context, userID, sessionID string) (*practice.PracticeSessionView, error)
	SubmitAnswer(ctx context.Context, input practice.SubmitAnswerInput) (*practice.SubmitResult, error)
	GetSummary(ctx context.Context, userID, sessionID string) (*practice.SessionSummary, error)
	ListWrongBook(ctx context.Context, filter practice.WrongBookFilter) ([]practice.WrongBookItem, error)
}

type LearnerPracticeHandler struct {
	practiceService PracticeService
}

func NewLearnerPracticeHandler(practiceService PracticeService) *LearnerPracticeHandler {
	return &LearnerPracticeHandler{practiceService: practiceService}
}

func (h *LearnerPracticeHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	if h.practiceService == nil {
		http.Error(w, "practice service unavailable", http.StatusInternalServerError)
		return
	}
	claims := middleware.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var body struct {
		ExamID            string   `json:"exam_id"`
		Mode              string   `json:"mode"`
		QuestionTypes     []string `json:"question_types"`
		Difficulty        string   `json:"difficulty"`
		Count             int      `json:"count"`
		SubjectIDs        []string `json:"subject_ids"`
		ChapterIDs        []string `json:"chapter_ids"`
		KnowledgePointIDs []string `json:"knowledge_point_ids"`
	}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}

	session, err := h.practiceService.CreateSession(r.Context(), practice.CreateSessionInput{
		UserID:            claims.UserID,
		ExamID:            body.ExamID,
		Mode:              body.Mode,
		QuestionTypes:     body.QuestionTypes,
		Difficulty:        body.Difficulty,
		Count:             body.Count,
		SubjectIDs:        body.SubjectIDs,
		ChapterIDs:        body.ChapterIDs,
		KnowledgePointIDs: body.KnowledgePointIDs,
	})
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, practice.ErrActiveEnrollmentRequired):
			status = http.StatusForbidden
		case errors.Is(err, practice.ErrDiagnosticRequired):
			status = http.StatusPreconditionFailed
		case errors.Is(err, practice.ErrNoPracticeQuestions):
			status = http.StatusBadRequest
		}
		writeJSON(w, status, map[string]any{
			"error": err.Error(),
			"meta":  map[string]any{},
			"data":  nil,
		})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"data":  session,
		"meta":  map[string]any{},
		"error": nil,
	})
}

func (h *LearnerPracticeHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	if h.practiceService == nil {
		http.Error(w, "practice service unavailable", http.StatusInternalServerError)
		return
	}
	claims := middleware.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	session, err := h.practiceService.GetSession(r.Context(), claims.UserID, r.PathValue("sessionId"))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, practice.ErrPracticeSessionNotFound) {
			status = http.StatusNotFound
		}
		writeJSON(w, status, map[string]any{
			"error": err.Error(),
			"meta":  map[string]any{},
			"data":  nil,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  session,
		"meta":  map[string]any{},
		"error": nil,
	})
}

func (h *LearnerPracticeHandler) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	if h.practiceService == nil {
		http.Error(w, "practice service unavailable", http.StatusInternalServerError)
		return
	}
	claims := middleware.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var body struct {
		Answer          any `json:"answer"`
		DurationSeconds int `json:"duration_seconds"`
	}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}

	result, err := h.practiceService.SubmitAnswer(r.Context(), practice.SubmitAnswerInput{
		UserID:          claims.UserID,
		SessionID:       r.PathValue("sessionId"),
		ItemID:          r.PathValue("itemId"),
		Answer:          body.Answer,
		DurationSeconds: body.DurationSeconds,
	})
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, practice.ErrPracticeItemNotFound) {
			status = http.StatusNotFound
		}
		writeJSON(w, status, map[string]any{
			"error": err.Error(),
			"meta":  map[string]any{},
			"data":  nil,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  result,
		"meta":  map[string]any{},
		"error": nil,
	})
}

func (h *LearnerPracticeHandler) Summary(w http.ResponseWriter, r *http.Request) {
	if h.practiceService == nil {
		http.Error(w, "practice service unavailable", http.StatusInternalServerError)
		return
	}
	claims := middleware.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	summary, err := h.practiceService.GetSummary(r.Context(), claims.UserID, r.PathValue("sessionId"))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, practice.ErrPracticeSessionNotFound) {
			status = http.StatusNotFound
		}
		writeJSON(w, status, map[string]any{
			"error": err.Error(),
			"meta":  map[string]any{},
			"data":  nil,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  summary,
		"meta":  map[string]any{},
		"error": nil,
	})
}

func (h *LearnerPracticeHandler) WrongBook(w http.ResponseWriter, r *http.Request) {
	if h.practiceService == nil {
		http.Error(w, "practice service unavailable", http.StatusInternalServerError)
		return
	}
	claims := middleware.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	items, err := h.practiceService.ListWrongBook(r.Context(), practice.WrongBookFilter{
		UserID:           claims.UserID,
		ExamID:           q.Get("exam_id"),
		SubjectID:        q.Get("subject_id"),
		ChapterID:        q.Get("chapter_id"),
		KnowledgePointID: q.Get("knowledge_point_id"),
		Status:           q.Get("status"),
		Page:             page,
		PageSize:         pageSize,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error": err.Error(),
			"meta":  map[string]any{},
			"data":  nil,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  items,
		"meta":  map[string]any{"count": len(items)},
		"error": nil,
	})
}
