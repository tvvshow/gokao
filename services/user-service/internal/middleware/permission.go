package middleware

import (
	"net/http"
	"strings"
	"time"

	"user-service/internal/models"
	"user-service/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/gaokaohub/pkg/auth"
)

// Permission 权限中间件
type Permission struct {
	userService    *services.UserService
	roleService    *services.RoleService
	authMiddleware *auth.AuthMiddleware
}

// NewPermission 创建权限中间件
func NewPermission(userService *services.UserService, roleService *services.RoleService, jwtSecret string) *Permission {
	return &Permission{
		userService:    userService,
		roleService:    roleService,
		authMiddleware: auth.NewAuthMiddleware(jwtSecret),
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

		// 检查用户是否存在并获取完整用户信息
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

// validateToken 方法已移除 - 现在使用共享认证包进行JWT验证

// CustomPermission 自定义权限检查
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

func (p *Permission) CustomPermission(checkFunc func(c *gin.Context, userID string) bool) gin.HandlerFunc {
    return func(c *gin.Context) {
        val, exists := c.Get("user_id")
        if !exists {
            c.JSON(http.StatusUnauthorized, models.APIResponse{
                Success: false,
                Message: "未登录",
            })
            c.Abort()
            return
        }

        var userIDStr string
        switch v := val.(type) {
        case uuid.UUID:
            userIDStr = v.String()
        case string:
            userIDStr = v
        default:
            c.JSON(http.StatusBadRequest, models.APIResponse{
                Success: false,
                Message: "无效的用户ID类型",
            })
            c.Abort()
            return
        }

        if !checkFunc(c, userIDStr) {
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
