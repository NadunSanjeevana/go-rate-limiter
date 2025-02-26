package main

import (
	"context"
	"net/http"

	"github.com/NadunSanjeevana/go-rate-limiter/middleware"
	"github.com/NadunSanjeevana/go-rate-limiter/pkg/redisclient"
	"github.com/NadunSanjeevana/go-rate-limiter/utils"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	redisclient.InitRedis()
	r := gin.Default()

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Apply JWT validation & rate limiting (except login route)
	r.Use(middleware.AuthMiddleware())

	// Public route - Generate JWT (Login)
	r.POST("/login", func(c *gin.Context) {
		type LoginRequest struct {
			Username string `json:"username"`
			Role     string `json:"role"`
		}

		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		// Generate JWT for user
		token, err := utils.GenerateJWT(req.Username, req.Role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": token})
	})

	r.POST("/logout", func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(400, gin.H{"error": "Missing token"})
			return
		}

		token = token[len("Bearer "):]

		ctx := context.Background()
		err := utils.BlacklistToken(ctx, token)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to logout"})
			return
		}

		c.JSON(200, gin.H{"message": "Logged out successfully"})
	})


	// Protected route - Requires valid JWT & rate limiting
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	r.Run(":8080") // Start server
}
