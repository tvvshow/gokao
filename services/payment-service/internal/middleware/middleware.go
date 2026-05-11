package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/tvvshow/gokao/pkg/response"
	"github.com/tvvshow/gokao/services/payment-service/internal/config"
	"golang.org/x/time/rate"

	pkgmiddleware "github.com/tvvshow/gokao/pkg/middleware"
)

// CORS 跨域中间件（委托 pkg/middleware 统一实现）
func CORS() gin.HandlerFunc {
	return pkgmiddleware.CORS(pkgmiddleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "X-Requested-With", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           3600,
	})
}

// RequestID 请求ID中间件（委托 pkg/middleware 统一实现）
func RequestID() gin.HandlerFunc {
	return pkgmiddleware.RequestID()
}

// RateLimit 限流中间件
func RateLimit(cfg config.RateLimitConfig) gin.HandlerFunc {
	if !cfg.Enable {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	limiter := rate.NewLimiter(rate.Limit(cfg.RPS), cfg.Burst)

	return func(c *gin.Context) {
		if !limiter.Allow() {
			response.AbortWithError(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "rate limit exceeded", nil)
			return
		}
		c.Next()
	}
}

// Auth JWT认证中间件
func Auth(cfg config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			response.AbortWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "missing or invalid authorization header", nil)
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if tokenString == "" {
			response.AbortWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "empty bearer token", nil)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.Secret), nil
		})
		if err != nil || !token.Valid {
			response.AbortWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token", nil)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.AbortWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token claims", nil)
			return
		}

		userID := claimString(claims, "user_id")
		if userID == "" {
			userID = claimString(claims, "sub")
		}
		if userID == "" {
			response.AbortWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "token missing user identity", nil)
			return
		}

		c.Set("user_id", userID)
		if username := claimString(claims, "username"); username != "" {
			c.Set("username", username)
		}
		if role := claimString(claims, "role"); role != "" {
			c.Set("role", role)
		}

		c.Next()
	}
}

func claimString(claims jwt.MapClaims, key string) string {
	raw, ok := claims[key]
	if !ok || raw == nil {
		return ""
	}
	v, ok := raw.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(v)
}

// AdminOnly 管理员权限中间件
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			response.AbortWithError(c, http.StatusForbidden, "INSUFFICIENT_PRIVILEGES", "admin access required", nil)
			return
		}
		c.Next()
	}
}

// VIPOnly VIP权限中间件
func VIPOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			response.AbortWithError(c, http.StatusForbidden, "AUTHENTICATION_REQUIRED", "authentication required", nil)
			return
		}

		// 允许 admin 和 vip 用户访问
		if role != "admin" && role != "vip" {
			response.AbortWithError(c, http.StatusForbidden, "VIP_REQUIRED", "VIP membership required", nil)
			return
		}
		c.Next()
	}
}

// GetUserID 从上下文获取用户ID
func GetUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		return userID.(string)
	}
	return ""
}

// GetUsername 从上下文获取用户名
func GetUsername(c *gin.Context) string {
	if username, exists := c.Get("username"); exists {
		return username.(string)
	}
	return ""
}

// GetRole 从上下文获取用户角色
func GetRole(c *gin.Context) string {
	if role, exists := c.Get("role"); exists {
		return role.(string)
	}
	return ""
}

// GetRequestID 从上下文获取请求ID
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		return requestID.(string)
	}
	return ""
}

// SecurityHeaders 已废弃 —— payment-service main.go 现在直接用 pkg/middleware.SecurityHeaders。
// 该函数从未被 main.go 接入（dead code），删除避免与 pkg 版本漂移。

// Logging 日志中间件
func Logging() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC3339),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}
