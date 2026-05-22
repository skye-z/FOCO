package interactive

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	cachepkg "foco/backend/api/internal/cache"
)

type UnitView struct {
	UnitVersionID string       `json:"unit_version_id"`
	Title         string       `json:"title"`
	Steps         []StepSchema `json:"steps"`
}

type UnitSummary struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	StepCount int    `json:"step_count"`
}

type UnitAttempt struct {
	ID            string `json:"id"`
	UnitVersionID string `json:"unit_version_id"`
	UserID        string `json:"user_id"`
	Status        string `json:"status"`
}

type StepFeedback struct {
	IsCorrect     bool   `json:"is_correct"`
	AllowContinue bool   `json:"allow_continue"`
	Hint          string `json:"hint"`
}

type ConceptCard struct {
	ID            string         `json:"id"`
	AttemptID     string         `json:"attempt_id"`
	UnitVersionID string         `json:"unit_version_id"`
	Content       map[string]any `json:"content"`
}

type CompletionSummary struct {
	AttemptID   string       `json:"attempt_id"`
	Status      string       `json:"status"`
	ConceptCard *ConceptCard `json:"concept_card"`
}

type Service struct {
	repo  Repository
	cache *cachepkg.Manager
}

func NewService(repo Repository, caches ...*cachepkg.Manager) *Service {
	var cacheManager *cachepkg.Manager
	if len(caches) > 0 {
		cacheManager = caches[0]
	}
	return &Service{repo: repo, cache: cacheManager}
}

func (s *Service) ListUnits(ctx context.Context) ([]UnitSummary, error) {
	if s.cache == nil {
		return s.repo.ListUnits(ctx)
	}
	var result []UnitSummary
	err := s.cache.GetJSON(ctx, interactiveNamespace(), "learner-list", 10*time.Minute, &result, func(ctx context.Context) (any, error) {
		return s.repo.ListUnits(ctx)
	})
	return result, err
}

func (s *Service) GetUnit(ctx context.Context, unitVersionID string) (*UnitView, error) {
	if s.cache == nil {
		return s.repo.GetUnit(ctx, unitVersionID)
	}
	var result *UnitView
	err := s.cache.GetJSON(ctx, interactiveNamespace(), "unit:"+unitVersionID, 10*time.Minute, &result, func(ctx context.Context) (any, error) {
		return s.repo.GetUnit(ctx, unitVersionID)
	})
	return result, err
}

func (s *Service) StartAttempt(ctx context.Context, unitVersionID, userID string) (*UnitAttempt, error) {
	return s.repo.CreateAttempt(ctx, unitVersionID, userID)
}

func (s *Service) SubmitStepAction(ctx context.Context, attemptID, stepID string, payload map[string]any) (*StepFeedback, error) {
	step, err := s.repo.GetStep(ctx, stepID)
	if err != nil {
		return nil, err
	}
	result := EvaluateStep(*step, payload)
	feedback := &StepFeedback{
		IsCorrect:     result.IsCorrect,
		AllowContinue: true,
		Hint:          result.Hint,
	}
	if err := s.repo.SaveStepAction(ctx, attemptID, stepID, payload); err != nil {
		return nil, err
	}
	if err := s.repo.SaveStepFeedback(ctx, attemptID, stepID, feedback); err != nil {
		return nil, err
	}
	return feedback, nil
}

func (s *Service) CompleteAttempt(ctx context.Context, attemptID string) (*CompletionSummary, error) {
	scope, err := s.repo.GetAttemptScope(ctx, attemptID)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CompleteAttempt(ctx, attemptID); err != nil {
		return nil, err
	}
	conceptCard, err := s.repo.CreateConceptCard(ctx, attemptID)
	if err != nil {
		return nil, err
	}
	if scope != nil {
		s.invalidateLearner(ctx, scope.UserID, scope.ExamID)
	}
	return &CompletionSummary{
		AttemptID:   attemptID,
		Status:      "completed",
		ConceptCard: conceptCard,
	}, nil
}

