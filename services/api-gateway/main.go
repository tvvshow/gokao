package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	// NEW: for Request-ID generation
	"crypto/rand"
	"encoding/hex"

	// NEW: Prometheus metrics
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	// NEW: Redis for caching
	"github.com/redis/go-redis/v9"

	// NEW: Swagger/OpenAPI integration
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// Security middleware
	"github.com/tvvshow/gokao/pkg/middleware"

	// Shared Redis init (pkg/database.OpenRedis + pkg/config.LoadRedis)
	sharedcfg "github.com/tvvshow/gokao/pkg/config"
	shareddb "github.com/tvvshow/gokao/pkg/database"

	// Unified error handling
	"github.com/tvvshow/gokao/pkg/errors"

	// JWT 认证
	authPkg "github.com/tvvshow/gokao/pkg/auth"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// @title Gaokao API Gateway
// @version 0.1.0
// @description API gateway for GaokaoHub providing health, readiness, metrics, and v1 endpoints.
// @BasePath /
// @schemes http
// NEW: tag descriptions
// @tag.name System
// @tag.description System endpoints for health, readiness, and root.
// @tag.name API v1
// @tag.description Version 1 public API.

// NEW: default rate limiter configuration
const (
	defaultRatePerSec          = 10 // tokens per second
	defaultBurst               = 20 // max burst tokens
	defaultPublicRate          = 100
	defaultPublicBurst         = 200
	defaultRecommendationRate  = 20
	defaultRecommendationBurst = 40
)

// PingResponse is the API v1 ping response schema for Swagger.
type PingResponse struct {
	Message string `json:"message" example:"pong"`
}

// ErrorResponse is the common error response schema for Swagger.
type ErrorResponse struct {
	Error string `json:"error" example:"too many requests"`
}

func abortWithAPIError(c *gin.Context, logger *logrus.Logger, apiErr *errors.ErrorResponse) {
	if apiErr == nil {
		apiErr = errors.InternalServerError("Internal server error")
	}
	errors.HandleError(c, apiErr, logger)
	c.Abort()
}

// NEW: structured access log middleware (method, path, status, latency_ms, request_id)
func accessLogMiddleware() gin.HandlerFunc { // NEW: access log
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		entry := map[string]any{
			"ts":         time.Now().Format(time.RFC3339Nano),
			"level":      "info",
			"msg":        "access",
			"method":     c.Request.Method,
			"path":       path,
			"status":     c.Writer.Status(),
			"latency_ms": float64(latency.Microseconds()) / 1000.0,
			"request_id": c.GetString("request_id"),
			"trace_id":   c.GetString("trace_id"),
			"client_ip":  c.ClientIP(),
			"user_id":    c.GetString("user_id"),
		}
		if b, err := json.Marshal(entry); err == nil {
			log.Println(string(b))
		} else {
			// fallback plain text
			log.Printf("access method=%s path=%s status=%d latency=%s reqid=%s", entry["method"], entry["path"], entry["status"], latency.String(), entry["request_id"]) //nolint:forcetypeassert
		}
	}
}

// NEW: metrics collector per-router to avoid global registry conflicts
type metrics struct {
	registry        *prometheus.Registry
	reqCounter      *prometheus.CounterVec
	reqDurationHist *prometheus.HistogramVec
}

// NEW: Redis cache manager
type cacheManager struct {
	client *redis.Client
	logger *logrus.Logger
}

