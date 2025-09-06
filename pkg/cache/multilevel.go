package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// MultiLevelCache 多级缓存实现（L1内存 + L2Redis）
type MultiLevelCache struct {
	redisClient *redis.Client
	localCache  *LocalCache
	stats       *CacheStats
	statsMutex  sync.RWMutex
}

// LocalCache 本地内存缓存实现
// 使用LRU策略和TTL管理
type LocalCache struct {
	cache    sync.Map
	maxSize  int
	evictMutex sync.Mutex
	accessTime map[string]time.Time
}

// NewMultiLevelCache 创建多级缓存管理器
func NewMultiLevelCache(config *CacheConfig) (*MultiLevelCache, error) {
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

	// 初始化本地缓存
	localCache := &LocalCache{
		maxSize:     config.LocalCacheSize,
		accessTime:  make(map[string]time.Time),
	}

	return &MultiLevelCache{
		redisClient: redisClient,
		localCache:  localCache,
		stats:       &CacheStats{},
	}, nil
}

// Get 多级缓存获取（L1 -> L2）
func (mlc *MultiLevelCache) Get(ctx context.Context, key string, target interface{}) (bool, error) {
	mlc.statsMutex.Lock()
	mlc.stats.TotalGets++
	mlc.statsMutex.Unlock()

	// 1. 首先尝试L1本地缓存
	if value, found := mlc.localCache.Get(key); found {
		mlc.statsMutex.Lock()
		mlc.stats.Hits++
		mlc.stats.LocalHits++
		mlc.statsMutex.Unlock()

		if err := json.Unmarshal(value, target); err != nil {
			return false, fmt.Errorf("failed to unmarshal local cache: %w", err)
		}
		return true, nil
	}

	// 2. 尝试L2 Redis缓存
	data, err := mlc.redisClient.Get(ctx, key).Bytes()
	if err == redis.Nil {
		mlc.statsMutex.Lock()
		mlc.stats.Misses++
		mlc.statsMutex.Unlock()
		return false, nil
	}
	if err != nil {
		mlc.statsMutex.Lock()
		mlc.stats.Errors++
		mlc.statsMutex.Unlock()
		return false, fmt.Errorf("failed to get from Redis: %w", err)
	}

	// 3. 解析Redis数据
	if err := json.Unmarshal(data, target); err != nil {
		return false, fmt.Errorf("failed to unmarshal Redis cache: %w", err)
	}

	// 4. 将Redis数据缓存到L1本地缓存（缓存回填）
	mlc.localCache.Set(key, data)

	mlc.statsMutex.Lock()
	mlc.stats.Hits++
	mlc.stats.RedisHits++
	mlc.statsMutex.Unlock()

	return true, nil
}

// Set 多级缓存设置（L1 + L2）
func (mlc *MultiLevelCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	mlc.statsMutex.Lock()
	mlc.stats.TotalSets++
	mlc.statsMutex.Unlock()

	// 序列化数据
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache value: %w", err)
	}

	// 1. 设置L1本地缓存
	mlc.localCache.Set(key, data)

	// 2. 设置L2 Redis缓存
	if ttl == 0 {
		ttl = 24 * time.Hour // 默认TTL
	}

	if err := mlc.redisClient.Set(ctx, key, data, ttl).Err(); err != nil {
		mlc.statsMutex.Lock()
		mlc.stats.Errors++
		mlc.statsMutex.Unlock()
		return fmt.Errorf("failed to set Redis cache: %w", err)
	}

	return nil
}

// Delete 多级缓存删除（L1 + L2）
func (mlc *MultiLevelCache) Delete(ctx context.Context, key string) error {
	mlc.statsMutex.Lock()
	mlc.stats.TotalDeletes++
	mlc.statsMutex.Unlock()

	// 1. 删除L1本地缓存
	mlc.localCache.Delete(key)

	// 2. 删除L2 Redis缓存
	if err := mlc.redisClient.Del(ctx, key).Err(); err != nil {
		mlc.statsMutex.Lock()
		mlc.stats.Errors++
		mlc.statsMutex.Unlock()
		return fmt.Errorf("failed to delete Redis cache: %w", err)
	}

	return nil
}

// GetOrSet 获取或设置缓存（缓存穿透保护）
func (mlc *MultiLevelCache) GetOrSet(ctx context.Context, key string, ttl time.Duration, fn func() (interface{}, error), target interface{}) error {
	// 先尝试获取缓存
	found, err := mlc.Get(ctx, key, target)
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
	if err := mlc.Set(ctx, key, value, ttl); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	// 将结果设置到目标
	if target != nil {
		if err := convertValue(value, target); err != nil {
			return err
		}
	}

	return nil
}

// LocalCache 方法实现

// Get 从本地缓存获取
func (lc *LocalCache) Get(key string) ([]byte, bool) {
	value, found := lc.cache.Load(key)
	if found {
		lc.evictMutex.Lock()
		lc.accessTime[key] = time.Now()
		lc.evictMutex.Unlock()
		return value.([]byte), true
	}
	return nil, false
}

// Set 设置本地缓存（带LRU淘汰）
func (lc *LocalCache) Set(key string, value []byte) {
	lc.evictMutex.Lock()
	defer lc.evictMutex.Unlock()

	// 检查是否需要淘汰
	if lc.cacheSize() >= lc.maxSize {
		lc.evictLRU()
	}

	lc.cache.Store(key, value)
	lc.accessTime[key] = time.Now()
}

// Delete 删除本地缓存
func (lc *LocalCache) Delete(key string) {
	lc.evictMutex.Lock()
	defer lc.evictMutex.Unlock()

	lc.cache.Delete(key)
	delete(lc.accessTime, key)
}

// cacheSize 获取当前缓存大小
func (lc *LocalCache) cacheSize() int {
	count := 0
	lc.cache.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

// evictLRU 淘汰最近最少使用的缓存
func (lc *LocalCache) evictLRU() {
	if len(lc.accessTime) == 0 {
		return
	}

	// 找到最久未访问的key
	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, accessTime := range lc.accessTime {
		if first || accessTime.Before(oldestTime) {
			oldestKey = key
			oldestTime = accessTime
			first = false
		}
	}

	// 淘汰该key
	if oldestKey != "" {
		lc.cache.Delete(oldestKey)
		delete(lc.accessTime, oldestKey)
	}
}

// Clear 清空本地缓存
func (lc *LocalCache) Clear() {
	lc.evictMutex.Lock()
	defer lc.evictMutex.Unlock()

	lc.cache.Range(func(key, value interface{}) bool {
		lc.cache.Delete(key)
		return true
	})
	lc.accessTime = make(map[string]time.Time)
}

// HealthCheck 健康检查
func (mlc *MultiLevelCache) HealthCheck(ctx context.Context) error {
	return mlc.redisClient.Ping(ctx).Err()
}

// Close 关闭缓存管理器
func (mlc *MultiLevelCache) Close() error {
	return mlc.redisClient.Close()
}

// GetStats 获取缓存统计信息
func (mlc *MultiLevelCache) GetStats() CacheStats {
	mlc.statsMutex.RLock()
	defer mlc.statsMutex.RUnlock()
	return *mlc.stats
}

// ResetStats 重置统计信息
func (mlc *MultiLevelCache) ResetStats() {
	mlc.statsMutex.Lock()
	defer mlc.statsMutex.Unlock()
	mlc.stats = &CacheStats{}
}