package cache

import (
	"context"
	"time"
)

// CacheInterface defines the unified caching interface
type CacheInterface interface {
	// Get retrieves a value from cache by key
	Get(ctx context.Context, key string) ([]byte, error)
	
	// Set stores a value in cache with TTL
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	
	// Delete removes a key from cache
	Delete(ctx context.Context, key string) error
	
	// Exists checks if a key exists in cache
	Exists(ctx context.Context, key string) (bool, error)
	
	// Clear removes all keys from cache
	Clear(ctx context.Context) error
	
	// HealthCheck verifies cache connectivity
	HealthCheck(ctx context.Context) error
	
	// Close closes the cache connection
	Close() error
}

// CacheError represents cache operation errors
type CacheError struct {
	Op  string // operation name
	Key string // cache key
	Err error  // underlying error
}

func (e *CacheError) Error() string {
	return "cache " + e.Op + " key=" + e.Key + ": " + e.Err.Error()
}

func (e *CacheError) Unwrap() error {
	return e.Err
}

// NewCacheError creates a new cache error
func NewCacheError(op, key string, err error) *CacheError {
	return &CacheError{
		Op:  op,
		Key: key,
		Err: err,
	}
}