// NEW: cache middleware function
func cacheMiddleware(cache *cacheManager, ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only cache GET requests
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		// Skip caching for certain paths
		fullPath := c.FullPath()
		if fullPath == "" {
			fullPath = c.Request.URL.Path
		}
		if strings.HasPrefix(fullPath, "/metrics") || strings.HasPrefix(fullPath, "/swagger") ||
			strings.HasPrefix(fullPath, "/healthz") || strings.HasPrefix(fullPath, "/readyz") {
			c.Next()
			return
		}

		// Generate cache key from request
		cacheKey := fmt.Sprintf("cache:%s:%s", c.Request.Method, c.Request.URL.RequestURI())

		// Try to get from cache
		cachedData, err := cache.client.Get(c.Request.Context(), cacheKey).Bytes()
		if err == nil {
			// Cache hit
			var response gin.H
			if err := json.Unmarshal(cachedData, &response); err == nil {
				// Debug 日志加 IsLevelEnabled 守卫 —— 默认 Info level 下避免每次 hit
				// 都构造 logrus.Fields{}（map alloc）。cache 路径是热路径，每请求都走。
				if cache.logger.IsLevelEnabled(logrus.DebugLevel) {
					cache.logger.WithFields(logrus.Fields{
						"cache_key": cacheKey,
						"hit":       true,
					}).Debug("Cache hit")
				}
				c.JSON(http.StatusOK, response)
				c.Abort()
				return
			}
		}

		// Cache miss — wrap response writer to capture body
		bw := &bodyWriter{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
		c.Writer = bw
		c.Next()

		// Cache the response if successful
		if c.Writer.Status() == http.StatusOK && bw.body.Len() > 0 {
			if err := cache.client.Set(c.Request.Context(), cacheKey, bw.body.Bytes(), ttl).Err(); err != nil {
				cache.logger.WithError(err).Warn("Failed to cache response")
			} else if cache.logger.IsLevelEnabled(logrus.DebugLevel) {
				// 同款 IsLevelEnabled 守卫：避免每次 miss 写缓存后都构造 Fields map。
				cache.logger.WithFields(logrus.Fields{
					"cache_key": cacheKey,
					"hit":       false,
					"ttl":       ttl.String(),
				}).Debug("Cache miss - response cached")
			}
		}
	}
}

// bodyWriter wraps gin.ResponseWriter to capture the response body for caching.
type bodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func newMetrics() *metrics {
	reg := prometheus.NewRegistry()
	mReq := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "gaokao",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)
	mDur := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "gaokao",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request duration in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)
	reg.MustRegister(mReq, mDur)
	return &metrics{registry: reg, reqCounter: mReq, reqDurationHist: mDur}
}

func (m *metrics) middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		status := c.Writer.Status()
		method := c.Request.Method
		// Prefer the matched route path; fallback to raw URL path
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		labels := prometheus.Labels{"method": method, "path": path, "status": fmt.Sprintf("%d", status)}
		m.reqCounter.With(labels).Inc()
		m.reqDurationHist.With(labels).Observe(time.Since(start).Seconds())
	}
}

// NEW: rate limiter structures
type rateBucket struct {
	mu     sync.Mutex
	tokens float64
	last   time.Time
}

// rateLimiter 维护按 client key（IP）分桶的令牌桶。
// sync.Map 长时间运行后会无限增长，cleanupLoop 周期淘汰空闲 bucket 防止内存泄漏。
type rateLimiter struct {
	rate    float64
	burst   float64
	m       sync.Map // key: client id (IP) -> *rateBucket
	idleTTL time.Duration
	now     func() time.Time
	stopCh  chan struct{}
	once    sync.Once
}

type routeRateLimitRule struct {
	name     string
	prefixes []string
	limiter  *rateLimiter
}

const (
	// rate limiter bucket TTL：超过 idleTTL 未访问的 IP bucket 会被淘汰。
	// 选 10min 是 P95 用户会话间隔的保守上界，命中无害（重建零成本）。
	rateLimiterDefaultIdleTTL = 10 * time.Minute
	// 扫描间隔：1min 一次足以让长尾不爆，且 sync.Map.Range 在万级桶下耗时可忽略（<1ms）。
	rateLimiterCleanupInterval = 1 * time.Minute
)

func newRateLimiter(ratePerSec, burst int) *rateLimiter {
	rl := newRateLimiterWithDeps(ratePerSec, burst, rateLimiterDefaultIdleTTL, time.Now)
	go rl.cleanupLoop(rateLimiterCleanupInterval)
	return rl
}

