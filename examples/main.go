package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/NadunSanjeevana/go-rate-limiter/pkg/middleware"
	"github.com/NadunSanjeevana/go-rate-limiter/pkg/ratelimit"
	"github.com/NadunSanjeevana/go-rate-limiter/pkg/redisclient"
	"github.com/NadunSanjeevana/go-rate-limiter/pkg/utils"
)

func main() {
	// Initialize Redis client
	redisOptions := &redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	}
	store := redisclient.NewRedisStore(redisOptions)

	// Configure rate limiter
	config := ratelimit.NewDefaultConfig()
	config.WindowSize = 10 * time.Second
	config.RateLimits = map[string]int{
		"free":    5,
		"premium": 20,
		"admin":   50,
	}

	// Create a leaky bucket rate limiter
	limiter := ratelimit.NewLeakyBucket(store, config)

	// Create Gin router
	r := gin.Default()

	// Create JWT auth middleware
	authMiddleware := middleware.NewAuthMiddleware(middleware.AuthMiddlewareOptions{
		ExcludePaths: []string{"/login", "/logout", "/metrics"},
		RedisClient:  store.GetClient(),
	})

	// Create rate limiter middleware
	rateLimiterMiddleware := middleware.NewRateLimiterMiddleware(
		limiter,
		middleware.WithIdentifierFunc(func(c *gin.Context) string {
			if username, exists := c.Get("username"); exists {
				return username.(string)
			}
			return c.ClientIP()
		}),
		middleware.WithRoleFunc(func(c *gin.Context) string {
			if role, exists := c.Get("role"); exists {
				return role.(string)
			}
			return "free"
		}),
	)

	// Expose Prometheus metrics
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Login endpoint
	r.POST("/login", func(c *gin.Context) {
		var loginData struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.ShouldBindJSON(&loginData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// In a real application, validate credentials here
		// For demo purposes, we just generate a token with a role

		// Determine role (in a real app, this would come from your user database)
		var role string
		if loginData.Username == "admin" {
			role = "admin"
		} else if loginData.Username == "premium" {
			role = "premium"
		} else {
			role = "free"
		}

		// Generate JWT token
		claims := map[string]interface{}{
			"username": loginData.Username,
			"role":     role,
		}
		token, err := utils.GenerateJWT(claims, 24*time.Hour)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"role":  role,
		})
	})

	// Protected routes group
	protected := r.Group("/")
	protected.Use(authMiddleware)
	protected.Use(rateLimiterMiddleware.Handle())

	protected.GET("/resource", func(c *gin.Context) {
		username, _ := c.Get("username")
		role, _ := c.Get("role")
		c.JSON(http.StatusOK, gin.H{
			"message":  "Access granted to protected resource",
			"username": username,
			"role":     role,
		})
	})

	// Start server
	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}