package home

import (
	"context"
	"testing"
	"time"

	diagnosticpkg "foco/backend/api/internal/domain/diagnostic"
	profilepkg "foco/backend/api/internal/domain/profile"
)

type serviceRepoStub struct {
	exam       *ExamOverview
	wallet     int
	streak     *profilepkg.StreakStats
	sessions   []profilepkg.PracticeSessionSummary
	kpResults  []profilepkg.KnowledgePointResult
	diagnostic *DiagnosticSummary
	labDone    int
}

func (s *serviceRepoStub) GetExamOverview(context.Context, string) (*ExamOverview, error) {
	return s.exam, nil
}

func (s *serviceRepoStub) GetWalletBalance(context.Context, string) (int, error) {
	return s.wallet, nil
}

func (s *serviceRepoStub) GetStreakStats(context.Context, string, string) (*profilepkg.StreakStats, error) {
	return s.streak, nil
}

func (s *serviceRepoStub) ListPracticeSessions(context.Context, string, string, int) ([]profilepkg.PracticeSessionSummary, error) {
	return s.sessions, nil
}

func (s *serviceRepoStub) ListKnowledgePointResults(context.Context, string, string) ([]profilepkg.KnowledgePointResult, error) {
	return s.kpResults, nil
}

func (s *serviceRepoStub) GetLatestDiagnosticSummary(context.Context, string, string) (*DiagnosticSummary, error) {
	return s.diagnostic, nil
}

func (s *serviceRepoStub) CountCompletedInteractiveAttempts(context.Context, string, string, time.Time, time.Time) (int, error) {
	return s.labDone, nil
}

func TestServiceBuildHomeUsesRealExamDateAndAggregatesDashboard(t *testing.T) {
	now := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	nextExam := now.Add(10 * 24 * time.Hour)
	repo := &serviceRepoStub{
		exam: &ExamOverview{
			ExamID:       "exam-1",
			ExamName:     "CFA 一级",
			NextExamDate: &nextExam,
		},
		wallet: 18,
		streak: &profilepkg.StreakStats{
			CurrentStreak: 4,
			BestStreak:    9,
			Status:        "active",
		},
		sessions: []profilepkg.PracticeSessionSummary{
			{
				ID:                   "session-1",
				Status:               "in_progress",
				TotalCount:           20,
				AnsweredCount:        8,
				CorrectCount:         6,
				XPEarned:             60,
				CoinsEarned:          6,
				TotalDurationSeconds: 1200,
				CreatedAt:            now.Add(-2 * time.Hour),
			},
			{
				ID:                   "session-2",
				Status:               "completed",
				TotalCount:           12,
				AnsweredCount:        12,
				CorrectCount:         10,
				XPEarned:             100,
				CoinsEarned:          10,
				TotalDurationSeconds: 1800,
				CreatedAt:            now.Add(-24 * time.Hour),
			},
		},
		kpResults: []profilepkg.KnowledgePointResult{
			{KnowledgePointID: "kp-1", KnowledgePointName: "年金终值", Correct: false, SubmittedAt: now.AddDate(0, 0, -1)},
			{KnowledgePointID: "kp-1", KnowledgePointName: "年金终值", Correct: true, SubmittedAt: now.AddDate(0, 0, -1)},
			{KnowledgePointID: "kp-2", KnowledgePointName: "收益率曲线", Correct: false, SubmittedAt: now.AddDate(0, 0, -2)},
			{KnowledgePointID: "kp-2", KnowledgePointName: "收益率曲线", Correct: false, SubmittedAt: now.AddDate(0, 0, -2)},
			{KnowledgePointID: "kp-3", KnowledgePointName: "信用利差", Correct: true, SubmittedAt: now.AddDate(0, 0, -1)},
		},
		diagnostic: &DiagnosticSummary{
			HasCompleted:                   true,
			SummaryText:                    "你在固定收益和收益率曲线方面基础偏弱，建议先从中等难度强化。",
			RecommendedDifficulty:          "medium",
			RecommendedSubjectNames:        []string{"固定收益"},
			RecommendedChapterNames:        []string{"收益率曲线"},
			RecommendedKnowledgePointNames: []string{"收益率曲线"},
		},
	}

	service := NewService(repo)
	payload, err := service.BuildHome(context.Background(), "user-1", "exam-1", now)
	if err != nil {
		t.Fatalf("BuildHome returned error: %v", err)
	}

	if payload.ExamName != "CFA 一级" {
		t.Fatalf("expected exam name, got %+v", payload)
	}
	if payload.CountdownDays == nil || *payload.CountdownDays != 10 {
		t.Fatalf("expected real countdown days from exam date, got %+v", payload)
	}
	if payload.Coins != 18 || payload.Streak != 4 {
		t.Fatalf("expected wallet and streak in payload, got %+v", payload)
	}
	if payload.Level != 2 || payload.XPCurrent != 60 || payload.XPTarget != 100 {
		t.Fatalf("expected level progress from real XP, got %+v", payload)
	}
	if len(payload.TodayTasks) == 0 || payload.TodayTasks[0].Type != "continue" {
		t.Fatalf("expected first today task to continue in-progress session, got %+v", payload.TodayTasks)
	}
	if len(payload.WeeklyActivity) != 7 {
		t.Fatalf("expected 7-day weekly activity, got %+v", payload.WeeklyActivity)
	}
	if payload.WeeklyActivity[6].QuestionCount != 8 || payload.WeeklyActivity[6].CompletedTasks != 0 {
		t.Fatalf("expected today rhythm to include real question/task counts, got %+v", payload.WeeklyActivity[6])
	}
	if len(payload.WeakPoints) == 0 || payload.WeakPoints[0].ID != "kp-2" {
		t.Fatalf("expected weakest point first, got %+v", payload.WeakPoints)
	}
	if payload.WeakPoints[0].ConfidenceScore == 0 || payload.WeakPoints[0].ForgettingDueAt == "" {
		t.Fatalf("expected weak point mastery details, got %+v", payload.WeakPoints[0])
	}
	if !payload.WeakPoints[0].ReviewDue || payload.WeakPoints[0].ReviewStage == "" || payload.WeakPoints[0].LastEvidenceAt == "" {
		t.Fatalf("expected Ebbinghaus review window details, got %+v", payload.WeakPoints[0])
	}
	if payload.DiagnosticSummary == nil || payload.DiagnosticSummary.RecommendedDifficulty != "medium" {
		t.Fatalf("expected diagnostic summary in home payload, got %+v", payload.DiagnosticSummary)
	}
	if payload.ProgressStats.CompletedQuestions != 20 || payload.ProgressStats.Accuracy != 80 || payload.ProgressStats.WrongCount != 2 {
		t.Fatalf("expected progress stats from real sessions and answers, got %+v", payload.ProgressStats)
	}
	if len(payload.Recommendations) == 0 || len(payload.Recommendations[0].Reasons) == 0 {
		t.Fatalf("expected explainable recommendations in home payload, got %+v", payload.Recommendations)
	}
	if payload.TodayTasks[1].Type != "spaced_review" || payload.TodayTasks[1].Reason == "" {
		t.Fatalf("expected today path to include Ebbinghaus spaced review, got %+v", payload.TodayTasks)
	}
	if payload.LearningReport.TrendSummary == "" || len(payload.LearningReport.NextActions) == 0 {
		t.Fatalf("expected learning report summary and next actions, got %+v", payload.LearningReport)
	}
}

