package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/nesiler/cestx/common"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	*redis.Client
}

// NewRedisClient creates a new Redis client instance with connection pooling.
func NewRedisClient(cfg *common.RedisConfig) (*redis.Client, error) {
	// Create a new Redis client with connection pool options
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Host + ":" + cfg.Port,
		PoolSize:     10,
		MinIdleConns: 2,
	})

	// Test the connection (ping) with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		common.Err("Failed to connect to Redis: %v", err)
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	common.Ok("Connected to Redis successfully!")

	return rdb, nil
}

// Close closes the Redis client connection pool.
// It's important to call this function during service shutdown to release resources.
func Close(rdb *redis.Client) error {
	if err := rdb.Close(); err != nil {
		common.Err("Failed to close Redis connection: %v", err)
		return fmt.Errorf("failed to close Redis connection: %w", err)
	}

	common.Ok("Closed Redis connection successfully!")

	return nil
}
