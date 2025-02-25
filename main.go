package main

import (
	"net/http"
	"time"

	"github.com/NadunSanjeevana/go-rate-limiter/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Apply rate limiter middleware: 5 requests per 10 seconds
	r.Use(middleware.RateLimiter(5, 10*time.Second))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	r.Run(":8080") // Start server
}
