package metrics

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// MetricsCollector 指标收集器
type MetricsCollector struct {
	cpuUsage      float64
	memoryUsage   float64
	dbConnections int
	dbMaxConn     int
	customMetrics map[string]float64
	mu            sync.RWMutex
}

// NewMetricsCollector 创建新的指标收集器
func NewMetricsCollector() *MetricsCollector {
	mc := &MetricsCollector{
		customMetrics: make(map[string]float64),
		dbMaxConn:     100, // 默认最大连接数
	}
	// 启动后台指标收集
	go mc.collectMetrics()
	return mc
}

// collectMetrics 后台收集系统指标
func (mc *MetricsCollector) collectMetrics() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		mc.mu.Lock()
		// 简化的CPU使用率估算
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		mc.memoryUsage = float64(m.Alloc) / float64(m.Sys) * 100
		mc.mu.Unlock()
	}
}

// GetCPUUsage 获取CPU使用率
func (mc *MetricsCollector) GetCPUUsage(ctx context.Context) (float64, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.cpuUsage, nil
}

// GetMemoryUsage 获取内存使用率
func (mc *MetricsCollector) GetMemoryUsage(ctx context.Context) (float64, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.memoryUsage, nil
}

// GetDatabaseConnectionUsage 获取数据库连接使用率
func (mc *MetricsCollector) GetDatabaseConnectionUsage(ctx context.Context) (float64, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	if mc.dbMaxConn == 0 {
		return 0, nil
	}
	return float64(mc.dbConnections) / float64(mc.dbMaxConn) * 100, nil
}

// GetMetricByQuery 通过查询获取指标值
func (mc *MetricsCollector) GetMetricByQuery(ctx context.Context, query string) (float64, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	if value, ok := mc.customMetrics[query]; ok {
		return value, nil
	}
	return 0, fmt.Errorf("metric not found: %s", query)
}

// SetCPUUsage 设置CPU使用率
func (mc *MetricsCollector) SetCPUUsage(usage float64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.cpuUsage = usage
}

// SetMemoryUsage 设置内存使用率
func (mc *MetricsCollector) SetMemoryUsage(usage float64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.memoryUsage = usage
}

// SetDatabaseConnections 设置数据库连接数
func (mc *MetricsCollector) SetDatabaseConnections(active, max int) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.dbConnections = active
	mc.dbMaxConn = max
}

// SetCustomMetric 设置自定义指标
func (mc *MetricsCollector) SetCustomMetric(name string, value float64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.customMetrics[name] = value
}

// GetAllMetrics 获取所有指标
func (mc *MetricsCollector) GetAllMetrics() map[string]float64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	result := make(map[string]float64)
	result["cpu_usage"] = mc.cpuUsage
	result["memory_usage"] = mc.memoryUsage
	result["db_connection_usage"] = float64(mc.dbConnections) / float64(mc.dbMaxConn) * 100
	for k, v := range mc.customMetrics {
		result[k] = v
	}
	return result
}

// StartMetricsCollection 启动指标收集
func (mc *MetricsCollector) StartMetricsCollection(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			mc.mu.Lock()
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			mc.memoryUsage = float64(m.Alloc) / float64(m.Sys) * 100
			mc.mu.Unlock()
		}
	}
}
