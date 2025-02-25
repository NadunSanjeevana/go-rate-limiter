package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/NadunSanjeevana/go-rate-limiter/pkg/redisclient"
	"github.com/gin-gonic/gin"
)

// Rate limiter middleware
func RateLimiter(maxRequests int, duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", clientIP)

		ctx := context.Background()

		// Increment request count in Redis
		count, err := redisclient.RedisClient.Incr(ctx, key).Result()
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
			return
		}

		// Set expiration if it's the first request
		if count == 1 {
			redisclient.RedisClient.Expire(ctx, key, duration)
		}

		// Check if limit exceeded
		if count > int64(maxRequests) {
			c.AbortWithStatusJSON(429, gin.H{"error": "Rate limit exceeded"})
			return
		}

		c.Next()
	}
}
