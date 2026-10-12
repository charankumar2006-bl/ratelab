package redisclient

import (
	"context"
	"testing"
)

func TestRedisConnection(t *testing.T) {
	ctx := context.Background()

	client := NewClient()
	defer client.Close()

	if err := Ping(ctx, client); err != nil {
		t.Fatalf("Redis connection failed: %v", err)
	}
}