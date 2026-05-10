package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/tvvshow/gokao/services/data-service/internal/database"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// CacheService 缓存服务
type CacheService struct {
	db     *database.DB
	logger *logrus.Logger
}

// NewCacheService 创建缓存服务实例
func NewCacheService(db *database.DB, logger *logrus.Logger) *CacheService {
	return &CacheService{
		db:     db,
		logger: logger,
	}
}

// CacheConfig 缓存配置
type CacheConfig struct {
	TTL               time.Duration
	Prefix            string
	EnableCompression bool
}

// DefaultCacheConfigs 默认缓存配置
var DefaultCacheConfigs = map[string]CacheConfig{
	"university":   {TTL: 10 * time.Minute, Prefix: "uni:", EnableCompression: true},
	"major":        {TTL: 10 * time.Minute, Prefix: "maj:", EnableCompression: true},
	"admission":    {TTL: 5 * time.Minute, Prefix: "adm:", EnableCompression: true},
	"search":       {TTL: 2 * time.Minute, Prefix: "search:", EnableCompression: false},
	"statistics":   {TTL: 30 * time.Minute, Prefix: "stats:", EnableCompression: true},
	"hot_searches": {TTL: 10 * time.Minute, Prefix: "hot:", EnableCompression: false},
	"autocomplete": {TTL: 5 * time.Minute, Prefix: "ac:", EnableCompression: false},
}

// Set 设置缓存
func (s *CacheService) Set(ctx context.Context, key string, value interface{}, cacheType string) error {
	if s.db.Redis == nil || !s.db.Config.CacheEnabled {
		return nil
	}

	config, exists := DefaultCacheConfigs[cacheType]
	if !exists {
		config = CacheConfig{TTL: s.db.Config.CacheDefaultTTL, Prefix: "default:"}
	}

	// 序列化数据
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化缓存数据失败: %w", err)
	}

	// 压缩数据（如果启用）
	if config.EnableCompression && len(data) > 1024 {
		// 这里可以添加压缩逻辑，如gzip
		s.logger.Debugf("缓存数据大小: %d bytes", len(data))
	}

	// 构建最终的缓存键
	cacheKey := config.Prefix + key

	// 设置缓存
	err = s.db.Redis.Set(ctx, cacheKey, data, config.TTL).Err()
	if err != nil {
		s.logger.Warnf("设置缓存失败: %v", err)
		return err
	}

	s.logger.Debugf("设置缓存成功: %s", cacheKey)
	return nil
}

// Get 获取缓存
func (s *CacheService) Get(ctx context.Context, key string, result interface{}, cacheType string) error {
	if s.db.Redis == nil || !s.db.Config.CacheEnabled {
		return redis.Nil
	}

	config, exists := DefaultCacheConfigs[cacheType]
	if !exists {
		config = CacheConfig{TTL: s.db.Config.CacheDefaultTTL, Prefix: "default:"}
	}

	// 构建最终的缓存键
	cacheKey := config.Prefix + key

	// 获取缓存
	data, err := s.db.Redis.Get(ctx, cacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			s.logger.Debugf("缓存未命中: %s", cacheKey)
		} else {
			s.logger.Warnf("获取缓存失败: %v", err)
		}
		return err
	}

	// 解压缩数据（如果启用）
	if config.EnableCompression {
		// 这里可以添加解压缩逻辑
	}

	// 反序列化数据
	err = json.Unmarshal([]byte(data), result)
	if err != nil {
		s.logger.Errorf("反序列化缓存数据失败: %v", err)
		return fmt.Errorf("反序列化缓存数据失败: %w", err)
	}

	s.logger.Debugf("缓存命中: %s", cacheKey)
	return nil
}

// Delete 删除缓存
func (s *CacheService) Delete(ctx context.Context, key string, cacheType string) error {
	if s.db.Redis == nil || !s.db.Config.CacheEnabled {
		return nil
	}

	config, exists := DefaultCacheConfigs[cacheType]
	if !exists {
		config = CacheConfig{Prefix: "default:"}
	}

	cacheKey := config.Prefix + key
	err := s.db.Redis.Del(ctx, cacheKey).Err()
	if err != nil {
		s.logger.Warnf("删除缓存失败: %v", err)
		return err
	}

	s.logger.Debugf("删除缓存成功: %s", cacheKey)
	return nil
}

