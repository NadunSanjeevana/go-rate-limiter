package redisclient

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9" // Use v9 instead of v8
)

// RedisClient is a global Redis client instance
var RedisClient *redis.Client

// InitRedis initializes the Redis client
func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Update as needed
	})

	// Test connection
	if err := RedisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
}

// CloseRedis closes the Redis client
func CloseRedis() {
	if err := RedisClient.Close(); err != nil {
		log.Printf("Error closing Redis connection: %v", err)
	}
}
