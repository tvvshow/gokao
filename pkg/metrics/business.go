package metrics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// BusinessMetrics 业务指标收集器
type BusinessMetrics struct {
	// 用户相关指标
	userRegistrations *prometheus.CounterVec
	userLogins        *prometheus.CounterVec
	userSessions      *prometheus.GaugeVec

	// 推荐相关指标
	recommendationRequests *prometheus.CounterVec
	recommendationLatency  *prometheus.HistogramVec
	recommendationSuccess  *prometheus.CounterVec

	// 支付相关指标
	paymentRequests *prometheus.CounterVec
	paymentAmounts  *prometheus.HistogramVec
	paymentSuccess  *prometheus.CounterVec

	// 数据服务指标
	dataQueries     *prometheus.CounterVec
	queryLatency    *prometheus.HistogramVec
	cacheHitRate    *prometheus.GaugeVec

	// 业务健康指标
	serviceAvailability *prometheus.GaugeVec
	concurrentUsers     prometheus.Gauge

	mu sync.Mutex
}

// NewBusinessMetrics 创建业务指标收集器
func NewBusinessMetrics() *BusinessMetrics {
	return &BusinessMetrics{
		userRegistrations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gaokao_user_registrations_total",
				Help: "Total number of user registrations",
			},
			[]string{"province", "channel"},
		),

		userLogins: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gaokao_user_logins_total",
				Help: "Total number of user logins",
			},
			[]string{"province", "status"},
		),

		userSessions: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gaokao_active_sessions",
				Help: "Number of active user sessions",
			},
			[]string{"province"},
		),

		recommendationRequests: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gaokao_recommendation_requests_total",
				Help: "Total number of recommendation requests",
			},
			[]string{"algorithm", "user_type"},
		),

		recommendationLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gaokao_recommendation_latency_seconds",
				Help:    "Recommendation request latency in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"algorithm"},
		),

		recommendationSuccess: promauto.NewCounterVec(
			prometheus.CounterOpts{
	极			Name: "gaokao_recommendation_success_total",
				Help: "Total number of successful recommendations",
			},
			[]string{"algorithm", "satisfaction"},
		),

		paymentRequests: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gaokao_payment_requests_total",
		极		Help: "Total number of payment requests",
			},
			[]string{"method", "status"},
		),

		paymentAmounts: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gaokao_payment_amounts",
				Help:    "Distribution of payment amounts",
				Buckets: []float64{10, 50, 100, 200, 500, 1000, 2000},
			},
			[]string{"method"},
		),

		paymentSuccess: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gaokao_payment_success_total",
				Help: "Total number of successful payments",
			},
			[]string{"method"},
		),

		dataQueries: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gaokao_data_queries_total",
				Help: "Total number of data queries",
			},
			[]string{"type", "source"},
		),

		queryLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gaokao_query_latency_seconds",
				Help:    "Data query latency in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"type"},
		),

		cacheHitRate: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gaokao_cache_hit_rate",
				Help: "Cache hit rate percentage",
			},
			[]string{"level"},
		),

		serviceAvailability: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gaokao_service_availability",
				Help: "Service availability status (1=available, 0=unavailable)",
			},
			[]string{"service"},
		),

		concurrentUsers: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "gaokao_concurrent_users",
				Help: "Number of concurrent users",
			},
		),
	}
}

// User Metrics

// RecordUserRegistration 记录用户注册
func (m *BusinessMetrics) RecordUserRegistration(province, channel string) {
	m.userRegistrations.WithLabelValues(province, channel).Inc()
}

// RecordUserLogin 记录用户登录
func (m *BusinessMetrics) RecordUserLogin(province, status string) {
	m.userLogins.WithLabelValues(province, status).Inc()
}

// SetActiveSessions 设置活跃会话数
func (m *BusinessMetrics) SetActiveSessions(province string, count int) {
	m.userSessions.WithLabelValues(province).Set(float64(count))
}

// SetConcurrentUsers 设置并发用户数
func (m *BusinessMetrics) SetConcurrentUsers(count int) {
	m.concurrentUsers.Set(float64(count))
}

