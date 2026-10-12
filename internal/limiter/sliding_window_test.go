package limiter

import (
	"context"
	"fmt"
	"testing"
	"time"

	redisclient "ratelab/internal/redis"
)

func TestSlidingWindowLimit(t *testing.T) {
	ctx := context.Background()

	client := redisclient.NewClient()
	defer client.Close()

	if err := redisclient.Ping(ctx, client); err != nil {
		t.Fatalf("Redis connection failed: %v", err)
	}

	clientID := fmt.Sprintf("sliding-test-%d", time.Now().UnixNano())
	key := "ratelab:sliding:" + clientID
	defer client.Del(ctx, key)

	limit := 5
	window := 60 * time.Second

	limiter := NewSlidingWindow(client, limit, window)

	for i := 1; i <= limit; i++ {
		result, err := limiter.Allow(ctx, clientID)
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
		if !result.Allowed {
			t.Fatalf("request %d should have been allowed", i)
		}
	}

	result, err := limiter.Allow(ctx, clientID)
	if err != nil {
		t.Fatalf("sixth request failed: %v", err)
	}
	if result.Allowed {
		t.Fatal("sixth request should have been rejected")
	}
}

func TestSlidingWindowSeparatesClients(t *testing.T) {
	ctx := context.Background()

	client := redisclient.NewClient()
	defer client.Close()

	if err := redisclient.Ping(ctx, client); err != nil {
		t.Fatalf("Redis connection failed: %v", err)
	}

	clientA := fmt.Sprintf("sliding-client-A-%d", time.Now().UnixNano())
	clientB := fmt.Sprintf("sliding-client-B-%d", time.Now().UnixNano())

	defer client.Del(ctx,
		"ratelab:sliding:"+clientA,
		"ratelab:sliding:"+clientB,
	)

	limiter := NewSlidingWindow(client, 1, 60*time.Second)

	first, err := limiter.Allow(ctx, clientA)
	if err != nil {
		t.Fatalf("client A request failed: %v", err)
	}
	if !first.Allowed {
		t.Fatal("client A's first request should be allowed")
	}

	second, err := limiter.Allow(ctx, clientA)
	if err != nil {
		t.Fatalf("client A second request failed: %v", err)
	}
	if second.Allowed {
		t.Fatal("client A's second request should be rejected")
	}

	otherClient, err := limiter.Allow(ctx, clientB)
	if err != nil {
		t.Fatalf("client B request failed: %v", err)
	}
	if !otherClient.Allowed {
		t.Fatal("client B's first request should be allowed")
	}
}
