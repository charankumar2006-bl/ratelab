package limiter

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type SlidingWindow struct {
	client *redis.Client
	limit  int
	window time.Duration
}

func NewSlidingWindow(
	client *redis.Client,
	limit int,
	window time.Duration,
) *SlidingWindow {
	return &SlidingWindow{
		client: client,
		limit:  limit,
		window: window,
	}
}

func (sw *SlidingWindow) Allow(
	ctx context.Context,
	clientID string,
) (Result, error) {
	if sw.limit <= 0 || sw.window <= 0 {
		return Result{}, fmt.Errorf("limit and window must be positive")
	}

	now := time.Now()
	nowMillis := now.UnixMilli()
	windowMillis := sw.window.Milliseconds()

	// Keep each client's request log separate.
	key := fmt.Sprintf("ratelab:sliding:%s", clientID)

	// Atomically remove old requests, count recent requests,
	// and add the new request only if it is allowed.
	script := redis.NewScript(`
		local key = KEYS[1]
		local now = tonumber(ARGV[1])
		local window = tonumber(ARGV[2])
		local limit = tonumber(ARGV[3])
		local member = ARGV[4]

		redis.call("ZREMRANGEBYSCORE", key, "-inf", now - window)

		local count = redis.call("ZCARD", key)

		if count >= limit then
			return {0, count}
		end

		redis.call("ZADD", key, now, member)
		redis.call("PEXPIRE", key, window)

		return {1, count + 1}
	`)

	member := strconv.FormatInt(nowMillis, 10) + "-" +
		strconv.FormatInt(time.Now().UnixNano(), 10)

	values, err := script.Run(
		ctx,
		sw.client,
		[]string{key},
		nowMillis,
		windowMillis,
		sw.limit,
		member,
	).Slice()

	if err != nil {
		return Result{}, fmt.Errorf("execute sliding window: %w", err)
	}

	if len(values) != 2 {
		return Result{}, fmt.Errorf("unexpected sliding window response")
	}

	allowedValue, err := redisInt64(values[0])
	if err != nil {
		return Result{}, err
	}

	count, err := redisInt64(values[1])
	if err != nil {
		return Result{}, err
	}

	remaining := sw.limit - int(count)
	if remaining < 0 {
		remaining = 0
	}

	return Result{
		Allowed:   allowedValue == 1,
		Limit:     sw.limit,
		Remaining: remaining,
		ResetAt:   now.Add(sw.window),
	}, nil
}

// redisInt64 converts a Redis Lua numeric result to int64.
func redisInt64(value interface{}) (int64, error) {
	switch v := value.(type) {
	case int64:
		return v, nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	case []byte:
		return strconv.ParseInt(string(v), 10, 64)
	default:
		return 0, fmt.Errorf("unexpected Redis numeric value: %T", value)
	}
}
