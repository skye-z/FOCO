package interactive

import (
	"context"
	"errors"
	"testing"
)

type serviceTestRepo struct {
	unit          *UnitView
	step          *StepSchema
	attempt       *UnitAttempt
	conceptCard   *ConceptCard
	adminUnits    []AdminUnitSummary
	completedWith string
	savedPayload  map[string]any
	savedFeedback *StepFeedback
}

func (r *serviceTestRepo) ListUnits(context.Context) ([]UnitSummary, error) {
	return nil, nil
}

func (r *serviceTestRepo) GetUnit(context.Context, string) (*UnitView, error) {
	if r.unit == nil {
		return nil, errors.New("unit not found")
	}
	return r.unit, nil
}

func (r *serviceTestRepo) CreateAttempt(_ context.Context, unitVersionID, userID string) (*UnitAttempt, error) {
	if r.attempt != nil {
		return r.attempt, nil
	}
	return &UnitAttempt{ID: "attempt_1", UnitVersionID: unitVersionID, UserID: userID, Status: "in_progress"}, nil
}

func (r *serviceTestRepo) GetAttemptScope(context.Context, string) (*AttemptScope, error) {
	return &AttemptScope{UserID: "user_1", ExamID: "exam_1"}, nil
}

func (r *serviceTestRepo) GetStep(context.Context, string) (*StepSchema, error) {
	if r.step == nil {
		return nil, errors.New("step not found")
	}
	return r.step, nil
}

func (r *serviceTestRepo) SaveStepAction(_ context.Context, _, _ string, payload map[string]any) error {
	r.savedPayload = payload
	return nil
}

func (r *serviceTestRepo) SaveStepFeedback(_ context.Context, _, _ string, feedback *StepFeedback) error {
	r.savedFeedback = feedback
	return nil
}

func (r *serviceTestRepo) CompleteAttempt(_ context.Context, attemptID string) error {
	r.completedWith = attemptID
	return nil
}

func (r *serviceTestRepo) CreateConceptCard(context.Context, string) (*ConceptCard, error) {
	if r.conceptCard == nil {
		return nil, errors.New("concept card missing")
	}
	return r.conceptCard, nil
}

func (r *serviceTestRepo) AdminListUnits(context.Context, string, string) ([]AdminUnitSummary, error) {
	return r.adminUnits, nil
}

func (r *serviceTestRepo) AdminCreateUnit(context.Context, string, string, string) (*AdminUnitSummary, error) {
	return nil, nil
}

func (r *serviceTestRepo) AdminListVersions(context.Context, string) ([]AdminVersionSummary, error) {
	return nil, nil
}

func (r *serviceTestRepo) AdminCreateVersion(context.Context, string) (*AdminVersionDetail, error) {
	return nil, nil
}

func (r *serviceTestRepo) AdminGetVersionDetail(context.Context, string) (*AdminVersionDetail, error) {
	return nil, nil
}

func (r *serviceTestRepo) AdminUpdateVersion(context.Context, string, string, []StepSchema) (*AdminVersionDetail, error) {
	return nil, nil
}

func (r *serviceTestRepo) AdminPublishVersion(context.Context, string) (*AdminVersionDetail, error) {
	return nil, nil
}

func (r *serviceTestRepo) AdminDeleteUnit(context.Context, string) error {
	return nil
}

func TestServiceSubmitStepActionPersistsFeedback(t *testing.T) {
	repo := &serviceTestRepo{
		step: &StepSchema{
			ID:         "step_1",
			WidgetType: "ordering_matching",
			EvaluationConfig: map[string]any{
				"correct_order": []any{"pv", "fv", "r"},
			},
		},
	}

	service := NewService(repo)
	feedback, err := service.SubmitStepAction(context.Background(), "attempt_1", "step_1", map[string]any{
		"ordered_ids": []any{"pv", "fv", "r"},
	})
	if err != nil {
		t.Fatalf("SubmitStepAction returned error: %v", err)
	}

	if !feedback.IsCorrect || !feedback.AllowContinue {
		t.Fatalf("expected correct feedback with allow continue, got %+v", feedback)
	}
	if repo.savedFeedback == nil || !repo.savedFeedback.IsCorrect {
		t.Fatalf("expected feedback to be persisted, got %+v", repo.savedFeedback)
	}
	if repo.savedPayload == nil {
		t.Fatalf("expected step payload to be saved")
	}
}

func TestServiceSubmitStepActionAllowsContinueAfterIncorrectAnswer(t *testing.T) {
	repo := &serviceTestRepo{
		step: &StepSchema{
			ID:         "step_1",
			WidgetType: "highlight_marking",
			EvaluationConfig: map[string]any{
				"expected_highlights": []any{"FIFO", "减值损失计入利润表"},
			},
		},
	}

	service := NewService(repo)
	feedback, err := service.SubmitStepAction(context.Background(), "attempt_1", "step_1", map[string]any{
		"marked_ids": []any{"FIFO"},
	})
	if err != nil {
		t.Fatalf("SubmitStepAction returned error: %v", err)
	}

	if feedback.IsCorrect || !feedback.AllowContinue {
		t.Fatalf("expected incorrect feedback to still allow continue, got %+v", feedback)
	}
}

func TestServiceCompleteAttemptReturnsConceptCard(t *testing.T) {
	repo := &serviceTestRepo{
		conceptCard: &ConceptCard{
			ID:            "card_1",
			AttemptID:     "attempt_1",
			UnitVersionID: "unit_v1",
			Content: map[string]any{
				"summary": "You can now explain the TVM chain from PV to discount rate.",
			},
		},
	}

	service := NewService(repo)
	summary, err := service.CompleteAttempt(context.Background(), "attempt_1")
	if err != nil {
		t.Fatalf("CompleteAttempt returned error: %v", err)
	}

	if summary.Status != "completed" {
		t.Fatalf("expected completed status, got %q", summary.Status)
	}
	if summary.ConceptCard == nil {
		t.Fatal("expected completion summary to include concept card")
	}
	if summary.ConceptCard.AttemptID != "attempt_1" {
		t.Fatalf("expected concept card for attempt_1, got %+v", summary.ConceptCard)
	}
	if repo.completedWith != "attempt_1" {
		t.Fatalf("expected attempt completion to persist attempt_1, got %q", repo.completedWith)
	}
}

func TestServiceAdminListUnitsPreservesUnitIDs(t *testing.T) {
	repo := &serviceTestRepo{
		adminUnits: []AdminUnitSummary{
			{ID: "unit-1", Title: "Unit 1"},
			{ID: "unit-2", Title: "Unit 2"},
		},
	}

	service := NewService(repo)
	units, err := service.AdminListUnits(context.Background(), "", "")
	if err != nil {
		t.Fatalf("AdminListUnits returned error: %v", err)
	}
	if len(units) != 2 {
		t.Fatalf("expected 2 units, got %d", len(units))
	}
	if units[0].ID == "" || units[1].ID == "" {
		t.Fatalf("expected listed units to preserve IDs, got %+v", units)
	}
}
