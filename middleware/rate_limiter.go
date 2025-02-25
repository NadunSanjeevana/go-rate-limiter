package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// Redis client setup
var redisClient = redis.NewClient(&redis.Options{
	Addr: "localhost:6379", // Change if Redis is running on another host
})

// Rate limiter middleware
func RateLimiter(maxRequests int, duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", clientIP)

		ctx := context.Background()

		// Increment request count in Redis
		count, err := redisClient.Incr(ctx, key).Result()
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
			return
		}

		// Set expiration if it's the first request
		if count == 1 {
			redisClient.Expire(ctx, key, duration)
		}

		// Check if limit exceeded
		if count > int64(maxRequests) {
			c.AbortWithStatusJSON(429, gin.H{"error": "Rate limit exceeded"})
			return
		}

		c.Next()
	}
}
