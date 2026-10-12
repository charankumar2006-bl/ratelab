package limiter

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestFixedWindowLimit(t *testing.T) {
	ctx := context.Background()

	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer client.Close()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("Redis is unavailable: %v", err)
	}

	clientID := fmt.Sprintf("test-client-%d", time.Now().UnixNano())

	limiter := NewFixedWindow(client, 5, 60*time.Second)

	defer func() {
		pattern := "ratelab:fixed:" + clientID + ":*"
		keys, err := client.Keys(ctx, pattern).Result()
		if err == nil && len(keys) > 0 {
			client.Del(ctx, keys...)
		}
	}()

	// The first five requests should be allowed.
	for i := 1; i <= 5; i++ {
		result, err := limiter.Allow(ctx, clientID)
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}

		if !result.Allowed {
			t.Fatalf("request %d should have been allowed", i)
		}

		if result.Remaining != 5-i {
			t.Errorf(
				"request %d: expected remaining %d, got %d",
				i,
				5-i,
				result.Remaining,
			)
		}
	}

	// The sixth request should be rejected.
	result, err := limiter.Allow(ctx, clientID)
	if err != nil {
		t.Fatalf("sixth request failed: %v", err)
	}

	if result.Allowed {
		t.Fatal("sixth request should have been rejected")
	}

	if result.Remaining != 0 {
		t.Errorf("expected remaining 0, got %d", result.Remaining)
	}
}

func TestFixedWindowSeparatesClients(t *testing.T) {
	ctx := context.Background()

	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer client.Close()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("Redis is unavailable: %v", err)
	}

	limiter := NewFixedWindow(client, 1, 60*time.Second)

	clientA := fmt.Sprintf("client-a-%d", time.Now().UnixNano())
	clientB := fmt.Sprintf("client-b-%d", time.Now().UnixNano())

	defer func() {
		for _, clientID := range []string{clientA, clientB} {
			pattern := "ratelab:fixed:" + clientID + ":*"
			keys, err := client.Keys(ctx, pattern).Result()
			if err == nil && len(keys) > 0 {
				client.Del(ctx, keys...)
			}
		}
	}()

	resultA, err := limiter.Allow(ctx, clientA)
	if err != nil {
		t.Fatal(err)
	}
	if !resultA.Allowed {
		t.Fatal("first request from client A should be allowed")
	}

	resultA, err = limiter.Allow(ctx, clientA)
	if err != nil {
		t.Fatal(err)
	}
	if resultA.Allowed {
		t.Fatal("second request from client A should be rejected")
	}

	resultB, err := limiter.Allow(ctx, clientB)
	if err != nil {
		t.Fatal(err)
	}
	if !resultB.Allowed {
		t.Fatal("first request from client B should be allowed")
	}
}
