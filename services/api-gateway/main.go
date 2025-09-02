package main

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "os"
    "time"
    "io"
    "net/url"

    // NEW: for Request-ID generation
    "crypto/rand"
    "encoding/hex"

    // NEW: missing imports for addr formatting and signal handling
    "fmt"
    "os/signal"
    "syscall"

    // NEW: for rate limiter
    "sync"

    // NEW: enable/disable swagger via env flag
    "strings"

    // NEW: Prometheus metrics
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"

    // NEW: Redis for caching
    "github.com/go-redis/redis/v8"

    // NEW: Swagger/OpenAPI integration
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"

    // NEW: generated docs package (Option A: committed to repo)
    "gaokao/docs"
    
    // Security middleware
    "gaokao/internal/middleware"

    // Unified error handling
    "gaokao/pkg/errors"

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
    defaultRatePerSec = 10 // tokens per second
    defaultBurst      = 20 // max burst tokens
)

// NEW: response schemas for Swagger
// NEW: API v1 ping response schema
type PingResponse struct {
    Message string `json:"message" example:"pong"`
}

// NEW: common error response schema (e.g., rate limiting)
type ErrorResponse struct {
    Error string `json:"error" example:"too many requests"`
}

// 已删除简单的corsMiddleware()，使用security中间件的CORS实现

// NEW: structured access log middleware (method, path, status, latency_ms, request_id)
func accessLogMiddleware() gin.HandlerFunc { // NEW: access log
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()
        latency := time.Since(start)
        entry := map[string]any{
            "ts":         time.Now().Format(time.RFC3339Nano),
            "level":      "info",
            "msg":        "access",
            "method":     c.Request.Method,
            "path":       c.Request.URL.Path,
            "status":     c.Writer.Status(),
            "latency_ms": float64(latency.Microseconds()) / 1000.0,
            "request_id": c.Request.Header.Get("X-Request-ID"),
        }
        if b, err := json.Marshal(entry); err == nil {
            log.Println(string(b))
        } else {
            // fallback plain text
            log.Printf("access method=%s path=%s status=%d latency=%s reqid=%s", entry["method"], entry["path"], entry["status"], latency.String(), entry["request_id"]) // nolint:forcetypeassert
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
                cache.logger.WithFields(logrus.Fields{
                    "cache_key": cacheKey,
                    "hit":       true,
                }).Debug("Cache hit")
                c.JSON(http.StatusOK, response)
                c.Abort()
                return
            }
        }

        // Cache miss, continue processing
        c.Next()

        // Cache the response if successful
        if c.Writer.Status() == http.StatusOK {
            // Capture response data
            responseData, exists := c.Get("response_data")
            if exists {
                if data, ok := responseData.([]byte); ok {
                    err := cache.client.Set(c.Request.Context(), cacheKey, data, ttl).Err()
                    if err != nil {
                        cache.logger.WithError(err).Warn("Failed to cache response")
                    } else {
                        cache.logger.WithFields(logrus.Fields{
                            "cache_key": cacheKey,
                            "hit":       false,
                            "ttl":       ttl.String(),
                        }).Debug("Cache miss - response cached")
                    }
                }
            }
        }
    }
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

type rateLimiter struct {
    rate  float64
    burst float64
    m     sync.Map // key: client id (IP) -> *rateBucket
}

func newRateLimiter(ratePerSec, burst int) *rateLimiter {
    rl := &rateLimiter{
        rate:  float64(ratePerSec),
        burst: float64(burst),
    }
    return rl
}

func (rl *rateLimiter) allow(key string) (ok bool, retryAfterSec int) {
    if key == "" {
        key = "unknown"
    }
    now := time.Now()
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
        b.tokens -= 1
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

func rateLimitMiddlewareWithConfig(ratePerSec, burst int) gin.HandlerFunc {
    rl := newRateLimiter(ratePerSec, burst)
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
            key := c.ClientIP()
            if ok, retry := rl.allow(key); !ok {
                c.Writer.Header().Set("Retry-After", fmt.Sprintf("%d", retry))
                c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
                return
            }
        }
        c.Next()
    }
}

func setupRouter() *gin.Engine { // NEW: expose router for tests
    return setupRouterWithLimiter(defaultRatePerSec, defaultBurst)
}

