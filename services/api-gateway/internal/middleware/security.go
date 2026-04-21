package middleware

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/oktetopython/gaokao/pkg/auth"
	"github.com/oktetopython/gaokao/pkg/errors"
)

// SecurityConfig 安全配置
type SecurityConfig struct {
	JWTSecret          string
	RateLimitEnabled   bool
	RateLimitPerMinute int
	RateLimitBurst     int
	SecurityHeaders    bool
	CSPPolicy          string
	HSTSMaxAge         int
}

// SecurityMiddleware 安全中间件
type SecurityMiddleware struct {
	config  *SecurityConfig
	limiter *rate.Limiter
}

// NewSecurityMiddleware 创建安全中间件
func NewSecurityMiddleware(config *SecurityConfig) *SecurityMiddleware {
	var limiter *rate.Limiter
	if config.RateLimitEnabled {
		limiter = rate.NewLimiter(
			rate.Every(time.Minute/time.Duration(config.RateLimitPerMinute)),
			config.RateLimitBurst,
		)
	}

	return &SecurityMiddleware{
		config:  config,
		limiter: limiter,
	}
}

// SecurityHeaders 安全头中间件
func (s *SecurityMiddleware) SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		if s.config.SecurityHeaders {
			// HSTS
			if s.config.HSTSMaxAge > 0 {
				c.Header("Strict-Transport-Security",
					fmt.Sprintf("max-age=%d; includeSubDomains", s.config.HSTSMaxAge))
			}

			// CSP
			if s.config.CSPPolicy != "" {
				c.Header("Content-Security-Policy", s.config.CSPPolicy)
			}

			// 其他安全头
			c.Header("X-Content-Type-Options", "nosniff")
			c.Header("X-Frame-Options", "DENY")
			c.Header("X-XSS-Protection", "1; mode=block")
			c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
			c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		}

		c.Next()
	}
}

// RateLimit 速率限制中间件
func (s *SecurityMiddleware) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if s.config.RateLimitEnabled && s.limiter != nil {
			if !s.limiter.Allow() {
				apiErr := errors.TooManyRequestsError(60).WithRequestID(c.GetString("request_id"))
				c.JSON(http.StatusTooManyRequests, apiErr)
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

// JWTAuth JWT认证中间件 - 使用共享认证包
func (s *SecurityMiddleware) JWTAuth() gin.HandlerFunc {
	authMiddleware := auth.NewAuthMiddleware(s.config.JWTSecret)
	return authMiddleware.RequireAuth()
}

// OptionalJWTAuth 可选JWT认证中间件 - 使用共享认证包
func (s *SecurityMiddleware) OptionalJWTAuth() gin.HandlerFunc {
	authMiddleware := auth.NewAuthMiddleware(s.config.JWTSecret)
	return authMiddleware.OptionalAuth()
}

// RequireRole 角色验证中间件
func (s *SecurityMiddleware) RequireRole(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "role_required",
				"message": "User role is required",
			})
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "invalid_role",
				"message": "Invalid user role",
			})
			c.Abort()
			return
		}

		// 检查用户角色是否在允许的角色列表中
		for _, requiredRole := range requiredRoles {
			if role == requiredRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error":   "insufficient_permissions",
			"message": "Insufficient permissions for this operation",
		})
		c.Abort()
	}
}

// CORS 跨域中间件
func (s *SecurityMiddleware) CORS(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 检查origin是否在允许列表中
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin || allowedOrigin == "*" {
				allowed = true
				break
			}
		}

		// 清除所有现有的CORS头，防止重复
		header := c.Writer.Header()
		for key := range header {
			if strings.HasPrefix(key, "Access-Control-") {
				header.Del(key)
			}
		}

		// 只有在允许的情况下才设置CORS头
		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "86400")
		}

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RequestID 请求ID中间件
func (s *SecurityMiddleware) RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Int63())
}
