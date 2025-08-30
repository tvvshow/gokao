package services

import (
	"context"
	"data-service/internal/database"
	"runtime"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// PerformanceService 性能监控服务
type PerformanceService struct {
	db      *database.DB
	logger  *logrus.Logger
	metrics *PerformanceMetrics
	mu      sync.RWMutex
}

// NewPerformanceService 创建性能监控服务实例
func NewPerformanceService(db *database.DB, logger *logrus.Logger) *PerformanceService {
	return &PerformanceService{
		db:      db,
		logger:  logger,
		metrics: NewPerformanceMetrics(),
	}
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	// API请求指标
	RequestCount    int64                    `json:"request_count"`
	RequestDuration map[string]*DurationStat `json:"request_duration"`
	ErrorCount      map[string]int64         `json:"error_count"`
	
	// 数据库指标
	DBQueryCount    int64                    `json:"db_query_count"`
	DBQueryDuration map[string]*DurationStat `json:"db_query_duration"`
	DBErrorCount    int64                    `json:"db_error_count"`
	
	// 缓存指标
	CacheHitCount  int64 `json:"cache_hit_count"`
	CacheMissCount int64 `json:"cache_miss_count"`
	CacheSetCount  int64 `json:"cache_set_count"`
	
	// 系统指标
	SystemMetrics *SystemMetrics `json:"system_metrics"`
	
	// 自定义指标
	CustomMetrics map[string]interface{} `json:"custom_metrics"`
	
	// 最后更新时间
	LastUpdated time.Time `json:"last_updated"`
	
	mu sync.RWMutex
}

// DurationStat 持续时间统计
type DurationStat struct {
	Count    int64         `json:"count"`
	Total    time.Duration `json:"total"`
	Min      time.Duration `json:"min"`
	Max      time.Duration `json:"max"`
	Average  time.Duration `json:"average"`
	P50      time.Duration `json:"p50"`
	P95      time.Duration `json:"p95"`
	P99      time.Duration `json:"p99"`
	Recent   []time.Duration `json:"-"` // 最近的样本，用于计算百分位数
}

// SystemMetrics 系统指标
type SystemMetrics struct {
	// 内存使用
	MemoryUsage    uint64  `json:"memory_usage"`     // 字节
	MemoryPercent  float64 `json:"memory_percent"`   // 百分比
	
	// CPU使用
	CPUPercent     float64 `json:"cpu_percent"`      // 百分比
	CPUCores       int     `json:"cpu_cores"`        // CPU核心数
	
	// Goroutine
	GoroutineCount int     `json:"goroutine_count"`  // Goroutine数量
	
	// GC统计
	GCCount        uint32  `json:"gc_count"`         // GC次数
	GCPauseTotal   uint64  `json:"gc_pause_total"`   // GC暂停总时间(纳秒)
	GCPauseAvg     uint64  `json:"gc_pause_avg"`     // GC平均暂停时间(纳秒)
	
	// 堆统计
	HeapAlloc      uint64  `json:"heap_alloc"`       // 堆分配内存
	HeapSys        uint64  `json:"heap_sys"`         // 堆系统内存
	HeapIdle       uint64  `json:"heap_idle"`        // 堆空闲内存
	HeapInuse      uint64  `json:"heap_inuse"`       // 堆使用内存
	
	Timestamp      time.Time `json:"timestamp"`
}

// NewPerformanceMetrics 创建性能指标实例
func NewPerformanceMetrics() *PerformanceMetrics {
	return &PerformanceMetrics{
		RequestDuration: make(map[string]*DurationStat),
		ErrorCount:      make(map[string]int64),
		DBQueryDuration: make(map[string]*DurationStat),
		CustomMetrics:   make(map[string]interface{}),
		SystemMetrics:   &SystemMetrics{},
		LastUpdated:     time.Now(),
	}
}

// RecordRequest 记录API请求
func (s *PerformanceService) RecordRequest(endpoint string, duration time.Duration, success bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.metrics.RequestCount++
	
	if !success {
		s.metrics.ErrorCount[endpoint]++
	}
	
	if stat, exists := s.metrics.RequestDuration[endpoint]; exists {
		stat.Record(duration)
	} else {
		stat := NewDurationStat()
		stat.Record(duration)
		s.metrics.RequestDuration[endpoint] = stat
	}
	
	s.metrics.LastUpdated = time.Now()
}

// RecordDBQuery 记录数据库查询
func (s *PerformanceService) RecordDBQuery(operation string, duration time.Duration, success bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.metrics.DBQueryCount++
	
	if !success {
		s.metrics.DBErrorCount++
	}
	
	if stat, exists := s.metrics.DBQueryDuration[operation]; exists {
		stat.Record(duration)
	} else {
		stat := NewDurationStat()
		stat.Record(duration)
		s.metrics.DBQueryDuration[operation] = stat
	}
	
	s.metrics.LastUpdated = time.Now()
}

// RecordCacheHit 记录缓存命中
func (s *PerformanceService) RecordCacheHit() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.metrics.CacheHitCount++
	s.metrics.LastUpdated = time.Now()
}

