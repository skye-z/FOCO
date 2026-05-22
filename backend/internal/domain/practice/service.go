package practice

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	cachepkg "foco/backend/api/internal/cache"
	"github.com/google/uuid"
)

var (
	ErrActiveEnrollmentRequired     = errors.New("active enrollment required")
	ErrDiagnosticRequired           = errors.New("diagnostic required before intelligent practice")
	ErrNoPracticeQuestions          = errors.New("no practice questions available for the selected scope")
	ErrPracticeSessionNotFound      = errors.New("practice session not found")
	ErrPracticeItemNotFound         = errors.New("practice item not found")
	ErrPracticeItemAlreadySubmitted = errors.New("practice item already submitted")
)

type Repository interface {
	FindEnrollment(ctx context.Context, userID, examID string) (*EnrollmentRef, error)
	FindLatestDiagnosticRecommendation(ctx context.Context, userID, examID string) (*DiagnosticRecommendation, error)
	ListCandidateQuestions(ctx context.Context, input CreateSessionInput) ([]CandidateQuestion, error)
	CreateSession(ctx context.Context, session *PracticeSessionRecord, items []PracticeSessionItemRecord) error
	GetSession(ctx context.Context, userID, sessionID string) (*PracticeSessionView, error)
	GetSubmission(ctx context.Context, userID, sessionID, itemID string) (*SubmissionRecord, error)
	SaveSubmission(ctx context.Context, sessionID, itemID string, correct bool, userAnswer []string, durationSeconds, xpEarned, coinsEarned int) error
	ListKnowledgePoints(ctx context.Context, ids []string) ([]KnowledgePointRef, error)
	GetSummary(ctx context.Context, userID, sessionID string) (*SessionSummary, error)
	ListWrongBook(ctx context.Context, filter WrongBookFilter) ([]WrongBookItem, error)
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

func (s *Service) CreateSession(ctx context.Context, input CreateSessionInput) (*PracticeSessionView, error) {
	if input.Count <= 0 {
		input.Count = 15
	}
	if input.Count > 50 {
		input.Count = 50
	}

	if input.Mode == "" {
		input.Mode = "manual"
	}

	enrollment, err := s.repo.FindEnrollment(ctx, input.UserID, input.ExamID)
	if err != nil {
		return nil, err
	}
	if enrollment == nil {
		return nil, ErrActiveEnrollmentRequired
	}

	resolvedScope, err := s.resolveScope(ctx, input)
	if err != nil {
		return nil, err
	}
	attempts := s.buildCandidateAttempts(input, resolvedScope)

	var (
		candidates      []CandidateQuestion
		resolvedAttempt CreateSessionInput
		found           bool
	)
	for _, attempt := range attempts {
		items, err := s.repo.ListCandidateQuestions(ctx, attempt)
		if err != nil {
			return nil, err
		}
		if len(items) == 0 {
			continue
		}
		candidates = items
		resolvedAttempt = attempt
		found = true
		break
	}
	if !found {
		return nil, ErrNoPracticeQuestions
	}

	now := time.Now().UTC()
	scopeJSON, _ := json.Marshal(map[string]any{
		"mode":                resolvedAttempt.Mode,
		"exam_id":             resolvedAttempt.ExamID,
		"question_types":      resolvedAttempt.QuestionTypes,
		"difficulty":          resolvedAttempt.Difficulty,
		"count":               resolvedAttempt.Count,
		"subject_ids":         resolvedAttempt.SubjectIDs,
		"chapter_ids":         resolvedAttempt.ChapterIDs,
		"knowledge_point_ids": resolvedAttempt.KnowledgePointIDs,
	})
	session := &PracticeSessionRecord{
		ID:               uuid.New().String(),
		UserID:           input.UserID,
		ExamID:           input.ExamID,
		ExamEnrollmentID: enrollment.ID,
		ScopeJSON:        string(scopeJSON),
		Status:           "in_progress",
		TotalCount:       len(candidates),
		StartedAt:        now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	items := make([]PracticeSessionItemRecord, 0, len(candidates))
	viewItems := make([]PracticeSessionItemView, 0, len(candidates))
	for index, candidate := range candidates {
		options, labelByOptionID := parseOptions(candidate.Options)
		correctLabels := parseCorrectLabels(candidate.CorrectAnswer, labelByOptionID)
		correctLabelsJSON, _ := json.Marshal(correctLabels)
		knowledgePointIDsJSON, _ := json.Marshal(candidate.KnowledgePointIDs)
		optionsJSON, _ := json.Marshal(options)
		stem := extractText(candidate.Stem)
		explanation := extractText(candidate.Explanation)

		itemID := uuid.New().String()
		items = append(items, PracticeSessionItemRecord{
			ID:                itemID,
			SessionID:         session.ID,
			QuestionID:        candidate.QuestionID,
			QuestionVersionID: candidate.QuestionVersionID,
			SubjectID:         candidate.SubjectID,
			ChapterID:         candidate.ChapterID,
			QuestionType:      candidate.QuestionType,
			Score:             1,
			Position:          index + 1,
			Stem:              stem,
			Options:           string(optionsJSON),
			CorrectLabels:     string(correctLabelsJSON),
			Explanation:       explanation,
			KnowledgePointIDs: string(knowledgePointIDsJSON),
			CreatedAt:         now,
		})
		viewItems = append(viewItems, PracticeSessionItemView{
			ItemID:            itemID,
			QuestionVersionID: candidate.QuestionVersionID,
			QuestionType:      candidate.QuestionType,
			Score:             1,
			Content: PracticeItemContent{
				Stem:    stem,
				Options: options,
			},
		})
	}

	if err := s.repo.CreateSession(ctx, session, items); err != nil {
		return nil, err
	}
	s.invalidateLearner(ctx, input.UserID, input.ExamID)
	s.invalidate(ctx, practiceSessionNamespace(session.ID))

	return &PracticeSessionView{
		SessionID:  session.ID,
		ExamName:   enrollment.ExamName,
		TotalCount: len(viewItems),
		Items:      viewItems,
	}, nil
}

func (s *Service) buildCandidateAttempts(input CreateSessionInput, scope *ResolvedScope) []CreateSessionInput {
	base := input
	base.Difficulty = scope.Difficulty
	base.SubjectIDs = append([]string(nil), scope.SubjectIDs...)
	base.ChapterIDs = append([]string(nil), scope.ChapterIDs...)
	base.KnowledgePointIDs = append([]string(nil), scope.KnowledgePointIDs...)

	attempts := []CreateSessionInput{base}
	if input.Mode != "intelligent" {
		return attempts
	}

	current := base
	relaxations := []func(*CreateSessionInput){
		func(next *CreateSessionInput) {
			next.KnowledgePointIDs = nil
		},
		func(next *CreateSessionInput) {
			next.ChapterIDs = nil
			next.KnowledgePointIDs = nil
		},
		func(next *CreateSessionInput) {
			next.SubjectIDs = nil
			next.ChapterIDs = nil
			next.KnowledgePointIDs = nil
		},
		func(next *CreateSessionInput) {
			next.Difficulty = ""
		},
	}

	for _, relax := range relaxations {
		next := current
		relax(&next)
		if samePracticeScope(current, next) {
			continue
		}
		attempts = append(attempts, next)
		current = next
	}

	return attempts
}

func samePracticeScope(a, b CreateSessionInput) bool {
	return a.Difficulty == b.Difficulty &&
		slices.Equal(a.SubjectIDs, b.SubjectIDs) &&
		slices.Equal(a.ChapterIDs, b.ChapterIDs) &&
		slices.Equal(a.KnowledgePointIDs, b.KnowledgePointIDs)
}

func (s *Service) resolveScope(ctx context.Context, input CreateSessionInput) (*ResolvedScope, error) {
	scope := &ResolvedScope{
		Difficulty:        input.Difficulty,
		SubjectIDs:        append([]string(nil), input.SubjectIDs...),
		ChapterIDs:        append([]string(nil), input.ChapterIDs...),
		KnowledgePointIDs: append([]string(nil), input.KnowledgePointIDs...),
	}
	if input.Mode != "intelligent" {
		return scope, nil
	}

	recommendation, err := s.repo.FindLatestDiagnosticRecommendation(ctx, input.UserID, input.ExamID)
	if err != nil {
		return nil, err
	}
	if recommendation == nil {
		return nil, ErrDiagnosticRequired
	}

	if recommendation.RecommendedDifficulty != "" {
		scope.Difficulty = recommendation.RecommendedDifficulty
	}
	if len(recommendation.RecommendedSubjectIDs) > 0 {
		scope.SubjectIDs = append([]string(nil), recommendation.RecommendedSubjectIDs...)
	}
	if len(recommendation.RecommendedChapterIDs) > 0 {
		scope.ChapterIDs = append([]string(nil), recommendation.RecommendedChapterIDs...)
	}
	if len(recommendation.RecommendedKnowledgePointIDs) > 0 {
		scope.KnowledgePointIDs = append([]string(nil), recommendation.RecommendedKnowledgePointIDs...)
	}
	return scope, nil
}

func (s *Service) GetSession(ctx context.Context, userID, sessionID string) (*PracticeSessionView, error) {
	if s.cache != nil {
		var cached *PracticeSessionView
		err := s.cache.GetJSON(ctx, practiceSessionNamespace(sessionID), "session:"+userID, 10*time.Minute, &cached, func(ctx context.Context) (any, error) {
			return s.getSessionUncached(ctx, userID, sessionID)
		})
		return cached, err
	}
	return s.getSessionUncached(ctx, userID, sessionID)
}

func (s *Service) getSessionUncached(ctx context.Context, userID, sessionID string) (*PracticeSessionView, error) {
	session, err := s.repo.GetSession(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrPracticeSessionNotFound
	}
	return session, nil
}

func (s *Service) ListWrongBook(ctx context.Context, filter WrongBookFilter) ([]WrongBookItem, error) {
	if s.cache == nil {
		return s.repo.ListWrongBook(ctx, filter)
	}
	var result []WrongBookItem
	err := s.cache.GetJSON(ctx, learnerNamespace(filter.UserID, filter.ExamID), "wrong-book:"+wrongBookCacheKey(filter), 2*time.Minute, &result, func(ctx context.Context) (any, error) {
		return s.repo.ListWrongBook(ctx, filter)
	})
	return result, err
}

func (s *Service) SubmitAnswer(ctx context.Context, input SubmitAnswerInput) (*SubmitResult, error) {
	record, err := s.repo.GetSubmission(ctx, input.UserID, input.SessionID, input.ItemID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, ErrPracticeItemNotFound
	}

	selectedAnswers := normalizeAnswer(input.Answer)
	knowledgePoints, err := s.repo.ListKnowledgePoints(ctx, record.KnowledgePointIDs)
	if err != nil {
		return nil, err
	}

	if record.SubmittedAt != nil && record.IsCorrect != nil {
		return nil, ErrPracticeItemAlreadySubmitted
	}

	isCorrect := slices.Equal(selectedAnswers, normalizeLabels(record.CorrectLabels))
	xpEarned := 2
	coinsEarned := 0
	if isCorrect {
		xpEarned = 10
		coinsEarned = 1
	}

	if err := s.repo.SaveSubmission(
		ctx,
		input.SessionID,
		input.ItemID,
		isCorrect,
		selectedAnswers,
		input.DurationSeconds,
		xpEarned,
		coinsEarned,
	); err != nil {
		return nil, err
	}
	s.invalidateLearner(ctx, input.UserID, record.ExamID)
	s.invalidate(ctx, practiceSessionNamespace(input.SessionID))

	return &SubmitResult{
		IsCorrect:       isCorrect,
		CorrectAnswer:   correctAnswerPayload(record.QuestionType, record.CorrectLabels),
		Explanation:     record.Explanation,
		KnowledgePoints: knowledgePoints,
		XpEarned:        xpEarned,
	}, nil
}

func (s *Service) GetSummary(ctx context.Context, userID, sessionID string) (*SessionSummary, error) {
	if s.cache != nil {
		var cached *SessionSummary
		err := s.cache.GetJSON(ctx, practiceSessionNamespace(sessionID), "summary:"+userID, 30*time.Second, &cached, func(ctx context.Context) (any, error) {
			return s.getSummaryUncached(ctx, userID, sessionID)
		})
		return cached, err
	}
	return s.getSummaryUncached(ctx, userID, sessionID)
}

func (s *Service) getSummaryUncached(ctx context.Context, userID, sessionID string) (*SessionSummary, error) {
	summary, err := s.repo.GetSummary(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}
	if summary == nil {
		return nil, ErrPracticeSessionNotFound
	}
	return summary, nil
}

func (s *Service) invalidate(ctx context.Context, namespaces ...string) {
	if s.cache != nil {
		s.cache.Invalidate(ctx, namespaces...)
	}
}

func (s *Service) invalidateLearner(ctx context.Context, userID, examID string) {
	namespaces := []string{learnerNamespace(userID, examID)}
	if examID != "" {
		namespaces = append(namespaces, learnerNamespace(userID, ""))
	}
	s.invalidate(ctx, namespaces...)
}

func learnerNamespace(userID, examID string) string {
	return "learner:user:" + userID + ":exam:" + examID
}

func practiceSessionNamespace(sessionID string) string {
	return "practice:session:" + sessionID
}

func wrongBookCacheKey(filter WrongBookFilter) string {
	return filter.ExamID + "|" + filter.SubjectID + "|" + filter.ChapterID + "|" + filter.KnowledgePointID + "|" + filter.Status + "|" + strconv.Itoa(filter.Page) + "|" + strconv.Itoa(filter.PageSize)
}

type rawOption struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

func parseOptions(raw string) ([]QuestionOption, map[string]string) {
	var payload struct {
		Choices []rawOption `json:"choices"`
	}
	_ = json.Unmarshal([]byte(raw), &payload)

	options := make([]QuestionOption, 0, len(payload.Choices))
	labelByOptionID := make(map[string]string, len(payload.Choices))
	for index, choice := range payload.Choices {
		label := string(rune('A' + index))
		labelByOptionID[choice.ID] = label
		options = append(options, QuestionOption{
			Label: label,
			Text:  choice.Text,
		})
	}
	return options, labelByOptionID
}

func parseCorrectLabels(raw string, labelByOptionID map[string]string) []string {
	var payload struct {
		SelectedOptionIDs []string `json:"selected_option_ids"`
	}
	_ = json.Unmarshal([]byte(raw), &payload)

	labels := make([]string, 0, len(payload.SelectedOptionIDs))
	for _, optionID := range payload.SelectedOptionIDs {
		if label, ok := labelByOptionID[optionID]; ok {
			labels = append(labels, label)
			continue
		}
		labels = append(labels, strings.ToUpper(strings.TrimSpace(optionID)))
	}
	return normalizeLabels(labels)
}

func extractText(raw string) string {
	var payload struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err == nil && strings.TrimSpace(payload.Text) != "" {
		return payload.Text
	}
	return strings.TrimSpace(raw)
}

func normalizeAnswer(answer any) []string {
	switch typed := answer.(type) {
	case string:
		if strings.TrimSpace(typed) == "" {
			return []string{}
		}
		return normalizeLabels([]string{typed})
	case []string:
		return normalizeLabels(typed)
	case []any:
		labels := make([]string, 0, len(typed))
		for _, item := range typed {
			if value, ok := item.(string); ok {
				labels = append(labels, value)
			}
		}
		return normalizeLabels(labels)
	default:
		return []string{}
	}
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

func correctAnswerPayload(questionType string, correctLabels []string) any {
	normalized := normalizeLabels(correctLabels)
	if questionType == "single_choice" && len(normalized) == 1 {
		return normalized[0]
	}
	return normalized
}

func durationMinutes(seconds int) int {
	if seconds <= 0 {
		return 0
	}
	return int(math.Ceil(float64(seconds) / 60.0))
}
