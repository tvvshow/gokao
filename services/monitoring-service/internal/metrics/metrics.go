package metrics

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// MetricsCollector 指标收集器
type MetricsCollector struct {
	// HTTP指标
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestsInFlight prometheus.Gauge

	// 业务指标
	userRegistrations    prometheus.Counter
	userLogins          prometheus.Counter
	paymentTransactions *prometheus.CounterVec
	membershipActivations prometheus.Counter
	recommendationRequests prometheus.Counter

	// 系统指标
	cpuUsage    prometheus.Gauge
	memoryUsage prometheus.Gauge
	diskUsage   *prometheus.GaugeVec
	networkIO   *prometheus.CounterVec

	// 数据库指标
	dbConnections     *prometheus.GaugeVec
	dbQueryDuration   *prometheus.HistogramVec
	dbQueriesTotal    *prometheus.CounterVec
	dbConnectionsIdle prometheus.Gauge

	// Redis指标
	redisConnections  prometheus.Gauge
	redisCommandsTotal *prometheus.CounterVec
	redisKeyCount     prometheus.Gauge

	// 应用指标
	goroutineCount    prometheus.Gauge
	gcDuration        prometheus.Histogram
	heapSize          prometheus.Gauge
	stackSize         prometheus.Gauge
}

// NewMetricsCollector 创建指标收集器
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		// HTTP指标
		httpRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status_code"},
		),
		httpRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		httpRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Number of HTTP requests currently being processed",
			},
		),

		// 业务指标
		userRegistrations: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "user_registrations_total",
				Help: "Total number of user registrations",
			},
		),
		userLogins: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "user_logins_total",
				Help: "Total number of user logins",
			},
		),
		paymentTransactions: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "payment_transactions_total",
				Help: "Total number of payment transactions",
			},
			[]string{"status", "payment_method"},
		),
		membershipActivations: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "membership_activations_total",
				Help: "Total number of membership activations",
			},
		),
		recommendationRequests: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "recommendation_requests_total",
				Help: "Total number of recommendation requests",
			},
		),

		// 系统指标
		cpuUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "system_cpu_usage_percent",
				Help: "Current CPU usage percentage",
			},
		),
		memoryUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "system_memory_usage_percent",
				Help: "Current memory usage percentage",
			},
		),
		diskUsage: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "system_disk_usage_percent",
				Help: "Current disk usage percentage",
			},
			[]string{"device", "mountpoint"},
		),
		networkIO: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "system_network_io_bytes_total",
				Help: "Total network I/O in bytes",
			},
			[]string{"device", "direction"},
		),

		// 数据库指标
		dbConnections: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "database_connections",
				Help: "Number of database connections",
			},
			[]string{"database", "state"},
		),
		dbQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "database_query_duration_seconds",
				Help:    "Database query duration in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5},
			},
			[]string{"database", "operation"},
		),
		dbQueriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "database_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"database", "operation", "status"},
		),

		// Redis指标
		redisConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "redis_connections",
				Help: "Number of Redis connections",
			},
		),
		redisCommandsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "redis_commands_total",
				Help: "Total number of Redis commands",
			},
			[]string{"command", "status"},
		),
		redisKeyCount: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "redis_keys_count",
				Help: "Number of keys in Redis",
			},
		),

		// 应用指标
		goroutineCount: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "go_goroutines",
				Help: "Number of goroutines",
			},
		),
		gcDuration: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "go_gc_duration_seconds",
				Help:    "Garbage collection duration in seconds",
				Buckets: []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1},
			},
		),
		heapSize: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "go_heap_size_bytes",
				Help: "Current heap size in bytes",
			},
		),
		stackSize: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "go_stack_size_bytes",
				Help: "Current stack size in bytes",
			},
		),
	}
}

