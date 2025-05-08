package ratelimit

import (
	"context"
	"fmt"
	"time"
	"strconv"

	"github.com/redis/go-redis/v9"
)

// LeakyBucket implements a leaky bucket rate limiter
type LeakyBucket struct {
	config *Config
	store  Store
}

// NewLeakyBucket creates a new leaky bucket rate limiter
func NewLeakyBucket(store Store, config *Config) *LeakyBucket {
	if config == nil {
		config = NewDefaultConfig()
	}
	return &LeakyBucket{
		config: config,
		store:  store,
	}
}

// Allow checks if a request is allowed based on the leaky bucket algorithm
func (lb *LeakyBucket) Allow(ctx context.Context, identifier string, role string) (bool, int, error) {
	key := fmt.Sprintf("bucket:%s", identifier)
	maxRequests := lb.config.GetLimit(role)

	// Get current time
	currentTime := time.Now().Unix()

	// Start a Redis transaction
	pipe := lb.store.GetClient().TxPipeline()
	
	// Retrieve bucket state from Redis
	lastRequestTime := pipe.HGet(ctx, key, "last_request_time")
	bucketLevel := pipe.HGet(ctx, key, "level")
	
	// Execute the transaction
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return false, 0, err
	}

	// Parse values
	var lastReqTime int64 = 0
	var level int64 = 0

	if lastTimeStr, err := lastRequestTime.Result(); err == nil && err != redis.Nil {
		lastReqTime, _ = strconv.ParseInt(lastTimeStr, 10, 64)
	}

	if levelStr, err := bucketLevel.Result(); err == nil && err != redis.Nil {
		level, _ = strconv.ParseInt(levelStr, 10, 64)
	}

	// Time elapsed since last request
	timeElapsed := currentTime - lastReqTime
	windowInSeconds := int64(lb.config.WindowSize / time.Second)

	// Leak requests at a fixed rate
	if timeElapsed > 0 {
		leakedRequests := timeElapsed / windowInSeconds
		if leakedRequests > 0 {
			level = level - leakedRequests
			if level < 0 {
				level = 0
			}
		}
	}

	// Check if request can be allowed
	if level >= int64(maxRequests) {
		if lb.config.EnableMetrics {
			BlockedRequests.WithLabelValues(role).Inc()
		}
		return false, 0, nil
	}

	// Increment the bucket level
	level++

	// Save updated bucket state to Redis
	pipe = lb.store.GetClient().TxPipeline()
	pipe.HSet(ctx, key, "last_request_time", currentTime)
	pipe.HSet(ctx, key, "level", level)
	pipe.Expire(ctx, key, lb.config.WindowSize)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return false, 0, err
	}

	remaining := maxRequests - int(level)
	
	// Update metrics if enabled
	if lb.config.EnableMetrics {
		RequestsTotal.WithLabelValues(role).Inc()
	}

	return true, remaining, nil
}