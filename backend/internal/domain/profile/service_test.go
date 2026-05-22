package profile

import (
	"context"
	"testing"
	"time"
)

type serviceRepoStub struct {
	walletBalance int
	streak        *StreakStats
	sessions      []PracticeSessionSummary
	kpResults     []KnowledgePointResult
	graph         *KnowledgeGraph
}

func (s *serviceRepoStub) GetWalletBalance(context.Context, string) (int, error) {
	return s.walletBalance, nil
}

func (s *serviceRepoStub) GetStreakStats(context.Context, string, string) (*StreakStats, error) {
	return s.streak, nil
}

func (s *serviceRepoStub) ListPracticeSessions(context.Context, string, string, int) ([]PracticeSessionSummary, error) {
	return s.sessions, nil
}

func (s *serviceRepoStub) ListKnowledgePointResults(context.Context, string, string) ([]KnowledgePointResult, error) {
	return s.kpResults, nil
}

func (s *serviceRepoStub) BuildKnowledgeGraph(context.Context, string) (*KnowledgeGraph, error) {
	if s.graph != nil {
		return s.graph, nil
	}
	return &KnowledgeGraph{Nodes: []KnowledgeGraphNode{}, Edges: []KnowledgeGraphEdge{}}, nil
}

func TestServiceBuildProfileAggregatesRealData(t *testing.T) {
	now := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	repo := &serviceRepoStub{
		walletBalance: 26,
		streak: &StreakStats{
			CurrentStreak: 3,
			BestStreak:    7,
			Status:        "active",
		},
		sessions: []PracticeSessionSummary{
			{
				ID:                   "session-1",
				CreatedAt:            now.Add(-24 * time.Hour),
				TotalCount:           10,
				AnsweredCount:        10,
				CorrectCount:         7,
				XPEarned:             70,
				CoinsEarned:          7,
				TotalDurationSeconds: 900,
			},
			{
				ID:                   "session-2",
				CreatedAt:            now.Add(-10 * 24 * time.Hour),
				TotalCount:           8,
				AnsweredCount:        8,
				CorrectCount:         6,
				XPEarned:             60,
				CoinsEarned:          6,
				TotalDurationSeconds: 600,
			},
			{
				ID:                   "session-3",
				CreatedAt:            now.Add(-40 * 24 * time.Hour),
				TotalCount:           5,
				AnsweredCount:        5,
				CorrectCount:         5,
				XPEarned:             50,
				CoinsEarned:          5,
				TotalDurationSeconds: 300,
			},
		},
		kpResults: []KnowledgePointResult{
			{KnowledgePointID: "kp-1", KnowledgePointName: "年金终值", Correct: false},
			{KnowledgePointID: "kp-1", KnowledgePointName: "年金终值", Correct: false},
			{KnowledgePointID: "kp-1", KnowledgePointName: "年金终值", Correct: true},
			{KnowledgePointID: "kp-2", KnowledgePointName: "收益率曲线", Correct: true},
			{KnowledgePointID: "kp-2", KnowledgePointName: "收益率曲线", Correct: true},
			{KnowledgePointID: "kp-3", KnowledgePointName: "信用利差", Correct: false},
		},
		graph: &KnowledgeGraph{
			Nodes: []KnowledgeGraphNode{
				{ID: "exam:exam-1", Type: "exam", RefID: "exam-1", Label: "基金从业"},
				{ID: "kp:kp-1", Type: "knowledge_point", RefID: "kp-1", Label: "年金终值"},
				{ID: "kp:kp-x", Type: "knowledge_point", RefID: "kp-x", Label: "未学习知识点"},
			},
			Edges: []KnowledgeGraphEdge{
				{ID: "exam-kp-1", Source: "exam:exam-1", Target: "kp:kp-1", Type: "contains"},
				{ID: "exam-kp-x", Source: "exam:exam-1", Target: "kp:kp-x", Type: "contains"},
			},
		},
	}

	service := NewService(repo)
	payload, err := service.BuildProfile(context.Background(), "user-1", "exam-1", now)
	if err != nil {
		t.Fatalf("BuildProfile returned error: %v", err)
	}

	if payload.Archive.TotalSessions != 3 {
		t.Fatalf("expected 3 sessions, got %+v", payload.Archive)
	}
	if payload.Archive.TotalAnswered != 23 || payload.Archive.TotalCorrect != 18 {
		t.Fatalf("expected aggregate answered/correct 23/18, got %+v", payload.Archive)
	}
	if payload.Archive.Accuracy != 78 {
		t.Fatalf("expected overall accuracy 78, got %+v", payload.Archive)
	}
	if payload.Archive.CoinsBalance != 26 || payload.Archive.CurrentStreak != 3 || payload.Archive.BestStreak != 7 {
		t.Fatalf("expected wallet/streaks to be preserved, got %+v", payload.Archive)
	}
	if payload.Records.MonthlySessions != 2 || payload.Records.MonthlyAnswered != 18 || payload.Records.MonthlyXP != 130 {
		t.Fatalf("expected monthly stats from recent sessions, got %+v", payload.Records)
	}
	if len(payload.Records.RecentSessions) != 3 || payload.Records.RecentSessions[0].ID != "session-1" {
		t.Fatalf("expected recent sessions to preserve newest-first ordering, got %+v", payload.Records.RecentSessions)
	}
	if len(payload.Portrait.WeakPoints) < 2 {
		t.Fatalf("expected weak points to be computed, got %+v", payload.Portrait)
	}
	if payload.Portrait.WeakPoints[0].KnowledgePointID != "kp-3" || payload.Portrait.WeakPoints[0].Accuracy != 0 {
		t.Fatalf("expected weakest point first, got %+v", payload.Portrait.WeakPoints)
	}
	if len(payload.Portrait.StrongPoints) == 0 || payload.Portrait.StrongPoints[0].KnowledgePointID != "kp-2" {
		t.Fatalf("expected strongest point ranking, got %+v", payload.Portrait.StrongPoints)
	}
	if len(payload.Portrait.KnowledgeGraph.Nodes) != 3 {
		t.Fatalf("expected graph nodes to be preserved, got %+v", payload.Portrait.KnowledgeGraph.Nodes)
	}
	var learnedPoint, untouchedPoint *KnowledgeGraphNode
	for i := range payload.Portrait.KnowledgeGraph.Nodes {
		node := &payload.Portrait.KnowledgeGraph.Nodes[i]
		switch node.RefID {
		case "kp-1":
			learnedPoint = node
		case "kp-x":
			untouchedPoint = node
		}
	}
	if learnedPoint == nil || learnedPoint.Mastery == nil || *learnedPoint.Mastery != 33 || learnedPoint.Attempts != 3 {
		t.Fatalf("expected learned point mastery to be merged, got %+v", learnedPoint)
	}
	if untouchedPoint == nil || untouchedPoint.Mastery != nil || untouchedPoint.Attempts != 0 {
		t.Fatalf("expected untouched point to stay unlearned, got %+v", untouchedPoint)
	}
}
