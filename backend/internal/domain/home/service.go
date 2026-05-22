package home

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	cachepkg "foco/backend/api/internal/cache"
	diagnosticpkg "foco/backend/api/internal/domain/diagnostic"
	profilepkg "foco/backend/api/internal/domain/profile"
)

type ExamOverview struct {
	ExamID           string
	ExamName         string
	NextExamDate     *time.Time
	NextNextExamDate *time.Time
}

type TodayTask struct {
	ID                string         `json:"id"`
	Title             string         `json:"title"`
	Type              string         `json:"type"`
	TypeLabel         string         `json:"typeLabel"`
	Duration          string         `json:"duration"`
	Status            string         `json:"status"`
	ProgressPercent   int            `json:"progressPercent"`
	XPRewardPreview   int            `json:"xpRewardPreview"`
	CoinRewardPreview int            `json:"coinRewardPreview"`
	Reason            string         `json:"reason"`
	ActionHref        string         `json:"actionHref"`
	ActionTarget      map[string]any `json:"actionTarget,omitempty"`
}

type WeeklyActivity struct {
	Day            string `json:"day"`
	Date           string `json:"date"`
	Minutes        int    `json:"minutes"`
	CompletedTasks int    `json:"completedTasks"`
	QuestionCount  int    `json:"questionCount"`
	XPEarned       int    `json:"xpEarned"`
	KeptStreak     bool   `json:"keptStreak"`
}

type WeakPoint struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Mastery         int    `json:"mastery"`
	ConfidenceScore int    `json:"confidenceScore"`
	Attempts        int    `json:"attempts"`
	CorrectCount    int    `json:"correctCount"`
	ForgettingDueAt string `json:"forgettingDueAt,omitempty"`
	LastEvidenceAt  string `json:"lastEvidenceAt,omitempty"`
	ReviewStage     string `json:"reviewStage"`
	ReviewDue       bool   `json:"reviewDue"`
	IntervalDays    int    `json:"intervalDays"`
	Source          string `json:"source"`
}

type RecommendationReason struct {
	ReasonCode string         `json:"reason_code"`
	ReasonText string         `json:"reason_text"`
	Evidence   map[string]any `json:"evidence"`
}

type Recommendation struct {
	ID               string                 `json:"id"`
	TaskType         string                 `json:"task_type"`
	TaskTypeLabel    string                 `json:"task_type_label"`
	Title            string                 `json:"title"`
	EstimatedMinutes int                    `json:"estimated_minutes"`
	PriorityScore    int                    `json:"priority_score"`
	Status           string                 `json:"status"`
	ActionHref       string                 `json:"action_href"`
	ActionTarget     map[string]any         `json:"action_target,omitempty"`
	Reasons          []RecommendationReason `json:"reasons"`
}

type ProgressStats struct {
	CompletedQuestions int    `json:"completedQuestions"`
	Accuracy           int    `json:"accuracy"`
	LastStudiedAt      string `json:"lastStudiedAt,omitempty"`
	WrongCount         int    `json:"wrongCount"`
	TotalXP            int    `json:"totalXP"`
	CompletedSessions  int    `json:"completedSessions"`
}

type LearningReport struct {
	PeriodLabel  string   `json:"periodLabel"`
	CoreMetrics  []Metric `json:"coreMetrics"`
	WeakSummary  string   `json:"weakSummary"`
	TrendSummary string   `json:"trendSummary"`
	NextActions  []string `json:"nextActions"`
	GeneratedAt  string   `json:"generatedAt"`
}

type Metric struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Delta string `json:"delta,omitempty"`
}

type Payload struct {
	ExamName          string             `json:"examName"`
	NextExamDate      string             `json:"nextExamDate,omitempty"`
	NextNextExamDate  string             `json:"nextNextExamDate,omitempty"`
	CountdownDays     *int               `json:"countdownDays,omitempty"`
	Level             int                `json:"level"`
	LevelTitle        string             `json:"levelTitle"`
	Coins             int                `json:"coins"`
	Streak            int                `json:"streak"`
	XPCurrent         int                `json:"xpCurrent"`
	XPTarget          int                `json:"xpTarget"`
	DiagnosticSummary *DiagnosticSummary `json:"diagnosticSummary,omitempty"`
	VolatilityAlert   *VolatilityAlert   `json:"volatilityAlert,omitempty"`
	TodayTasks        []TodayTask        `json:"todayTasks"`
	WeeklyActivity    []WeeklyActivity   `json:"weeklyActivity"`
	WeakPoints        []WeakPoint        `json:"weakPoints"`
	Recommendations   []Recommendation   `json:"recommendations"`
	ProgressStats     ProgressStats      `json:"progressStats"`
	LearningReport    LearningReport     `json:"learningReport"`
}

type DiagnosticSummary = diagnosticpkg.ProfileSummary

