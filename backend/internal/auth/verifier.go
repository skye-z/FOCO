package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	cachepkg "foco/backend/api/internal/cache"
)

type RoleRepository interface {
	ListRolesByUserID(ctx context.Context, userID string) ([]string, error)
}

type SupabaseTokenVerifier struct {
	baseURL  string
	apiKey   string
	client   *http.Client
	roleRepo RoleRepository
	cache    *cachepkg.Manager
}

func NewSupabaseTokenVerifier(baseURL, apiKey string, client *http.Client, roleRepo RoleRepository, cacheManagers ...*cachepkg.Manager) *SupabaseTokenVerifier {
	if client == nil {
		client = http.DefaultClient
	}
	var cacheManager *cachepkg.Manager
	if len(cacheManagers) > 0 {
		cacheManager = cacheManagers[0]
	}
	return &SupabaseTokenVerifier{
		baseURL:  strings.TrimRight(baseURL, "/"),
		apiKey:   apiKey,
		client:   client,
		roleRepo: roleRepo,
		cache:    cacheManager,
	}
}

func (v *SupabaseTokenVerifier) VerifyToken(ctx context.Context, token string) (*Claims, error) {
	if v.cache != nil {
		var cached Claims
		err := v.cache.GetJSON(ctx, "auth:token", tokenCacheKey(token), 2*time.Minute, &cached, func(ctx context.Context) (any, error) {
			return v.verifyTokenUncached(ctx, token)
		})
		if err != nil {
			return nil, err
		}
		return &cached, nil
	}
	return v.verifyTokenUncached(ctx, token)
}

func (v *SupabaseTokenVerifier) verifyTokenUncached(ctx context.Context, token string) (*Claims, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.baseURL+"/auth/v1/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if v.apiKey != "" {
		req.Header.Set("apikey", v.apiKey)
	}

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("supabase auth returned %d", resp.StatusCode)
	}

	var payload struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	claims := &Claims{
		UserID: payload.ID,
		Email:  payload.Email,
	}

	if v.roleRepo != nil {
		roles, err := v.roleRepo.ListRolesByUserID(ctx, payload.ID)
		if err != nil {
			return nil, err
		}
		claims.Roles = roles
	}
	return claims, nil
}

func tokenCacheKey(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
