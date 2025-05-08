package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/NadunSanjeevana/go-rate-limiter/pkg/utils"
)

// AuthMiddlewareOptions defines options for the authentication middleware
type AuthMiddlewareOptions struct {
	// ExcludePaths lists paths that don't require authentication
	ExcludePaths []string
	// RedisClient for checking blacklisted tokens
	RedisClient *redis.Client
	// ExtractRoleFn extracts the role from JWT claims
	ExtractRoleFn func(map[string]interface{}) string
}

// NewAuthMiddleware creates a new JWT authentication middleware
func NewAuthMiddleware(options AuthMiddlewareOptions) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if the path is excluded from authentication
		for _, path := range options.ExcludePaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}

		// Extract JWT token
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Missing token"})
			return
		}

		// Remove "Bearer " prefix if present
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		// Check if token is blacklisted
		if options.RedisClient != nil {
			exists, err := options.RedisClient.Exists(context.Background(), fmt.Sprintf("blacklist:%s", tokenString)).Result()
			if err == nil && exists > 0 {
				c.AbortWithStatusJSON(401, gin.H{"error": "Token has been revoked"})
				return
			}
		}

		// Parse and validate token
		claims, err := utils.ParseJWT(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
			return
		}

		// Extract username and role
		username, ok := claims["username"].(string)
		if !ok {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token claims"})
			return
		}

		var role string
		if options.ExtractRoleFn != nil {
			role = options.ExtractRoleFn(claims)
		} else if r, ok := claims["role"].(string); ok {
			role = r
		} else {
			role = "free" // Default role
		}

		// Set user info in context
		c.Set("username", username)
		c.Set("role", role)
		c.Set("claims", claims)

		c.Next()
	}
}