// newRateLimiterWithDeps 暴露 idleTTL / now 注入入口供测试快进时间，
// 不启动 cleanup goroutine，测试用 evictIdle() 直接同步触发淘汰。
func newRateLimiterWithDeps(ratePerSec, burst int, idleTTL time.Duration, nowFn func() time.Time) *rateLimiter {
	if nowFn == nil {
		nowFn = time.Now
	}
	return &rateLimiter{
		rate:    float64(ratePerSec),
		burst:   float64(burst),
		idleTTL: idleTTL,
		now:     nowFn,
		stopCh:  make(chan struct{}),
	}
}

// cleanupLoop 周期触发淘汰；stopCh 关闭即退出，sync.Once 保证 stop 幂等。
func (rl *rateLimiter) cleanupLoop(interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-rl.stopCh:
			return
		case <-t.C:
			rl.evictIdle()
		}
	}
}

// evictIdle 扫描所有 bucket，淘汰超过 idleTTL 未访问的；返回淘汰条数便于测试。
// 锁顺序：先 bucket.mu 读 last，释放后再 m.Delete，避免与 allow() 并发死锁。
func (rl *rateLimiter) evictIdle() int {
	evicted := 0
	now := rl.now()
	rl.m.Range(func(k, v any) bool {
		b := v.(*rateBucket)
		b.mu.Lock()
		idle := now.Sub(b.last) > rl.idleTTL
		b.mu.Unlock()
		if idle {
			rl.m.Delete(k)
			evicted++
		}
		return true
	})
	return evicted
}

// stop 关闭 cleanup goroutine；多次调用安全。
func (rl *rateLimiter) stop() {
	rl.once.Do(func() { close(rl.stopCh) })
}

func (rl *rateLimiter) allow(key string) (ok bool, retryAfterSec int) {
	if key == "" {
		key = "unknown"
	}
	now := rl.now()
	v, _ := rl.m.LoadOrStore(key, &rateBucket{tokens: rl.burst, last: now})
	b := v.(*rateBucket)

	b.mu.Lock()
	defer b.mu.Unlock()

	// Refill
	elapsed := now.Sub(b.last).Seconds()
	if rl.rate > 0 && elapsed > 0 {
		b.tokens += elapsed * rl.rate
		if b.tokens > rl.burst {
			b.tokens = rl.burst
		}
	}
	b.last = now

	if b.tokens >= 1 {
		b.tokens--
		return true, 0
	}

	// Compute retry-after seconds conservatively
	if rl.rate <= 0 {
		return false, 1
	}
	missing := 1 - b.tokens
	sec := int(missing/rl.rate + 0.999) // ceil without importing math
	if sec < 1 {
		sec = 1
	}
	return false, sec
}

func rateLimitMiddlewareWithConfig(ratePerSec, burst int, logger *logrus.Logger) gin.HandlerFunc {
	limiters := buildRouteRateLimitRules(ratePerSec, burst)
	return func(c *gin.Context) {
		// Only limit non-OPTIONS (preflight handled earlier by CORS)
		if c.Request.Method != http.MethodOptions {
			// skip rate limiting for Prometheus metrics endpoint and Swagger docs
			full := c.FullPath()
			if full == "" {
				full = c.Request.URL.Path
			}
			if full == "/metrics" || full == "/swagger/*any" {
				c.Next()
				return
			}

			limiter := selectRateLimiter(limiters, c.Request.URL.Path)
			key := c.ClientIP()
			if ok, retry := limiter.allow(key); !ok {
				abortWithAPIError(c, logger, errors.TooManyRequestsError(retry))
				return
			}
		}
		c.Next()
	}
}

