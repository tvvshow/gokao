package cache

import (
	"context"
	"sync"
	"time"
)

// MemoryCache implements CacheInterface using in-memory storage
type MemoryCache struct {
	data   sync.Map
	expiry sync.Map // stores expiration times
	mu     sync.RWMutex
}

type memoryCacheItem struct {
	Value     []byte
	ExpiresAt time.Time
}

// NewMemoryCache creates a new memory cache instance
func NewMemoryCache() *MemoryCache {
	mc := &MemoryCache{}
	// Start cleanup goroutine
	go mc.cleanup()
	return mc
}

// Get retrieves a value from memory cache
func (mc *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	value, exists := mc.data.Load(key)
	if !exists {
		return nil, NewCacheError("get", key, ErrKeyNotFound)
	}
	
	item := value.(memoryCacheItem)
	
	// Check expiration
	if !item.ExpiresAt.IsZero() && time.Now().After(item.ExpiresAt) {
		mc.data.Delete(key)
		return nil, NewCacheError("get", key, ErrKeyExpired)
	}
	
	return item.Value, nil
}

// Set stores a value in memory cache with TTL
func (mc *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	item := memoryCacheItem{
		Value: make([]byte, len(value)),
	}
	copy(item.Value, value)
	
	if ttl > 0 {
		item.ExpiresAt = time.Now().Add(ttl)
	}
	
	mc.data.Store(key, item)
	return nil
}

// Delete removes a key from memory cache
func (mc *MemoryCache) Delete(ctx context.Context, key string) error {
	mc.data.Delete(key)
	return nil
}

// Exists checks if a key exists in memory cache
func (mc *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	value, exists := mc.data.Load(key)
	if !exists {
		return false, nil
	}
	
	item := value.(memoryCacheItem)
	
	// Check expiration
	if !item.ExpiresAt.IsZero() && time.Now().After(item.ExpiresAt) {
		mc.data.Delete(key)
		return false, nil
	}
	
	return true, nil
}

// Clear removes all keys from memory cache
func (mc *MemoryCache) Clear(ctx context.Context) error {
	mc.data.Range(func(key, value interface{}) bool {
		mc.data.Delete(key)
		return true
	})
	return nil
}

// HealthCheck verifies memory cache is working
func (mc *MemoryCache) HealthCheck(ctx context.Context) error {
	// Memory cache is always healthy
	return nil
}

// Close closes the memory cache (no-op)
func (mc *MemoryCache) Close() error {
	return nil
}

// cleanup removes expired items periodically
func (mc *MemoryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		now := time.Now()
		mc.data.Range(func(key, value interface{}) bool {
			item := value.(memoryCacheItem)
			if !item.ExpiresAt.IsZero() && now.After(item.ExpiresAt) {
				mc.data.Delete(key)
			}
			return true
		})
	}
}