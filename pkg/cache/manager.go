package cache

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// CacheStats 缓存统计信息
type CacheStats struct {
	Hits         int64
	Misses       int64
	LocalHits    int64
	RedisHits    int64
	Errors       int64
	Evictions    int64
	TotalGets    int64
	TotalSets    int64
	TotalDeletes int64
}

// CacheConfig 缓存配置
type CacheConfig struct {
	// Redis配置
	RedisURL      string
	RedisPassword string
	RedisDB       int

	// 本地缓存配置
	LocalCacheEnabled bool
	LocalCacheTTL     time.Duration
	LocalCacheSize    int

	// Redis缓存配置
	RedisDefaultTTL time.Duration
	RedisMaxRetries int
	RedisPoolSize   int
}

// DefaultConfig 默认缓存配置
func DefaultConfig() *CacheConfig {
	return &CacheConfig{
		RedisURL:          "localhost:6379",
		RedisPassword:     "",
		RedisDB:           0,
		LocalCacheEnabled: true,
		LocalCacheTTL:     5 * time.Minute,
		LocalCacheSize:    1000,
		RedisDefaultTTL:   24 * time.Hour,
		RedisMaxRetries:   3,
		RedisPoolSize:     10,
	}
}

// CacheManager 兼容层：统一复用 MultiLevelCache，避免双实现分叉。
type CacheManager struct {
	backend *MultiLevelCache
}

// NewCacheManager 创建新的缓存管理器
func NewCacheManager(config *CacheConfig) (*CacheManager, error) {
	if config == nil {
		config = DefaultConfig()
	}

	backend, err := NewMultiLevelCache(config)
	if err != nil {
		return nil, err
	}

	return &CacheManager{backend: backend}, nil
}

// Get 获取缓存值（多级缓存：本地内存 -> Redis）
func (cm *CacheManager) Get(ctx context.Context, key string, target interface{}) (bool, error) {
	return cm.backend.Get(ctx, key, target)
}

// Set 设置缓存值（多级缓存）
func (cm *CacheManager) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return cm.backend.Set(ctx, key, value, ttl)
}

// Delete 删除缓存（多级缓存）
func (cm *CacheManager) Delete(ctx context.Context, key string) error {
	return cm.backend.Delete(ctx, key)
}

// GetOrSet 获取或设置缓存（缓存穿透保护）
func (cm *CacheManager) GetOrSet(ctx context.Context, key string, ttl time.Duration, fn func() (interface{}, error), target interface{}) error {
	return cm.backend.GetOrSet(ctx, key, ttl, fn, target)
}

// GetStats 获取缓存统计信息
func (cm *CacheManager) GetStats() CacheStats {
	return cm.backend.GetStats()
}

// ResetStats 重置统计信息
func (cm *CacheManager) ResetStats() {
	cm.backend.ResetStats()
}

// ClearLocalCache 清空本地缓存
func (cm *CacheManager) ClearLocalCache() {
	cm.backend.ClearLocalCache()
}

// HealthCheck 健康检查
func (cm *CacheManager) HealthCheck(ctx context.Context) error {
	return cm.backend.HealthCheck(ctx)
}

// Close 关闭缓存管理器
func (cm *CacheManager) Close() error {
	return cm.backend.Close()
}

// CacheMiddleware Gin中间件，用于缓存管理
func (cm *CacheManager) CacheMiddleware(ttl time.Duration, keyGenerator func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		cacheKey := keyGenerator(c)
		if cacheKey == "" {
			c.Next()
			return
		}

		var response gin.H
		found, err := cm.Get(c.Request.Context(), cacheKey, &response)
		if err != nil {
			log.Printf("Cache get error: %v", err)
			c.Next()
			return
		}

		if found {
			c.JSON(200, response)
			c.Abort()
			return
		}

		c.Next()

		if c.Writer.Status() == 200 {
			if err := cm.Set(c.Request.Context(), cacheKey, response, ttl); err != nil {
				log.Printf("Cache set error: %v", err)
			}
		}
	}
}

// convertValue 转换值类型
func convertValue(src, dst interface{}) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}
