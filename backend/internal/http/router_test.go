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
