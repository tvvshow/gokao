package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/oktetopython/gaokao/pkg/auth"
)

// JWTAuth JWT认证中间件 (已废弃 - 请使用 pkg/auth)
// DEPRECATED: Use github.com/oktetopython/gaokao/pkg/auth.AuthMiddleware instead
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

// RequestLogger 请求日志中间件
func RequestLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
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

// RateLimiter 简单的内存限流中间件（生产环境建议使用Redis）
func RateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 这里可以实现基于IP或用户的限流逻辑
		// 目前暂时跳过，后续可以集成Redis限流
		c.Next()
	}
}