type VolatilityAlert struct {
	ShouldRetest bool   `json:"shouldRetest"`
	Message      string `json:"message"`
	MinAccuracy  int    `json:"minAccuracy"`
	MaxAccuracy  int    `json:"maxAccuracy"`
}

type Repository interface {
	GetExamOverview(ctx context.Context, examID string) (*ExamOverview, error)
	GetWalletBalance(ctx context.Context, userID string) (int, error)
	GetStreakStats(ctx context.Context, userID, examID string) (*profilepkg.StreakStats, error)
	GetLatestDiagnosticSummary(ctx context.Context, userID, examID string) (*DiagnosticSummary, error)
	ListPracticeSessions(ctx context.Context, userID, examID string, limit int) ([]profilepkg.PracticeSessionSummary, error)
	ListKnowledgePointResults(ctx context.Context, userID, examID string) ([]profilepkg.KnowledgePointResult, error)
	CountCompletedInteractiveAttempts(ctx context.Context, userID, examID string, since, until time.Time) (int, error)
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

func (s *Service) BuildHome(ctx context.Context, userID, examID string, now time.Time) (*Payload, error) {
	if s.cache != nil {
		var result *Payload
		key := "home:" + now.Format("2006-01-02")
		err := s.cache.GetJSON(ctx, learnerNamespace(userID, examID), key, 90*time.Second, &result, func(ctx context.Context) (any, error) {
			return s.buildHomeUncached(ctx, userID, examID, now)
		})
		return result, err
	}
	return s.buildHomeUncached(ctx, userID, examID, now)
}

func (s *Service) buildHomeUncached(ctx context.Context, userID, examID string, now time.Time) (*Payload, error) {
	exam, err := s.repo.GetExamOverview(ctx, examID)
	if err != nil {
		return nil, err
	}
	coins, err := s.repo.GetWalletBalance(ctx, userID)
	if err != nil {
		return nil, err
	}
	streak, err := s.repo.GetStreakStats(ctx, userID, examID)
	if err != nil {
		return nil, err
	}
	diagnosticSummary, err := s.repo.GetLatestDiagnosticSummary(ctx, userID, examID)
	if err != nil {
		return nil, err
	}
	sessions, err := s.repo.ListPracticeSessions(ctx, userID, examID, 10000)
	if err != nil {
		return nil, err
	}
	kpResults, err := s.repo.ListKnowledgePointResults(ctx, userID, examID)
	if err != nil {
		return nil, err
	}
	dayStart, dayEnd := dayBounds(now)
	completedInteractiveAttempts, err := s.repo.CountCompletedInteractiveAttempts(ctx, userID, examID, dayStart, dayEnd)
	if err != nil {
		return nil, err
	}

	totalXP := 0
	currentStreak := 0
	if streak != nil {
		currentStreak = streak.CurrentStreak
	}
	for _, session := range sessions {
		totalXP += session.XPEarned
	}

	level := totalXP/100 + 1
	xpCurrent := totalXP % 100
	xpTarget := 100

	weakPoints := buildWeakPoints(kpResults, diagnosticSummary, now)
	countdownDays := countdownDays(now, exam)
	weeklyActivity := buildWeeklyActivity(now, sessions)
	progressStats := buildProgressStats(sessions, kpResults)
	recommendations := buildRecommendations(exam, sessions, weakPoints, diagnosticSummary, countdownDays, buildVolatilityAlert(sessions, diagnosticCompletedAt(diagnosticSummary)))
	todayTasks := buildTodayTasks(now, exam, sessions, weakPoints, countdownDays, recommendations, completedInteractiveAttempts)

	return &Payload{
		ExamName:          valueOrEmpty(exam, func(item *ExamOverview) string { return item.ExamName }),
		NextExamDate:      timeString(exam, func(item *ExamOverview) *time.Time { return item.NextExamDate }),
		NextNextExamDate:  timeString(exam, func(item *ExamOverview) *time.Time { return item.NextNextExamDate }),
		CountdownDays:     countdownDays,
		Level:             level,
		LevelTitle:        levelTitle(level),
		Coins:             coins,
		Streak:            currentStreak,
		XPCurrent:         xpCurrent,
		XPTarget:          xpTarget,
		DiagnosticSummary: diagnosticSummary,
		VolatilityAlert:   buildVolatilityAlert(sessions, diagnosticCompletedAt(diagnosticSummary)),
		TodayTasks:        todayTasks,
		WeeklyActivity:    weeklyActivity,
		WeakPoints:        weakPoints,
		Recommendations:   recommendations,
		ProgressStats:     progressStats,
		LearningReport:    buildLearningReport(now, weeklyActivity, progressStats, weakPoints, recommendations),
	}, nil
}

func (s *Service) BuildRecommendations(ctx context.Context, userID, examID string, now time.Time) ([]Recommendation, error) {
	payload, err := s.BuildHome(ctx, userID, examID, now)
	if err != nil {
		return nil, err
	}
	return payload.Recommendations, nil
}

func learnerNamespace(userID, examID string) string {
	return "learner:user:" + userID + ":exam:" + examID
}

func buildTodayTasks(now time.Time, exam *ExamOverview, sessions []profilepkg.PracticeSessionSummary, weakPoints []WeakPoint, countdownDays *int, recommendations []Recommendation, completedInteractiveAttempts int) []TodayTask {
	tasks := make([]TodayTask, 0, 4)

	if len(sessions) > 0 {
		latest := sessions[0]
		if latest.Status == "in_progress" && latest.AnsweredCount < latest.TotalCount {
			remaining := latest.TotalCount - latest.AnsweredCount
			tasks = append(tasks, TodayTask{
				ID:                "continue:" + latest.ID,
				Title:             "继续上次练习",
				Type:              "continue",
				TypeLabel:         "继续练习",
				Duration:          fmt.Sprintf("%d 分钟", max(remaining*2, 10)),
				Status:            "in_progress",
				ProgressPercent:   percentage(latest.AnsweredCount, latest.TotalCount),
				XPRewardPreview:   max(remaining*5, 20),
				CoinRewardPreview: max(remaining/2, 2),
				Reason:            "检测到未完成会话，继续完成能保留上下文并推进今日路径。",
				ActionHref:        "/practice/" + latest.ID,
				ActionTarget: map[string]any{
					"session_id": latest.ID,
				},
			})
		}
	}

	for _, point := range weakPoints {
		if len(tasks) >= 3 {
			break
		}
		if point.Source == "diagnostic_profile" && !point.ReviewDue {
			continue
		}
		tasks = append(tasks, TodayTask{
			ID:                "weak:" + point.ID,
			Title:             "间隔复习：" + point.Name,
			Type:              "spaced_review",
			TypeLabel:         "记忆复习",
			Duration:          "10 分钟",
			Status:            "pending",
			ProgressPercent:   0,
			XPRewardPreview:   40,
			CoinRewardPreview: 4,
			Reason:            ebbinghausTaskReason(point),
			ActionHref:        "/practice/setup",
			ActionTarget: map[string]any{
				"knowledge_point_id": point.ID,
				"review_stage":       point.ReviewStage,
				"forgetting_due_at":  point.ForgettingDueAt,
			},
		})
	}

	if countdownDays != nil && *countdownDays <= 30 {
		tasks = append(tasks, TodayTask{
			ID:                "sprint",
			Title:             "考前冲刺：" + valueOrEmpty(exam, func(item *ExamOverview) string { return item.ExamName }),
			Type:              "sprint",
			TypeLabel:         "考前冲刺",
			Duration:          "20 分钟",
			Status:            "pending",
			ProgressPercent:   0,
			XPRewardPreview:   60,
			CoinRewardPreview: 6,
			Reason:            fmt.Sprintf("距离考试仅剩 %d 天，适合提高综合题量。", *countdownDays),
			ActionHref:        "/practice/setup",
		})
	}

	for _, recommendation := range recommendations {
		if len(tasks) >= 4 {
			break
		}
		if recommendation.TaskType == "continue_practice" || recommendation.TaskType == "spaced_review" || recommendation.TaskType == "exam_sprint" {
			continue
		}
		tasks = append(tasks, TodayTask{
			ID:                "rec:" + recommendation.ID,
			Title:             recommendation.Title,
			Type:              recommendation.TaskType,
			TypeLabel:         recommendation.TaskTypeLabel,
			Duration:          fmt.Sprintf("%d 分钟", recommendation.EstimatedMinutes),
			Status:            "pending",
			ProgressPercent:   0,
			XPRewardPreview:   max(recommendation.EstimatedMinutes*3, 30),
			CoinRewardPreview: max(recommendation.EstimatedMinutes/5, 3),
			Reason:            firstReasonText(recommendation.Reasons),
			ActionHref:        recommendation.ActionHref,
			ActionTarget:      recommendation.ActionTarget,
		})
	}

	for len(tasks) < 4 {
		tasks = append(tasks, TodayTask{
			ID:                fmt.Sprintf("mixed:%d", len(tasks)+1),
			Title:             "混合练习：" + valueOrEmpty(exam, func(item *ExamOverview) string { return item.ExamName }),
			Type:              "mixed",
			TypeLabel:         "新题练习",
			Duration:          "15 分钟",
			Status:            "pending",
			ProgressPercent:   0,
			XPRewardPreview:   45,
			CoinRewardPreview: 4,
			Reason:            "保持题型覆盖，给推荐引擎补充新的答题证据。",
			ActionHref:        "/practice/setup",
		})
	}

	return markCompletedTodayTasks(tasks, sessions, now, completedInteractiveAttempts)
}

func markCompletedTodayTasks(tasks []TodayTask, sessions []profilepkg.PracticeSessionSummary, now time.Time, completedInteractiveAttempts int) []TodayTask {
	completedCount := completedPracticeSessionsOn(sessions, now)
	if completedCount == 0 && completedInteractiveAttempts == 0 {
		return tasks
	}

	out := append([]TodayTask(nil), tasks...)
	for i := range out {
		if completedCount == 0 || out[i].Status == "completed" || !isPracticePathTask(out[i]) {
			continue
		}
		out[i].Status = "completed"
		out[i].ProgressPercent = 100
		completedCount--
	}
	for i := range out {
		if completedInteractiveAttempts == 0 || out[i].Status == "completed" || !isInteractivePathTask(out[i]) {
			continue
		}
		out[i].Status = "completed"
		out[i].ProgressPercent = 100
		completedInteractiveAttempts--
	}
	return out
}

func completedPracticeSessionsOn(sessions []profilepkg.PracticeSessionSummary, now time.Time) int {
	count := 0
	for _, session := range sessions {
		if session.Status != "completed" {
			continue
		}
		completedAt := session.CreatedAt
		if session.CompletedAt != nil {
			completedAt = *session.CompletedAt
		}
		if sameDay(completedAt, now) {
			count++
		}
	}
	return count
}

func sameDay(left time.Time, right time.Time) bool {
	localLeft := left.In(right.Location())
	return localLeft.Year() == right.Year() && localLeft.YearDay() == right.YearDay()
}

func isPracticePathTask(task TodayTask) bool {
	return task.ActionHref == "/practice/setup" || strings.HasPrefix(task.ActionHref, "/practice/")
}

func isInteractivePathTask(task TodayTask) bool {
	return task.ActionHref == "/labs" || strings.HasPrefix(task.ActionHref, "/labs/")
}

func dayBounds(now time.Time) (time.Time, time.Time) {
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return start, start.AddDate(0, 0, 1)
}

func buildWeeklyActivity(now time.Time, sessions []profilepkg.PracticeSessionSummary) []WeeklyActivity {
	type dayStat struct {
		minutes        int
		completedTasks int
		questionCount  int
		xpEarned       int
	}

	byDate := make(map[string]dayStat, 7)
	windowStart := now.AddDate(0, 0, -6)
	for _, session := range sessions {
		if session.CreatedAt.Before(windowStart) {
			continue
		}
		key := session.CreatedAt.Format("2006-01-02")
		existing := byDate[key]
		existing.minutes += durationMinutes(session.TotalDurationSeconds)
		existing.questionCount += session.AnsweredCount
		existing.xpEarned += session.XPEarned
		if session.Status == "completed" {
			existing.completedTasks++
		}
		byDate[key] = existing
	}

	result := make([]WeeklyActivity, 0, 7)
	for i := 6; i >= 0; i-- {
		day := now.AddDate(0, 0, -i)
		key := day.Format("2006-01-02")
		stat := byDate[key]
		result = append(result, WeeklyActivity{
			Day:            weekdayLabel(day.Weekday()),
			Date:           key,
			Minutes:        stat.minutes,
			CompletedTasks: stat.completedTasks,
			QuestionCount:  stat.questionCount,
			XPEarned:       stat.xpEarned,
			KeptStreak:     stat.minutes > 0 || stat.questionCount > 0 || stat.completedTasks > 0,
		})
	}
	return result
}

func buildWeakPoints(results []profilepkg.KnowledgePointResult, diagnosticSummary *DiagnosticSummary, now time.Time) []WeakPoint {
	type weakPointStat struct {
		id             string
		name           string
		attempts       int
		correct        int
		lastEvidenceAt time.Time
	}

	byPoint := map[string]*weakPointStat{}
	for _, result := range results {
		stat, ok := byPoint[result.KnowledgePointID]
		if !ok {
			stat = &weakPointStat{
				id:   result.KnowledgePointID,
				name: result.KnowledgePointName,
			}
			byPoint[result.KnowledgePointID] = stat
		}
		stat.attempts++
		if result.Correct {
			stat.correct++
		}
		if result.SubmittedAt.After(stat.lastEvidenceAt) {
			stat.lastEvidenceAt = result.SubmittedAt
		}
	}

	weakPoints := make([]WeakPoint, 0, len(byPoint))
	for _, stat := range byPoint {
		mastery := 0
		if stat.attempts > 0 {
			mastery = stat.correct * 100 / stat.attempts
		}
		intervalDays, forgettingDueAt, reviewStage, reviewDue := ebbinghausReviewWindow(now, stat.lastEvidenceAt, mastery, stat.attempts)
		weakPoints = append(weakPoints, WeakPoint{
			ID:              stat.id,
			Name:            stat.name,
			Mastery:         mastery,
			ConfidenceScore: min(95, 20+stat.attempts*15),
			Attempts:        stat.attempts,
			CorrectCount:    stat.correct,
			ForgettingDueAt: forgettingDueAt,
			LastEvidenceAt:  timeStringFromValue(stat.lastEvidenceAt),
			ReviewStage:     reviewStage,
			ReviewDue:       reviewDue,
			IntervalDays:    intervalDays,
			Source:          "practice_answers",
		})
	}

	if diagnosticSummary != nil {
		existing := make(map[string]bool, len(weakPoints))
		for _, point := range weakPoints {
			existing[point.ID] = true
		}
		for _, point := range diagnosticSummary.KnowledgePoints {
			if existing[point.KnowledgePointID] {
				continue
			}
			weakPoints = append(weakPoints, WeakPoint{
				ID:              point.KnowledgePointID,
				Name:            point.KnowledgePointName,
				Mastery:         point.MasteryScore,
				ConfidenceScore: point.ConfidenceScore,
				ForgettingDueAt: point.ForgettingDueAt,
				ReviewStage:     reviewStageFromDueDate(now, point.ForgettingDueAt),
				ReviewDue:       isDueDate(now, point.ForgettingDueAt),
				Source:          "diagnostic_profile",
			})
			existing[point.KnowledgePointID] = true
		}

		for index, pointID := range diagnosticSummary.RecommendedKnowledgePointIDs {
			if existing[pointID] {
				continue
			}
			name := pointID
			if index < len(diagnosticSummary.RecommendedKnowledgePointNames) {
				name = diagnosticSummary.RecommendedKnowledgePointNames[index]
			}
			weakPoints = append(weakPoints, WeakPoint{
				ID:              pointID,
				Name:            name,
				Mastery:         45,
				ConfidenceScore: 60,
				ForgettingDueAt: now.AddDate(0, 0, 3).Format("2006-01-02"),
				ReviewStage:     "第 3 天复习",
				ReviewDue:       false,
				IntervalDays:    3,
				Source:          "diagnostic_profile",
			})
			existing[pointID] = true
		}
	}

	slices.SortFunc(weakPoints, func(a, b WeakPoint) int {
		if a.ReviewDue != b.ReviewDue {
			if a.ReviewDue {
				return -1
			}
			return 1
		}
		if a.Mastery != b.Mastery {
			return a.Mastery - b.Mastery
		}
		return compareStrings(a.Name, b.Name)
	})
	if len(weakPoints) > 4 {
		weakPoints = weakPoints[:4]
	}
	return weakPoints
}

func buildRecommendations(exam *ExamOverview, sessions []profilepkg.PracticeSessionSummary, weakPoints []WeakPoint, diagnosticSummary *DiagnosticSummary, countdownDays *int, volatilityAlert *VolatilityAlert) []Recommendation {
	recommendations := make([]Recommendation, 0, 6)

	if diagnosticSummary == nil {
		recommendations = append(recommendations, Recommendation{
			ID:               "diagnostic-initial",
			TaskType:         "diagnostic",
			TaskTypeLabel:    "诊断测评",
			Title:            "完成入门诊断，生成初始强弱项",
			EstimatedMinutes: 25,
			PriorityScore:    96,
			Status:           "active",
			ActionHref:       "/diagnostic",
			Reasons: []RecommendationReason{{
				ReasonCode: "diagnostic_missing",
				ReasonText: "当前考试还没有诊断画像，推荐先完成测评以校准练习路径。",
				Evidence: map[string]any{
					"has_profile": false,
				},
			}},
		})
	}

	if volatilityAlert != nil && volatilityAlert.ShouldRetest {
		recommendations = append(recommendations, Recommendation{
			ID:               "diagnostic-refresh",
			TaskType:         "diagnostic_refresh",
			TaskTypeLabel:    "重新测评",
			Title:            "重新诊断近期波动知识点",
			EstimatedMinutes: 20,
			PriorityScore:    94,
			Status:           "active",
			ActionHref:       "/diagnostic",
			Reasons: []RecommendationReason{{
				ReasonCode: "diagnostic_refresh_needed",
				ReasonText: "最近练习正确率波动较大，需要重新校准画像。",
				Evidence: map[string]any{
					"min_accuracy": volatilityAlert.MinAccuracy,
					"max_accuracy": volatilityAlert.MaxAccuracy,
				},
			}},
		})
	}

	if len(sessions) > 0 {
		latest := sessions[0]
		if latest.Status == "in_progress" && latest.AnsweredCount < latest.TotalCount {
			recommendations = append(recommendations, Recommendation{
				ID:               "continue-" + latest.ID,
				TaskType:         "continue_practice",
				TaskTypeLabel:    "继续练习",
				Title:            "完成上次未结束的练习",
				EstimatedMinutes: max((latest.TotalCount-latest.AnsweredCount)*2, 10),
				PriorityScore:    90,
				Status:           "active",
				ActionHref:       "/practice/" + latest.ID,
				ActionTarget: map[string]any{
					"session_id": latest.ID,
				},
				Reasons: []RecommendationReason{{
					ReasonCode: "unfinished_path_step",
					ReasonText: "存在未完成练习，完成后可立即获得进度与奖励反馈。",
					Evidence: map[string]any{
						"answered_count": latest.AnsweredCount,
						"total_count":    latest.TotalCount,
						"progress":       percentage(latest.AnsweredCount, latest.TotalCount),
					},
				}},
			})
		}
	}

	for index, point := range weakPoints {
		if index >= 3 {
			break
		}
		score := 88 - index*5
		if point.Source == "diagnostic_profile" && !point.ReviewDue {
			recommendations = append(recommendations, Recommendation{
				ID:               "weak-" + point.ID,
				TaskType:         "diagnostic_weakness",
				TaskTypeLabel:    "弱项巩固",
				Title:            "巩固薄弱知识点：" + point.Name,
				EstimatedMinutes: 12,
				PriorityScore:    score,
				Status:           "active",
				ActionHref:       "/practice/setup",
				ActionTarget: map[string]any{
					"knowledge_point_id": point.ID,
				},
				Reasons: []RecommendationReason{
					{
						ReasonCode: "diagnostic_weak_point",
						ReasonText: fmt.Sprintf("诊断显示 %s 掌握度 %d%%，建议针对性练习巩固。", point.Name, point.Mastery),
						Evidence: map[string]any{
							"knowledge_point_id": point.ID,
							"mastery_score":      point.Mastery,
							"confidence_score":   point.ConfidenceScore,
						},
					},
				},
			})
			continue
		}
		recommendations = append(recommendations, Recommendation{
			ID:               "weak-" + point.ID,
			TaskType:         "spaced_review",
			TaskTypeLabel:    "记忆复习",
			Title:            "按艾宾浩斯窗口复习：" + point.Name,
			EstimatedMinutes: 12,
			PriorityScore:    score,
			Status:           "active",
			ActionHref:       "/practice/setup",
			ActionTarget: map[string]any{
				"knowledge_point_id": point.ID,
			},
			Reasons: []RecommendationReason{
				{
					ReasonCode: "ebbinghaus_review_due",
					ReasonText: ebbinghausTaskReason(point),
					Evidence: map[string]any{
						"knowledge_point_id": point.ID,
						"mastery_score":      point.Mastery,
						"attempts":           point.Attempts,
						"correct_count":      point.CorrectCount,
						"last_evidence_at":   point.LastEvidenceAt,
						"review_stage":       point.ReviewStage,
						"interval_days":      point.IntervalDays,
					},
				},
				{
					ReasonCode: "entering_forgetting_window",
					ReasonText: "该知识点已进入复习窗口，适合用短练习巩固。",
					Evidence: map[string]any{
						"forgetting_due_at": point.ForgettingDueAt,
						"confidence_score":  point.ConfidenceScore,
						"review_due":        point.ReviewDue,
					},
				},
			},
		})

		if point.Mastery < 50 {
			recommendations = append(recommendations, Recommendation{
				ID:               "concept-" + point.ID,
				TaskType:         "concept_review",
				TaskTypeLabel:    "概念回看",
				Title:            "用交互实验重建概念：" + point.Name,
				EstimatedMinutes: 15,
				PriorityScore:    score - 2,
				Status:           "active",
				ActionHref:       "/labs",
				ActionTarget: map[string]any{
					"knowledge_point_id": point.ID,
				},
				Reasons: []RecommendationReason{{
					ReasonCode: "prerequisite_not_mastered",
					ReasonText: "连续薄弱更像概念理解问题，建议先回到交互单元而不是继续盲刷。",
					Evidence: map[string]any{
						"knowledge_point_id": point.ID,
						"mastery_score":      point.Mastery,
					},
				}},
			})
		}
	}

	if countdownDays != nil && *countdownDays <= 30 {
		recommendations = append(recommendations, Recommendation{
			ID:               "exam-sprint",
			TaskType:         "exam_sprint",
			TaskTypeLabel:    "冲刺",
			Title:            "考前综合冲刺：" + valueOrEmpty(exam, func(item *ExamOverview) string { return item.ExamName }),
			EstimatedMinutes: 20,
			PriorityScore:    82,
			Status:           "active",
			ActionHref:       "/practice/setup",
			Reasons: []RecommendationReason{{
				ReasonCode: "exam_urgency",
				ReasonText: fmt.Sprintf("距离考试还有 %d 天，综合练习优先级上升。", *countdownDays),
				Evidence: map[string]any{
					"countdown_days": *countdownDays,
				},
			}},
		})
	}

	recommendations = append(recommendations, Recommendation{
		ID:               "balanced-new-practice",
		TaskType:         "new_practice",
		TaskTypeLabel:    "新题练习",
		Title:            "补充一组新题，更新推荐证据",
		EstimatedMinutes: 15,
		PriorityScore:    70,
		Status:           "active",
		ActionHref:       "/practice/setup",
		Reasons: []RecommendationReason{{
			ReasonCode: "variety_balance",
			ReasonText: "今日路径需要保留新题输入，让画像持续吸收新证据。",
			Evidence: map[string]any{
				"recent_sessions": len(sessions),
			},
		}},
	})

	slices.SortFunc(recommendations, func(a, b Recommendation) int {
		if a.PriorityScore != b.PriorityScore {
			return b.PriorityScore - a.PriorityScore
		}
		return compareStrings(a.ID, b.ID)
	})
	if len(recommendations) > 5 {
		recommendations = recommendations[:5]
	}
	return recommendations
}

func buildProgressStats(sessions []profilepkg.PracticeSessionSummary, results []profilepkg.KnowledgePointResult) ProgressStats {
	stats := ProgressStats{}
	wrongSeen := map[string]bool{}
	totalCorrect := 0
	for _, session := range sessions {
		stats.CompletedQuestions += session.AnsweredCount
		totalCorrect += session.CorrectCount
		stats.TotalXP += session.XPEarned
		if session.Status == "completed" {
			stats.CompletedSessions++
		}
		if stats.LastStudiedAt == "" || session.CreatedAt.Format("2006-01-02") > stats.LastStudiedAt {
			stats.LastStudiedAt = session.CreatedAt.Format("2006-01-02")
		}
	}

	for _, result := range results {
		if result.Correct {
			continue
		}
		wrongSeen[result.KnowledgePointID] = true
	}
	stats.Accuracy = percentage(totalCorrect, stats.CompletedQuestions)
	stats.WrongCount = len(wrongSeen)
	return stats
}

func buildLearningReport(now time.Time, weekly []WeeklyActivity, stats ProgressStats, weakPoints []WeakPoint, recommendations []Recommendation) LearningReport {
	totalMinutes := 0
	totalTasks := 0
	totalQuestions := 0
	activeDays := 0
	for _, day := range weekly {
		totalMinutes += day.Minutes
		totalTasks += day.CompletedTasks
		totalQuestions += day.QuestionCount
		if day.KeptStreak {
			activeDays++
		}
	}

	weakSummary := "当前还没有足够的答题证据形成薄弱点摘要。"
	if len(weakPoints) > 0 {
		weakSummary = fmt.Sprintf("优先关注 %s，当前掌握度 %d%%、置信度 %d%%。", weakPoints[0].Name, weakPoints[0].Mastery, weakPoints[0].ConfidenceScore)
	}

	trendSummary := "本周还没有稳定学习节奏。"
	if activeDays > 0 {
		trendSummary = fmt.Sprintf("最近 7 天学习 %d 天，累计 %d 分钟、完成 %d 个任务、作答 %d 题。", activeDays, totalMinutes, totalTasks, totalQuestions)
	}

	nextActions := make([]string, 0, 3)
	for _, recommendation := range recommendations {
		if len(nextActions) >= 3 {
			break
		}
		nextActions = append(nextActions, recommendation.Title)
	}
	if len(nextActions) == 0 {
		nextActions = append(nextActions, "完成一组新题练习，建立初始学习记录。")
	}

	return LearningReport{
		PeriodLabel: "最近 7 天",
		CoreMetrics: []Metric{
			{Label: "学习时长", Value: fmt.Sprintf("%d 分钟", totalMinutes)},
			{Label: "完成任务", Value: fmt.Sprintf("%d 个", totalTasks)},
			{Label: "作答题数", Value: fmt.Sprintf("%d 题", stats.CompletedQuestions)},
			{Label: "正确率", Value: fmt.Sprintf("%d%%", stats.Accuracy)},
		},
		WeakSummary:  weakSummary,
		TrendSummary: trendSummary,
		NextActions:  nextActions,
		GeneratedAt:  now.Format(time.RFC3339),
	}
}

func diagnosticCompletedAt(summary *DiagnosticSummary) time.Time {
	if summary == nil || !summary.HasCompleted || summary.CompletedAt == "" {
		return time.Time{}
	}
	t, _ := time.Parse(time.RFC3339, summary.CompletedAt)
	return t
}

func buildVolatilityAlert(sessions []profilepkg.PracticeSessionSummary, lastDiagnosticAt time.Time) *VolatilityAlert {
	var recent []profilepkg.PracticeSessionSummary
	for _, s := range sessions {
		if !lastDiagnosticAt.IsZero() && !s.CreatedAt.After(lastDiagnosticAt) {
			continue
		}
		recent = append(recent, s)
	}
	if len(recent) < 2 {
		return nil
	}

	limit := len(recent)
	if limit > 10 {
		limit = 10
	}

	minAccuracy := 101
	maxAccuracy := -1
	for _, session := range recent[:limit] {
		accuracy := sessionAccuracy(session)
		if accuracy < minAccuracy {
			minAccuracy = accuracy
		}
		if accuracy > maxAccuracy {
			maxAccuracy = accuracy
		}
	}

	if maxAccuracy-minAccuracy < 20 {
		return nil
	}

	return &VolatilityAlert{
		ShouldRetest: true,
		Message:      "近期出现水平剧烈波动，建议重新测验。",
		MinAccuracy:  minAccuracy,
		MaxAccuracy:  maxAccuracy,
	}
}

func sessionAccuracy(session profilepkg.PracticeSessionSummary) int {
	if session.AnsweredCount <= 0 {
		return 0
	}
	return session.CorrectCount * 100 / session.AnsweredCount
}

func percentage(current, total int) int {
	if total <= 0 {
		return 0
	}
	return current * 100 / total
}

func ebbinghausReviewWindow(now, lastEvidenceAt time.Time, mastery int, attempts int) (int, string, string, bool) {
	intervalDays := ebbinghausIntervalDays(mastery, attempts)
	if lastEvidenceAt.IsZero() {
		lastEvidenceAt = now
	}
	dueAt := lastEvidenceAt.AddDate(0, 0, intervalDays)
	return intervalDays, dueAt.Format("2006-01-02"), ebbinghausStageLabel(intervalDays), !dueAt.After(now.Add(12 * time.Hour))
}

func ebbinghausIntervalDays(mastery int, attempts int) int {
	if attempts <= 0 {
		return 1
	}
	stage := attempts
	if mastery < 50 {
		stage--
	}
	if mastery >= 80 {
		stage++
	}
	if stage < 1 {
		stage = 1
	}

	intervals := []int{1, 2, 4, 7, 15, 30}
	if stage > len(intervals) {
		stage = len(intervals)
	}
	return intervals[stage-1]
}

func ebbinghausStageLabel(intervalDays int) string {
	switch intervalDays {
	case 1:
		return "第 1 天复习"
	case 2:
		return "第 2 天复习"
	case 4:
		return "第 4 天复习"
	case 7:
		return "第 7 天复习"
	case 15:
		return "第 15 天复习"
	default:
		return "第 30 天复习"
	}
}

func ebbinghausTaskReason(point WeakPoint) string {
	if point.ForgettingDueAt == "" {
		return fmt.Sprintf("%s 当前掌握度 %d%%，需要纳入间隔复习。", point.Name, point.Mastery)
	}
	status := "即将进入"
	if point.ReviewDue {
		status = "已经进入"
	}
	return fmt.Sprintf("%s %s艾宾浩斯%s窗口（到期 %s），当前掌握度 %d%%、置信度 %d%%。", point.Name, status, point.ReviewStage, point.ForgettingDueAt, point.Mastery, point.ConfidenceScore)
}

func reviewStageFromDueDate(now time.Time, dueDate string) string {
	due, err := time.Parse("2006-01-02", dueDate)
	if err != nil {
		return "间隔复习"
	}
	days := int(due.Sub(now).Hours() / 24)
	switch {
	case days <= 1:
		return "第 1 天复习"
	case days <= 2:
		return "第 2 天复习"
	case days <= 4:
		return "第 4 天复习"
	case days <= 7:
		return "第 7 天复习"
	case days <= 15:
		return "第 15 天复习"
	default:
		return "第 30 天复习"
	}
}

func isDueDate(now time.Time, dueDate string) bool {
	due, err := time.Parse("2006-01-02", dueDate)
	if err != nil {
		return false
	}
	return !due.After(now.Add(12 * time.Hour))
}

func timeStringFromValue(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02")
}

func firstReasonText(reasons []RecommendationReason) string {
	if len(reasons) == 0 {
		return "系统根据近期学习记录生成。"
	}
	return reasons[0].ReasonText
}

func countdownDays(now time.Time, exam *ExamOverview) *int {
	if exam == nil || exam.NextExamDate == nil {
		return nil
	}

	days := int(exam.NextExamDate.Sub(now).Hours() / 24)
	if days < 0 {
		days = 0
	}
	return &days
}

func levelTitle(level int) string {
	switch {
	case level >= 20:
		return "冲刺者"
	case level >= 10:
		return "进阶者"
	case level >= 5:
		return "探索者"
	default:
		return "初学者"
	}
}

func weekdayLabel(weekday time.Weekday) string {
	switch weekday {
	case time.Monday:
		return "周一"
	case time.Tuesday:
		return "周二"
	case time.Wednesday:
		return "周三"
	case time.Thursday:
		return "周四"
	case time.Friday:
		return "周五"
	case time.Saturday:
		return "周六"
	default:
		return "周日"
	}
}

func durationMinutes(seconds int) int {
	if seconds <= 0 {
		return 0
	}
	return (seconds + 59) / 60
}

func valueOrEmpty[T any](value *T, getter func(*T) string) string {
	if value == nil {
		return ""
	}
	return getter(value)
}

func timeString[T any](value *T, getter func(*T) *time.Time) string {
	if value == nil {
		return ""
	}
	raw := getter(value)
	if raw == nil {
		return ""
	}
	return raw.Format("2006-01-02")
}

func compareStrings(a, b string) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}
