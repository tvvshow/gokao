package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/tvvshow/gokao/pkg/auth"
	"github.com/tvvshow/gokao/services/user-service/internal/models"
	"github.com/tvvshow/gokao/services/user-service/internal/services"
)

// Permission 权限中间件
type Permission struct {
	userService    *services.UserService
	roleService    *services.RoleService
	authMiddleware *auth.AuthMiddleware
	redis          *redis.Client
}

// NewPermission 创建权限中间件
func NewPermission(userService *services.UserService, roleService *services.RoleService, jwtSecret string, redisClient *redis.Client) *Permission {
	return &Permission{
		userService:    userService,
		roleService:    roleService,
		authMiddleware: auth.NewAuthMiddleware(jwtSecret),
		redis:          redisClient,
	}
}

// RequireAuth 需要登录认证 - 使用共享认证包
func (p *Permission) RequireAuth() gin.HandlerFunc {
	// 首先使用共享认证包进行基础JWT验证
	baseAuth := p.authMiddleware.RequireAuth()

	return func(c *gin.Context) {
		// 执行基础认证
		baseAuth(c)

		// 如果基础认证失败，直接返回
		if c.IsAborted() {
			return
		}

		// 获取用户ID（由共享认证包设置）
		userIDStr, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "用户ID未找到",
			})
			c.Abort()
			return
		}

		// 解析用户ID
		userID, err := uuid.Parse(userIDStr.(string))
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "无效的用户ID",
			})
			c.Abort()
			return
		}

		// 检查用户是否存在并获取完整用户信息（Redis 缓存优先）
		user, err := p.getCachedUser(c.Request.Context(), userID)
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

const authCacheTTL = 5 * time.Minute

func (p *Permission) cacheKey(userID uuid.UUID) string {
	return fmt.Sprintf("auth:user:%s", userID.String())
}

// getCachedUser 从 Redis 缓存获取用户，缓存未命中则查库并回填。
func (p *Permission) getCachedUser(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	key := p.cacheKey(userID)

	if p.redis != nil {
		if cached, err := p.redis.Get(ctx, key).Bytes(); err == nil {
			var user models.User
			if json.Unmarshal(cached, &user) == nil {
				return &user, nil
			}
		}
	}

	user, err := p.userService.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	if p.redis != nil {
		if data, err := json.Marshal(user); err == nil {
			p.redis.Set(ctx, key, data, authCacheTTL)
		}
	}

	return user, nil
}

// AdminOnly 仅管理员可访问
func (p *Permission) AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID
		val, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "未登录",
			})
			c.Abort()
			return
		}

		var userID uuid.UUID
		switch v := val.(type) {
		case uuid.UUID:
			userID = v
		case string:
			parsed, err := uuid.Parse(v)
			if err != nil {
				c.JSON(http.StatusBadRequest, models.APIResponse{
					Success: false,
					Message: "无效的用户ID",
				})
				c.Abort()
				return
			}
			userID = parsed
		default:
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Message: "无效的用户ID类型",
			})
			c.Abort()
			return
		}

		// 使用缓存检查用户角色
		roles, err := p.roleService.GetUserRoles(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "获取用户角色失败",
			})
			c.Abort()
			return
		}

		// 检查是否有管理员角色
		isAdmin := false
		for _, role := range roles {
			if role.Name == models.UserRoleAdmin {
				isAdmin = true
				break
			}
		}

		if !isAdmin {
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

// validateToken 方法已移除 - 现在使用共享认证包进行JWT验证

// RequirePermission 检查特定权限 - 使用缓存优化
func (p *Permission) RequirePermission(permissionName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID
		val, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "未登录",
			})
			c.Abort()
			return
		}

		var userID uuid.UUID
		switch v := val.(type) {
		case uuid.UUID:
			userID = v
		case string:
			parsed, err := uuid.Parse(v)
			if err != nil {
				c.JSON(http.StatusBadRequest, models.APIResponse{
					Success: false,
					Message: "无效的用户ID",
				})
				c.Abort()
				return
			}
			userID = parsed
		default:
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Message: "无效的用户ID类型",
			})
			c.Abort()
			return
		}

		// 使用缓存检查用户权限
		hasPermission, err := p.roleService.HasPermission(userID, permissionName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "权限检查失败",
			})
			c.Abort()
			return
		}

		if !hasPermission {
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
