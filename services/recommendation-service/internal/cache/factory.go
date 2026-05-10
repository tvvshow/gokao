package cache

import (
	"fmt"
	"github.com/tvvshow/gokao/services/recommendation-service/internal/config"
)

// NewCache creates a cache instance based on configuration
func NewCache(cfg *config.RedisConfig) (CacheInterface, error) {
	if !cfg.Enabled {
		return NewMemoryCache(), nil
	}

	// Build Redis address
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	
	// Try to create Redis cache
	redisCache, err := NewRedisCache(addr, cfg.Password, cfg.DB)
	if err != nil {
		// Fallback to memory cache if Redis fails
		return NewMemoryCache(), fmt.Errorf("Redis unavailable, falling back to memory cache: %w", err)
	}
	
	return redisCache, nil
}

// NewCacheWithFallback creates a cache with automatic fallback
func NewCacheWithFallback(cfg *config.RedisConfig) CacheInterface {
	cache, err := NewCache(cfg)
	if err != nil {
		// Always return memory cache as fallback
		return NewMemoryCache()
	}
	return cache
}