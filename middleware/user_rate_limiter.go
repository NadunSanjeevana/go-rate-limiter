package middleware

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/NadunSanjeevana/go-rate-limiter/pkg/redisclient"
	"github.com/NadunSanjeevana/go-rate-limiter/utils"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Define Prometheus metrics
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_requests_total",
			Help: "Total number of requests per role",
		},
		[]string{"role"},
	)

	blockedRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_blocked_total",
			Help: "Total number of blocked requests due to rate limiting",
		},
		[]string{"role"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "rate_limit_request_duration_seconds",
			Help:    "Histogram of request duration per role",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"role"},
	)
)

// Register Prometheus metrics
func init() {
	prometheus.MustRegister(requestsTotal)
	prometheus.MustRegister(blockedRequests)
	prometheus.MustRegister(requestDuration)
}

// Define rate limits for different user roles
var rateLimits = map[string]int{
	"free":    5,  // 5 requests per 10 sec
	"premium": 20, // 20 requests per 10 sec
	"admin":   50, // 50 requests per 10 sec
}


// Leaky Bucket Algorithm Rate Limiting
func applyLeakyBucketRateLimit(c *gin.Context, username, role string) {
	startTime := time.Now()
	ctx := context.Background()

	key := fmt.Sprintf("bucket:%s", username)
	maxRequests := rateLimits[role]

	// Get current time
	currentTime := time.Now().Unix()

	// Retrieve bucket state from Redis
	lastRequestTime, _ := redisclient.RedisClient.HGet(ctx, key, "last_request_time").Int64()
	bucketLevel, _ := redisclient.RedisClient.HGet(ctx, key, "level").Int64()

	// Time elapsed since last request
	timeElapsed := currentTime - lastRequestTime

	// Leak requests at a fixed rate
	if timeElapsed > 10 { // 1 request per 10 seconds
		bucketLevel = bucketLevel - (timeElapsed / 10)
		if bucketLevel < 0 {
			bucketLevel = 0
		}
	}

	// Check if request can be allowed
	if bucketLevel >= int64(maxRequests) {
		log.Printf("[BLOCKED] User: %s, Role: %s, Rate Limit Exceeded!", username, role)
		blockedRequests.WithLabelValues(role).Inc()
		c.AbortWithStatusJSON(429, gin.H{"error": "Rate limit exceeded"})
		return
	}

	// Increment the bucket level
	bucketLevel++

	// Save updated bucket state to Redis
	redisclient.RedisClient.HSet(ctx, key, "last_request_time", currentTime)
	redisclient.RedisClient.HSet(ctx, key, "level", bucketLevel)

	// Set expiration for the bucket
	redisclient.RedisClient.Expire(ctx, key, 10*time.Second)

	// Log allowed request
	log.Printf("[ALLOWED] User: %s, Role: %s, Remaining: %d", username, role, maxRequests-int(bucketLevel))

	// Increment Prometheus metrics
	requestsTotal.WithLabelValues(role).Inc()
	duration := time.Since(startTime).Seconds()
	requestDuration.WithLabelValues(role).Observe(duration)

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