// RecordCacheMiss 记录缓存未命中
func (s *PerformanceService) RecordCacheMiss() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.metrics.CacheMissCount++
	s.metrics.LastUpdated = time.Now()
}

// RecordCacheSet 记录缓存设置
func (s *PerformanceService) RecordCacheSet() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.metrics.CacheSetCount++
	s.metrics.LastUpdated = time.Now()
}

// UpdateSystemMetrics 更新系统指标
func (s *PerformanceService) UpdateSystemMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.metrics.SystemMetrics = &SystemMetrics{
		MemoryUsage:    m.Alloc,
		MemoryPercent:  float64(m.Alloc) / float64(m.Sys) * 100,
		CPUCores:       runtime.NumCPU(),
		GoroutineCount: runtime.NumGoroutine(),
		GCCount:        m.NumGC,
		GCPauseTotal:   m.PauseTotalNs,
		GCPauseAvg:     m.PauseTotalNs / uint64(m.NumGC+1),
		HeapAlloc:      m.HeapAlloc,
		HeapSys:        m.HeapSys,
		HeapIdle:       m.HeapIdle,
		HeapInuse:      m.HeapInuse,
		Timestamp:      time.Now(),
	}
	
	s.metrics.LastUpdated = time.Now()
}

// GetMetrics 获取性能指标
func (s *PerformanceService) GetMetrics() *PerformanceMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// 更新系统指标
	s.UpdateSystemMetrics()
	
	// 深拷贝指标数据
	metrics := &PerformanceMetrics{
		RequestCount:    s.metrics.RequestCount,
		RequestDuration: make(map[string]*DurationStat),
		ErrorCount:      make(map[string]int64),
		DBQueryCount:    s.metrics.DBQueryCount,
		DBQueryDuration: make(map[string]*DurationStat),
		DBErrorCount:    s.metrics.DBErrorCount,
		CacheHitCount:   s.metrics.CacheHitCount,
		CacheMissCount:  s.metrics.CacheMissCount,
		CacheSetCount:   s.metrics.CacheSetCount,
		SystemMetrics:   s.metrics.SystemMetrics,
		CustomMetrics:   make(map[string]interface{}),
		LastUpdated:     s.metrics.LastUpdated,
	}
	
	// 拷贝请求持续时间统计
	for k, v := range s.metrics.RequestDuration {
		metrics.RequestDuration[k] = v.Copy()
	}
	
	// 拷贝错误统计
	for k, v := range s.metrics.ErrorCount {
		metrics.ErrorCount[k] = v
	}
	
	// 拷贝数据库查询持续时间统计
	for k, v := range s.metrics.DBQueryDuration {
		metrics.DBQueryDuration[k] = v.Copy()
	}
	
	// 拷贝自定义指标
	for k, v := range s.metrics.CustomMetrics {
		metrics.CustomMetrics[k] = v
	}
	
	return metrics
}