func (s *Service) AdminListUnits(ctx context.Context, examID, subjectID string) ([]AdminUnitSummary, error) {
	if s.cache == nil {
		return s.repo.AdminListUnits(ctx, examID, subjectID)
	}
	var result []AdminUnitSummary
	err := s.cache.GetJSON(ctx, interactiveNamespace(), "admin-list:"+examID+":"+subjectID, 5*time.Minute, &result, func(ctx context.Context) (any, error) {
		return s.repo.AdminListUnits(ctx, examID, subjectID)
	})
	return result, err
}

func (s *Service) AdminCreateUnit(ctx context.Context, examID, subjectID, title string) (*AdminUnitSummary, error) {
	unit, err := s.repo.AdminCreateUnit(ctx, examID, subjectID, title)
	s.invalidateInteractiveIfOK(ctx, err)
	return unit, err
}

func (s *Service) AdminListVersions(ctx context.Context, unitID string) ([]AdminVersionSummary, error) {
	if s.cache == nil {
		return s.repo.AdminListVersions(ctx, unitID)
	}
	var result []AdminVersionSummary
	err := s.cache.GetJSON(ctx, interactiveNamespace(), "versions:"+unitID, 5*time.Minute, &result, func(ctx context.Context) (any, error) {
		return s.repo.AdminListVersions(ctx, unitID)
	})
	return result, err
}

func (s *Service) AdminCreateVersion(ctx context.Context, unitID string) (*AdminVersionDetail, error) {
	detail, err := s.repo.AdminCreateVersion(ctx, unitID)
	s.invalidateInteractiveIfOK(ctx, err)
	return detail, err
}

func (s *Service) AdminGetVersionDetail(ctx context.Context, versionID string) (*AdminVersionDetail, error) {
	if s.cache == nil {
		return s.repo.AdminGetVersionDetail(ctx, versionID)
	}
	var result *AdminVersionDetail
	err := s.cache.GetJSON(ctx, interactiveNamespace(), "admin-version:"+versionID, 5*time.Minute, &result, func(ctx context.Context) (any, error) {
		return s.repo.AdminGetVersionDetail(ctx, versionID)
	})
	return result, err
}

func (s *Service) AdminUpdateVersion(ctx context.Context, versionID, title string, steps []StepSchema) (*AdminVersionDetail, error) {
	detail, err := s.repo.AdminUpdateVersion(ctx, versionID, title, steps)
	s.invalidateInteractiveIfOK(ctx, err)
	return detail, err
}

func (s *Service) AdminPublishVersion(ctx context.Context, versionID string) (*AdminVersionDetail, error) {
	detail, err := s.repo.AdminPublishVersion(ctx, versionID)
	s.invalidateInteractiveIfOK(ctx, err)
	return detail, err
}

func (s *Service) AdminDeleteUnit(ctx context.Context, unitID string) error {
	err := s.repo.AdminDeleteUnit(ctx, unitID)
	s.invalidateInteractiveIfOK(ctx, err)
	return err
}

func (s *Service) invalidateInteractiveIfOK(ctx context.Context, err error) {
	if err == nil && s.cache != nil {
		s.cache.Invalidate(ctx, interactiveNamespace(), contentNamespace())
	}
}

func (s *Service) invalidateLearner(ctx context.Context, userID, examID string) {
	if s.cache == nil {
		return
	}
	namespaces := []string{learnerNamespace(userID, examID)}
	if examID != "" {
		namespaces = append(namespaces, learnerNamespace(userID, ""))
	}
	s.cache.Invalidate(ctx, namespaces...)
}

func interactiveNamespace() string { return "interactive:all" }
func contentNamespace() string     { return "content:all" }
func learnerNamespace(userID, examID string) string {
	return "learner:user:" + userID + ":exam:" + examID
}

func newUUIDLikeString() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4],
		b[4:6],
		b[6:8],
		b[8:10],
		b[10:16],
	)
}
