package cache

import "errors"

// Common cache errors
var (
	ErrKeyNotFound = errors.New("key not found")
	ErrKeyExpired  = errors.New("key expired")
	ErrConnection  = errors.New("cache connection failed")
	ErrTimeout     = errors.New("cache operation timeout")
)