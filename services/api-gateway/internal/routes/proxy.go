package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

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
			BaseURL: getEnv("USER_SERVICE_URL", "http://user-service:8081"),
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
			BaseURL: getEnv("PAYMENT_SERVICE_URL", "http://payment-service:8083"),
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
func (pm *ProxyManager) SetupProxyRoutes(router *gin.Engine) {
	api := router.Group("/v1")

	// 用户服务路由
	userGroup := api.Group("/users")
	userGroup.Use(pm.createProxy("user"))
	userGroup.Any("/*path", func(c *gin.Context) {})

	// 数据服务路由
	dataGroup := api.Group("/data")
	dataGroup.Use(pm.createProxy("data"))
	dataGroup.Any("/*path", func(c *gin.Context) {})

	// 支付服务路由
	paymentGroup := api.Group("/payments")
	paymentGroup.Use(pm.createProxy("payment"))
	paymentGroup.Any("/*path", func(c *gin.Context) {})

	// 推荐服务路由
	recommendationGroup := api.Group("/recommendations")
	recommendationGroup.Use(pm.createProxy("recommendation"))
	recommendationGroup.Any("/*path", func(c *gin.Context) {})
}

// createProxy 创建代理中间件
func (pm *ProxyManager) createProxy(serviceName string) gin.HandlerFunc {
	service, exists := pm.services[serviceName]
	if !exists {
		return func(c *gin.Context) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "service_not_found",
				"message": fmt.Sprintf("Service %s not found", serviceName),
			})
			c.Abort()
		}
	}

	// 解析目标URL
	targetURL, err := url.Parse(service.BaseURL)
	if err != nil {
		pm.logger.WithError(err).Errorf("Failed to parse service URL: %s", service.BaseURL)
		return func(c *gin.Context) {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "service_configuration_error",
				"message": "Service configuration error",
			})
			c.Abort()
		}
	}

	// 创建反向代理
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// 自定义Director函数
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// 修改请求路径
		req.URL.Path = strings.TrimPrefix(req.URL.Path, service.Prefix)
		if req.URL.Path == "" {
			req.URL.Path = "/"
		}

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

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)

		response := gin.H{
			"error":   "service_unavailable",
			"message": fmt.Sprintf("Service %s is currently unavailable", serviceName),
			"service": serviceName,
		}

		if jsonBytes, err := json.Marshal(response); err == nil {
			w.Write(jsonBytes)
		}
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

		// 设置超时
		ctx := c.Request.Context()
		if service.Timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, service.Timeout)
			defer cancel()
			c.Request = c.Request.WithContext(ctx)
		}

		// 记录请求日志
		pm.logger.WithFields(logrus.Fields{
			"service":    serviceName,
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"user_id":    c.GetString("user_id"),
			"request_id": c.GetString("request_id"),
		}).Info("Proxying request to service")

		// 执行代理
		proxy.ServeHTTP(c.Writer, c.Request)

		// 记录响应日志
		duration := time.Since(startTime)
		pm.logger.WithFields(logrus.Fields{
			"service":     serviceName,
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"status_code": c.Writer.Status(),
			"duration":    duration.String(),
			"user_id":     c.GetString("user_id"),
			"request_id":  c.GetString("request_id"),
		}).Info("Request completed")

		// 阻止Gin继续处理
		c.Abort()
	}
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
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
