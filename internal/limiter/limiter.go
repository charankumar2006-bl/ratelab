package limiter

import (
	"context"
	"time"
)

// Result describes the decision made by a rate limiter.
type Result struct {
	Allowed   bool
	Limit     int
	Remaining int
	ResetAt   time.Time
}

// RateLimiter defines the behavior every rate-limiting
// algorithm must implement.
type RateLimiter interface {
	Allow(ctx context.Context, clientID string) (Result, error)
}
