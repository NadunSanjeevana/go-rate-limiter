package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/NadunSanjeevana/go-rate-limiter/utils"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// Redis client
var redisClient = redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
})

// Define rate limits for different user roles
var rateLimits = map[string]int{
	"free":    5,  // 5 requests per 10 sec
	"premium": 20, // 20 requests per 10 sec
	"admin":   50, // 50 requests per 10 sec
}

// Middleware for JWT validation and rate limiting
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip token check for login route
		if c.Request.URL.Path == "/login" {
			c.Next()
			return
		}

		// Extract JWT token
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Missing token"})
			return
		}

		// Remove "Bearer " prefix if present
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		// Parse and validate token
		claims, err := utils.ParseJWT(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
			return
		}

		username := claims["username"].(string)
		role := claims["role"].(string)

		// Apply rate limiting per user
		applyRateLimit(c, username, role)

		// Continue processing request
		c.Next()
	}
}

// Apply rate limiting based on user role
func applyRateLimit(c *gin.Context, username, role string) {
	ctx := context.Background()
	key := fmt.Sprintf("rate_limit:%s", username)

	// Get rate limit for user role
	maxRequests, exists := rateLimits[role]
	if !exists {
		maxRequests = 5 // Default limit
	}

	// Increment request count in Redis
	count, err := redisClient.Incr(ctx, key).Result()
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
		return
	}

	// Set expiration if first request
	if count == 1 {
		redisClient.Expire(ctx, key, 10*time.Second)
	}

	// Check if limit exceeded
	if count > int64(maxRequests) {
		c.AbortWithStatusJSON(429, gin.H{"error": "Rate limit exceeded"})
		return
	}
}