// DeletePattern 按模式删除缓存
func (s *CacheService) DeletePattern(ctx context.Context, pattern string, cacheType string) error {
	if s.db.Redis == nil || !s.db.Config.CacheEnabled {
		return nil
	}

	config, exists := DefaultCacheConfigs[cacheType]
	if !exists {
		config = CacheConfig{Prefix: "default:"}
	}

	searchPattern := config.Prefix + pattern
	keys, err := s.db.Redis.Keys(ctx, searchPattern).Result()
	if err != nil {
		return fmt.Errorf("获取匹配键失败: %w", err)
	}

	if len(keys) > 0 {
		err = s.db.Redis.Del(ctx, keys...).Err()
		if err != nil {
			return fmt.Errorf("删除缓存失败: %w", err)
		}
		s.logger.Debugf("删除 %d 个缓存键: %s", len(keys), searchPattern)
	}

	return nil
}

// Exists 检查缓存是否存在
func (s *CacheService) Exists(ctx context.Context, key string, cacheType string) bool {
	if s.db.Redis == nil || !s.db.Config.CacheEnabled {
		return false
	}

	config, exists := DefaultCacheConfigs[cacheType]
	if !exists {
		config = CacheConfig{Prefix: "default:"}
	}

	cacheKey := config.Prefix + key
	count, err := s.db.Redis.Exists(ctx, cacheKey).Result()
	if err != nil {
		s.logger.Warnf("检查缓存存在性失败: %v", err)
		return false
	}

	return count > 0
}

// SetWithExpiration 设置带过期时间的缓存
func (s *CacheService) SetWithExpiration(ctx context.Context, key string, value interface{}, expiration time.Duration, cacheType string) error {
	if s.db.Redis == nil || !s.db.Config.CacheEnabled {
		return nil
	}

	config, exists := DefaultCacheConfigs[cacheType]
	if !exists {
		config = CacheConfig{Prefix: "default:"}
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化缓存数据失败: %w", err)
	}

	cacheKey := config.Prefix + key
	err = s.db.Redis.Set(ctx, cacheKey, data, expiration).Err()
	if err != nil {
		s.logger.Warnf("设置缓存失败: %v", err)
		return err
	}

	return nil
}

// GetMultiple 批量获取缓存
func (s *CacheService) GetMultiple(ctx context.Context, keys []string, cacheType string) (map[string]interface{}, error) {
	if s.db.Redis == nil || !s.db.Config.CacheEnabled {
		return nil, redis.Nil
	}

	config, exists := DefaultCacheConfigs[cacheType]
	if !exists {
		config = CacheConfig{Prefix: "default:"}
	}

	// 构建缓存键
	cacheKeys := make([]string, len(keys))
	for i, key := range keys {
		cacheKeys[i] = config.Prefix + key
	}

	// 批量获取
	results, err := s.db.Redis.MGet(ctx, cacheKeys...).Result()
	if err != nil {
		return nil, fmt.Errorf("批量获取缓存失败: %w", err)
	}

	// 构建结果map
	resultMap := make(map[string]interface{})
	for i, result := range results {
		if result != nil {
			var data interface{}
			if err := json.Unmarshal([]byte(result.(string)), &data); err == nil {
				resultMap[keys[i]] = data
			}
		}
	}

	return resultMap, nil
}

// SetMultiple 批量设置缓存
func (s *CacheService) SetMultiple(ctx context.Context, keyValues map[string]interface{}, cacheType string) error {
	if s.db.Redis == nil || !s.db.Config.CacheEnabled {
		return nil
	}

	config, exists := DefaultCacheConfigs[cacheType]
	if !exists {
		config = CacheConfig{TTL: s.db.Config.CacheDefaultTTL, Prefix: "default:"}
	}

	// 使用管道批量设置
	pipe := s.db.Redis.Pipeline()

	for key, value := range keyValues {
		data, err := json.Marshal(value)
		if err != nil {
			s.logger.Errorf("序列化缓存数据失败: %v", err)
			continue
		}

		cacheKey := config.Prefix + key
		pipe.Set(ctx, cacheKey, data, config.TTL)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("批量设置缓存失败: %w", err)
	}

	return nil
}