func buildRouteRateLimitRules(defaultRatePerSec, defaultBurst int) []routeRateLimitRule {
	publicRate := getEnvAsIntOrDefault("RATE_LIMIT_PUBLIC_RPS", defaultPublicRate)
	publicBurst := getEnvAsIntOrDefault("RATE_LIMIT_PUBLIC_BURST", defaultPublicBurst)
	recommendationRate := getEnvAsIntOrDefault("RATE_LIMIT_RECOMMENDATION_RPS", defaultRecommendationRate)
	recommendationBurst := getEnvAsIntOrDefault("RATE_LIMIT_RECOMMENDATION_BURST", defaultRecommendationBurst)

	return []routeRateLimitRule{
		{
			name:     "public-data",
			prefixes: []string{"/api/v1/data", "/v1/data"},
			limiter:  newRateLimiter(publicRate, publicBurst),
		},
		{
			name:     "recommendation",
			prefixes: []string{"/api/v1/recommendations", "/v1/recommendations"},
			limiter:  newRateLimiter(recommendationRate, recommendationBurst),
		},
		{
			name:     "default",
			prefixes: nil,
			limiter:  newRateLimiter(defaultRatePerSec, defaultBurst),
		},
	}
}

func selectRateLimiter(rules []routeRateLimitRule, path string) *rateLimiter {
	for _, rule := range rules {
		for _, prefix := range rule.prefixes {
			if strings.HasPrefix(path, prefix) {
				return rule.limiter
			}
		}
	}
	for _, rule := range rules {
		if rule.name == "default" {
			return rule.limiter
		}
	}
	return newRateLimiter(defaultRatePerSec, defaultBurst)
}

func setupRouter() *gin.Engine { // NEW: expose router for tests
	return setupRouterWithLimiter(defaultRatePerSec, defaultBurst)
}

