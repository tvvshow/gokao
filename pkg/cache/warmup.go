package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// WarmupConfig 缓存预热配置
type WarmupConfig struct {
	Enabled          bool
	WarmupInterval   time.Duration // 预热间隔
	PreloadKeys      []string      // 预加载的key列表
	PreloadPatterns  []string      // 预加载的key模式
	Concurrency      int           // 并发预热数
	TTL              time.Duration // 预热数据的TTL
}

// WarmupManager 缓存预热管理器
type WarmupManager struct {
	cacheManager *MultiLevelCache
	config       *WarmupConfig
	warmupFuncs  []WarmupFunc
	stopChan     chan struct{}
	wg           sync.WaitGroup
}

// WarmupFunc 预热函数类型
type WarmupFunc func(ctx context.Context) (map[string]interface{}, error)

// NewWarmupManager 创建缓存预热管理器
func NewWarmupManager(cacheManager *MultiLevelCache, config *WarmupConfig) *WarmupManager {
	return &WarmupManager{
		cacheManager: cacheManager,
		config:       config,
		warmupFuncs:  make([]WarmupFunc, 0),
		stopChan:     make(chan struct{}),
	}
}

// RegisterWarmupFunc 注册预热函数
func (wm *WarmupManager) RegisterWarmupFunc(fn WarmupFunc) {
	wm.warmupFuncs = append(wm.warmupFuncs, fn)
}

// Start 启动缓存预热
func (wm *WarmupManager) Start() {
	if !wm.config.Enabled {
		log.Println("Cache warmup is disabled")
		return
	}

	// 立即执行一次预热
	wm.warmupOnce(context.Background())

	// 启动定时预热
	wm.wg.Add(1)
	go wm.warmupScheduler()
}

// Stop 停止缓存预热
func (wm *WarmupManager) Stop() {
	close(wm.stopChan)
	wm.wg.Wait()
}

// warmupScheduler 定时预热调度器
func (wm *WarmupManager) warmupScheduler() {
	defer wm.wg.Done()

	ticker := time.NewTicker(wm.config.WarmupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			wm.warmupOnce(ctx)
			cancel()
		case <-wm.stopChan:
			return
		}
	}
}

// warmupOnce 执行一次预热
func (wm *WarmupManager) warmupOnce(ctx context.Context) {
	log.Printf("Starting cache warmup at %s", time.Now().Format(time.RFC3339))

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, wm.config.Concurrency)

	// 执行所有注册的预热函数
	for _, warmupFunc := range wm.warmupFuncs {
		wg.Add(1)
		go func(fn WarmupFunc) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			data, err := fn(ctx)
			if err != nil {
				log.Printf("Warmup function failed: %v", err)
				return
			}

			// 缓存数据
			for key, value := range data {
				if err := wm.cacheManager.Set(ctx, key, value, wm.config.TTL); err != nil {
					log.Printf("Failed to cache key %s: %v", key, err)
				} else {
					log.Printf("Warmed up cache for key: %s", key)
				}
			}
		}(warmupFunc)
	}

	// 预加载指定的keys
	if len(wm.config.PreloadKeys) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			wm.preloadKeys(ctx)
		}()
	}

	// 预加载匹配模式的keys
	if len(wm.config.PreloadPatterns) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			wm.preloadPatterns(ctx)
		}()
	}

	wg.Wait()
	log.Printf("Cache warmup completed at %s", time.Now().Format(time.RFC3339))
}

// preloadKeys 预加载指定的keys
func (wm *WarmupManager) preloadKeys(ctx context.Context) {
	for _, key := range wm.config.PreloadKeys {
		// 检查key是否已存在
		exists, err := wm.cacheManager.redisClient.Exists(ctx, key).Result()
		if err != nil {
			log.Printf("Failed to check key existence %s: %v", key, err)
			continue
		}

		if exists == 0 {
			// Key不存在，可以触发相应的数据加载逻辑
			log.Printf("Key %s not found, may need data loading", key)
		}
	}
}

// preloadPatterns 预加载匹配模式的keys
func (wm *WarmupManager) preloadPatterns(ctx context.Context) {
	for _, pattern := range wm.config.PreloadPatterns {
		keys, err := wm.cacheManager.redisClient.Keys(ctx, pattern).Result()
		if err != nil {
			log.Printf("Failed to get keys for pattern %s: %v", pattern, err)
			continue
		}

		// 预加载这些keys到本地缓存
		for _, key := range keys {
			data, err := wm.cacheManager.redisClient.Get(ctx, key).Bytes()
			if err != nil {
				log.Printf("Failed to get key %s: %v", key, err)
				continue
			}

			// 加载到本地缓存
			wm.cacheManager.localCache.Set(key, data)
			log.Printf("Preloaded key to local cache: %s", key)
		}
	}
}

// DefaultWarmupConfig 默认预热配置
func DefaultWarmupConfig() *WarmupConfig {
	return &WarmupConfig{
		Enabled:         true,
		WarmupInterval:  1 * time.Hour,
		PreloadKeys:     []string{"universities:list", "majors:popular", "provinces:config"},
		PreloadPatterns: []string{"university:*", "major:*", "province:*"},
		Concurrency:     5,
		TTL:             24 * time.Hour,
	}
}

// ExampleWarmupFuncs 示例预热函数

// WarmupUniversities 预热大学数据
func WarmupUniversities(ctx context.Context) (map[string]interface{}, error) {
	// 这里应该是从数据库或其他数据源加载大学数据的逻辑
	// 返回示例数据
	return map[string]interface{}{
		"universities:list": []interface{}{
			map[string]interface{}{"id": 1, "name": "清华大学", "province": "北京"},
			map[string]interface{}{"id": 2, "name": "北京大学", "province": "北京"},
		},
		"university:1": map[string]interface{}{"id": 1, "name": "清华大学", "ranking": 1},
		"university:2": map[string]interface{}{"id": 2, "name": "北京大学", "ranking": 2},
	}, nil
}

// WarmupMajors 预热专业数据
func WarmupMajors(ctx context.Context) (map[string]interface{}, error) {
	// 这里应该是从数据库或其他数据源加载专业数据的逻辑
	return map[string]interface{}{
		"majors:popular": []interface{}{
			map[string]interface{}{"id": 1, "name": "计算机科学与技术", "category": "工学"},
			map[string]interface{}{"id": 2, "name": "电子信息工程", "category": "工学"},
		},
		"major:1": map[string]interface{}{"id": 1, "name": "计算机科学与技术", "employment_rate": 0.95},
		"major:2": map[string]interface{}{"id": 2, "name": "电子信息工程", "employment_rate": 0.92},
	}, nil
}

// WarmupProvinces 预热省份配置数据
func WarmupProvinces(ctx context.Context) (map[string]interface{}, error) {
	// 这里应该是从数据库或其他数据源加载省份配置的逻辑
	return map[string]interface{}{
		"provinces:config": map[string]interface{}{
			"北京": map[string]interface{}{"gaokao_score": 750, "admission_rules": "平行志愿"},
			"上海": map[string]interface{}{"gaokao_score": 660, "admission_rules": "平行志愿"},
		},
		"province:北京": map[string]interface{}{"gaokao_score": 750, "admission_rules": "平行志愿"},
		"province:上海": map[string]interface{}{"gaokao_score": 660, "admission_rules": "平行志愿"},
	}, nil
}