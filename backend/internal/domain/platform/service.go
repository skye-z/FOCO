package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	cachepkg "foco/backend/api/internal/cache"
	"xorm.io/xorm"
)

type Service struct {
	engine *xorm.Engine
	cache  *cachepkg.Manager
}

func NewService(engine *xorm.Engine, caches ...*cachepkg.Manager) *Service {
	var cacheManager *cachepkg.Manager
	if len(caches) > 0 {
		cacheManager = caches[0]
	}
	return &Service{engine: engine, cache: cacheManager}
}

func (s *Service) EnsureSchema() error {
	if s.engine == nil {
		return nil
	}
	_, err := s.engine.Exec(`
		CREATE TABLE IF NOT EXISTS admin_settings (
			key text PRIMARY KEY,
			value_json text NOT NULL,
			updated_at timestamptz NOT NULL DEFAULT now()
		)
	`)
	if err != nil {
		return fmt.Errorf("ensure admin_settings schema: %w", err)
	}
	return nil
}

func (s *Service) GetAdminSettings(ctx context.Context) (*AdminSettings, error) {
	if s.cache != nil {
		var result *AdminSettings
		err := s.cache.GetJSON(ctx, settingsNamespace(), "admin", 5*time.Minute, &result, func(ctx context.Context) (any, error) {
			return s.getAdminSettingsUncached(ctx)
		})
		return result, err
	}
	return s.getAdminSettingsUncached(ctx)
}

func (s *Service) getAdminSettingsUncached(ctx context.Context) (*AdminSettings, error) {
	return withAdminSettingsSchemaRetry(s.EnsureSchema, func() (*AdminSettings, error) {
		llm, err := s.loadLLMSettings(ctx)
		if err != nil {
			return nil, err
		}
		registrationOpen, err := s.loadRegistrationOpen(ctx)
		if err != nil {
			return nil, err
		}
		return &AdminSettings{
			LLM: LLMSettingsSummary{
				Provider:   llm.Provider,
				BaseURL:    llm.BaseURL,
				Model:      llm.Model,
				Enabled:    llm.Enabled,
				Configured: llm.APIKey != "",
			},
			RegistrationOpen: registrationOpen,
		}, nil
	})
}

func (s *Service) UpdateAdminSettings(ctx context.Context, input AdminSettingsUpdate) (*AdminSettings, error) {
	settings, err := withAdminSettingsSchemaRetry(s.EnsureSchema, func() (*AdminSettings, error) {
		if err := s.saveJSON(ctx, "llm_settings", input.LLM); err != nil {
			return nil, err
		}
		if err := s.saveJSON(ctx, "registration_open", map[string]bool{"value": input.RegistrationOpen}); err != nil {
			return nil, err
		}
		llm := input.LLM
		return &AdminSettings{
			LLM: LLMSettingsSummary{
				Provider:   llm.Provider,
				BaseURL:    llm.BaseURL,
				Model:      llm.Model,
				Enabled:    llm.Enabled,
				Configured: llm.APIKey != "",
			},
			RegistrationOpen: input.RegistrationOpen,
		}, nil
	})
	if err == nil && s.cache != nil {
		s.cache.Invalidate(ctx, settingsNamespace())
	}
	return settings, err
}

func (s *Service) GetPublicSettings(ctx context.Context) (*PublicSettings, error) {
	if s.cache != nil {
		var result *PublicSettings
		err := s.cache.GetJSON(ctx, settingsNamespace(), "public", 5*time.Minute, &result, func(ctx context.Context) (any, error) {
			return s.getPublicSettingsUncached(ctx)
		})
		return result, err
	}
	return s.getPublicSettingsUncached(ctx)
}

func (s *Service) getPublicSettingsUncached(ctx context.Context) (*PublicSettings, error) {
	return withAdminSettingsSchemaRetry(s.EnsureSchema, func() (*PublicSettings, error) {
		registrationOpen, err := s.loadRegistrationOpen(ctx)
		if err != nil {
			return nil, err
		}
		return &PublicSettings{RegistrationOpen: registrationOpen}, nil
	})
}

func settingsNamespace() string { return "platform:settings" }

func (s *Service) loadLLMSettings(ctx context.Context) (*LLMSettings, error) {
	stored := &AdminSetting{Key: "llm_settings"}
	has, err := s.engine.Context(ctx).ID(stored.Key).Get(stored)
	if err != nil {
		return nil, err
	}
	if !has {
		return &LLMSettings{}, nil
	}
	var result LLMSettings
	if err := json.Unmarshal([]byte(stored.ValueJSON), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *Service) loadRegistrationOpen(ctx context.Context) (bool, error) {
	stored := &AdminSetting{Key: "registration_open"}
	has, err := s.engine.Context(ctx).ID(stored.Key).Get(stored)
	if err != nil {
		return false, err
	}
	if !has {
		return true, nil
	}
	var payload struct {
		Value bool `json:"value"`
	}
	if err := json.Unmarshal([]byte(stored.ValueJSON), &payload); err != nil {
		return false, err
	}
	return payload.Value, nil
}

func (s *Service) saveJSON(ctx context.Context, key string, value any) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	stored := &AdminSetting{Key: key}
	has, err := s.engine.Context(ctx).ID(key).Get(stored)
	if err != nil {
		return err
	}
	now := time.Now()
	if has {
		stored.ValueJSON = string(b)
		stored.UpdatedAt = now
		_, err = s.engine.Context(ctx).ID(key).Cols("value_json", "updated_at").Update(stored)
		return err
	}
	_, err = s.engine.Context(ctx).Insert(&AdminSetting{Key: key, ValueJSON: string(b), UpdatedAt: now})
	return err
}

func withAdminSettingsSchemaRetry[T any](ensureSchema func() error, run func() (T, error)) (T, error) {
	value, err := run()
	if err == nil || !isMissingAdminSettingsRelationError(err) {
		return value, err
	}

	var zero T
	if ensureSchema == nil {
		return zero, err
	}
	if ensureErr := ensureSchema(); ensureErr != nil {
		return zero, ensureErr
	}
	return run()
}

func isMissingAdminSettingsRelationError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, `relation "admin_settings" does not exist`) ||
		strings.Contains(message, "no such table: admin_settings")
}
