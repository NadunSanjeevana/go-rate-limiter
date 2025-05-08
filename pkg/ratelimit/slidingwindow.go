package ratelimit

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

// SlidingWindow implements a sliding window rate limiter
type SlidingWindow struct {
	config *Config
	store  Store
}

// NewSlidingWindow creates a new sliding window rate limiter
func NewSlidingWindow(store Store, config *Config) *SlidingWindow {
	if config == nil {
		config = NewDefaultConfig()
	}
	return &SlidingWindow{
		config: config,
		store:  store,
	}
}

// Allow checks if a request is allowed based on the sliding window algorithm
func (sw *SlidingWindow) Allow(ctx context.Context, identifier string, role string) (bool, int, error) {
	key := fmt.Sprintf("sliding:%s", identifier)
	maxRequests := sw.config.GetLimit(role)

	// Get current timestamp and the start of the time window
	currentTime := time.Now().Unix()
	windowSize := int64(sw.config.WindowSize / time.Second)
	windowStartTime := currentTime - windowSize

	// Get the list of request timestamps from Redis
	timestamps, err := sw.store.GetClient().LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return false, 0, err
	}

	// Filter valid timestamps (within the current window)
	validCount := 0
	for _, timestampStr := range timestamps {
		ts, _ := strconv.ParseInt(timestampStr, 10, 64)
		if ts > windowStartTime {
			validCount++
		}
	}

	// Check if the number of requests within the time window exceeds the limit
	if validCount >= maxRequests {
		if sw.config.EnableMetrics {
			BlockedRequests.WithLabelValues(role).Inc()
		}
		return false, 0, nil
	}

	// Add the current timestamp to the list
	pipe := sw.store.GetClient().TxPipeline()
	pipe.LPush(ctx, key, currentTime)
	pipe.LTrim(ctx, key, 0, int64(maxRequests-1))
	pipe.Expire(ctx, key, sw.config.WindowSize)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return false, 0, err
	}

	remaining := maxRequests - validCount - 1

	// Update metrics if enabled
	if sw.config.EnableMetrics {
		RequestsTotal.WithLabelValues(role).Inc()
	}

	return true, remaining, nil
}