// NEW: initialize Redis cache
func initRedisCache() (*cacheManager, error) {
	// 通过 pkg/config.LoadRedis + pkg/database.OpenRedis 收敛 Redis 初始化（含 ping 校验、统一池参数）。
	redisCfg := sharedcfg.LoadRedis("localhost:6379", 0)
	client, err := shareddb.OpenRedis(redisCfg, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	return &cacheManager{
		client: client,
		logger: logger,
	}, nil
}

// NEW: allow tests to customize limiter
func setupRouterWithLimiter(ratePerSec, burst int) *gin.Engine {
	// Switch to gin.New to avoid duplicate logs, and add Recovery explicitly
	r := gin.New()
	r.Use(gin.Recovery())

	// 创建日志记录器
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	// 初始化Redis缓存
	cache, err := initRedisCache()
	if err != nil {
		log.Printf("Warning: Redis cache initialization failed: %v. Caching will be disabled.", err)
		cache = nil
	}

	// NEW: Register middlewares in order:
	// 1) request ID for correlation
	// 1.1) trace ID for cross-service tracing
	// 2) metrics (wrap around the whole chain)
	// 3) access logging (wraps request to capture latency)
	// 4) security headers (should be present on all responses)
	// 5) CORS (may abort preflight; security headers already set)
	// 6) Rate limiter (skip OPTIONS)
	// 7) Unified error handling
	r.Use(requestIDMiddleware())
	r.Use(traceIDMiddleware())
	m := newMetrics()
	r.Use(m.middleware())
	r.Use(accessLogMiddleware())

	// 添加缓存中间件（如果Redis可用）
	if cache != nil {
		cacheTTL := getEnvAsDuration("CACHE_TTL", "5m")
		r.Use(cacheMiddleware(cache, cacheTTL))
	}

	// 使用security中间件的CORS实现，不使用重复的securityHeadersMiddleware
	securityMiddleware := middleware.NewSecurityMiddleware(&middleware.SecurityConfig{
		JWTSecret:        os.Getenv("JWT_SECRET"),
		JWTIssuer:        os.Getenv("JWT_ISSUER"),
		JWTAudience:      os.Getenv("JWT_AUDIENCE"),
		RateLimitEnabled: false,
		SecurityHeaders:  true,
	})
	r.Use(securityMiddleware.SecurityHeaders())

	// 限制CORS来源，只允许特定的域名
	allowedOrigins := []string{
		"http://localhost:3000",     // 本地开发环境
		"http://127.0.0.1:3000",     // 本地开发环境
		"https://gaokaohub.com",     // 生产环境主域名
		"https://www.gaokaohub.com", // 生产环境www域名
	}

	// 从环境变量获取额外的允许来源
	if extraOrigins := os.Getenv("ALLOWED_ORIGINS"); extraOrigins != "" {
		extraOriginsList := strings.Split(extraOrigins, ",")
		allowedOrigins = append(allowedOrigins, extraOriginsList...)
	}

	r.Use(securityMiddleware.CORS(allowedOrigins))

	// 添加输入验证中间件
	inputValidationMiddleware := middleware.NewInputValidationMiddleware(&middleware.InputValidationConfig{
		MaxBodySize:         10 * 1024 * 1024, // 10MB
		Logger:              logger,
		AllowedContentTypes: []string{"application/json", "application/x-www-form-urlencoded", "multipart/form-data"},
	})
	r.Use(inputValidationMiddleware.Middleware())

	r.Use(rateLimitMiddlewareWithConfig(ratePerSec, burst, logger))
	r.Use(errors.ErrorHandlerMiddleware())

	// configure swagger base path
	// docs.SwaggerInfo.BasePath = "/"

	// NEW: /metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})))

	// NEW: Swagger UI and spec (controlled by ENABLE_SWAGGER env; default: enabled)
	// Set ENABLE_SWAGGER=0 or false to disable in production.
	sw := os.Getenv("ENABLE_SWAGGER")
	swaggerEnabled := sw == "" || strings.EqualFold(sw, "1") || strings.EqualFold(sw, "true") || strings.EqualFold(sw, "yes")
	if swaggerEnabled {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		// NEW: safety warning if Swagger is enabled under release mode
		if gin.Mode() == gin.ReleaseMode {
			log.Println("[WARN] Swagger UI is enabled in release mode. For production, set ENABLE_SWAGGER=0 or block /swagger upstream.")
		}
	}

	// @Summary Liveness probe
	// @Tags System
	// @Success 200 {string} string "ok"
	// @Router /healthz [get]
	r.GET("/healthz", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	r.GET("/health", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	// @Summary Readiness probe
	// @Tags System
	// @Success 200 {string} string "ready"
	// @Router /readyz [get]
	r.GET("/readyz", func(c *gin.Context) { c.String(http.StatusOK, "ready") })

	// @Summary Root
	// @Tags System
	// @Success 200 {string} string "GaokaoHub API Gateway"
	// @Router / [get]
	r.GET("/", func(c *gin.Context) { c.String(http.StatusOK, "GaokaoHub API Gateway") })

	// JWT 认证中间件
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "gaokao-dev-secret-key-change-in-production" //nolint:gosec // G101: dev-only fallback；生产必须 JWT_SECRET 注入，启动时 release 模式日志已警告
	}
	authMiddleware := authPkg.NewAuthMiddleware(jwtSecret)

	// Initialize ProxyManager and set up proxy routes
	proxyManager := NewProxyManager(logger)
	proxyManager.SetupProxyRoutes(r, authMiddleware)

	// Add a simple test route to verify proxy logic
	r.GET("/test/proxy", proxyManager.createProxy("data", "/test/proxy", "/api/v1"))

	return r
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	Name    string
	BaseURL string
	Prefix  string
	Timeout time.Duration
}

// ProxyManager 代理管理器
type ProxyManager struct {
	services map[string]*ServiceConfig
	logger   *logrus.Logger
}

// NewProxyManager 创建代理管理器
func NewProxyManager(logger *logrus.Logger) *ProxyManager {
	services := map[string]*ServiceConfig{
		"user": {
			Name:    "user-service",
			BaseURL: getEnv("USER_SERVICE_URL", "http://user-service:8083"),
			Prefix:  "/api/v1/users",
			Timeout: 30 * time.Second,
		},
		"data": {
			Name:    "data-service",
			BaseURL: getEnv("DATA_SERVICE_URL", "http://data-service:8082"),
			Prefix:  "/api/v1/data",
			Timeout: 30 * time.Second,
		},
		"payment": {
			Name:    "payment-service",
			BaseURL: getEnv("PAYMENT_SERVICE_URL", "http://payment-service:8085"),
			Prefix:  "/api/v1/payments",
			Timeout: 30 * time.Second,
		},
		"recommendation": {
			Name:    "recommendation-service",
			BaseURL: getEnv("RECOMMENDATION_SERVICE_URL", "http://recommendation-service:8084"),
			Prefix:  "/api/v1/recommendations",
			Timeout: 30 * time.Second,
		},
	}

	return &ProxyManager{
		services: services,
		logger:   logger,
	}
}

// SetupProxyRoutes 设置代理路由
func (pm *ProxyManager) SetupProxyRoutes(router *gin.Engine, auth *authPkg.AuthMiddleware) {
	pm.setupProxyRoutes(router, "/api/v1", auth)
	pm.setupProxyRoutes(router, "/v1", auth)
}

func (pm *ProxyManager) setupProxyRoutes(router *gin.Engine, basePrefix string, auth *authPkg.AuthMiddleware) {
	api := router.Group(basePrefix)

	// 兼容旧客户端直接调用 /auth/*。
	legacyAuthGroup := api.Group("/auth")
	legacyAuthGroup.Use(pm.createProxy("user", basePrefix+"/auth", "/api/v1/auth"))
	legacyAuthGroup.Any("/*path")

	// /users/auth/* 公开，其余 /users/* 维持鉴权。
	userGroup := api.Group("/users")
	userGroup.Any("/*path", pm.createUserProxy(basePrefix, auth))

	paymentGroup := api.Group("/payments")
	paymentGroup.Use(auth.RequireAuth(), pm.createProxy("payment", basePrefix+"/payments", "/api/v1/payments"))
	paymentGroup.Any("/*path")

	// 数据查询按前端设计支持匿名访问，有 token 时透传用户信息。
	dataGroup := api.Group("/data")
	dataGroup.Use(auth.OptionalAuth(), pm.createProxy("data", basePrefix+"/data", "/api/v1"))
	dataGroup.Any("/*path")

	recommendationGroup := api.Group("/recommendations")
	recommendationGroup.Use(auth.RequireAuth(), pm.createProxy("recommendation", basePrefix+"/recommendations", "/api/v1/recommendations"))
	recommendationGroup.Any("/*path")
}

func (pm *ProxyManager) createUserProxy(basePrefix string, auth *authPkg.AuthMiddleware) gin.HandlerFunc {
	authProxy := pm.createProxy("user", basePrefix+"/users/auth", "/api/v1/auth")
	userProxy := pm.createProxy("user", basePrefix+"/users", "/api/v1/users")
	requireAuth := auth.RequireAuth()
	authPrefix := strings.TrimRight(basePrefix, "/") + "/users/auth"

	return func(c *gin.Context) {
		if c.Request.URL.Path == authPrefix || strings.HasPrefix(c.Request.URL.Path, authPrefix+"/") {
			authProxy(c)
			return
		}

		requireAuth(c)
		if c.IsAborted() {
			return
		}

		userProxy(c)
	}
}

// createProxy 创建代理中间件
func (pm *ProxyManager) createProxy(serviceName, externalPrefix, backendPrefix string) gin.HandlerFunc {
	service, exists := pm.services[serviceName]
	if !exists {
		return func(c *gin.Context) {
			abortWithAPIError(c, pm.logger, errors.NotFoundError(fmt.Sprintf("service %s", serviceName)))
		}
	}

	// 解析目标URL
	targetURL, err := url.Parse(service.BaseURL)
	if err != nil {
		pm.logger.WithError(err).Errorf("Failed to parse service URL: %s", service.BaseURL)
		return func(c *gin.Context) {
			abortWithAPIError(c, pm.logger, errors.NewError(errors.ErrConfigError, "Service configuration error", map[string]interface{}{
				"service": serviceName,
			}))
		}
	}

	// 创建反向代理
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// 自定义Director函数
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// 按服务真实路由重写，避免前端 /api/v1/{service}/... 与后端 /api/v1/... 漂移。
		req.URL.Path = rewriteServicePath(req.URL.Path, externalPrefix, backendPrefix)

		// 添加请求头
		req.Header.Set("X-Forwarded-Service", service.Name)
		req.Header.Set("X-Gateway-Version", "1.0.0")

		// 传递用户信息
		if userID := req.Header.Get("X-User-ID"); userID != "" {
			req.Header.Set("X-User-ID", userID)
		}
		if username := req.Header.Get("X-Username"); username != "" {
			req.Header.Set("X-Username", username)
		}
		if role := req.Header.Get("X-User-Role"); role != "" {
			req.Header.Set("X-User-Role", role)
		}
	}

	// 自定义错误处理
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		pm.logger.WithError(err).Errorf("Proxy error for service %s", serviceName)
		requestID := r.Header.Get("X-Request-ID")
		apiErr := errors.ServiceUnavailableError(fmt.Sprintf("Service %s is currently unavailable", serviceName), 0).WithRequestID(requestID)
		apiErr.Details = map[string]interface{}{
			"service": serviceName,
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(apiErr.JSON())) //nolint:gosec // G705: apiErr.JSON 是结构化错误对象，已 Content-Type=application/json 序列化，不是用户输入回写
	}

	// 自定义响应修改
	proxy.ModifyResponse = func(resp *http.Response) error {
		// 移除后端服务的CORS头，避免与API Gateway的CORS中间件冲突
		for key := range resp.Header {
			if strings.HasPrefix(key, "Access-Control-") {
				resp.Header.Del(key)
			}
		}

		// 添加响应头
		resp.Header.Set("X-Served-By", service.Name)
		resp.Header.Set("X-Gateway-Timestamp", time.Now().UTC().Format(time.RFC3339))

		return nil
	}

	return func(c *gin.Context) {
		// 记录请求开始时间
		startTime := time.Now()

		// 获取用户信息并设置到请求头
		if userID, exists := c.Get("user_id"); exists {
			c.Request.Header.Set("X-User-ID", userID.(string))
		}
		if username, exists := c.Get("username"); exists {
			c.Request.Header.Set("X-Username", username.(string))
		}
		if role, exists := c.Get("role"); exists {
			c.Request.Header.Set("X-User-Role", role.(string))
		}
		if requestID, exists := c.Get("request_id"); exists {
			c.Request.Header.Set("X-Request-ID", requestID.(string))
		}
		if traceID, exists := c.Get("trace_id"); exists {
			c.Request.Header.Set("X-Trace-ID", traceID.(string))
		}

		// 设置超时
		ctx := c.Request.Context()
		if service.Timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, service.Timeout)
			defer cancel()
			c.Request = c.Request.WithContext(ctx)
		}

		// 记录请求日志
		// 共享 base entry（service/method/path/user_id/request_id/trace_id 6 字段），
		// 请求前后两条日志复用 —— 原实现两次 logrus.Fields{} 构造 + WithFields，每请求
		// 12 次 mapassign。改造后 base 字段只构造 1 次，post 阶段仅追加 2 个变量字段。
		baseLog := pm.logger.WithFields(logrus.Fields{
			"service":    serviceName,
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"user_id":    c.GetString("user_id"),
			"request_id": c.GetString("request_id"),
			"trace_id":   c.GetString("trace_id"),
		})
		baseLog.Info("Proxying request to service")

		// 执行代理
		proxy.ServeHTTP(c.Writer, c.Request)

		// 记录响应日志（在 base 上追加 status_code/duration 两个字段）
		duration := time.Since(startTime)
		baseLog.WithFields(logrus.Fields{
			"status_code": c.Writer.Status(),
			"duration":    duration.String(),
		}).Info("Request completed")

		// 阻止Gin继续处理
		c.Abort()
	}
}

