package practice

import (
	"context"
	"testing"
)

type serviceRepoStub struct {
	enrollment            *EnrollmentRef
	candidates            []CandidateQuestion
	candidateBatches      [][]CandidateQuestion
	profileRecommendation *DiagnosticRecommendation
	lastScope             *ResolvedScope
	scopes                []ResolvedScope
	wrongBook             []WrongBookItem
}

func (r *serviceRepoStub) FindEnrollment(context.Context, string, string) (*EnrollmentRef, error) {
	return r.enrollment, nil
}

func (r *serviceRepoStub) ListCandidateQuestions(_ context.Context, input CreateSessionInput) ([]CandidateQuestion, error) {
	scope := ResolvedScope{
		Difficulty:        input.Difficulty,
		SubjectIDs:        input.SubjectIDs,
		ChapterIDs:        input.ChapterIDs,
		KnowledgePointIDs: input.KnowledgePointIDs,
	}
	r.lastScope = &scope
	r.scopes = append(r.scopes, scope)
	if len(r.candidateBatches) > 0 {
		index := len(r.scopes) - 1
		if index >= len(r.candidateBatches) {
			return nil, nil
		}
		return r.candidateBatches[index], nil
	}
	return r.candidates, nil
}

func (r *serviceRepoStub) CreateSession(context.Context, *PracticeSessionRecord, []PracticeSessionItemRecord) error {
	return nil
}

func (r *serviceRepoStub) GetSession(context.Context, string, string) (*PracticeSessionView, error) {
	return nil, nil
}

func (r *serviceRepoStub) GetSubmission(context.Context, string, string, string) (*SubmissionRecord, error) {
	return nil, nil
}

func (r *serviceRepoStub) SaveSubmission(context.Context, string, string, bool, []string, int, int, int) error {
	return nil
}

func (r *serviceRepoStub) ListKnowledgePoints(context.Context, []string) ([]KnowledgePointRef, error) {
	return nil, nil
}

func (r *serviceRepoStub) GetSummary(context.Context, string, string) (*SessionSummary, error) {
	return nil, nil
}

func (r *serviceRepoStub) ListWrongBook(context.Context, WrongBookFilter) ([]WrongBookItem, error) {
	return r.wrongBook, nil
}

func (r *serviceRepoStub) FindLatestDiagnosticRecommendation(context.Context, string, string) (*DiagnosticRecommendation, error) {
	return r.profileRecommendation, nil
}

func TestServiceListWrongBookUsesRepository(t *testing.T) {
	repo := &serviceRepoStub{
		wrongBook: []WrongBookItem{
			{ID: "item-1", QuestionID: "question-1", Status: "open", ErrorCount: 2},
		},
	}
	service := NewService(repo)
	items, err := service.ListWrongBook(context.Background(), WrongBookFilter{UserID: "user-1", ExamID: "exam-1"})
	if err != nil {
		t.Fatalf("ListWrongBook returned error: %v", err)
	}
	if len(items) != 1 || items[0].QuestionID != "question-1" {
		t.Fatalf("expected wrong-book items from repository, got %+v", items)
	}
}

