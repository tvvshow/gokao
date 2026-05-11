package middleware

import (
	"context"
	"github.com/tvvshow/gokao/services/data-service/internal/handlers"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	pkgmiddleware "github.com/tvvshow/gokao/pkg/middleware"
)

// Logger 日志中间件
func Logger(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Set("trace_id", traceID)
		c.Header("X-Request-ID", requestID)
		c.Header("X-Trace-ID", traceID)

		// 记录请求开始时间
		start := time.Now()

		// 记录请求信息
		logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"trace_id":   traceID,
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"query":      c.Request.URL.RawQuery,
			"user_agent": c.Request.UserAgent(),
			"ip":         c.ClientIP(),
		}).Info("请求开始")

		// 处理请求
		c.Next()

		// 计算处理时间
		duration := time.Since(start)

		// 记录响应信息
		logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"trace_id":   traceID,
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"status":     c.Writer.Status(),
			"duration":   duration.String(),
			"size":       c.Writer.Size(),
		}).Info("请求完成")
	}
}

// CORS 跨域中间件（委托 pkg/middleware 统一实现）
func CORS() gin.HandlerFunc {
	return pkgmiddleware.CORS(pkgmiddleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "X-Request-ID", "X-Trace-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID", "X-Trace-ID"},
		AllowCredentials: true,
		MaxAge:           86400,
	})
}

// RateLimit 限流中间件
func RateLimit(logger *logrus.Logger) gin.HandlerFunc {
	// 简单的内存限流实现，生产环境建议使用Redis
	clientRequests := make(map[string][]time.Time)
	maxRequests := 100 // 每分钟最大请求数
	windowSize := time.Minute

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()

		// 清理过期记录
		if requests, exists := clientRequests[clientIP]; exists {
			var validRequests []time.Time
			for _, reqTime := range requests {
				if now.Sub(reqTime) < windowSize {
					validRequests = append(validRequests, reqTime)
				}
			}
			clientRequests[clientIP] = validRequests
		}

		// 检查当前请求数
		currentRequests := len(clientRequests[clientIP])
		if currentRequests >= maxRequests {
			logger.WithFields(logrus.Fields{
				"ip":               clientIP,
				"current_requests": currentRequests,
				"max_requests":     maxRequests,
			}).Warn("请求超出限制")

			c.JSON(http.StatusTooManyRequests, handlers.NewErrorResponseWithCode(
				"RATE_LIMIT_EXCEEDED",
				"请求过于频繁，请稍后再试",
			))
			c.Abort()
			return
		}

		// 记录当前请求
		clientRequests[clientIP] = append(clientRequests[clientIP], now)

		c.Next()
	}
}

// Recovery 错误恢复中间件
func Recovery(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.WithFields(logrus.Fields{
					"error":      err,
					"request_id": c.GetString("request_id"),
					"method":     c.Request.Method,
					"path":       c.Request.URL.Path,
				}).Error("服务器内部错误")

				c.JSON(http.StatusInternalServerError, handlers.NewErrorResponseWithCode(
					"INTERNAL_SERVER_ERROR",
					"服务器内部错误",
				))
				c.Abort()
			}
		}()

		c.Next()
	}
}

// ValidatePageSize 验证分页参数中间件
func ValidatePageSize(maxSize int) gin.HandlerFunc {
	return func(c *gin.Context) {
		pageSize := c.Query("page_size")
		if pageSize != "" {
			size, err := strconv.Atoi(pageSize)
			if err != nil || size < 1 || size > maxSize {
				c.JSON(http.StatusBadRequest, handlers.NewErrorResponseWithCode(
					"INVALID_PAGE_SIZE",
					"页面大小必须在1到"+strconv.Itoa(maxSize)+"之间",
				))
				c.Abort()
				return
			}
		}

		page := c.Query("page")
		if page != "" {
			pageNum, err := strconv.Atoi(page)
			if err != nil || pageNum < 1 {
				c.JSON(http.StatusBadRequest, handlers.NewErrorResponseWithCode(
					"INVALID_PAGE",
					"页码必须大于0",
				))
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// Timeout 请求超时中间件
func Timeout(timeout time.Duration, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置超时上下文
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// 替换请求上下文
		c.Request = c.Request.WithContext(ctx)

		// 创建一个通道来接收处理完成信号
		done := make(chan bool, 1)

		go func() {
			c.Next()
			done <- true
		}()

		select {
		case <-done:
			// 请求正常完成
			return
		case <-ctx.Done():
			// 请求超时
			logger.WithFields(logrus.Fields{
				"request_id": c.GetString("request_id"),
				"method":     c.Request.Method,
				"path":       c.Request.URL.Path,
				"timeout":    timeout.String(),
			}).Warn("请求超时")

			c.JSON(http.StatusRequestTimeout, handlers.NewErrorResponseWithCode(
				"REQUEST_TIMEOUT",
				"请求超时",
			))
			c.Abort()
		}
	}
}

// Security 已废弃 —— 改用 pkg/middleware.SecurityHeaders（统一头列表 + Referrer-Policy 补全）。
// 删除该函数避免漂移。如需配置驱动开关，用 pkg/middleware.NewSecurityMiddleware + method 形式。
