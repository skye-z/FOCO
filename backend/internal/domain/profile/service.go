package profile

import (
	"context"
	"time"

	cachepkg "foco/backend/api/internal/cache"
)

type StreakStats struct {
	CurrentStreak int    `xorm:"current_streak" json:"current_streak"`
	BestStreak    int    `xorm:"best_streak" json:"best_streak"`
	Status        string `xorm:"status" json:"status"`
}

type PracticeSessionSummary struct {
	ID                   string     `xorm:"id" json:"id"`
	Status               string     `xorm:"status" json:"status"`
	TotalCount           int        `xorm:"total_count" json:"total_count"`
	AnsweredCount        int        `xorm:"answered_count" json:"answered_count"`
	CorrectCount         int        `xorm:"correct_count" json:"correct_count"`
	XPEarned             int        `xorm:"xp_earned" json:"xp_earned"`
	CoinsEarned          int        `xorm:"coins_earned" json:"coins_earned"`
	TotalDurationSeconds int        `xorm:"total_duration_seconds" json:"total_duration_seconds"`
	CreatedAt            time.Time  `xorm:"created_at" json:"created_at"`
	CompletedAt          *time.Time `xorm:"completed_at" json:"completed_at,omitempty"`
}

type KnowledgePointResult struct {
	KnowledgePointID   string    `xorm:"knowledge_point_id" json:"knowledge_point_id"`
	KnowledgePointName string    `xorm:"knowledge_point_name" json:"knowledge_point_name"`
	Correct            bool      `xorm:"correct" json:"correct"`
	SubmittedAt        time.Time `xorm:"submitted_at" json:"submitted_at"`
}

type KnowledgePointStat struct {
	KnowledgePointID   string `json:"knowledge_point_id"`
	KnowledgePointName string `json:"knowledge_point_name"`
	Attempts           int    `json:"attempts"`
	CorrectCount       int    `json:"correct_count"`
	Accuracy           int    `json:"accuracy"`
}

type Archive struct {
	TotalSessions  int        `json:"total_sessions"`
	TotalAnswered  int        `json:"total_answered"`
	TotalCorrect   int        `json:"total_correct"`
	Accuracy       int        `json:"accuracy"`
	TotalXP        int        `json:"total_xp"`
	CoinsBalance   int        `json:"coins_balance"`
	CurrentStreak  int        `json:"current_streak"`
	BestStreak     int        `json:"best_streak"`
	StreakStatus   string     `json:"streak_status"`
	FirstStudiedAt *time.Time `json:"first_studied_at,omitempty"`
	LastStudiedAt  *time.Time `json:"last_studied_at,omitempty"`
}

type RecordSession struct {
	ID              string     `json:"id"`
	Status          string     `json:"status"`
	TotalCount      int        `json:"total_count"`
	AnsweredCount   int        `json:"answered_count"`
	CorrectCount    int        `json:"correct_count"`
	Accuracy        int        `json:"accuracy"`
	XPEarned        int        `json:"xp_earned"`
	CoinsEarned     int        `json:"coins_earned"`
	DurationMinutes int        `json:"duration_minutes"`
	CreatedAt       time.Time  `json:"created_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
}

type Records struct {
	MonthlySessions int             `json:"monthly_sessions"`
	MonthlyAnswered int             `json:"monthly_answered"`
	MonthlyXP       int             `json:"monthly_xp"`
	RecentSessions  []RecordSession `json:"recent_sessions"`
}

type Portrait struct {
	TotalAttempts   int                  `json:"total_attempts"`
	OverallAccuracy int                  `json:"overall_accuracy"`
	WeakPoints      []KnowledgePointStat `json:"weak_points"`
	StrongPoints    []KnowledgePointStat `json:"strong_points"`
	KnowledgeGraph  KnowledgeGraph       `json:"knowledge_graph"`
}

type Payload struct {
	Archive  Archive  `json:"archive"`
	Records  Records  `json:"records"`
	Portrait Portrait `json:"portrait"`
}

type KnowledgeGraphNode struct {
	ID          string `xorm:"id" json:"id"`
	Type        string `xorm:"type" json:"type"`
	RefID       string `xorm:"ref_id" json:"ref_id"`
	Label       string `xorm:"label" json:"label"`
	Description string `xorm:"description" json:"description,omitempty"`
	Group       string `xorm:"group" json:"group,omitempty"`
	Mastery     *int   `xorm:"-" json:"mastery,omitempty"`
	Attempts    int    `xorm:"-" json:"attempts"`
}

type KnowledgeGraphEdge struct {
	ID     string `xorm:"id" json:"id"`
	Source string `xorm:"source" json:"source"`
	Target string `xorm:"target" json:"target"`
	Type   string `xorm:"type" json:"type"`
	Label  string `xorm:"label" json:"label,omitempty"`
}

type KnowledgeGraph struct {
	Nodes []KnowledgeGraphNode `json:"nodes"`
	Edges []KnowledgeGraphEdge `json:"edges"`
}

type Repository interface {
	GetWalletBalance(ctx context.Context, userID string) (int, error)
	GetStreakStats(ctx context.Context, userID, examID string) (*StreakStats, error)
	ListPracticeSessions(ctx context.Context, userID, examID string, limit int) ([]PracticeSessionSummary, error)
	ListKnowledgePointResults(ctx context.Context, userID, examID string) ([]KnowledgePointResult, error)
	BuildKnowledgeGraph(ctx context.Context, examID string) (*KnowledgeGraph, error)
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

func (s *Service) BuildProfile(ctx context.Context, userID, examID string, now time.Time) (*Payload, error) {
	if s.cache != nil {
		var result *Payload
		err := s.cache.GetJSON(ctx, learnerNamespace(userID, examID), "profile", 2*time.Minute, &result, func(ctx context.Context) (any, error) {
			return s.buildProfileUncached(ctx, userID, examID, now)
		})
		return result, err
	}
	return s.buildProfileUncached(ctx, userID, examID, now)
}

func (s *Service) buildProfileUncached(ctx context.Context, userID, examID string, now time.Time) (*Payload, error) {
	walletBalance, err := s.repo.GetWalletBalance(ctx, userID)
	if err != nil {
		return nil, err
	}
	streak, err := s.repo.GetStreakStats(ctx, userID, examID)
	if err != nil {
		return nil, err
	}
	sessions, err := s.repo.ListPracticeSessions(ctx, userID, examID, 50)
	if err != nil {
		return nil, err
	}
	kpResults, err := s.repo.ListKnowledgePointResults(ctx, userID, examID)
	if err != nil {
		return nil, err
	}
	knowledgeGraph, err := s.repo.BuildKnowledgeGraph(ctx, examID)
	if err != nil {
		return nil, err
	}

	archive := buildArchive(walletBalance, streak, sessions)
	records := buildRecords(now, sessions)
	portrait := buildPortrait(kpResults, archive.Accuracy)
	portrait.KnowledgeGraph = buildUserKnowledgeGraph(knowledgeGraph, kpResults)

	return &Payload{
		Archive:  archive,
		Records:  records,
		Portrait: portrait,
	}, nil
}

func learnerNamespace(userID, examID string) string {
	return "learner:user:" + userID + ":exam:" + examID
}
