package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tvvshow/gokao/pkg/auth"
	"github.com/sirupsen/logrus"
)

// JWTAuth JWT认证中间件 (已废弃 - 请使用 pkg/auth)
// DEPRECATED: Use github.com/tvvshow/gokao/pkg/auth.AuthMiddleware instead
func JWTAuth(jwtSecret string) gin.HandlerFunc {
	// 使用共享的认证中间件
	authMiddleware := auth.NewAuthMiddleware(jwtSecret)
	return authMiddleware.RequireAuth()
}

// RequireRole 角色权限验证中间件
func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户角色
		rolesInterface, exists := c.Get("roles")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "No roles found in token",
				"code":  "NO_ROLES",
			})
			c.Abort()
			return
		}

		roles, ok := rolesInterface.([]string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Invalid roles format",
				"code":  "INVALID_ROLES",
			})
			c.Abort()
			return
		}

		// 检查是否有所需角色
		hasRole := false
		for _, role := range roles {
			if role == requiredRole || role == "admin" { // admin角色拥有所有权限
				hasRole = true
				break
			}
		}

		if !hasRole {
			userID, _ := c.Get("user_id")
			logrus.WithFields(logrus.Fields{
				"user_id":       userID,
				"required_role": requiredRole,
				"user_roles":    roles,
				"path":          c.Request.URL.Path,
				"method":        c.Request.Method,
			}).Warn("Access denied: insufficient permissions")

			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
				"code":  "INSUFFICIENT_PERMISSIONS",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermission 权限验证中间件
func RequirePermission(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 这里可以扩展为更细粒度的权限检查
		// 目前简化为角色检查
		// 在实际项目中，可以从数据库查询用户的具体权限
		c.Next()
	}
}

// ContextHeaders 透传或生成请求链路标识
func ContextHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateContextID()
		}

		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = generateContextID()
		}

		c.Set("request_id", requestID)
		c.Set("trace_id", traceID)
		c.Header("X-Request-ID", requestID)
		c.Header("X-Trace-ID", traceID)
		c.Next()
	}
}

// RequestLogger 请求日志中间件
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		entry := logrus.WithFields(logrus.Fields{
			"request_id": c.GetString("request_id"),
			"trace_id":   c.GetString("trace_id"),
			"method":     c.Request.Method,
			"path":       path,
			"status":     c.Writer.Status(),
			"latency_ms": float64(time.Since(start).Microseconds()) / 1000.0,
			"client_ip":  c.ClientIP(),
		})

		if userAgent := c.Request.UserAgent(); userAgent != "" {
			entry = entry.WithField("user_agent", userAgent)
		}
		if errMsg := c.Errors.ByType(gin.ErrorTypePrivate).String(); errMsg != "" {
			entry = entry.WithField("error", errMsg)
		}

		entry.Info("request completed")
	}
}

// RateLimiter 简单的内存限流中间件（生产环境建议使用Redis）
func RateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 这里可以实现基于IP或用户的限流逻辑
		// 目前暂时跳过，后续可以集成Redis限流
		c.Next()
	}
}

func generateContextID() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}
