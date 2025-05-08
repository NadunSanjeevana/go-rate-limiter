package redisclient

import (
	"github.com/redis/go-redis/v9"
)

// RedisStore implements the ratelimit.Store interface
type RedisStore struct {
	client *redis.Client
}

// NewRedisStore creates a new Redis store for rate limiting
func NewRedisStore(options *redis.Options) *RedisStore {
	if options == nil {
		options = &redis.Options{
			Addr: "localhost:6379",
			DB:   0,
		}
	}
	
	return &RedisStore{
		client: redis.NewClient(options),
	}
}

// GetClient returns the underlying Redis client
func (rs *RedisStore) GetClient() *redis.Client {
	return rs.client
}