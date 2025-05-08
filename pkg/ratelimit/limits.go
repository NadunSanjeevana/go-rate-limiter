package ratelimit

import "time"

// DefaultWindowSize defines the default time window for rate limiting (in seconds)
const DefaultWindowSize = 10 * time.Second

// DefaultRateLimits defines the default rate limits for different user roles
var DefaultRateLimits = map[string]int{
	"free":    5,  // 5 requests per window
	"premium": 20, // 20 requests per window
	"admin":   50, // 50 requests per window
}

// Config holds rate limiter configuration
type Config struct {
	// RateLimits defines the request limit per window for each role
	RateLimits map[string]int
	// WindowSize defines the time window for rate limiting
	WindowSize time.Duration
	// EnableMetrics determines whether to collect Prometheus metrics
	EnableMetrics bool
	// DefaultRole is used when a role is not specified
	DefaultRole string
	// DefaultLimit is used when a role's limit is not specified
	DefaultLimit int
}

// NewDefaultConfig returns a new Config with default values
func NewDefaultConfig() *Config {
	return &Config{
		RateLimits:    DefaultRateLimits,
		WindowSize:    DefaultWindowSize,
		EnableMetrics: true,
		DefaultRole:   "free",
		DefaultLimit:  5,
	}
}

// GetLimit returns the rate limit for the given role
func (c *Config) GetLimit(role string) int {
	if limit, exists := c.RateLimits[role]; exists {
		return limit
	}
	return c.DefaultLimit
}