// NEW: initialize Redis cache
func initRedisCache() (*cacheManager, error) {
    redisURL := getEnv("REDIS_URL", "localhost:6379")
    redisPassword := getEnv("REDIS_PASSWORD", "")
    redisDB := 0
    
    if dbStr := getEnv("REDIS_DB", ""); dbStr != "" {
        if db, err := strconv.Atoi(dbStr); err == nil {
            redisDB = db
        }
    }

    client := redis.NewClient(&redis.Options{
        Addr:     redisURL,
        Password: redisPassword,
        DB:       redisDB,
    })

    // Test connection
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := client.Ping(ctx).Err(); err != nil {
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
    // 2) metrics (wrap around the whole chain)
    // 3) access logging (wraps request to capture latency)
    // 4) security headers (should be present on all responses)
    // 5) CORS (may abort preflight; security headers already set)
    // 6) Rate limiter (skip OPTIONS)
    // 7) Unified error handling
    r.Use(requestIDMiddleware())
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
        JWTSecret:          os.Getenv("JWT_SECRET"),
        RateLimitEnabled:   false,
        SecurityHeaders:    true,
    })
    r.Use(securityMiddleware.SecurityHeaders())
    allowedOrigins := []string{"http://localhost:3000", "http://127.0.0.1:3000"}
    r.Use(securityMiddleware.CORS(allowedOrigins))
    r.Use(rateLimitMiddlewareWithConfig(ratePerSec, burst))
    r.Use(errors.ErrorHandler(logger))

    // configure swagger base path
    docs.SwaggerInfo.BasePath = "/"

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
            log.Println("[WARN] Swagger UI is enabled while GIN_MODE=release. For production, set ENABLE_SWAGGER=0 or block /swagger upstream.")
        }
    }

    // @Summary Liveness probe
    // @Tags System
    // @Success 200 {string} string "ok"
    // @Router /healthz [get]
    r.GET("/healthz", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

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

    v1 := r.Group("/api/v1")
    {
        // @Summary Ping
        // @Tags API v1
        // @Produce json
        // @Success 200 {object} PingResponse
        // @Failure 429 {object} ErrorResponse "Too Many Requests"
        // @Router /api/v1/ping [get]
        v1.GET("/ping", func(c *gin.Context) {
            c.JSON(http.StatusOK, gin.H{"message": "pong"})
        })

        // NEW: Proxy routes to microservices
        // Data service routes (universities, majors, recommendations)
        dataServiceGroup := v1.Group("/data")
        {
            dataServiceGroup.Any("/*path", proxyToService(getEnv("DATA_SERVICE_URL", "http://data-service:8082")))
        }

        // User service routes (auth, profile)
        userServiceGroup := v1.Group("/users")
        {
            userServiceGroup.Any("/*path", proxyToService(getEnv("USER_SERVICE_URL", "http://user-service:8081")))
        }

        // Payment service routes
        paymentServiceGroup := v1.Group("/payments")
        {
            paymentServiceGroup.Any("/*path", proxyToService(getEnv("PAYMENT_SERVICE_URL", "http://payment-service:8083")))
        }

        // Recommendation service routes
        recommendationServiceGroup := v1.Group("/recommendations")
        {
            recommendationServiceGroup.Any("/*path", proxyToService(getEnv("RECOMMENDATION_SERVICE_URL", "http://recommendation-service:8084")))
        }
    }

    return r
}

// NEW: Security headers middleware
func securityHeadersMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        h := c.Writer.Header()
        h.Set("X-Content-Type-Options", "nosniff")
        h.Set("X-Frame-Options", "DENY")
        h.Set("X-XSS-Protection", "0")
        h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
        c.Next()
    }
}

// NEW: resolve port from env with default
func getPortFromEnv() string {
    p := os.Getenv("PORT")
    if p == "" {
        return "8080"
    }
    return p
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
        ReadHeaderTimeout: 5 *time.Second,
        ReadTimeout:       10 *time.Second,
        WriteTimeout:      10 *time.Second,
        IdleTimeout:       60 *time.Second,
    }
}

// NEW: core runner that can be controlled via context (test-friendly)
func runWithShutdownContext(srv *http.Server, ctx context.Context) error {
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
    return runWithShutdownContext(srv, ctx)
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

func generateRequestID() string {
    b := make([]byte, 16)
    _, _ = rand.Read(b)
    return hex.EncodeToString(b)
}

// NEW: HTTP proxy function to forward requests to microservices
func proxyToService(targetURL string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Parse target URL
        target, err := url.Parse(targetURL)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid target URL"})
            return
        }

        // Build the full target URL with path
        path := c.Param("path")
        if path == "" {
            path = "/"
        }
        // For recommendation service, we need to prepend /api/v1/recommendations
        if strings.Contains(targetURL, ":10083") {
            target.Path = "/api/v1/recommendations" + path
        } else if strings.Contains(targetURL, ":10081") {
            // For user service, we need to prepend /api/v1 (path already contains /auth)
            target.Path = "/api/v1" + path
        } else {
            target.Path = path
        }
        target.RawQuery = c.Request.URL.RawQuery

        // Create new request
        req, err := http.NewRequest(c.Request.Method, target.String(), c.Request.Body)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
            return
        }

        // Copy headers
        for key, values := range c.Request.Header {
            for _, value := range values {
                req.Header.Add(key, value)
            }
        }

        // Forward request
        client := &http.Client{Timeout: 30 * time.Second}
        resp, err := client.Do(req)
        if err != nil {
            c.JSON(http.StatusBadGateway, gin.H{"error": "service unavailable"})
            return
        }
        defer resp.Body.Close()

        // Copy response headers (exclude CORS headers to avoid conflicts)
        for key, values := range resp.Header {
            // Skip CORS headers as they are handled by API Gateway middleware
            if strings.HasPrefix(strings.ToLower(key), "access-control-") {
                continue
            }
            for _, value := range values {
                c.Writer.Header().Add(key, value)
            }
        }

        // Set status code
        c.Status(resp.StatusCode)

        // Copy response body
        _, err = io.Copy(c.Writer, resp.Body)
        if err != nil {
            log.Printf("Error copying response body: %v", err)
        }
    }
}

// Doc-only placeholders (no-op) to make swagger annotations explicit when handlers are closures.
// These functions are never called.
// @Summary Liveness probe
// @Tags System
// @Success 200 {string} string "ok"
// @Router /healthz [get]
func _docHealthz() {}

// @Summary Readiness probe
// @Tags System
// @Success 200 {string} string "ready"
// @Router /readyz [get]
func _docReadyz() {}

// @Summary Ping
// @Tags API v1
// @Produce json
// @Success 200 {object} PingResponse
// @Failure 429 {object} ErrorResponse "Too Many Requests"
// @Router /api/v1/ping [get]
func _docPing() {}