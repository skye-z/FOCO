package middleware

import (
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type rateLimitRule struct {
	methods map[string]struct{}
	prefix  string
	limit   int
	window  time.Duration
}

type rateLimitBucket struct {
	windowStart time.Time
	count       int
}

type fixedWindowLimiter struct {
	now     func() time.Time
	mu      sync.Mutex
	buckets map[string]rateLimitBucket
	rules   []rateLimitRule
}

func RateLimitMiddleware(next http.Handler) http.Handler {
	limiter := &fixedWindowLimiter{
		now:     time.Now,
		buckets: make(map[string]rateLimitBucket),
		rules:   defaultRateLimitRules(),
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rule, ok := limiter.matchRule(r.Method, r.URL.Path)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		allowed, retryAfter := limiter.allow(clientKey(r)+"|"+rule.prefix, rule)
		if !allowed {
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (l *fixedWindowLimiter) matchRule(method, path string) (rateLimitRule, bool) {
	for _, rule := range l.rules {
		if _, ok := rule.methods[method]; !ok {
			continue
		}
		if strings.HasPrefix(path, rule.prefix) {
			return rule, true
		}
	}
	return rateLimitRule{}, false
}

func (l *fixedWindowLimiter) allow(key string, rule rateLimitRule) (bool, int) {
	now := l.now().UTC()
	windowStart := now.Truncate(rule.window)

	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.buckets) > 4096 {
		for bucketKey, bucket := range l.buckets {
			if bucket.windowStart.Before(windowStart.Add(-2 * rule.window)) {
				delete(l.buckets, bucketKey)
			}
		}
	}

	bucket := l.buckets[key]
	if bucket.windowStart.IsZero() || !bucket.windowStart.Equal(windowStart) {
		bucket = rateLimitBucket{windowStart: windowStart}
	}

	if bucket.count >= rule.limit {
		retryAfter := int(rule.window.Seconds() - now.Sub(windowStart).Seconds())
		if retryAfter < 1 {
			retryAfter = 1
		}
		l.buckets[key] = bucket
		return false, retryAfter
	}

	bucket.count++
	l.buckets[key] = bucket
	return true, 0
}

func defaultRateLimitRules() []rateLimitRule {
	return []rateLimitRule{
		newRateLimitRule([]string{http.MethodPost}, "/api/v1/seed/admin", envInt("API_RATE_LIMIT_SEED_ADMIN_PER_MINUTE", 5), time.Minute),
		newRateLimitRule([]string{http.MethodPost}, "/api/v1/admin/content-package/import", envInt("API_RATE_LIMIT_IMPORT_PER_MINUTE", 6), time.Minute),
		newRateLimitRule([]string{http.MethodPost, http.MethodPatch, http.MethodDelete}, "/api/v1/admin/", envInt("API_RATE_LIMIT_ADMIN_WRITE_PER_MINUTE", 90), time.Minute),
		newRateLimitRule([]string{http.MethodPost}, "/api/v1/learner/practice-sessions/", envInt("API_RATE_LIMIT_LEARNER_WRITE_PER_MINUTE", 120), time.Minute),
		newRateLimitRule([]string{http.MethodPost}, "/api/v1/learner/practice-sessions", envInt("API_RATE_LIMIT_LEARNER_WRITE_PER_MINUTE", 120), time.Minute),
		newRateLimitRule([]string{http.MethodPost}, "/api/v1/learner/interactive-unit-attempts/", envInt("API_RATE_LIMIT_LEARNER_WRITE_PER_MINUTE", 120), time.Minute),
	}
}

func newRateLimitRule(methods []string, prefix string, limit int, window time.Duration) rateLimitRule {
	methodSet := make(map[string]struct{}, len(methods))
	for _, method := range methods {
		methodSet[method] = struct{}{}
	}
	if limit < 1 {
		limit = 1
	}
	return rateLimitRule{
		methods: methodSet,
		prefix:  prefix,
		limit:   limit,
		window:  window,
	}
}

func envInt(name string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return fallback
	}
	return value
}

func clientKey(r *http.Request) string {
	if forwardedFor := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwardedFor != "" {
		if first := strings.TrimSpace(strings.Split(forwardedFor, ",")[0]); first != "" {
			return first
		}
	}

	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}

	if strings.TrimSpace(r.RemoteAddr) != "" {
		return strings.TrimSpace(r.RemoteAddr)
	}

	return "unknown"
}
