package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/NadunSanjeevana/go-rate-limiter/pkg/redisclient"
)

// BlacklistToken adds the JWT to the blacklist in Redis with an expiration time
func BlacklistToken(ctx context.Context, token string) error {
	// Create a unique key for the token in the Redis blacklist
	key := fmt.Sprintf("blacklist:%s", token)

	// Set the token with a 15-minute expiration
	err := redisclient.RedisClient.Set(ctx, key, "revoked", 15*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("failed to blacklist token: %v", err)
	}
	return nil
}
