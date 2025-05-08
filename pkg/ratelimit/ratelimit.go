package ratelimit

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// Limiter defines the interface for rate limiters
type Limiter interface {
	// Allow checks if a request is allowed and updates internal counters
	// Returns true if the request is allowed, false otherwise
	Allow(ctx context.Context, key string, role string) (bool, int, error)
}

// Store provides storage capabilities for rate limiting
type Store interface {
	// GetClient returns the underlying Redis client
	GetClient() *redis.Client
}

// Result represents the result of a rate limit check
type Result struct {
	// Allowed indicates whether the request is allowed
	Allowed bool
	// Remaining is the number of requests remaining in the current window
	Remaining int
	// RetryAfter indicates how long to wait before the next request (in seconds)
	RetryAfter int
}