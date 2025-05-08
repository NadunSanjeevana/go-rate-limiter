# Go Rate Limiter

A flexible and efficient rate limiter package for Go applications, with support for multiple rate limiting algorithms and integration with popular web frameworks like Gin.

## Features

- Multiple rate limiting algorithms:
  - **Leaky Bucket**: Smooths out traffic by processing requests at a constant rate
  - **Sliding Window**: Tracks requests within a moving time window for precise control
- Redis-backed storage for distributed environments
- Prometheus metrics integration
- Easy integration with the Gin web framework
- JWT authentication middleware with blacklisting support
- Customizable rate limits per user role

## Installation

```bash
go get -u github.com/NadunSanjeevana/go-rate-limiter
```

## Quick Start

```go
package main

import (
	"time"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/NadunSanjeevana/go-rate-limiter/pkg/middleware"
	"github.com/NadunSanjeevana/go-rate-limiter/pkg/ratelimit"
	"github.com/NadunSanjeevana/go-rate-limiter/pkg/redisclient"
)

func main() {
	// Initialize Redis store
	store := redisclient.NewRedisStore(&redis.Options{
		Addr: "localhost:6379",
	})

	// Configure rate limiter
	config := ratelimit.NewDefaultConfig()
	config.WindowSize = 10 * time.Second
	config.RateLimits = map[string]int{
		"free":    5,  // 5 requests per 10 seconds
		"premium": 20, // 20 requests per 10 seconds
	}

	// Create a leaky bucket rate limiter
	limiter := ratelimit.NewLeakyBucket(store, config)

	// Create Gin router
	r := gin.Default()

	// Create rate limiter middleware
	rateLimiterMiddleware := middleware.NewRateLimiterMiddleware(
		limiter,
		middleware.WithIdentifierFunc(func(c *gin.Context) string {
			return c.ClientIP() // Identify users by IP
		}),
	)

	// Apply middleware to your routes
	r.Use(rateLimiterMiddleware.Handle())

	// Define your routes
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello World!",
		})
	})

	r.Run(":8080")
}
```

## Rate Limiting Algorithms

### Leaky Bucket

The leaky bucket algorithm works by processing requests at a constant rate, smoothing out bursts of traffic. It's like a bucket with a small hole at the bottom - water (requests) pours in at the top, and leaks out at a steady rate.

```go
limiter := ratelimit.NewLeakyBucket(store, config)
```

### Sliding Window

The sliding window algorithm provides precise control by tracking requests within a moving time window. It's more responsive to changes in traffic patterns than fixed window approaches.

```go
limiter := ratelimit.NewSlidingWindow(store, config)
```

## Authentication Middleware

The package also includes JWT authentication middleware with role-based access control:

```go
authMiddleware := middleware.NewAuthMiddleware(middleware.AuthMiddlewareOptions{
    ExcludePaths: []string{"/login", "/public"},
    RedisClient:  store.GetClient(),
})

// Apply authentication middleware
r.Use(authMiddleware)
```

## Custom Configuration

You can fully customize the rate limiter with your own configuration:

```go
config := &ratelimit.Config{
    RateLimits: map[string]int{
        "free":     10,
        "premium":  50,
        "business": 100,
    },
    WindowSize:    30 * time.Second,
    EnableMetrics: true,
    DefaultRole:   "free",
    DefaultLimit:  10,
}
```

## Examples

Check out the `examples` directory for complete working examples including:

- Basic rate limiting with Gin
- Authentication with role-based rate limits
- Prometheus metrics integration

## License

MIT License - see [LICENSE](LICENSE) file for details.