func TestServiceCreateSessionUsesDiagnosticRecommendationForIntelligentMode(t *testing.T) {
	repo := &serviceRepoStub{
		enrollment: &EnrollmentRef{
			ID:       "enrollment-1",
			ExamID:   "exam-1",
			ExamName: "CFA 一级",
		},
		candidates: []CandidateQuestion{
			{
				QuestionID:        "question-1",
				QuestionVersionID: "version-1",
				SubjectID:         "subject-2",
				ChapterID:         "chapter-2",
				QuestionType:      "single_choice",
				Stem:              `{"text":"题目"}`,
				Options:           `{"choices":[{"id":"opt_a","text":"A"}]}`,
				CorrectAnswer:     `{"selected_option_ids":["opt_a"]}`,
				Explanation:       `{"text":"解析"}`,
				KnowledgePointIDs: []string{"kp-2"},
			},
		},
		profileRecommendation: &DiagnosticRecommendation{
			RecommendedDifficulty:        "medium",
			RecommendedSubjectIDs:        []string{"subject-2"},
			RecommendedChapterIDs:        []string{"chapter-2"},
			RecommendedKnowledgePointIDs: []string{"kp-2"},
		},
	}

	service := NewService(repo)
	_, err := service.CreateSession(context.Background(), CreateSessionInput{
		UserID: "user-1",
		ExamID: "exam-1",
		Mode:   "intelligent",
		Count:  10,
	})
	if err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}

	if repo.lastScope == nil {
		t.Fatalf("expected intelligent scope to be resolved")
	}
	if repo.lastScope.Difficulty != "medium" {
		t.Fatalf("expected difficulty from diagnostic recommendation, got %+v", repo.lastScope)
	}
	if len(repo.lastScope.SubjectIDs) == 0 || repo.lastScope.SubjectIDs[0] != "subject-2" {
		t.Fatalf("expected subject scope from diagnostic recommendation, got %+v", repo.lastScope)
	}
	if len(repo.lastScope.ChapterIDs) == 0 || repo.lastScope.ChapterIDs[0] != "chapter-2" {
		t.Fatalf("expected chapter scope from diagnostic recommendation, got %+v", repo.lastScope)
	}
	if len(repo.lastScope.KnowledgePointIDs) == 0 || repo.lastScope.KnowledgePointIDs[0] != "kp-2" {
		t.Fatalf("expected knowledge-point scope from diagnostic recommendation, got %+v", repo.lastScope)
	}
}

func TestServiceCreateSessionRelaxesIntelligentScopeWhenRecommendationHasNoQuestions(t *testing.T) {
	candidate := CandidateQuestion{
		QuestionID:        "question-1",
		QuestionVersionID: "version-1",
		SubjectID:         "subject-2",
		ChapterID:         "chapter-2",
		QuestionType:      "single_choice",
		Stem:              `{"text":"题目"}`,
		Options:           `{"choices":[{"id":"opt_a","text":"A"}]}`,
		CorrectAnswer:     `{"selected_option_ids":["opt_a"]}`,
		Explanation:       `{"text":"解析"}`,
		KnowledgePointIDs: []string{"kp-other"},
	}
	repo := &serviceRepoStub{
		enrollment: &EnrollmentRef{
			ID:       "enrollment-1",
			ExamID:   "exam-1",
			ExamName: "CFA 一级",
		},
		candidateBatches: [][]CandidateQuestion{
			nil,
			{candidate},
		},
		profileRecommendation: &DiagnosticRecommendation{
			RecommendedDifficulty:        "medium",
			RecommendedSubjectIDs:        []string{"subject-2"},
			RecommendedChapterIDs:        []string{"chapter-2"},
			RecommendedKnowledgePointIDs: []string{"kp-missing"},
		},
	}

	service := NewService(repo)
	view, err := service.CreateSession(context.Background(), CreateSessionInput{
		UserID: "user-1",
		ExamID: "exam-1",
		Mode:   "intelligent",
		Count:  10,
	})
	if err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}
	if view.TotalCount != 1 {
		t.Fatalf("expected relaxed scope to create a one-question session, got %+v", view)
	}
	if len(repo.scopes) != 2 {
		t.Fatalf("expected original and relaxed scope attempts, got %+v", repo.scopes)
	}
	if len(repo.scopes[0].KnowledgePointIDs) != 1 {
		t.Fatalf("expected first attempt to use recommended knowledge point scope, got %+v", repo.scopes)
	}
	if len(repo.scopes[1].KnowledgePointIDs) != 0 {
		t.Fatalf("expected second attempt to relax knowledge point scope, got %+v", repo.scopes)
	}
	if len(repo.scopes[1].SubjectIDs) != 1 || len(repo.scopes[1].ChapterIDs) != 1 {
		t.Fatalf("expected relaxed attempt to keep broader subject/chapter scope, got %+v", repo.scopes)
	}
}
