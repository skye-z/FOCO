package diagnostic

import (
	"context"
	"testing"
	"time"
)

type repoStub struct {
	existingAttempt *Attempt
	createdAttempt  *Attempt
	savedProfile    *Profile
	questions       []Question
}

func (r *repoStub) FindLatestAttempt(context.Context, string, string) (*Attempt, error) {
	return r.existingAttempt, nil
}

func (r *repoStub) CreateAttempt(_ context.Context, attempt *Attempt) error {
	copy := *attempt
	r.createdAttempt = &copy
	r.existingAttempt = &copy
	return nil
}

func (r *repoStub) UpdateAttempt(_ context.Context, attempt *Attempt) error {
	copy := *attempt
	r.existingAttempt = &copy
	return nil
}

func (r *repoStub) ListDiagnosticQuestions(context.Context, string, int) ([]Question, error) {
	return r.questions, nil
}

func (r *repoStub) SaveProfile(_ context.Context, profile *Profile) error {
	copy := *profile
	r.savedProfile = &copy
	return nil
}

func (r *repoStub) FindLatestProfile(context.Context, string, string) (*Profile, error) {
	return r.savedProfile, nil
}

func TestServiceGetCurrentCreatesInitialDiagnosticAttempt(t *testing.T) {
	repo := &repoStub{
		questions: []Question{
			{
				QuestionVersionID: "v1",
				SubjectID:         "subject-1",
				SubjectName:       "数量方法",
				ChapterID:         "chapter-1",
				ChapterName:       "年金终值",
				QuestionType:      "single_choice",
				Stem:              "题目一",
				Options:           []Option{{Label: "A", Text: "选项 A"}},
				CorrectLabels:     []string{"A"},
				KnowledgePoints:   []KnowledgePoint{{ID: "kp-1", Name: "年金终值"}},
			},
			{
				QuestionVersionID: "v2",
				SubjectID:         "subject-2",
				SubjectName:       "固定收益",
				ChapterID:         "chapter-2",
				ChapterName:       "收益率曲线",
				QuestionType:      "single_choice",
				Stem:              "题目二",
				Options:           []Option{{Label: "B", Text: "选项 B"}},
				CorrectLabels:     []string{"B"},
				KnowledgePoints:   []KnowledgePoint{{ID: "kp-2", Name: "收益率曲线"}},
			},
		},
	}

	service := NewService(repo)
	payload, err := service.GetCurrent(context.Background(), "user-1", "exam-1", time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("GetCurrent returned error: %v", err)
	}

	if payload.Status != "pending" {
		t.Fatalf("expected pending diagnostic attempt, got %+v", payload)
	}
	if repo.createdAttempt == nil {
		t.Fatalf("expected initial diagnostic attempt to be created")
	}
	if len(payload.Items) != 2 {
		t.Fatalf("expected representative questions to be attached, got %+v", payload.Items)
	}
}

func TestServiceSubmitBuildsProfileSummaryAndRecommendations(t *testing.T) {
	repo := &repoStub{
		existingAttempt: &Attempt{
			ID:     "attempt-1",
			UserID: "user-1",
			ExamID: "exam-1",
			Status: "pending",
			Items: []Question{
				{
					ID:                "item-1",
					QuestionVersionID: "v1",
					SubjectID:         "subject-1",
					SubjectName:       "数量方法",
					ChapterID:         "chapter-1",
					ChapterName:       "年金终值",
					QuestionType:      "single_choice",
					CorrectLabels:     []string{"A"},
					KnowledgePoints:   []KnowledgePoint{{ID: "kp-1", Name: "年金终值"}},
				},
				{
					ID:                "item-2",
					QuestionVersionID: "v2",
					SubjectID:         "subject-2",
					SubjectName:       "固定收益",
					ChapterID:         "chapter-2",
					ChapterName:       "收益率曲线",
					QuestionType:      "single_choice",
					CorrectLabels:     []string{"B"},
					KnowledgePoints:   []KnowledgePoint{{ID: "kp-2", Name: "收益率曲线"}},
				},
			},
		},
	}

	service := NewService(repo)
	payload, err := service.Submit(context.Background(), SubmitInput{
		UserID:    "user-1",
		AttemptID: "attempt-1",
		Answers: map[string][]string{
			"item-1": {"A"},
			"item-2": {"C"},
		},
	}, time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("Submit returned error: %v", err)
	}

	if !payload.HasCompleted {
		t.Fatalf("expected completed diagnostic result, got %+v", payload)
	}
	if payload.RecommendedDifficulty != "medium" {
		t.Fatalf("expected recommended difficulty medium, got %+v", payload)
	}
	if len(payload.RecommendedSubjectIDs) == 0 || payload.RecommendedSubjectIDs[0] != "subject-2" {
		t.Fatalf("expected weakest subject to drive recommendation, got %+v", payload)
	}
	if len(payload.RecommendedKnowledgePointIDs) == 0 || payload.RecommendedKnowledgePointIDs[0] != "kp-2" {
		t.Fatalf("expected weakest knowledge point to drive recommendation, got %+v", payload)
	}
	if repo.savedProfile == nil {
		t.Fatalf("expected learner profile snapshot to be saved")
	}
	if repo.savedProfile.Summary.OverallAccuracy != 50 {
		t.Fatalf("expected saved profile to preserve accuracy, got %+v", repo.savedProfile.Summary)
	}
}