// Recommendation Metrics

// RecordRecommendationRequest 记录推荐请求
func (m *BusinessMetrics) RecordRecommendationRequest(algorithm, userType string) {
	m.recommendationRequests.WithLabelValues(algorithm, userType).Inc()
}

// RecordRecommendationLatency 记录推荐延迟
func (m *BusinessMetrics) RecordRecommendationLatency(algorithm string, duration time.Duration) {
	m.recommendationLatency.WithLabelValues(algorithm).Observe(duration.Seconds())
}

// RecordRecommendationSuccess 记录推荐成功
func (m *BusinessMetrics) RecordRecommendationSuccess(algorithm, satisfaction string) {
	m.recommendationSuccess.WithLabelValues(algorithm, satisfaction).Inc()
}

// Payment Metrics

// RecordPaymentRequest 记录支付请求
func (m *BusinessMetrics) RecordPaymentRequest(method, status string) {
	m.paymentRequests.WithLabelValues(method, status).Inc()
}

// RecordPaymentAmount 记录支付金额
func (m *BusinessMetrics) RecordPaymentAmount(method string, amount float64) {
	m.paymentAmounts.WithLabelValues(method).Observe(amount)
}

// RecordPaymentSuccess 记录支付成功
func (m *BusinessMetrics) RecordPaymentSuccess(method string) {
	m.paymentSuccess.WithLabelValues(method).Inc()
}

// Data Service Metrics

// RecordDataQuery 记录数据查询
func (m *BusinessMetrics) RecordDataQuery(queryType, source string) {
	m.dataQueries.WithLabelValues(queryType, source).Inc()
}

// RecordQueryLatency 记录查询延迟
func (m *BusinessMetrics) RecordQueryLatency(queryType string, duration time.Duration) {
	m.queryLatency.WithLabelValues(queryType).Observe(duration.Seconds())
}

// SetCacheHitRate 设置缓存命中率
func (m *BusinessMetrics) SetCacheHitRate(level string, hitRate float64) {
	m.cacheHitRate.WithLabelValues(level).Set(hitRate)
}

// Service Health Metrics

// SetServiceAvailability 设置服务可用性
func (m *BusinessMetrics) SetServiceAvailability(service string, available bool) {
	var value float64
	if available {
		value = 1
	}
	m.serviceAvailability.WithLabelValues(service).Set(value)
}

// Business-specific metrics

// RecordVolunteerSubmission 记录志愿提交
func (m *BusinessMetrics) RecordVolunteerSubmission(province, schoolType string, count int) {
	// 这里可以添加自定义的志愿提交指标
}

// RecordAdmissionResult 记录录取结果
func (m *BusinessMetrics) RecordAdmissionResult(province, result string) {
	// 这里可以添加自定义的录取结果指标
}

// RecordScoreAnalysis 记录分数分析
func (m *BusinessMetrics) RecordScoreAnalysis(province string, score float64, ranking int) {
	// 这里可以添加自定义的分数分析指标
}

// MetricsMiddleware 指标收集中间件
// 这是一个示例中间件，需要在具体的web框架中实现
func MetricsMiddleware(metrics *BusinessMetrics) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// 包装ResponseWriter来捕获状态码
			wrapped := &responseWriterWrapper{ResponseWriter: w}
			
			next.ServeHTTP(wrapped, r)
			
			duration := time.Since(start)
			
			// 记录请求指标
			metrics.RecordDataQuery(
				r.URL.Path, 
				"api",
			)
			metrics.RecordQueryLatency(
				r.URL.Path, 
				duration,
			)
		})
	}
}

// responseWriterWrapper 包装ResponseWriter来捕获状态码
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterWrapper) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// GetDefaultMetrics 获取默认的业务指标实例（单例模式）
var (
	defaultMetrics     *BusinessMetrics
	defaultMetricsOnce sync.Once
)

func GetDefaultMetrics() *BusinessMetrics {
	defaultMetricsOnce.Do(func() {
		defaultMetrics = NewBusinessMetrics()
	})
	return defaultMetrics
}