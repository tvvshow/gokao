package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/tvvshow/gokao/pkg/response"
	"github.com/tvvshow/gokao/services/payment-service/internal/service"
)

// MembershipMiddleware 会员权限中间件
type MembershipMiddleware struct {
	membershipService *service.MembershipService
}

// NewMembershipMiddleware 创建会员权限中间件
func NewMembershipMiddleware(membershipService *service.MembershipService) *MembershipMiddleware {
	return &MembershipMiddleware{
		membershipService: membershipService,
	}
}

// RequireVIP 要求VIP会员权限
func (m *MembershipMiddleware) RequireVIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			response.AbortWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "用户未认证", nil)
			return
		}

		status, err := m.membershipService.GetMembershipStatus(c.Request.Context(), userID)
		if err != nil {
			response.AbortWithError(c, http.StatusInternalServerError, "MEMBERSHIP_CHECK_FAILED", "获取会员状态失败", nil)
			return
		}

		if !status.IsVIP {
			response.AbortWithError(c, http.StatusForbidden, "VIP_REQUIRED", "需要VIP会员权限", gin.H{"upgrade": true})
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
			response.AbortWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "用户未认证", nil)
			return
		}

		hasPermission, err := m.membershipService.CheckMembershipPermission(c.Request.Context(), userID, feature)
		if err != nil {
			response.AbortWithError(c, http.StatusInternalServerError, "PERMISSION_CHECK_FAILED", "权限检查失败", nil)
			return
		}

		if !hasPermission {
			response.AbortWithError(c, http.StatusForbidden, "FEATURE_PERMISSION_DENIED", "功能权限不足", gin.H{
				"feature": feature,
				"upgrade": true,
			})
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
			response.AbortWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "用户未认证", nil)
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
			switch {
			case strings.Contains(err.Error(), "limit exceeded"):
				response.AbortWithError(c, http.StatusForbidden, "QUERY_LIMIT_EXCEEDED", "查询次数已用完", gin.H{"upgrade": true})
			case strings.Contains(err.Error(), "VIP membership required"):
				response.AbortWithError(c, http.StatusForbidden, "VIP_REQUIRED", "需要VIP会员权限", gin.H{"upgrade": true})
			default:
				response.AbortWithError(c, http.StatusInternalServerError, "CONSUME_QUERY_FAILED", "消费查询次数失败", nil)
			}
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
			response.AbortWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "用户未认证", nil)
			return
		}

		err := m.membershipService.ConsumeDownload(c.Request.Context(), userID)
		if err != nil {
			switch {
			case strings.Contains(err.Error(), "limit exceeded"):
				response.AbortWithError(c, http.StatusForbidden, "DOWNLOAD_LIMIT_EXCEEDED", "下载次数已用完", gin.H{"upgrade": true})
			case strings.Contains(err.Error(), "VIP membership required"):
				response.AbortWithError(c, http.StatusForbidden, "VIP_REQUIRED", "需要VIP会员权限", gin.H{"upgrade": true})
			default:
				response.AbortWithError(c, http.StatusInternalServerError, "CONSUME_DOWNLOAD_FAILED", "消费下载次数失败", nil)
			}
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