func rewriteServicePath(requestPath, externalPrefix, backendPrefix string) string {
	rest := strings.TrimPrefix(requestPath, externalPrefix)
	if rest == requestPath {
		return requestPath
	}
	if rest == "" || rest == "/" {
		return strings.TrimRight(backendPrefix, "/")
	}
	if !strings.HasPrefix(rest, "/") {
		rest = "/" + rest
	}
	return strings.TrimRight(backendPrefix, "/") + rest
}

// LoadBalancer 负载均衡器
type LoadBalancer struct {
	services map[string][]*ServiceConfig
	current  map[string]int
}

// NewLoadBalancer 创建负载均衡器
func NewLoadBalancer() *LoadBalancer {
	return &LoadBalancer{
		services: make(map[string][]*ServiceConfig),
		current:  make(map[string]int),
	}
}

// AddService 添加服务实例
func (lb *LoadBalancer) AddService(serviceName string, config *ServiceConfig) {
	if lb.services[serviceName] == nil {
		lb.services[serviceName] = make([]*ServiceConfig, 0)
	}
	lb.services[serviceName] = append(lb.services[serviceName], config)
}

// GetService 获取服务实例（轮询）
func (lb *LoadBalancer) GetService(serviceName string) *ServiceConfig {
	services := lb.services[serviceName]
	if len(services) == 0 {
		return nil
	}

	// 轮询算法
	current := lb.current[serviceName]
	service := services[current]
	lb.current[serviceName] = (current + 1) % len(services)

	return service
}

