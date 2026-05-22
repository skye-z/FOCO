package account

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"xorm.io/xorm"
)

type SeedConfig struct {
	SupabaseURL    string
	ServiceRoleKey string
	Engine         *xorm.Engine
}

func SeedDefaultAdmin(cfg SeedConfig) (string, bool, error) {
	if cfg.Engine == nil || cfg.SupabaseURL == "" || cfg.ServiceRoleKey == "" {
		return "", false, fmt.Errorf("seed config incomplete")
	}

	email := "skai-zhang@hotmail.com"
	password := "DevAdmin@2026"

	profile := &Profile{}
	has, err := cfg.Engine.Table("auth.users").
		Where("email = ?", email).
		Cols("id").
		Get(profile)
	if err != nil {
		return "", false, fmt.Errorf("check existing user: %w", err)
	}
	if has && profile.Id != "" {
		if err := ensureAdminRoleXorm(cfg.Engine, profile.Id); err != nil {
			return "", false, fmt.Errorf("ensure admin role: %w", err)
		}
		return email, false, nil
	}

	body, _ := json.Marshal(map[string]any{
		"email":         email,
		"password":      password,
		"email_confirm": true,
	})
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, cfg.SupabaseURL+"/auth/v1/admin/users", bytes.NewReader(body))
	req.Header.Set("apikey", cfg.ServiceRoleKey)
	req.Header.Set("Authorization", "Bearer "+cfg.ServiceRoleKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", false, fmt.Errorf("create user via admin api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errBody map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		return "", false, fmt.Errorf("supabase admin create user returned %d: %v", resp.StatusCode, errBody)
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", false, fmt.Errorf("decode created user: %w", err)
	}

	_, err = cfg.Engine.Insert(&Profile{
		Id:          result.ID,
		DisplayName: "Admin",
		Status:      "active",
		CreatedAt:   time.Now(),
	})
	if err != nil {
		return email, true, fmt.Errorf("insert profile: %w", err)
	}

	if err := ensureAdminRoleXorm(cfg.Engine, result.ID); err != nil {
		return email, true, fmt.Errorf("ensure admin role: %w", err)
	}

	return email, true, nil
}

func ensureAdminRoleXorm(engine *xorm.Engine, userID string) error {
	count, err := engine.Table("user_roles").Where("user_id = ? AND role = 'admin'", userID).Count()
	if err != nil {
		return fmt.Errorf("check admin role: %w", err)
	}
	if count > 0 {
		return nil
	}
	_, err = engine.Exec(`
		INSERT INTO user_roles (id, user_id, role, granted_at, created_at)
		VALUES (gen_random_uuid(), ?::uuid, 'admin', now(), now())
		ON CONFLICT DO NOTHING
	`, userID)
	if err != nil {
		return fmt.Errorf("insert admin role: %w", err)
	}
	return nil
}
