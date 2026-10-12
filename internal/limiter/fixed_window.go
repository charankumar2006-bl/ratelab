package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type FixedWindow struct {
	client *redis.Client
	limit  int
	window time.Duration
}

func NewFixedWindow(
	client *redis.Client,
	limit int,
	window time.Duration,
) *FixedWindow {
	return &FixedWindow{
		client: client,
		limit:  limit,
		window: window,
	}
}

func (fw *FixedWindow) Allow(
	ctx context.Context,
	clientID string,
) (Result, error) {
	now := time.Now()
	windowSeconds := int64(fw.window.Seconds())

	if fw.limit <= 0 || windowSeconds <= 0 {
		return Result{}, fmt.Errorf("limit and window must be positive")
	}

	windowStart := now.Unix() / windowSeconds * windowSeconds
	resetAt := time.Unix(windowStart+windowSeconds, 0)

	key := fmt.Sprintf(
		"ratelab:fixed:%s:%d",
		clientID,
		windowStart,
	)

	// Atomically increment the counter and set its expiration
	// only when the counter is first created.
	script := redis.NewScript(`
		local count = redis.call("INCR", KEYS[1])
		if count == 1 then
			redis.call("EXPIRE", KEYS[1], ARGV[1])
		end
		return count
	`)

	count, err := script.Run(
		ctx,
		fw.client,
		[]string{key},
		windowSeconds,
	).Int64()

	if err != nil {
		return Result{}, fmt.Errorf("increment request counter: %w", err)
	}

	remaining := fw.limit - int(count)
	if remaining < 0 {
		remaining = 0
	}

	return Result{
		Allowed:   count <= int64(fw.limit),
		Limit:     fw.limit,
		Remaining: remaining,
		ResetAt:   resetAt,
	}, nil
}
