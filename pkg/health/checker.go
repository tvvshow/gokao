package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// HealthStatus 健康状态
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusDegraded  HealthStatus = "degraded"
	StatusUnhealthy HealthStatus = "unhealthy"
	StatusUnknown   HealthStatus = "unknown"
)

// CheckResult 健康检查结果
type CheckResult struct {
	Name     string       `json:"name"`
	Status   HealthStatus `json:"status"`
	Message  string       `json:"message"`
	Duration time.Duration `json:"duration"`
	Error    string       `json:"error,omitempty"`
}

// HealthChecker 健康检查器
type HealthChecker struct {
	checks     []HealthCheck
	results    map[string]CheckResult
	resultsMux sync.RWMutex
}

// HealthCheck 健康检查接口
type HealthCheck interface {
	Name() string
	Check(ctx context.Context) CheckResult
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checks:  make([]HealthCheck, 0),
		results: make(map[string]CheckResult),
	}
}

// Register 注册健康检查
func (hc *HealthChecker) Register(check HealthCheck) {
	hc.checks = append(hc.checks, check)
}

// CheckAll 执行所有健康检查
func (hc *HealthChecker) CheckAll(ctx context.Context) map[string]CheckResult {
	results := make(map[string]CheckResult)
	var wg sync.WaitGroup
	var mux sync.Mutex

	for _, check := range hc.checks {
		wg.Add(1)
		go func(c HealthCheck) {
			defer wg.Done()
			result := c.Check(ctx)
			mux.Lock()
			results[c.Name()] = result
			mux.Unlock()
		}(check)
	}

	wg.Wait()

	hc.resultsMux.Lock()
	hc.results = results
	hc.resultsMux.Unlock()

	return results
}

// GetResults 获取最近的健康检查结果
func (hc *HealthChecker) GetResults() map[string]CheckResult {
	hc.resultsMux.RLock()
	defer hc.resultsMux.RUnlock()
	return hc.results
}

// OverallStatus 获取整体健康状态
func (hc *HealthChecker) OverallStatus() HealthStatus {
	hc.resultsMux.RLock()
	defer hc.resultsMux.RUnlock()

	if len(hc.results) == 0 {
		return StatusUnknown
	}

	for _, result := range hc.results {
		if result.Status == StatusUnhealthy {
			return StatusUnhealthy
		}
	}

	for _, result := range hc.results {
		if result.Status == StatusDegraded {
			return StatusDegraded
		}
	}

	return StatusHealthy
}

// HTTPHandler HTTP健康检查处理器
func (hc *HealthChecker) HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		results := hc.CheckAll(ctx)
		
		w.Header().Set("Content-Type", "application/json")
		
		status := hc.OverallStatus()
		if status == StatusUnhealthy {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else if status == StatusDegraded {
			w.WriteHeader(http.StatusOK) // 或者206 Partial Content
		} else {
			w.WriteHeader(http.StatusOK)
		}
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  status,
			"results": results,
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}
}

// 具体的健康检查实现

// DatabaseHealthCheck 数据库健康检查
type DatabaseHealthCheck struct {
	DB *gorm.DB
}

func (d *DatabaseHealthCheck) Name() string {
	return "database"
}

func (d *DatabaseHealthCheck) Check(ctx context.Context) CheckResult {
	start := time.Now()
	
	var result int
	err := d.DB.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error
	
	duration := time.Since(start)
	
	if err != nil {
		return CheckResult{
			Name:     d.Name(),
			Status:   StatusUnhealthy,
			Message:  "Database connection failed",
			Duration: duration,
			Error:    err.Error(),
		}
	}
	
	if result != 1 {
		return CheckResult{
			Name:     d.Name(),
			Status:   StatusDegraded,
			Message:  "Database query returned unexpected result",
			Duration: duration,
			Error:    fmt.Sprintf("Expected 1, got %d", result),
		}
	}
	
	return CheckResult{
		Name:     d.Name(),
		Status:   StatusHealthy,
		Message:  "Database is healthy",
		Duration: duration,
	}
}

// RedisHealthCheck Redis健康检查
type RedisHealthCheck struct {
	Client *redis.Client
}

func (r *RedisHealthCheck) Name() string {
	return "redis"
}

func (r *RedisHealthCheck) Check(ctx context.Context) CheckResult {
	start := time.Now()
	
	err := r.Client.Ping(ctx).Err()
	duration := time.Since(start)
	
	if err != nil {
		return CheckResult{
			Name:     r.Name(),
			Status:   StatusUnhealthy,
			Message:  "Redis connection failed",
			Duration: duration,
			Error:    err.Error(),
		}
	}
	
	return CheckResult{
		Name:     r.Name(),
		Status:   StatusHealthy,
		Message:  "Redis is healthy",
		Duration: duration,
	}
}

// APIConnectivityHealthCheck API连通性检查
type APIConnectivityHealthCheck struct {
	Name        string
	URL         string
	Timeout     time.Duration
	ExpectCode  int
}

func (a *APIConnectivityHealthCheck) Check(ctx context.Context) CheckResult {
	start := time.Now()
	
	client := &http.Client{
		Timeout: a.Timeout,
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", a.URL, nil)
	if err != nil {
		return CheckResult{
			Name:     a.Name,
			Status:   StatusUnhealthy,
			Message:  "Failed to create request",
			Duration: time.Since(start),
			Error:    err.Error(),
		}
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return CheckResult{
			Name:     a.Name,
			Status:   StatusUnhealthy,
			Message:  "API request failed",
			Duration: time.Since(start),
			Error:    err.Error(),
		}
	}
	defer resp.Body.Close()
	
	duration := time.Since(start)
	
	if resp.StatusCode != a.ExpectCode {
		return CheckResult{
			Name:     a.Name,
			Status:   StatusDegraded,
			Message:  fmt.Sprintf("Unexpected status code: %d", resp.StatusCode),
			Duration: duration,
			Error:    fmt.Sprintf("Expected %d, got %d", a.ExpectCode, resp.StatusCode),
		}
	}
	
	return CheckResult{
		Name:     a.Name,
		Status:   StatusHealthy,
		Message:  "API connectivity is healthy",
		Duration: duration,
	}
}

// DiskSpaceHealthCheck 磁盘空间检查
type DiskSpaceHealthCheck struct {
	Path      string
	Threshold uint64 // 阈值（字节）
}

func (d *DiskSpaceHealthCheck) Name() string {
	return "disk_space"
}

func (d *DiskSpaceHealthCheck) Check(ctx context.Context) CheckResult {
	// 在实际实现中，这里应该使用系统调用来检查磁盘空间
	// 这里返回一个模拟的健康状态
	return CheckResult{
		Name:     d.Name(),
		Status:   StatusHealthy,
		Message:  "Disk space is sufficient",
		Duration: 0,
	}
}

// MemoryUsageHealthCheck 内存使用检查
type MemoryUsageHealthCheck struct {
	Threshold float64 // 内存使用率阈值（0-1）
}

func (m *MemoryUsageHealthCheck) Name() string {
	return "memory_usage"
}

func (m *MemoryUsageHealthCheck) Check(ctx context.Context) CheckResult {
	// 在实际实现中，这里应该使用系统调用来检查内存使用
	// 这里返回一个模拟的健康状态
	return CheckResult{
		Name:     m.Name(),
		Status:   StatusHealthy,
		Message:  "Memory usage is normal",
		Duration: 0,
	}
}