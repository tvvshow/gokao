package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// CacheManager 统一的缓存管理器
type CacheManager struct {
	redisClient *redis.Client
	localCache  sync.Map // 本地内存缓存
	stats       *CacheStats
	statsMutex  sync.RWMutex
}

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

// NewCacheManager 创建新的缓存管理器
func NewCacheManager(config *CacheConfig) (*CacheManager, error) {
	// 初始化Redis客户端
	redisOpts := &redis.Options{
		Addr:         config.RedisURL,
		Password:     config.RedisPassword,
		DB:           config.RedisDB,
		MaxRetries:   config.RedisMaxRetries,
		PoolSize:     config.RedisPoolSize,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	redisClient := redis.NewClient(redisOpts)

	// 测试Redis连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &CacheManager{
		redisClient: redisClient,
		stats:       &CacheStats{},
	}, nil
}

// Get 获取缓存值（多级缓存：本地内存 -> Redis）
func (cm *CacheManager) Get(ctx context.Context, key string, target interface{}) (bool, error) {
	cm.statsMutex.Lock()
	cm.stats.TotalGets++
	cm.statsMutex.Unlock()

	// 1. 首先尝试本地缓存
	if value, found := cm.localCache.Load(key); found {
		cm.statsMutex.Lock()
		cm.stats.Hits++
		cm.stats.LocalHits++
		cm.statsMutex.Unlock()

		if err := json.Unmarshal(value.([]byte), target); err != nil {
			return false, fmt.Errorf("failed to unmarshal local cache: %w", err)
		}
		return true, nil
	}

	// 2. 尝试Redis缓存
	data, err := cm.redisClient.Get(ctx, key).Bytes()
	if err == redis.Nil {
		cm.statsMutex.Lock()
		cm.stats.Misses++
		cm.statsMutex.Unlock()
		return false, nil
	}
	if err != nil {
		cm.statsMutex.Lock()
		cm.stats.Errors++
		cm.statsMutex.Unlock()
		return false, fmt.Errorf("failed to get from Redis: %w", err)
	}

	// 3. 解析Redis数据
	if err := json.Unmarshal(data, target); err != nil {
		return false, fmt.Errorf("failed to unmarshal Redis cache: %w", err)
	}

	// 4. 将Redis数据缓存到本地
	cm.localCache.Store(key, data)

	cm.statsMutex.Lock()
	cm.stats.Hits++
	cm.stats.RedisHits++
	cm.statsMutex.Unlock()

	return true, nil
}

// Set 设置缓存值（多级缓存）
func (cm *CacheManager) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	cm.statsMutex.Lock()
	cm.stats.TotalSets++
	cm.statsMutex.Unlock()

	// 序列化数据
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache value: %w", err)
	}

	// 1. 设置本地缓存
	cm.localCache.Store(key, data)

	// 2. 设置Redis缓存
	if ttl == 0 {
		ttl = 24 * time.Hour // 默认TTL
	}

	if err := cm.redisClient.Set(ctx, key, data, ttl).Err(); err != nil {
		cm.statsMutex.Lock()
		cm.stats.Errors++
		cm.statsMutex.Unlock()
		return fmt.Errorf("failed to set Redis cache: %w", err)
	}

	return nil
}

// Delete 删除缓存（多级缓存）
func (cm *CacheManager) Delete(ctx context.Context, key string) error {
	cm.statsMutex.Lock()
	cm.stats.TotalDeletes++
	cm.statsMutex.Unlock()

	// 1. 删除本地缓存
	cm.localCache.Delete(key)

	// 2. 删除Redis缓存
	if err := cm.redisClient.Del(ctx, key).Err(); err != nil {
		cm.statsMutex.Lock()
		cm.stats.Errors++
		cm.statsMutex.Unlock()
		return fmt.Errorf("failed to delete Redis cache: %w", err)
	}

	return nil
}

// GetOrSet 获取或设置缓存（缓存穿透保护）
func (cm *CacheManager) GetOrSet(ctx context.Context, key string, ttl time.Duration, fn func() (interface{}, error), target interface{}) error {
	// 先尝试获取缓存
	found, err := cm.Get(ctx, key, target)
	if err != nil {
		return err
	}
	if found {
		return nil
	}

	// 缓存未命中，执行回调函数获取数据
	value, err := fn()
	if err != nil {
		return err
	}

	// 设置缓存
	if err := cm.Set(ctx, key, value, ttl); err != nil {
		log.Printf("Warning: failed to set cache for key %s: %v", key, err)
	}

	// 将结果设置到目标
	if target != nil {
		if err := convertValue(value, target); err != nil {
			return err
		}
	}

	return nil
}

// convertValue 转换值类型
func convertValue(src, dst interface{}) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}

// GetStats 获取缓存统计信息
func (cm *CacheManager) GetStats() CacheStats {
	cm.statsMutex.RLock()
	defer cm.statsMutex.RUnlock()
	return *cm.stats
}

// ResetStats 重置统计信息
func (cm *CacheManager) ResetStats() {
	cm.statsMutex.Lock()
	defer cm.statsMutex.Unlock()
	cm.stats = &CacheStats{}
}

// ClearLocalCache 清空本地缓存
func (cm *CacheManager) ClearLocalCache() {
	cm.localCache.Range(func(key, value interface{}) bool {
		cm.localCache.Delete(key)
		return true
	})
}

// HealthCheck 健康检查
func (cm *CacheManager) HealthCheck(ctx context.Context) error {
	return cm.redisClient.Ping(ctx).Err()
}

// Close 关闭缓存管理器
func (cm *CacheManager) Close() error {
	return cm.redisClient.Close()
}

// CacheMiddleware Gin中间件，用于缓存管理
func (cm *CacheManager) CacheMiddleware(ttl time.Duration, keyGenerator func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只缓存GET请求
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		// 生成缓存键
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

		// 继续处理请求
		c.Next()

		// 检查响应状态
		if c.Writer.Status() == 200 {
			// 获取响应数据并缓存
			// 这里需要根据实际响应格式进行调整
			if err := cm.Set(c.Request.Context(), cacheKey, response, ttl); err != nil {
				log.Printf("Cache set error: %v", err)
			}
		}
	}
}