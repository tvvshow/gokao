package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware JWT认证中间件
type AuthMiddleware struct {
	jwtSecret   string
	redis       *redis.Client
	tokenExpiry time.Duration
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(jwtSecret string, redis *redis.Client) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret:   jwtSecret,
		redis:       redis,
		tokenExpiry: 24 * time.Hour, // 默认24小时过期
	}
}

// JWTClaims JWT声明
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// RequireAuth 需要认证的中间件
func (a *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := a.extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "缺少认证令牌",
				"code":  "MISSING_TOKEN",
			})
			c.Abort()
			return
		}

		claims, err := a.validateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "无效的认证令牌",
				"code":  "INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		// 检查令牌是否在黑名单中
		if a.isTokenBlacklisted(c.Request.Context(), token) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "令牌已失效",
				"code":  "TOKEN_BLACKLISTED",
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("token", token)

		c.Next()
	}
}

// OptionalAuth 可选认证的中间件
func (a *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := a.extractToken(c)
		if token != "" {
			claims, err := a.validateToken(token)
			if err == nil && !a.isTokenBlacklisted(c.Request.Context(), token) {
				c.Set("user_id", claims.UserID)
				c.Set("username", claims.Username)
				c.Set("role", claims.Role)
				c.Set("token", token)
			}
		}
		c.Next()
	}
}

// RequireRole 需要特定角色的中间件
func (a *AuthMiddleware) RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("role")
		if userRole == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "用户未认证",
				"code":  "UNAUTHENTICATED",
			})
			c.Abort()
			return
		}

		// 检查用户角色是否在允许的角色列表中
		hasRole := false
		for _, role := range roles {
			if userRole == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "权限不足",
				"code":  "INSUFFICIENT_PERMISSIONS",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RefreshToken 刷新令牌中间件
func (a *AuthMiddleware) RefreshToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := a.extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "缺少认证令牌",
				"code":  "MISSING_TOKEN",
			})
			c.Abort()
			return
		}

		claims, err := a.validateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "无效的认证令牌",
				"code":  "INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		// 检查令牌是否即将过期（剩余时间少于1小时）
		if time.Until(claims.ExpiresAt.Time) < time.Hour {
			newToken, err := a.generateToken(claims.UserID, claims.Username, claims.Role)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "生成新令牌失败",
					"code":  "TOKEN_GENERATION_FAILED",
				})
				c.Abort()
				return
			}

			// 将旧令牌加入黑名单
			a.blacklistToken(c.Request.Context(), token, claims.ExpiresAt.Time)

			// 返回新令牌
			c.Header("X-New-Token", newToken)
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("token", token)

		c.Next()
	}
}

// GenerateToken 生成JWT令牌
func (a *AuthMiddleware) GenerateToken(userID, username, role string) (string, error) {
	return a.generateToken(userID, username, role)
}

// generateToken 内部生成令牌方法
func (a *AuthMiddleware) generateToken(userID, username, role string) (string, error) {
	now := time.Now()
	claims := &JWTClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(a.tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "gaokao-system",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.jwtSecret))
}

// extractToken 从请求中提取令牌
func (a *AuthMiddleware) extractToken(c *gin.Context) string {
	// 从Authorization头提取
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	// 从查询参数提取
	if token := c.Query("token"); token != "" {
		return token
	}

	// 从Cookie提取
	if token, err := c.Cookie("token"); err == nil {
		return token
	}

	return ""
}

// validateToken 验证令牌
func (a *AuthMiddleware) validateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// isTokenBlacklisted 检查令牌是否在黑名单中
func (a *AuthMiddleware) isTokenBlacklisted(ctx context.Context, token string) bool {
	if a.redis == nil {
		return false
	}

	key := fmt.Sprintf("blacklist:token:%s", token)
	exists, err := a.redis.Exists(ctx, key).Result()
	if err != nil {
		return false
	}

	return exists > 0
}

// blacklistToken 将令牌加入黑名单
func (a *AuthMiddleware) blacklistToken(ctx context.Context, token string, expiry time.Time) error {
	if a.redis == nil {
		return nil
	}

	key := fmt.Sprintf("blacklist:token:%s", token)
	ttl := time.Until(expiry)
	if ttl <= 0 {
		return nil
	}

	return a.redis.Set(ctx, key, "1", ttl).Err()
}

// BlacklistToken 公开的黑名单令牌方法
func (a *AuthMiddleware) BlacklistToken(ctx context.Context, token string) error {
	claims, err := a.validateToken(token)
	if err != nil {
		return err
	}

	return a.blacklistToken(ctx, token, claims.ExpiresAt.Time)
}

// ValidateToken 公开的令牌验证方法
func (a *AuthMiddleware) ValidateToken(token string) (*JWTClaims, error) {
	return a.validateToken(token)
}

// SetTokenExpiry 设置令牌过期时间
func (a *AuthMiddleware) SetTokenExpiry(expiry time.Duration) {
	a.tokenExpiry = expiry
}

// GetUserFromToken 从令牌中获取用户信息
func (a *AuthMiddleware) GetUserFromToken(token string) (userID, username, role string, err error) {
	claims, err := a.validateToken(token)
	if err != nil {
		return "", "", "", err
	}

	return claims.UserID, claims.Username, claims.Role, nil
}

// AdminOnly 仅管理员访问的中间件
func (a *AuthMiddleware) AdminOnly() gin.HandlerFunc {
	return a.RequireRole("admin", "super_admin")
}

// VIPOnly 仅VIP用户访问的中间件
func (a *AuthMiddleware) VIPOnly() gin.HandlerFunc {
	return a.RequireRole("vip", "admin", "super_admin")
}

// TeacherOnly 仅教师访问的中间件
func (a *AuthMiddleware) TeacherOnly() gin.HandlerFunc {
	return a.RequireRole("teacher", "admin", "super_admin")
}