func TestServiceBuildHomeMarksCompletedTodayPathTasks(t *testing.T) {
	now := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	completedAt := now.Add(-30 * time.Minute)
	service := NewService(&serviceRepoStub{
		exam: &ExamOverview{ExamID: "exam-1", ExamName: "CFA 一级"},
		sessions: []profilepkg.PracticeSessionSummary{
			{
				ID:            "session-1",
				Status:        "completed",
				TotalCount:    10,
				AnsweredCount: 10,
				CorrectCount:  8,
				CreatedAt:     now.Add(-45 * time.Minute),
				CompletedAt:   &completedAt,
			},
		},
	})

	payload, err := service.BuildHome(context.Background(), "user-1", "exam-1", now)
	if err != nil {
		t.Fatalf("BuildHome returned error: %v", err)
	}
	if len(payload.TodayTasks) == 0 {
		t.Fatalf("expected today tasks")
	}
	completedIndex := -1
	for index, task := range payload.TodayTasks {
		if task.Status == "completed" {
			completedIndex = index
			break
		}
	}
	if completedIndex == -1 || payload.TodayTasks[completedIndex].ProgressPercent != 100 {
		t.Fatalf("expected one today practice path task completed, got %+v", payload.TodayTasks)
	}
	if payload.TodayTasks[completedIndex].ActionHref != "/practice/setup" {
		t.Fatalf("expected completed task to be a practice path task, got %+v", payload.TodayTasks[completedIndex])
	}
	for index, task := range payload.TodayTasks {
		if index != completedIndex && task.Status == "completed" {
			t.Fatalf("expected only one completed task to be consumed, got %+v", payload.TodayTasks)
		}
	}
	if payload.TodayTasks[0].ActionHref == "/diagnostic" && payload.TodayTasks[0].Status == "completed" {
		t.Fatalf("expected only one completed task to be consumed, got %+v", payload.TodayTasks)
	}
}