// GetSummary 获取性能摘要
func (s *PerformanceService) GetSummary() map[string]interface{} {
	metrics := s.GetMetrics()
	
	// 计算缓存命中率
	totalCacheRequests := metrics.CacheHitCount + metrics.CacheMissCount
	var cacheHitRate float64
	if totalCacheRequests > 0 {
		cacheHitRate = float64(metrics.CacheHitCount) / float64(totalCacheRequests) * 100
	}
	
	// 计算平均响应时间
	var avgResponseTime time.Duration
	totalDuration := time.Duration(0)
	totalRequests := int64(0)
	for _, stat := range metrics.RequestDuration {
		totalDuration += stat.Total
		totalRequests += stat.Count
	}
	if totalRequests > 0 {
		avgResponseTime = totalDuration / time.Duration(totalRequests)
	}
	
	// 计算错误率
	var errorRate float64
	totalErrors := int64(0)
	for _, count := range metrics.ErrorCount {
		totalErrors += count
	}
	if metrics.RequestCount > 0 {
		errorRate = float64(totalErrors) / float64(metrics.RequestCount) * 100
	}
	
	return map[string]interface{}{
		"total_requests":       metrics.RequestCount,
		"avg_response_time_ms": avgResponseTime.Milliseconds(),
		"error_rate_percent":   errorRate,
		"cache_hit_rate_percent": cacheHitRate,
		"db_query_count":       metrics.DBQueryCount,
		"memory_usage_mb":      float64(metrics.SystemMetrics.MemoryUsage) / 1024 / 1024,
		"goroutine_count":      metrics.SystemMetrics.GoroutineCount,
		"gc_count":            metrics.SystemMetrics.GCCount,
		"last_updated":        metrics.LastUpdated,
	}
}

// ResetMetrics 重置指标
func (s *PerformanceService) ResetMetrics() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.metrics = NewPerformanceMetrics()
	s.logger.Info("性能指标已重置")
}

// SetCustomMetric 设置自定义指标
func (s *PerformanceService) SetCustomMetric(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.metrics.CustomMetrics[key] = value
	s.metrics.LastUpdated = time.Now()
}

// StartPeriodicCollection 开始定期收集系统指标
func (s *PerformanceService) StartPeriodicCollection(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	s.logger.Infof("开始定期收集性能指标，间隔: %v", interval)
	
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("停止性能指标收集")
			return
		case <-ticker.C:
			s.UpdateSystemMetrics()
		}
	}
}

// NewDurationStat 创建持续时间统计实例
func NewDurationStat() *DurationStat {
	return &DurationStat{
		Min:    time.Duration(0),
		Max:    time.Duration(0),
		Recent: make([]time.Duration, 0, 1000), // 保留最近1000个样本
	}
}

// Record 记录持续时间
func (d *DurationStat) Record(duration time.Duration) {
	d.Count++
	d.Total += duration
	
	if d.Count == 1 || duration < d.Min {
		d.Min = duration
	}
	if duration > d.Max {
		d.Max = duration
	}
	
	d.Average = d.Total / time.Duration(d.Count)
	
	// 保留最近的样本用于百分位数计算
	d.Recent = append(d.Recent, duration)
	if len(d.Recent) > 1000 {
		d.Recent = d.Recent[1:]
	}
	
	// 计算百分位数（简化实现）
	d.calculatePercentiles()
}

// calculatePercentiles 计算百分位数
func (d *DurationStat) calculatePercentiles() {
	if len(d.Recent) == 0 {
		return
	}
	
	// 简单排序（生产环境中应使用更高效的算法）
	samples := make([]time.Duration, len(d.Recent))
	copy(samples, d.Recent)
	
	// 冒泡排序（简化实现）
	for i := 0; i < len(samples)-1; i++ {
		for j := 0; j < len(samples)-1-i; j++ {
			if samples[j] > samples[j+1] {
				samples[j], samples[j+1] = samples[j+1], samples[j]
			}
		}
	}
	
	// 计算百分位数
	n := len(samples)
	d.P50 = samples[n*50/100]
	d.P95 = samples[n*95/100]
	d.P99 = samples[n*99/100]
}

// Copy 创建持续时间统计的副本
func (d *DurationStat) Copy() *DurationStat {
	return &DurationStat{
		Count:   d.Count,
		Total:   d.Total,
		Min:     d.Min,
		Max:     d.Max,
		Average: d.Average,
		P50:     d.P50,
		P95:     d.P95,
		P99:     d.P99,
	}
}