// 辅助函数
// NEW: resolve port from env with default
func getPortFromEnv() string {
	p := os.Getenv("SERVER_PORT")
	if p == "" {
		p = os.Getenv("PORT")
	}
	if p == "" {
		return "8080"
	}
	return strings.TrimPrefix(p, ":")
}

// NEW: build listen addr from port
func getAddr(port string) string {
	return fmt.Sprintf(":%s", port)
}

// NEW: construct *http.Server with sensible timeouts to improve robustness
func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

// NEW: core runner that can be controlled via context (test-friendly)
func runWithShutdownContext(ctx context.Context, srv *http.Server) error {
	// Start server
	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		// Graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

// Wrapper that wires OS signals to the context
func runWithGracefulShutdown(srv *http.Server) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return runWithShutdownContext(ctx, srv)
}

func main() {
	r := setupRouter()

	port := getPortFromEnv()
	addr := getAddr(port)

	srv := newHTTPServer(addr, r)
	if err := runWithGracefulShutdown(srv); err != nil {
		panic(err)
	}
}

// NEW: Request ID middleware
func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.Request.Header.Get("X-Request-ID")
		if rid == "" {
			rid = generateRequestID()
		}
		c.Set("request_id", rid)
		c.Writer.Header().Set("X-Request-ID", rid)
		c.Next()
	}
}

func traceIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tid := c.Request.Header.Get("X-Trace-ID")
		if tid == "" {
			tid = generateTraceID()
		}
		c.Set("trace_id", tid)
		c.Writer.Header().Set("X-Trace-ID", tid)
		c.Next()
	}
}

func generateRequestID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func generateTraceID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Helper functions
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func getEnvAsDuration(key, defaultValue string) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		value = defaultValue
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		log.Printf("Invalid duration for %s: %s, using default %s", key, value, defaultValue) //nolint:gosec // G706: 内部启动配置日志，key 来自代码常量、value 是 env 变量内容，不是用户输入
		duration, _ = time.ParseDuration(defaultValue)
	}
	return duration
}

// Doc-only placeholders (no-op) to make swagger annotations explicit when handlers are closures.
// These functions are never called.
// @Summary Liveness probe
// @Tags System
// @Success 200 {string} string "ok"
// @Router /healthz [get]
//
//nolint:unused // Swagger annotation placeholder for closure-based handler.
func _docHealthz() {}

// @Summary Legacy liveness probe
// @Tags System
// @Success 200 {string} string "ok"
// @Router /health [get]
//
//nolint:unused // Swagger annotation placeholder for closure-based handler.
func _docHealth() {}

// @Summary Readiness probe
// @Tags System
// @Success 200 {string} string "ready"
// @Router /readyz [get]
//
//nolint:unused // Swagger annotation placeholder for closure-based handler.
func _docReadyz() {}

// @Summary Ping
// @Tags API v1
// @Produce json
// @Success 200 {object} PingResponse
// @Failure 429 {object} ErrorResponse "Too Many Requests"
// @Router /api/v1/ping [get]
//
//nolint:unused // Swagger annotation placeholder for closure-based handler.
func _docPing() {}
