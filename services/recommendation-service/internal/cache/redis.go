package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache implements CacheInterface using Redis
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(addr, password string, db int) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{client: client}, nil
}

// Get retrieves a value from Redis cache
func (rc *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := rc.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, NewCacheError("get", key, ErrKeyNotFound)
		}
		return nil, NewCacheError("get", key, err)
	}
	return []byte(val), nil
}

// Set stores a value in Redis cache with TTL
func (rc *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	err := rc.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return NewCacheError("set", key, err)
	}
	return nil
}

// Delete removes a key from Redis cache
func (rc *RedisCache) Delete(ctx context.Context, key string) error {
	err := rc.client.Del(ctx, key).Err()
	if err != nil {
		return NewCacheError("delete", key, err)
	}
	return nil
}

// Exists checks if a key exists in Redis cache
func (rc *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := rc.client.Exists(ctx, key).Result()
	if err != nil {
		return false, NewCacheError("exists", key, err)
	}
	return count > 0, nil
}

// Clear removes all keys from Redis cache (current DB)
func (rc *RedisCache) Clear(ctx context.Context) error {
	err := rc.client.FlushDB(ctx).Err()
	if err != nil {
		return NewCacheError("clear", "*", err)
	}
	return nil
}

// HealthCheck verifies Redis connectivity
func (rc *RedisCache) HealthCheck(ctx context.Context) error {
	err := rc.client.Ping(ctx).Err()
	if err != nil {
		return NewCacheError("healthcheck", "ping", err)
	}
	return nil
}

// Close closes the Redis connection
func (rc *RedisCache) Close() error {
	return rc.client.Close()
}

// GetClient returns the underlying Redis client (for advanced operations)
func (rc *RedisCache) GetClient() *redis.Client {
	return rc.client
}