func TestServiceBuildHomeMarksCompletedInteractivePathTasks(t *testing.T) {
	now := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	service := NewService(&serviceRepoStub{
		exam:    &ExamOverview{ExamID: "exam-1", ExamName: "CFA 一级"},
		labDone: 1,
		diagnostic: &DiagnosticSummary{
			HasCompleted: true,
			KnowledgePoints: []diagnosticpkg.KnowledgeMastery{
				{
					KnowledgePointID:   "kp-1",
					KnowledgePointName: "存货减值",
					MasteryScore:       30,
					ConfidenceScore:    80,
					ForgettingDueAt:    "2026-05-22",
				},
			},
		},
	})

	payload, err := service.BuildHome(context.Background(), "user-1", "exam-1", now)
	if err != nil {
		t.Fatalf("BuildHome returned error: %v", err)
	}

	found := false
	for _, task := range payload.TodayTasks {
		if task.ActionHref == "/labs" {
			found = true
			if task.Status != "completed" || task.ProgressPercent != 100 {
				t.Fatalf("expected interactive path task completed, got %+v", task)
			}
		}
	}
	if !found {
		t.Fatalf("expected today path to include interactive task, got %+v", payload.TodayTasks)
	}
}

func TestServiceBuildHomeUsesDiagnosticProfileWhenPracticeEvidenceIsMissing(t *testing.T) {
	now := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	service := NewService(&serviceRepoStub{
		exam: &ExamOverview{ExamID: "exam-1", ExamName: "CFA 一级"},
		diagnostic: &DiagnosticSummary{
			HasCompleted: true,
			KnowledgePoints: []diagnosticpkg.KnowledgeMastery{
				{
					KnowledgePointID:   "kp-diagnostic",
					KnowledgePointName: "财报比率分析",
					MasteryScore:       38,
					ConfidenceScore:    82,
					ForgettingDueAt:    "2026-05-25",
				},
			},
		},
	})

	payload, err := service.BuildHome(context.Background(), "user-1", "exam-1", now)
	if err != nil {
		t.Fatalf("BuildHome returned error: %v", err)
	}

	if len(payload.WeakPoints) == 0 || payload.WeakPoints[0].Source != "diagnostic_profile" {
		t.Fatalf("expected diagnostic profile to populate weak points, got %+v", payload.WeakPoints)
	}
	if payload.WeakPoints[0].ConfidenceScore != 82 || payload.WeakPoints[0].ForgettingDueAt != "2026-05-25" {
		t.Fatalf("expected diagnostic mastery fields to be preserved, got %+v", payload.WeakPoints[0])
	}
}

func TestServiceBuildHomeFlagsAccuracyVolatilityOverRecentTenSessions(t *testing.T) {
	now := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	sessions := make([]profilepkg.PracticeSessionSummary, 0, 10)
	accuracies := []int{92, 90, 88, 86, 84, 82, 80, 78, 74, 52}
	for index, accuracy := range accuracies {
		sessions = append(sessions, profilepkg.PracticeSessionSummary{
			ID:            "session-" + string(rune('a'+index)),
			Status:        "completed",
			TotalCount:    10,
			AnsweredCount: 10,
			CorrectCount:  accuracy / 10,
			CreatedAt:     now.Add(time.Duration(-index) * time.Hour),
		})
	}

	service := NewService(&serviceRepoStub{
		exam:     &ExamOverview{ExamID: "exam-1", ExamName: "CFA 一级"},
		sessions: sessions,
	})

	payload, err := service.BuildHome(context.Background(), "user-1", "exam-1", now)
	if err != nil {
		t.Fatalf("BuildHome returned error: %v", err)
	}

	if payload.VolatilityAlert == nil || !payload.VolatilityAlert.ShouldRetest {
		t.Fatalf("expected volatility alert when recent 10 sessions swing by >=20 points, got %+v", payload.VolatilityAlert)
	}
	recommendations, err := service.BuildRecommendations(context.Background(), "user-1", "exam-1", now)
	if err != nil {
		t.Fatalf("BuildRecommendations returned error: %v", err)
	}
	if len(recommendations) == 0 || recommendations[0].TaskType != "diagnostic" {
		t.Fatalf("expected missing diagnostic recommendation to lead, got %+v", recommendations)
	}
}
