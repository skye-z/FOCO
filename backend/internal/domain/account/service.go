package account

import (
	"context"
	"time"

	cachepkg "foco/backend/api/internal/cache"
)

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

func (s *Service) BootstrapLearner(ctx context.Context, input BootstrapInput) ([]string, error) {
	if err := s.repo.EnsureProfile(ctx, input.UserID, input.DisplayName); err != nil {
		return nil, err
	}
	if err := s.repo.EnsureUserRole(ctx, input.UserID, "learner"); err != nil {
		return nil, err
	}
	s.invalidate(ctx, accountUserNamespace(input.UserID), accountAdminUsersNamespace(), accountStatsNamespace())
	return s.repo.ListRolesByUserID(ctx, input.UserID)
}

func (s *Service) ListExams(ctx context.Context) ([]ExamSummary, error) {
	if s.cache == nil {
		return s.repo.ListActiveExams(ctx)
	}
	var result []ExamSummary
	err := s.cache.GetJSON(ctx, accountContentNamespace(), "active-exams", 10*time.Minute, &result, func(ctx context.Context) (any, error) {
		return s.repo.ListActiveExams(ctx)
	})
	return result, err
}

func (s *Service) ListAdminUsers(ctx context.Context) ([]AdminUserSummary, error) {
	if s.cache == nil {
		return s.repo.ListAdminUsers(ctx)
	}
	var result []AdminUserSummary
	err := s.cache.GetJSON(ctx, accountAdminUsersNamespace(), "list", 30*time.Second, &result, func(ctx context.Context) (any, error) {
		return s.repo.ListAdminUsers(ctx)
	})
	return result, err
}

func (s *Service) GetActiveEnrollment(ctx context.Context, userID string) (*ActiveEnrollment, error) {
	if s.cache == nil {
		return s.repo.GetActiveEnrollmentByUserID(ctx, userID)
	}
	var result *ActiveEnrollment
	err := s.cache.GetJSON(ctx, accountUserNamespace(userID), "active-enrollment", 2*time.Minute, &result, func(ctx context.Context) (any, error) {
		return s.repo.GetActiveEnrollmentByUserID(ctx, userID)
	})
	return result, err
}

func (s *Service) EnsureEnrollment(ctx context.Context, input CreateEnrollmentInput) (*ActiveEnrollment, error) {
	enrollment, err := s.repo.EnsureEnrollment(ctx, input.UserID, input.ExamID)
	if err != nil {
		return nil, err
	}
	s.invalidate(ctx, accountUserNamespace(input.UserID), learnerUserExamNamespace(input.UserID, input.ExamID), accountStatsNamespace())
	return enrollment, nil
}

func (s *Service) GetPlatformStats(ctx context.Context) (*PlatformStats, error) {
	if s.cache != nil {
		var result *PlatformStats
		err := s.cache.GetJSON(ctx, accountStatsNamespace(), "overview", 2*time.Minute, &result, func(ctx context.Context) (any, error) {
			return s.getPlatformStatsUncached(ctx)
		})
		return result, err
	}
	return s.getPlatformStatsUncached(ctx)
}

func (s *Service) getPlatformStatsUncached(ctx context.Context) (*PlatformStats, error) {
	exams, err := s.repo.CountExams(ctx)
	if err != nil {
		return nil, err
	}
	users, err := s.repo.CountProfiles(ctx)
	if err != nil {
		return nil, err
	}
	active, err := s.repo.CountActiveUsersSince(ctx, time.Now().AddDate(0, 0, -7))
	if err != nil {
		return nil, err
	}
	return &PlatformStats{
		TotalExams:       exams,
		TotalUsers:       users,
		ActiveUsers7Days: active,
	}, nil
}

func (s *Service) invalidate(ctx context.Context, namespaces ...string) {
	if s.cache != nil {
		s.cache.Invalidate(ctx, namespaces...)
	}
}

func accountContentNamespace() string           { return "content:all" }
func accountAdminUsersNamespace() string        { return "account:admin-users" }
func accountStatsNamespace() string             { return "account:stats" }
func accountUserNamespace(userID string) string { return "account:user:" + userID }
func learnerUserExamNamespace(userID, examID string) string {
	return "learner:user:" + userID + ":exam:" + examID
}
