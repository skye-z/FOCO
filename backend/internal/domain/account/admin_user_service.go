package account

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	cachepkg "foco/backend/api/internal/cache"
)

type AdminUserService struct {
	repo           Repository
	supabaseURL    string
	serviceRoleKey string
	httpClient     *http.Client
	cache          *cachepkg.Manager
}

func NewAdminUserService(repo Repository, supabaseURL, serviceRoleKey string, httpClient *http.Client, caches ...*cachepkg.Manager) *AdminUserService {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	var cacheManager *cachepkg.Manager
	if len(caches) > 0 {
		cacheManager = caches[0]
	}
	return &AdminUserService{
		repo:           repo,
		supabaseURL:    strings.TrimRight(supabaseURL, "/"),
		serviceRoleKey: serviceRoleKey,
		httpClient:     httpClient,
		cache:          cacheManager,
	}
}

func (s *AdminUserService) ListAdminUsers(ctx context.Context) ([]AdminUserSummary, error) {
	return s.repo.ListAdminUsers(ctx)
}

func (s *AdminUserService) GrantRole(ctx context.Context, userID, role, grantedBy string) error {
	if role != "admin" && role != "editor" && role != "learner" {
		return fmt.Errorf("invalid role")
	}
	err := s.repo.GrantRole(ctx, userID, role, grantedBy)
	s.invalidateUserCaches(ctx, userID, err)
	return err
}

func (s *AdminUserService) DisableUser(ctx context.Context, userID string) error {
	if err := s.requireSupabaseAdmin(); err != nil {
		return err
	}

	body, _ := json.Marshal(map[string]any{
		"ban_duration": "876000h",
	})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPut, s.supabaseURL+"/auth/v1/admin/users/"+userID, bytes.NewReader(body))
	req.Header.Set("apikey", s.serviceRoleKey)
	req.Header.Set("Authorization", "Bearer "+s.serviceRoleKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("disable user via supabase: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("disable user via supabase returned %d", resp.StatusCode)
	}

	err = s.repo.DisableProfile(ctx, userID)
	s.invalidateUserCaches(ctx, userID, err)
	return err
}

func (s *AdminUserService) ResetPassword(ctx context.Context, userID, newPassword string) error {
	if err := s.requireSupabaseAdmin(); err != nil {
		return err
	}
	if len(strings.TrimSpace(newPassword)) < 8 {
		return fmt.Errorf("password too short")
	}

	body, _ := json.Marshal(map[string]any{
		"password": newPassword,
	})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPut, s.supabaseURL+"/auth/v1/admin/users/"+userID, bytes.NewReader(body))
	req.Header.Set("apikey", s.serviceRoleKey)
	req.Header.Set("Authorization", "Bearer "+s.serviceRoleKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("reset password via supabase: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("reset password via supabase returned %d", resp.StatusCode)
	}

	s.invalidateUserCaches(ctx, userID, nil)
	return nil
}

func (s *AdminUserService) requireSupabaseAdmin() error {
	if s.supabaseURL == "" || s.serviceRoleKey == "" {
		return fmt.Errorf("supabase admin config incomplete")
	}
	return nil
}

func (s *AdminUserService) invalidateUserCaches(ctx context.Context, userID string, err error) {
	if err != nil || s.cache == nil {
		return
	}
	s.cache.Invalidate(ctx, accountAdminUsersNamespace(), accountUserNamespace(userID), accountStatsNamespace())
}
