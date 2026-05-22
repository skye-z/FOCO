package cache

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

type cachePayload struct {
	Value string `json:"value"`
}

func TestManagerUsesRedisAsSecondLevel(t *testing.T) {
	server := runMiniRedisOrSkip(t)
	ctx := context.Background()

	first := NewManager("redis://" + server.Addr() + "/0")
	defer first.Close()
	second := NewManager("redis://" + server.Addr() + "/0")
	defer second.Close()

	loads := 0
	var out cachePayload
	if err := first.GetJSON(ctx, "test:redis", "payload", time.Minute, &out, func(context.Context) (any, error) {
		loads++
		return cachePayload{Value: "from-loader"}, nil
	}); err != nil {
		t.Fatalf("first load: %v", err)
	}
	if out.Value != "from-loader" || loads != 1 {
		t.Fatalf("unexpected first value=%+v loads=%d", out, loads)
	}

	var fromSecond cachePayload
	if err := second.GetJSON(ctx, "test:redis", "payload", time.Minute, &fromSecond, func(context.Context) (any, error) {
		loads++
		return cachePayload{Value: "should-not-load"}, nil
	}); err != nil {
		t.Fatalf("second load: %v", err)
	}
	if fromSecond.Value != "from-loader" || loads != 1 {
		t.Fatalf("expected Redis hit, got value=%+v loads=%d", fromSecond, loads)
	}
}

func TestManagerInvalidationBumpsNamespaceAcrossManagers(t *testing.T) {
	server := runMiniRedisOrSkip(t)
	ctx := context.Background()

	first := NewManager("redis://" + server.Addr() + "/0")
	defer first.Close()
	second := NewManager("redis://" + server.Addr() + "/0")
	defer second.Close()

	loads := 0
	loader := func(value string) func(context.Context) (any, error) {
		return func(context.Context) (any, error) {
			loads++
			return cachePayload{Value: value}, nil
		}
	}

	var before cachePayload
	if err := second.GetJSON(ctx, "test:invalidate", "payload", time.Minute, &before, loader("before")); err != nil {
		t.Fatalf("before load: %v", err)
	}
	first.Invalidate(ctx, "test:invalidate")

	var after cachePayload
	if err := second.GetJSON(ctx, "test:invalidate", "payload", time.Minute, &after, loader("after")); err != nil {
		t.Fatalf("after load: %v", err)
	}
	if after.Value != "after" || loads != 2 {
		t.Fatalf("expected namespace invalidation, got value=%+v loads=%d", after, loads)
	}
}

func runMiniRedisOrSkip(t *testing.T) *miniredis.Miniredis {
	t.Helper()
	server, err := miniredis.Run()
	if err == nil {
		t.Cleanup(server.Close)
		return server
	}
	if strings.Contains(err.Error(), "operation not permitted") {
		t.Skipf("local TCP listener is blocked in this environment: %v", err)
	}
	t.Fatalf("could not start miniredis: %v", err)
	return nil
}
