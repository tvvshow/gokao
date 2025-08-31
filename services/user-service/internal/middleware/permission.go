package middleware

import (
	"net/http"
	"strings"
	"time"

	"user-service/internal/models"
	"user-service/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Permission 权限中间件
type Permission struct {
	userService *services.UserService
	jwtSecret   string
}

// NewPermission 创建权限中间件
func NewPermission(userService *services.UserService, jwtSecret string) *Permission {
	return &Permission{
		userService: userService,
		jwtSecret:   jwtSecret,
	}
}

// RequireAuth 需要登录认证
func (p *Permission) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取token
		token := p.extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "未提供认证token",
			})
			c.Abort()
			return
		}

		// 验证token
		claims, err := p.validateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "无效的认证token: " + err.Error(),
			})
			c.Abort()
			return
		}

		// 检查用户是否存在
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "无效的用户ID",
			})
			c.Abort()
			return
		}

		user, err := p.userService.GetUserByID(userID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "用户不存在",
			})
			c.Abort()
			return
		}

		// 检查用户状态
		if user.Status != models.UserStatusActive {
			c.JSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Message: "用户账户已被禁用",
			})
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		c.Set("user_id", user.ID)
		c.Set("user", user)
		c.Next()
	}
}

// RequireMembership 需要会员权限
func (p *Permission) RequireMembership(minLevel string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "未登录",
			})
			c.Abort()
			return
		}

		// 获取用户会员信息（存根实现）
		_ = userID // 避免未使用变量警告
		// 存根实现：假设用户有基础会员权限
		currentLevel := "basic"

		// 检查会员等级（简单字符串比较）
		levelOrder := map[string]int{
			"free":       0,
			"basic":      1,
			"premium":    2,
			"enterprise": 3,
		}

		if levelOrder[currentLevel] < levelOrder[minLevel] {
			c.JSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Message: "权限不足，需要更高级别的会员",
				Data: gin.H{
					"required_level": minLevel,
					"current_level":  currentLevel,
				},
			})
			c.Abort()
			return
		}

		// 检查会员是否过期（存根实现：假设未过期）
		// 在实际实现中这里会检查用户的会员过期时间

		// 设置会员信息到上下文
		c.Set("membership_level", currentLevel)
		c.Next()
	}
}

// RequireFeature 需要特定功能权限
func (p *Permission) RequireFeature(featureName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "未登录",
			})
			c.Abort()
			return
		}

		// 检查功能权限（存根实现）
		_ = userID // 避免未使用变量警告
		// 存根实现：假设用户有所有功能权限
		hasPermission := true

		if !hasPermission {
			c.JSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Message: "没有访问该功能的权限",
				Data: gin.H{
					"feature": featureName,
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimit 使用量限制
func (p *Permission) RateLimit(feature string, maxCount int, duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "未登录",
			})
			c.Abort()
			return
		}

		// 检查使用量限制（存根实现）
		_ = userID // 避免未使用变量警告
		// 存根实现：假设用户未达到使用量限制
		canUse := true
		currentCount := 0

		if !canUse {
			c.JSON(http.StatusTooManyRequests, models.APIResponse{
				Success: false,
				Message: "使用量已达上限",
				Data: gin.H{
					"feature":       feature,
					"current_count": currentCount,
					"max_count":     maxCount,
					"duration":      duration.String(),
				},
			})
			c.Abort()
			return
		}

		// 记录使用量（存根实现）
		// 在实际实现中这里会记录用户的功能使用量

		c.Next()
	}
}

// AdminOnly 仅管理员可访问
func (p *Permission) AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "未登录",
			})
			c.Abort()
			return
		}

		u := user.(*models.User)
		// 检查用户角色（存根实现：从用户的角色字段获取）
		userRole := "student" // 默认角色，实际应该从数据库获取
		if len(u.Roles) > 0 {
			userRole = u.Roles[0].Name
		}

		if userRole != models.UserRoleAdmin {
			c.JSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Message: "需要管理员权限",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// extractToken 从请求中提取token
func (p *Permission) extractToken(c *gin.Context) string {
	// 从Authorization header提取
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}

	// 从查询参数提取
	token := c.Query("token")
	if token != "" {
		return token
	}

	// 从Cookie提取
	token, _ = c.Cookie("access_token")
	return token
}

// validateToken 验证JWT token
func (p *Permission) validateToken(tokenString string) (*models.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(p.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*models.JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrInvalidKey
}

// CustomPermission 自定义权限检查
func (p *Permission) CustomPermission(checkFunc func(c *gin.Context, userID string) bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "未登录",
			})
			c.Abort()
			return
		}

		if !checkFunc(c, userID.(string)) {
			c.JSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Message: "权限不足",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetMembershipLevelHandler 获取会员等级信息的中间件
func (p *Permission) GetMembershipLevelHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if exists {
			// 存根实现：设置默认会员信息
			_ = userID
			c.Set("membership_level", "basic")
			c.Set("membership_features", []string{"basic_access"})
		}
		c.Next()
	}
}
