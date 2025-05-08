package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/NadunSanjeevana/go-rate-limiter/pkg/ratelimit"
)

// RateLimiterMiddleware creates a middleware function for rate limiting
type RateLimiterMiddleware struct {
	limiter      ratelimit.Limiter
	identifierFn func(*gin.Context) string
	roleFn       func(*gin.Context) string
}

// RateLimiterOption defines options for the rate limiter middleware
type RateLimiterOption func(*RateLimiterMiddleware)

// WithIdentifierFunc sets a custom function to extract identifier from the request
func WithIdentifierFunc(fn func(*gin.Context) string) RateLimiterOption {
	return func(rlm *RateLimiterMiddleware) {
		rlm.identifierFn = fn
	}
}

// WithRoleFunc sets a custom function to extract role from the request
func WithRoleFunc(fn func(*gin.Context) string) RateLimiterOption {
	return func(rlm *RateLimiterMiddleware) {
		rlm.roleFn = fn
	}
}

// NewRateLimiterMiddleware creates a new rate limiter middleware
func NewRateLimiterMiddleware(limiter ratelimit.Limiter, options ...RateLimiterOption) *RateLimiterMiddleware {
	rlm := &RateLimiterMiddleware{
		limiter: limiter,
		// Default identifier function uses the client IP
		identifierFn: func(c *gin.Context) string {
			return c.ClientIP()
		},
		// Default role function returns "free"
		roleFn: func(c *gin.Context) string {
			return "free"
		},
	}

	// Apply options
	for _, option := range options {
		option(rlm)
	}

	return rlm
}

// Handle returns a gin middleware function
func (rlm *RateLimiterMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		
		identifier := rlm.identifierFn(c)
		role := rlm.roleFn(c)
		
		// Check rate limit
		allowed, remaining, err := rlm.limiter.Allow(context.Background(), identifier, role)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Rate limiter error"})
			return
		}
		
		if !allowed {
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("Retry-After", "10")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			return
		}
		
		// Set headers
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		
		// Track request duration for metrics
		duration := time.Since(startTime).Seconds()
		ratelimit.RequestDuration.WithLabelValues(role).Observe(duration)
		
		c.Next()
	}
}