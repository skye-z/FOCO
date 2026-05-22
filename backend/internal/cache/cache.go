package cache

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const defaultPrefix = "foco"

type Manager struct {
	local  *localStore
	redis  *redis.Client
	prefix string
}

func NewManager(redisURL string) *Manager {
	manager := &Manager{
		local:  newLocalStore(),
		prefix: defaultPrefix,
	}
	redisURL = strings.TrimSpace(redisURL)
	if redisURL == "" {
		return manager
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Printf("cache: invalid REDIS_URL, falling back to local cache: %v", err)
		return manager
	}
	opts.PoolSize = max(opts.PoolSize, 20)
	opts.MinIdleConns = max(opts.MinIdleConns, 2)
	manager.redis = redis.NewClient(opts)
	return manager
}

func (m *Manager) Enabled() bool {
	return m != nil
}

func (m *Manager) RedisEnabled() bool {
	return m != nil && m.redis != nil
}

func (m *Manager) Close() error {
	if m == nil || m.redis == nil {
		return nil
	}
	return m.redis.Close()
}

func (m *Manager) GetJSON(ctx context.Context, namespace, logicalKey string, ttl time.Duration, target any, loader func(context.Context) (any, error)) error {
	if m == nil || ttl <= 0 {
		value, err := loader(ctx)
		if err != nil {
			return err
		}
		return decodeInto(value, target)
	}

	version := m.version(ctx, namespace)
	key := m.dataKey(namespace, version, logicalKey)

	if payload, ok := m.local.get(key); ok {
		return json.Unmarshal(payload, target)
	}

	if m.redis != nil {
		payload, err := m.redis.Get(ctx, key).Bytes()
		if err == nil {
			m.local.set(key, payload, ttl)
			return json.Unmarshal(payload, target)
		}
	}

	value, err := loader(ctx)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	m.local.set(key, payload, ttl)
	if m.redis != nil {
		_ = m.redis.Set(ctx, key, payload, ttl).Err()
	}
	return json.Unmarshal(payload, target)
}

func (m *Manager) Invalidate(ctx context.Context, namespaces ...string) {
	if m == nil {
		return
	}
	for _, namespace := range namespaces {
		namespace = strings.TrimSpace(namespace)
		if namespace == "" {
			continue
		}
		next := m.local.bumpVersion(namespace)
		if m.redis != nil {
			if value, err := m.redis.Incr(ctx, m.versionKey(namespace)).Result(); err == nil {
				next = value
				m.local.setVersion(namespace, next)
			}
		}
		m.local.deletePrefix(m.namespacePrefix(namespace))
	}
}

func (m *Manager) version(ctx context.Context, namespace string) int64 {
	if m.redis != nil {
		if raw, err := m.redis.Get(ctx, m.versionKey(namespace)).Result(); err == nil {
			parsed, parseErr := strconv.ParseInt(raw, 10, 64)
			if parseErr == nil {
				m.local.setVersion(namespace, parsed)
				return parsed
			}
		}
	}

	if version, ok := m.local.getVersion(namespace); ok {
		return version
	}
	var version int64
	m.local.setVersion(namespace, version)
	return version
}

func (m *Manager) dataKey(namespace string, version int64, logicalKey string) string {
	return m.namespacePrefix(namespace) + ":" + strconv.FormatInt(version, 10) + ":" + hash(logicalKey)
}

func (m *Manager) namespacePrefix(namespace string) string {
	return m.prefix + ":cache:data:" + namespace
}

func (m *Manager) versionKey(namespace string) string {
	return m.prefix + ":cache:ver:" + namespace
}

func decodeInto(value any, target any) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(payload, target)
}

func hash(value string) string {
	sum := sha1.Sum([]byte(value))
	return hex.EncodeToString(sum[:])
}

type localStore struct {
	mu       sync.RWMutex
	values   map[string]localEntry
	versions map[string]int64
}

type localEntry struct {
	payload   []byte
	expiresAt time.Time
}

func newLocalStore() *localStore {
	return &localStore{
		values:   map[string]localEntry{},
		versions: map[string]int64{},
	}
}

func (s *localStore) get(key string) ([]byte, bool) {
	now := time.Now()
	s.mu.RLock()
	entry, ok := s.values[key]
	s.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if now.After(entry.expiresAt) {
		s.mu.Lock()
		delete(s.values, key)
		s.mu.Unlock()
		return nil, false
	}
	return append([]byte(nil), entry.payload...), true
}

func (s *localStore) set(key string, payload []byte, ttl time.Duration) {
	if ttl <= 0 {
		return
	}
	s.mu.Lock()
	s.values[key] = localEntry{
		payload:   append([]byte(nil), payload...),
		expiresAt: time.Now().Add(ttl),
	}
	s.mu.Unlock()
}

func (s *localStore) getVersion(namespace string) (int64, bool) {
	s.mu.RLock()
	version, ok := s.versions[namespace]
	s.mu.RUnlock()
	return version, ok
}

func (s *localStore) setVersion(namespace string, version int64) {
	s.mu.Lock()
	s.versions[namespace] = version
	s.mu.Unlock()
}

func (s *localStore) bumpVersion(namespace string) int64 {
	s.mu.Lock()
	next := s.versions[namespace] + 1
	s.versions[namespace] = next
	s.mu.Unlock()
	return next
}

func (s *localStore) deletePrefix(prefix string) {
	s.mu.Lock()
	for key := range s.values {
		if strings.HasPrefix(key, prefix) {
			delete(s.values, key)
		}
	}
	s.mu.Unlock()
}
