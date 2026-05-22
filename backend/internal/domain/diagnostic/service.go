package diagnostic

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	cachepkg "foco/backend/api/internal/cache"
	"github.com/google/uuid"
)

var (
	ErrDiagnosticAttemptNotFound = errors.New("diagnostic attempt not found")
	ErrNoDiagnosticQuestions     = errors.New("no diagnostic questions available")
)

type Repository interface {
	FindLatestAttempt(ctx context.Context, userID, examID string) (*Attempt, error)
	FindAttemptByID(ctx context.Context, userID, attemptID string) (*Attempt, error)
	CreateAttempt(ctx context.Context, attempt *Attempt) error
	UpdateAttempt(ctx context.Context, attempt *Attempt) error
	ListDiagnosticQuestions(ctx context.Context, examID string, limit int) ([]Question, error)
	SaveProfile(ctx context.Context, profile *Profile) error
	FindLatestProfile(ctx context.Context, userID, examID string) (*Profile, error)
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

func (s *Service) GetCurrent(ctx context.Context, userID, examID string, now time.Time) (*CurrentPayload, error) {
	if s.cache != nil {
		var result *CurrentPayload
		err := s.cache.GetJSON(ctx, learnerNamespace(userID, examID), "diagnostic-current", 90*time.Second, &result, func(ctx context.Context) (any, error) {
			return s.getCurrentUncached(ctx, userID, examID, now)
		})
		return result, err
	}
	return s.getCurrentUncached(ctx, userID, examID, now)
}

func (s *Service) getCurrentUncached(ctx context.Context, userID, examID string, now time.Time) (*CurrentPayload, error) {
	attempt, err := s.repo.FindLatestAttempt(ctx, userID, examID)
	if err != nil {
		return nil, err
	}

	if attempt != nil && attempt.Status == StatusPending {
		return &CurrentPayload{
			Status:    StatusPending,
			AttemptID: attempt.ID,
			Items:     attempt.Items,
		}, nil
	}

	profile, err := s.repo.FindLatestProfile(ctx, userID, examID)
	if err != nil {
		return nil, err
	}
	if profile != nil {
		return &CurrentPayload{
			Status:  StatusCompleted,
			Summary: &profile.Summary,
		}, nil
	}

	return s.Restart(ctx, userID, examID, TriggerTypeInitialAuto, now)
}

func (s *Service) Restart(ctx context.Context, userID, examID, triggerType string, now time.Time) (*CurrentPayload, error) {
	existing, err := s.repo.FindLatestAttempt(ctx, userID, examID)
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.Status == StatusPending {
		return nil, errors.New("diagnostic attempt already in progress")
	}
	questions, err := s.repo.ListDiagnosticQuestions(ctx, examID, 0)
	if err != nil {
		return nil, err
	}
	if len(questions) == 0 {
		return nil, ErrNoDiagnosticQuestions
	}

	startedAt := now.UTC()
	attempt := &Attempt{
		ID:          uuid.New().String(),
		UserID:      userID,
		ExamID:      examID,
		TriggerType: triggerType,
		Status:      StatusPending,
		StartedAt:   &startedAt,
		Items:       questions,
	}
	if err := s.repo.CreateAttempt(ctx, attempt); err != nil {
		return nil, err
	}
	s.invalidateLearner(ctx, userID, examID)

	return &CurrentPayload{
		Status:    StatusPending,
		AttemptID: attempt.ID,
		Items:     attempt.Items,
	}, nil
}

func (s *Service) Submit(ctx context.Context, input SubmitInput, now time.Time) (*ProfileSummary, error) {
	attempt, err := s.findAttemptByID(ctx, input.UserID, input.AttemptID)
	if err != nil {
		return nil, err
	}
	if attempt == nil {
		return nil, ErrDiagnosticAttemptNotFound
	}

	summary := buildProfileSummary(attempt.Items, input.Answers, now)
	completedAt := now.UTC()
	attempt.Status = StatusCompleted
	attempt.CompletedAt = &completedAt
	attempt.Result = &summary

	if err := s.repo.UpdateAttempt(ctx, attempt); err != nil {
		return nil, err
	}

	profileVersion := 1
	latestProfile, err := s.repo.FindLatestProfile(ctx, input.UserID, attempt.ExamID)
	if err != nil {
		return nil, err
	}
	if latestProfile != nil {
		profileVersion = latestProfile.ProfileVersion + 1
	}

	profile := &Profile{
		ID:              uuid.New().String(),
		UserID:          input.UserID,
		ExamID:          attempt.ExamID,
		ProfileVersion:  profileVersion,
		Summary:         summary,
		ConfidenceScore: overallConfidence(summary.KnowledgePoints),
		ComputedAt:      completedAt,
	}
	if err := s.repo.SaveProfile(ctx, profile); err != nil {
		return nil, err
	}
	s.invalidateLearner(ctx, input.UserID, attempt.ExamID)

	return &summary, nil
}

func (s *Service) invalidateLearner(ctx context.Context, userID, examID string) {
	if s.cache != nil {
		s.cache.Invalidate(ctx, learnerNamespace(userID, examID))
	}
}

func learnerNamespace(userID, examID string) string {
	return "learner:user:" + userID + ":exam:" + examID
}

func (s *Service) findAttemptByID(ctx context.Context, userID, attemptID string) (*Attempt, error) {
	return s.repo.FindAttemptByID(ctx, userID, attemptID)
}

func buildProfileSummary(items []Question, answers map[string][]string, now time.Time) ProfileSummary {
	subjects := map[string]*areaAccumulator{}
	chapters := map[string]*areaAccumulator{}
	knowledgePoints := map[string]*areaAccumulator{}

	totalCorrect := 0
	for _, item := range items {
		userAnswer := normalizeLabels(answers[item.ID])
		isCorrect := slices.Equal(userAnswer, normalizeLabels(item.CorrectLabels))
		if isCorrect {
			totalCorrect++
		}

		accumulate(subjects, item.SubjectID, item.SubjectName, isCorrect)
		accumulate(chapters, item.ChapterID, item.ChapterName, isCorrect)
		for _, kp := range item.KnowledgePoints {
			accumulate(knowledgePoints, kp.ID, kp.Name, isCorrect)
		}
	}

	subjectSummaries, recommendedSubjectIDs, recommendedSubjectNames := buildAreaSummaries(subjects)
	chapterSummaries, recommendedChapterIDs, recommendedChapterNames := buildAreaSummaries(chapters)
	kpMasteries, recommendedKnowledgePointIDs, recommendedKnowledgePointNames := buildKnowledgeMasteries(knowledgePoints, now)

	overallAccuracy := percentage(totalCorrect, len(items))

	return ProfileSummary{
		HasCompleted:                   true,
		CompletedAt:                    now.UTC().Format(time.RFC3339),
		OverallAccuracy:                overallAccuracy,
		SummaryText:                    buildSummaryText(overallAccuracy, recommendedSubjectNames, recommendedChapterNames),
		RecommendedDifficulty:          recommendedDifficulty(overallAccuracy),
		RecommendedSubjectIDs:          recommendedSubjectIDs,
		RecommendedSubjectNames:        recommendedSubjectNames,
		RecommendedChapterIDs:          recommendedChapterIDs,
		RecommendedChapterNames:        recommendedChapterNames,
		RecommendedKnowledgePointIDs:   recommendedKnowledgePointIDs,
		RecommendedKnowledgePointNames: recommendedKnowledgePointNames,
		Subjects:                       subjectSummaries,
		Chapters:                       chapterSummaries,
		KnowledgePoints:                kpMasteries,
	}
}

func accumulate(bucket map[string]*areaAccumulator, id, name string, correct bool) {
	if strings.TrimSpace(id) == "" {
		return
	}
	item, ok := bucket[id]
	if !ok {
		item = &areaAccumulator{id: id, name: name}
		bucket[id] = item
	}
	item.attempts++
	if correct {
		item.correct++
	}
}

type areaAccumulator struct {
	id       string
	name     string
	attempts int
	correct  int
}

func buildAreaSummaries(bucket map[string]*areaAccumulator) ([]AreaSummary, []string, []string) {
	stats := make([]AreaSummary, 0, len(bucket))
	for _, item := range bucket {
		stats = append(stats, AreaSummary{
			ID:       item.id,
			Name:     item.name,
			Accuracy: percentage(item.correct, item.attempts),
			Attempts: item.attempts,
		})
	}
	slices.SortFunc(stats, func(a, b AreaSummary) int {
		if a.Accuracy != b.Accuracy {
			return a.Accuracy - b.Accuracy
		}
		if a.Attempts != b.Attempts {
			return b.Attempts - a.Attempts
		}
		return strings.Compare(a.Name, b.Name)
	})

	recommendedIDs := make([]string, 0, 2)
	recommendedNames := make([]string, 0, 2)
	for index := range stats {
		if index >= 2 {
			break
		}
		if stats[index].Accuracy >= 80 {
			continue
		}
		stats[index].Recommended = true
		recommendedIDs = append(recommendedIDs, stats[index].ID)
		recommendedNames = append(recommendedNames, stats[index].Name)
	}
	return stats, recommendedIDs, recommendedNames
}

func buildKnowledgeMasteries(bucket map[string]*areaAccumulator, now time.Time) ([]KnowledgeMastery, []string, []string) {
	stats := make([]KnowledgeMastery, 0, len(bucket))
	for _, item := range bucket {
		mastery := percentage(item.correct, item.attempts)
		confidence := 20 + mastery/2 + item.attempts*10
		if confidence > 100 {
			confidence = 100
		}
		forgettingDue := now.AddDate(0, 0, 3)
		if mastery >= 60 {
			forgettingDue = now.AddDate(0, 0, 7)
		}
		stats = append(stats, KnowledgeMastery{
			KnowledgePointID:   item.id,
			KnowledgePointName: item.name,
			MasteryScore:       mastery,
			ConfidenceScore:    confidence,
			ForgettingDueAt:    forgettingDue.Format("2006-01-02"),
		})
	}

	slices.SortFunc(stats, func(a, b KnowledgeMastery) int {
		if a.MasteryScore != b.MasteryScore {
			return a.MasteryScore - b.MasteryScore
		}
		if a.ConfidenceScore != b.ConfidenceScore {
			return a.ConfidenceScore - b.ConfidenceScore
		}
		return strings.Compare(a.KnowledgePointName, b.KnowledgePointName)
	})

	recommendedIDs := make([]string, 0, 5)
	recommendedNames := make([]string, 0, 5)
	for index := range stats {
		if index >= 5 {
			break
		}
		recommendedIDs = append(recommendedIDs, stats[index].KnowledgePointID)
		recommendedNames = append(recommendedNames, stats[index].KnowledgePointName)
	}
	return stats, recommendedIDs, recommendedNames
}

func buildSummaryText(overallAccuracy int, weakSubjects, weakChapters []string) string {
	switch {
	case len(weakSubjects) > 0 && len(weakChapters) > 0:
		return "当前诊断显示你在" + weakSubjects[0] + "（如" + weakChapters[0] + "）需要优先补强，建议先从中等难度巩固基础。"
	case len(weakSubjects) > 0:
		return "当前诊断显示你在" + weakSubjects[0] + "需要优先补强，建议先围绕薄弱科目稳定正确率。"
	case overallAccuracy < 60:
		return "当前诊断显示整体基础仍需巩固，建议先从中等难度开始，逐步稳定正确率。"
	default:
		return "当前诊断显示基础较稳，可以结合智能练习继续提升薄弱知识点。"
	}
}

func recommendedDifficulty(overallAccuracy int) string {
	switch {
	case overallAccuracy < 45:
		return "easy"
	case overallAccuracy < 75:
		return "medium"
	default:
		return "hard"
	}
}

func overallConfidence(items []KnowledgeMastery) int {
	if len(items) == 0 {
		return 0
	}
	total := 0
	for _, item := range items {
		total += item.ConfidenceScore
	}
	return total / len(items)
}

func percentage(correct, total int) int {
	if total <= 0 {
		return 0
	}
	return correct * 100 / total
}

func normalizeLabels(labels []string) []string {
	normalized := make([]string, 0, len(labels))
	seen := map[string]struct{}{}
	for _, label := range labels {
		value := strings.ToUpper(strings.TrimSpace(label))
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	slices.Sort(normalized)
	return normalized
}
