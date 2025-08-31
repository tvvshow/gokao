package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gaokaohub/payment-service/internal/services"
)

// MembershipMiddleware 会员权限中间件
type MembershipMiddleware struct {
	membershipService *services.MembershipService
}

// NewMembershipMiddleware 创建会员权限中间件
func NewMembershipMiddleware(membershipService *services.MembershipService) *MembershipMiddleware {
	return &MembershipMiddleware{
		membershipService: membershipService,
	}
}

// RequireVIP 要求VIP会员权限
func (m *MembershipMiddleware) RequireVIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "用户未认证",
				"code":  "UNAUTHORIZED",
			})
			c.Abort()
			return
		}

		status, err := m.membershipService.GetMembershipStatus(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "获取会员状态失败",
				"code":  "MEMBERSHIP_CHECK_FAILED",
			})
			c.Abort()
			return
		}

		if !status.IsVIP {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "需要VIP会员权限",
				"code":    "VIP_REQUIRED",
				"upgrade": true,
			})
			c.Abort()
			return
		}

		// 将会员状态存储到上下文中，供后续处理器使用
		c.Set("membership_status", status)
		c.Next()
	}
}

// RequireFeature 要求特定功能权限
func (m *MembershipMiddleware) RequireFeature(feature string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "用户未认证",
				"code":  "UNAUTHORIZED",
			})
			c.Abort()
			return
		}

		hasPermission, err := m.membershipService.CheckMembershipPermission(c.Request.Context(), userID, feature)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "权限检查失败",
				"code":  "PERMISSION_CHECK_FAILED",
			})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "功能权限不足",
				"code":    "FEATURE_PERMISSION_DENIED",
				"feature": feature,
				"upgrade": true,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ConsumeQuery 消费查询次数中间件
func (m *MembershipMiddleware) ConsumeQuery() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "用户未认证",
				"code":  "UNAUTHORIZED",
			})
			c.Abort()
			return
		}

		// 检查是否需要消费查询次数（某些接口可能不需要）
		skipConsume := c.GetBool("skip_query_consume")
		if skipConsume {
			c.Next()
			return
		}

		err := m.membershipService.ConsumeQuery(c.Request.Context(), userID)
		if err != nil {
			if strings.Contains(err.Error(), "limit exceeded") {
				c.JSON(http.StatusForbidden, gin.H{
					"error":   "查询次数已用完",
					"code":    "QUERY_LIMIT_EXCEEDED",
					"upgrade": true,
				})
			} else if strings.Contains(err.Error(), "VIP membership required") {
				c.JSON(http.StatusForbidden, gin.H{
					"error":   "需要VIP会员权限",
					"code":    "VIP_REQUIRED",
					"upgrade": true,
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "消费查询次数失败",
					"code":  "CONSUME_QUERY_FAILED",
				})
			}
			c.Abort()
			return
		}

		c.Next()
	}
}

// ConsumeDownload 消费下载次数中间件
func (m *MembershipMiddleware) ConsumeDownload() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "用户未认证",
				"code":  "UNAUTHORIZED",
			})
			c.Abort()
			return
		}

		err := m.membershipService.ConsumeDownload(c.Request.Context(), userID)
		if err != nil {
			if strings.Contains(err.Error(), "limit exceeded") {
				c.JSON(http.StatusForbidden, gin.H{
					"error":   "下载次数已用完",
					"code":    "DOWNLOAD_LIMIT_EXCEEDED",
					"upgrade": true,
				})
			} else if strings.Contains(err.Error(), "VIP membership required") {
				c.JSON(http.StatusForbidden, gin.H{
					"error":   "需要VIP会员权限",
					"code":    "VIP_REQUIRED",
					"upgrade": true,
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "消费下载次数失败",
					"code":  "CONSUME_DOWNLOAD_FAILED",
				})
			}
			c.Abort()
			return
		}

		c.Next()
	}
}

// CheckMembershipExpiry 检查会员到期中间件
func (m *MembershipMiddleware) CheckMembershipExpiry() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			c.Next()
			return
		}

		status, err := m.membershipService.GetMembershipStatus(c.Request.Context(), userID)
		if err != nil {
			// 不阻断请求，只是无法获取会员状态
			c.Next()
			return
		}

		// 如果会员即将到期（7天内），添加提醒信息
		if status.IsVIP && status.RemainingDays <= 7 && status.RemainingDays > 0 {
			c.Header("X-Membership-Warning", "会员即将到期，请及时续费")
			c.Header("X-Remaining-Days", string(rune(status.RemainingDays)))
		}

		// 如果会员已过期，添加过期信息
		if status.IsVIP && status.RemainingDays <= 0 {
			c.Header("X-Membership-Expired", "会员已过期")
		}

		c.Set("membership_status", status)
		c.Next()
	}
}

// SkipQueryConsume 跳过查询次数消费的标记中间件
func SkipQueryConsume() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("skip_query_consume", true)
		c.Next()
	}
}
