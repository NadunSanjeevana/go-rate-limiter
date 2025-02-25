package main

import (
	"net/http"

	"github.com/NadunSanjeevana/go-rate-limiter/middleware"
	"github.com/NadunSanjeevana/go-rate-limiter/utils"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

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

	// Protected route - Requires valid JWT & rate limiting
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	r.Run(":8080") // Start server
}