// GetCacheStats 获取缓存统计信息
func (s *CacheService) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	if s.db.Redis == nil {
		return nil, fmt.Errorf("Redis未配置")
	}

	info, err := s.db.Redis.Info(ctx, "memory", "stats", "keyspace").Result()
	if err != nil {
		return nil, fmt.Errorf("获取Redis信息失败: %w", err)
	}

	stats := make(map[string]interface{})

	// 解析Redis INFO输出
	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.Contains(line, ":") && !strings.HasPrefix(line, "#") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				stats[parts[0]] = parts[1]
			}
		}
	}

	// 添加自定义统计
	stats["cache_enabled"] = s.db.Config.CacheEnabled
	stats["cache_configs"] = DefaultCacheConfigs

	return stats, nil
}

// WarmupCache 缓存预热
func (s *CacheService) WarmupCache(ctx context.Context) error {
	if s.db.Redis == nil || !s.db.Config.CacheEnabled {
		return nil
	}

	s.logger.Info("开始缓存预热")

	// 预热常用的静态数据
	go s.warmupStaticData(ctx)

	// 预热热门搜索数据
	go s.warmupHotSearches(ctx)

	// 预热统计数据
	go s.warmupStatistics(ctx)

	s.logger.Info("缓存预热任务已启动")
	return nil
}

// warmupStaticData 预热静态数据
func (s *CacheService) warmupStaticData(ctx context.Context) {
	// 预热省份列表
	provinces := []string{
		"北京市", "天津市", "河北省", "山西省", "内蒙古自治区",
		// ... 其他省份
	}
	s.Set(ctx, "provinces", provinces, "statistics")

	// 预热院校类型
	universityTypes := map[string]string{
		"undergraduate": "本科院校",
		"graduate":      "研究生院校",
		"vocational":    "高职院校",
	}
	s.Set(ctx, "university_types", universityTypes, "statistics")

	// 预热院校层次
	universityLevels := map[string]string{
		"985":                "985工程",
		"211":                "211工程",
		"double_first_class": "双一流",
		"ordinary":           "普通院校",
	}
	s.Set(ctx, "university_levels", universityLevels, "statistics")

	s.logger.Debug("静态数据预热完成")
}

// warmupHotSearches 预热热门搜索
func (s *CacheService) warmupHotSearches(ctx context.Context) {
	// 这里可以预热一些常见的搜索关键词
	hotKeywords := []string{
		"北京大学", "清华大学", "复旦大学", "上海交通大学",
		"计算机科学与技术", "软件工程", "临床医学", "金融学",
	}

	for _, keyword := range hotKeywords {
		cacheKey := fmt.Sprintf("search_suggestions:%s", keyword)
		// 这里可以调用相应的服务生成搜索建议并缓存
		suggestions := []string{keyword} // 简化实现
		s.Set(ctx, cacheKey, suggestions, "autocomplete")
	}

	s.logger.Debug("热门搜索预热完成")
}

// warmupStatistics 预热统计数据
func (s *CacheService) warmupStatistics(ctx context.Context) {
	// 预热基础统计数据
	stats := map[string]interface{}{
		"total_universities":   3000,
		"total_majors":         12000,
		"total_admission_data": 500000,
	}
	s.Set(ctx, "basic_stats", stats, "statistics")

	s.logger.Debug("统计数据预热完成")
}

// ClearAllCache 清空所有缓存
func (s *CacheService) ClearAllCache(ctx context.Context) error {
	if s.db.Redis == nil {
		return nil
	}

	err := s.db.Redis.FlushDB(ctx).Err()
	if err != nil {
		return fmt.Errorf("清空缓存失败: %w", err)
	}

	s.logger.Info("所有缓存已清空")
	return nil
}

// RefreshCache 刷新特定类型的缓存
func (s *CacheService) RefreshCache(ctx context.Context, cacheType string) error {
	if s.db.Redis == nil || !s.db.Config.CacheEnabled {
		return nil
	}

	_, exists := DefaultCacheConfigs[cacheType]
	if !exists {
		return fmt.Errorf("未知的缓存类型: %s", cacheType)
	}

	// 删除该类型的所有缓存
	err := s.DeletePattern(ctx, "*", cacheType)
	if err != nil {
		return fmt.Errorf("删除缓存失败: %w", err)
	}

	s.logger.Infof("已刷新缓存类型: %s", cacheType)
	return nil
}
