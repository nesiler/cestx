package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nesiler/cestx/common"
	"github.com/redis/go-redis/v9"
)

// It takes the Redis client, context, key, value (which can be any serializable type), and expiration duration.
func Set(ctx context.Context, rdb *redis.Client, key string, value interface{}, expiration time.Duration) error {
	// Serialize the value to JSON
	data, err := json.Marshal(value)
	if err != nil {
		return common.Err("Failed to marshal value for Redis: %v", err)
	}

	// Set the value in Redis with expiration
	err = rdb.Set(ctx, key, data, expiration).Err()
	if err != nil {
		return common.Err("Failed to set value in Redis: %v", err)
	}

	return nil
}

// Get retrieves a value from Redis.
// It takes the Redis client, context, key, and a target pointer to store the retrieved value (must be a pointer).
func Get(ctx context.Context, rdb *redis.Client, key string, target interface{}) error {
	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return common.Err("Key '%s' not found in Redis", key)
		}
		return common.Err("Failed to get value from Redis: %v", err)
	}

	if err := json.Unmarshal([]byte(val), target); err != nil {
		return common.Err("Failed to unmarshal Redis value: %v", err)
	}

	return nil
}

// Delete deletes a key from Redis.
func Delete(ctx context.Context, rdb *redis.Client, key string) error {
	if err := rdb.Del(ctx, key).Err(); err != nil {
		return common.Err("Failed to delete key '%s' from Redis: %v", key, err)
	}
	return nil
}

// Incr increments a counter in Redis and returns the new value.
func Incr(ctx context.Context, rdb *redis.Client, key string) (int64, error) {
	newVal, err := rdb.Incr(ctx, key).Result()
	if err != nil {
		return 0, common.Err("Failed to increment key '%s' in Redis: %v", key, err)
	}
	return newVal, nil
}
