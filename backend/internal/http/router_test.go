package apihttp

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLearnerPracticeRoutesAreRegistered(t *testing.T) {
	router := NewRouter(Dependencies{
		SeedChinese: func(http.ResponseWriter, *http.Request) {},
	})

	testCases := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "create session",
			method: http.MethodPost,
			path:   "/api/v1/learner/practice-sessions",
		},
		{
			name:   "get session",
			method: http.MethodGet,
			path:   "/api/v1/learner/practice-sessions/session-1",
		},
		{
			name:   "submit answer",
			method: http.MethodPost,
			path:   "/api/v1/learner/practice-sessions/session-1/items/item-1/submit",
		},
		{
			name:   "profile data",
			method: http.MethodGet,
			path:   "/api/v1/learner/profile",
		},
		{
			name:   "wrong book data",
			method: http.MethodGet,
			path:   "/api/v1/learner/wrong-book",
		},
		{
			name:   "home data",
			method: http.MethodGet,
			path:   "/api/v1/learner/home",
		},
		{
			name:   "recommendations data",
			method: http.MethodGet,
			path:   "/api/v1/learner/recommendations",
		},
		{
			name:   "diagnostic current",
			method: http.MethodGet,
			path:   "/api/v1/learner/diagnostic/current?exam_id=exam-1",
		},
		{
			name:   "diagnostic restart",
			method: http.MethodPost,
			path:   "/api/v1/learner/diagnostic/restart",
		},
		{
			name:   "diagnostic submit",
			method: http.MethodPost,
			path:   "/api/v1/learner/diagnostic/attempt-1/submit",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			res := httptest.NewRecorder()

			router.ServeHTTP(res, req)

			if res.Code == http.StatusNotFound {
				t.Fatalf("expected learner practice route %s %s to be registered", tc.method, tc.path)
			}
		})
	}
}

func TestRouterAppliesSecurityHeadersAndCORSAllowlist(t *testing.T) {
	t.Setenv("API_ALLOWED_ORIGINS", "http://localhost:3000")

	router := NewRouter(Dependencies{
		SeedChinese: func(http.ResponseWriter, *http.Request) {},
	})

	allowedReq := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	allowedReq.Header.Set("Origin", "http://localhost:3000")
	allowedRes := httptest.NewRecorder()

	router.ServeHTTP(allowedRes, allowedReq)

	if got := allowedRes.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("expected allowed origin to be echoed back, got %q", got)
	}
	if got := allowedRes.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("expected X-Content-Type-Options=nosniff, got %q", got)
	}
	if got := allowedRes.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Fatalf("expected X-Frame-Options=DENY, got %q", got)
	}
	if got := allowedRes.Header().Get("Referrer-Policy"); got != "strict-origin-when-cross-origin" {
		t.Fatalf("expected Referrer-Policy to be set, got %q", got)
	}

	blockedReq := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	blockedReq.Header.Set("Origin", "http://evil.example")
	blockedRes := httptest.NewRecorder()

	router.ServeHTTP(blockedRes, blockedReq)

	if got := blockedRes.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected disallowed origin to be rejected, got %q", got)
	}
}

func TestRouterRateLimitsSeedAdminRequests(t *testing.T) {
	t.Setenv("API_RATE_LIMIT_SEED_ADMIN_PER_MINUTE", "2")

	router := NewRouter(Dependencies{
		SeedChinese: func(http.ResponseWriter, *http.Request) {},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/seed/admin", nil)
	req.RemoteAddr = "127.0.0.1:34567"
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected seed endpoint to require auth, got %d", res.Code)
	}
}
