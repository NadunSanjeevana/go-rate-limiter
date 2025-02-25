package middleware

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/NadunSanjeevana/go-rate-limiter/pkg/redisclient"
	"github.com/NadunSanjeevana/go-rate-limiter/utils"
	"github.com/gin-gonic/gin"
)

// Define rate limits for different user roles
var rateLimits = map[string]int{
	"free":    5,  // 5 requests per 10 sec
	"premium": 20, // 20 requests per 10 sec
	"admin":   50, // 50 requests per 10 sec
}


// Leaky Bucket Algorithm Rate Limiting
func applyLeakyBucketRateLimit(c *gin.Context, username, role string) {
	ctx := context.Background()

	// key := fmt.Sprintf("rate_limit:%s", username)
	maxRequests, exists := rateLimits[role]
	if !exists {
		maxRequests = 5 // Default for unknown roles
	}

	// Check current state of the leaky bucket in Redis
	currentTime := time.Now().Unix()
	bucketKey := fmt.Sprintf("bucket:%s", username)
	lastRequestTime, _ := redisclient.RedisClient.HGet(ctx, bucketKey, "last_request_time").Int64()
	bucketLevel, _ := redisclient.RedisClient.HGet(ctx, bucketKey, "level").Int64()

	// Time elapsed since last request
	timeElapsed := currentTime - lastRequestTime

	// Leak the bucket based on time elapsed
	if timeElapsed > 10 { // Assuming 10 seconds leak rate
		bucketLevel = bucketLevel - (int64(timeElapsed) / 10) // Leaks at 1 per 10 seconds
		if bucketLevel < 0 {
			bucketLevel = 0
		}
	}

	// Check if the bucket has room for more requests
	if bucketLevel >= int64(maxRequests) {
		c.AbortWithStatusJSON(429, gin.H{"error": "Rate limit exceeded"})
		return
	}

	// Increment the bucket level
	bucketLevel++

	// Save the updated bucket state in Redis
	redisclient.RedisClient.HSet(ctx, bucketKey, "last_request_time", currentTime)
	redisclient.RedisClient.HSet(ctx, bucketKey, "level", bucketLevel)

	// Set expiration for the bucket (to reset it after a certain period)
	redisclient.RedisClient.Expire(ctx, bucketKey, 10*time.Second)

	c.Next()
}

// Sliding Window Rate Limiting
func applySlidingWindowRateLimit(c *gin.Context, username, role string) {
	ctx := context.Background()
	key := fmt.Sprintf("rate_limit:%s", username)

	maxRequests, exists := rateLimits[role]
	if !exists {
		maxRequests = 5 // Default limit for unknown roles
	}

	// Get current timestamp and the start of the time window (e.g., 10 seconds ago)
	currentTime := time.Now().Unix()
	windowStartTime := currentTime - 10

	// Get the list of request timestamps from Redis
	timestamps, err := redisclient.RedisClient.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
		return
	}

	// Remove timestamps that are outside of the time window
	var validTimestamps []string
	for _, timestamp := range timestamps {
		ts, _ := strconv.ParseInt(timestamp, 10, 64)
		if ts > windowStartTime {
			validTimestamps = append(validTimestamps, timestamp)
		}
	}

	// Check if the number of requests within the time window exceeds the limit
	if len(validTimestamps) >= maxRequests {
		c.AbortWithStatusJSON(429, gin.H{"error": "Rate limit exceeded"})
		return
	}

	// Add the current timestamp to the list
	redisclient.RedisClient.LPush(ctx, key, currentTime)
	// Trim the list to only keep timestamps within the time window
	redisclient.RedisClient.LTrim(ctx, key, 0, int64(maxRequests-1))

	// Set expiration for the timestamp list
	redisclient.RedisClient.Expire(ctx, key, 10*time.Second)

	c.Next()
}

// Middleware for JWT validation and rate limiting
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip token check for login route
		if c.Request.URL.Path == "/login"|| c.Request.URL.Path == "/logout" {
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

		if isTokenBlacklisted(tokenString) {
			c.AbortWithStatusJSON(401, gin.H{"error": "Token has been revoked"})
			return
		}

		// Parse and validate token
		claims, err := utils.ParseJWT(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
			return
		}

		username := claims["username"].(string)
		role := claims["role"].(string)

		// Apply rate limiting per user
		// applyRateLimit(c, username, role)

		// applySlidingWindowRateLimit(c, username, role)

		applyLeakyBucketRateLimit(c, username, role)

		// Continue processing request
		c.Next()
	}
}

func isTokenBlacklisted(token string) bool {
	ctx := context.Background()
	exists, err := redisclient.RedisClient.Exists(ctx, fmt.Sprintf("blacklist:%s", token)).Result()
	return err == nil && exists > 0
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
	count, err := redisclient.RedisClient.Incr(ctx, key).Result()
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
		return
	}

	// Set expiration if first request
	if count == 1 {
		redisclient.RedisClient.Expire(ctx, key, 10*time.Second)
	}

	// Check if limit exceeded
	if count > int64(maxRequests) {
		c.AbortWithStatusJSON(429, gin.H{"error": "Rate limit exceeded"})
		return
	}
	remaining := maxRequests - int(count)
	c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
}
