package home

import (
	"context"
	"encoding/json"
	"time"

	contentpkg "foco/backend/api/internal/domain/content"
	diagnosticpkg "foco/backend/api/internal/domain/diagnostic"
	profilepkg "foco/backend/api/internal/domain/profile"
	"xorm.io/xorm"
)

type XormRepository struct {
	engine      *xorm.Engine
	profileRepo *profilepkg.XormRepository
}

func NewXormRepository(engine *xorm.Engine) *XormRepository {
	return &XormRepository{
		engine:      engine,
		profileRepo: profilepkg.NewXormRepository(engine),
	}
}

func (r *XormRepository) GetExamOverview(ctx context.Context, examID string) (*ExamOverview, error) {
	if examID == "" {
		return nil, nil
	}

	var exam contentpkg.Exam
	has, err := r.engine.Context(ctx).ID(examID).Get(&exam)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}

	return &ExamOverview{
		ExamID:           exam.Id,
		ExamName:         exam.Name,
		NextExamDate:     exam.NextExamDate,
		NextNextExamDate: exam.NextNextExamDate,
	}, nil
}

func (r *XormRepository) GetWalletBalance(ctx context.Context, userID string) (int, error) {
	return r.profileRepo.GetWalletBalance(ctx, userID)
}

func (r *XormRepository) GetStreakStats(ctx context.Context, userID, examID string) (*profilepkg.StreakStats, error) {
	return r.profileRepo.GetStreakStats(ctx, userID, examID)
}

func (r *XormRepository) GetLatestDiagnosticSummary(ctx context.Context, userID, examID string) (*DiagnosticSummary, error) {
	var row struct {
		ProfileSummary string `xorm:"profile_summary"`
	}
	has, err := r.engine.Context(ctx).SQL(`
		select profile_summary::text as profile_summary
		from learner_profiles
		where user_id = ?::uuid and exam_id = ?::uuid
		order by profile_version desc
		limit 1
	`, userID, examID).Get(&row)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}

	var summary diagnosticpkg.ProfileSummary
	if err := json.Unmarshal([]byte(row.ProfileSummary), &summary); err != nil {
		return nil, err
	}
	return &summary, nil
}

func (r *XormRepository) ListPracticeSessions(ctx context.Context, userID, examID string, limit int) ([]profilepkg.PracticeSessionSummary, error) {
	return r.profileRepo.ListPracticeSessions(ctx, userID, examID, limit)
}

func (r *XormRepository) ListKnowledgePointResults(ctx context.Context, userID, examID string) ([]profilepkg.KnowledgePointResult, error) {
	return r.profileRepo.ListKnowledgePointResults(ctx, userID, examID)
}

func (r *XormRepository) CountCompletedInteractiveAttempts(ctx context.Context, userID, examID string, since, until time.Time) (int, error) {
	var row struct {
		Count int `xorm:"count"`
	}
	has, err := r.engine.Context(ctx).SQL(`
		select count(1)::int as count
		from unit_attempts ua
		join interactive_unit_versions iuv on iuv.id = ua.unit_version_id
		join interactive_units iu on iu.id = iuv.interactive_unit_id
		where ua.user_id = ?::uuid
		  and iu.exam_id = ?::uuid
		  and ua.status = 'completed'
		  and ua.completed_at >= ?
		  and ua.completed_at < ?
	`, userID, examID, since, until).Get(&row)
	if err != nil {
		return 0, err
	}
	if !has {
		return 0, nil
	}
	return row.Count, nil
}