// RecordHTTPRequest 记录HTTP请求指标
func (m *MetricsCollector) RecordHTTPRequest(method, endpoint, statusCode string, duration time.Duration) {
	m.httpRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
	m.httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// IncHTTPRequestsInFlight 增加正在处理的HTTP请求数
func (m *MetricsCollector) IncHTTPRequestsInFlight() {
	m.httpRequestsInFlight.Inc()
}

// DecHTTPRequestsInFlight 减少正在处理的HTTP请求数
func (m *MetricsCollector) DecHTTPRequestsInFlight() {
	m.httpRequestsInFlight.Dec()
}

// RecordUserRegistration 记录用户注册
func (m *MetricsCollector) RecordUserRegistration() {
	m.userRegistrations.Inc()
}

// RecordUserLogin 记录用户登录
func (m *MetricsCollector) RecordUserLogin() {
	m.userLogins.Inc()
}

// RecordPaymentTransaction 记录支付交易
func (m *MetricsCollector) RecordPaymentTransaction(status, paymentMethod string) {
	m.paymentTransactions.WithLabelValues(status, paymentMethod).Inc()
}

// RecordMembershipActivation 记录会员激活
func (m *MetricsCollector) RecordMembershipActivation() {
	m.membershipActivations.Inc()
}

// RecordRecommendationRequest 记录推荐请求
func (m *MetricsCollector) RecordRecommendationRequest() {
	m.recommendationRequests.Inc()
}

// RecordDatabaseQuery 记录数据库查询
func (m *MetricsCollector) RecordDatabaseQuery(database, operation, status string, duration time.Duration) {
	m.dbQueriesTotal.WithLabelValues(database, operation, status).Inc()
	m.dbQueryDuration.WithLabelValues(database, operation).Observe(duration.Seconds())
}

// SetDatabaseConnections 设置数据库连接数
func (m *MetricsCollector) SetDatabaseConnections(database, state string, count float64) {
	m.dbConnections.WithLabelValues(database, state).Set(count)
}

// RecordRedisCommand 记录Redis命令
func (m *MetricsCollector) RecordRedisCommand(command, status string) {
	m.redisCommandsTotal.WithLabelValues(command, status).Inc()
}

// SetRedisConnections 设置Redis连接数
func (m *MetricsCollector) SetRedisConnections(count float64) {
	m.redisConnections.Set(count)
}

// SetRedisKeyCount 设置Redis键数量
func (m *MetricsCollector) SetRedisKeyCount(count float64) {
	m.redisKeyCount.Set(count)
}

// CollectSystemMetrics 收集系统指标
func (m *MetricsCollector) CollectSystemMetrics(ctx context.Context) error {
	// CPU使用率
	cpuPercent, err := cpu.PercentWithContext(ctx, time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		m.cpuUsage.Set(cpuPercent[0])
	}

	// 内存使用率
	memInfo, err := mem.VirtualMemoryWithContext(ctx)
	if err == nil {
		m.memoryUsage.Set(memInfo.UsedPercent)
	}

	// 磁盘使用率
	diskInfo, err := disk.PartitionsWithContext(ctx, false)
	if err == nil {
		for _, partition := range diskInfo {
			usage, err := disk.UsageWithContext(ctx, partition.Mountpoint)
			if err == nil {
				m.diskUsage.WithLabelValues(partition.Device, partition.Mountpoint).Set(usage.UsedPercent)
			}
		}
	}

	// 网络I/O
	netInfo, err := net.IOCountersWithContext(ctx, true)
	if err == nil {
		for _, netStat := range netInfo {
			m.networkIO.WithLabelValues(netStat.Name, "sent").Add(float64(netStat.BytesSent))
			m.networkIO.WithLabelValues(netStat.Name, "recv").Add(float64(netStat.BytesRecv))
		}
	}

	return nil
}

// CollectGoMetrics 收集Go运行时指标
func (m *MetricsCollector) CollectGoMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	m.goroutineCount.Set(float64(runtime.NumGoroutine()))
	m.heapSize.Set(float64(memStats.HeapAlloc))
	m.stackSize.Set(float64(memStats.StackInuse))
}

// StartMetricsCollection 启动指标收集
func (m *MetricsCollector) StartMetricsCollection(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.CollectSystemMetrics(ctx)
			m.CollectGoMetrics()
		}
	}
}
