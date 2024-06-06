package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/nesiler/cestx/common"
	"github.com/redis/go-redis/v9"
)

// RedisClient manages the connection to a Redis server.
type RedisClient struct {
	Client *redis.Client
	Config *Config
}

// NewRedisClient creates a new Redis client instance with connection pooling.
func NewRedisClient(cfg *Config) (*RedisClient, error) {
	// Create a new Redis client with connection pool options
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Address,
		Password:     cfg.Password, // No password set
		DB:           cfg.DB,       // Use default DB
		PoolSize:     10,           // Maximum number of connections in the pool (adjust as needed)
		MinIdleConns: 2,            // Minimum number of idle connections (adjust as needed)
		// TODO Check other options like Dialer, TLSConfig, etc.
	})

	// Test the connection (ping) with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		common.Err("Failed to connect to Redis: %v", err)
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	common.Ok("Connected to Redis successfully!")

	return &RedisClient{
		Client: client,
		Config: cfg,
	}, nil
}

// Close closes the Redis client connection pool.
// It's important to call this function during service shutdown to release resources.
func (c *RedisClient) Close() error {
	common.Info("Closing Redis connection pool...")
	if err := c.Client.Close(); err != nil {
		common.Err("Error closing Redis connection pool: %v", err)
		return err
	}
	common.Ok("Redis connection pool closed successfully.")
	return nil
}