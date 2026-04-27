package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/oktetopython/gaokao/services/payment-service/internal/models"
)

// MembershipService 定义会员处理器依赖的服务接口。
type MembershipService interface {
	GetPlans(ctx context.Context) ([]*models.MembershipPlan, error)
	Subscribe(ctx context.Context, userID, orderNo string) error
	GetMembershipStatus(ctx context.Context, userID string) (*models.MembershipStatusResponse, error)
	RenewMembership(ctx context.Context, userID, planCode string) (string, error)
	CancelMembership(ctx context.Context, userID string) error
	GetMemberBenefits(ctx context.Context, userID string) (map[string]interface{}, error)
	UpdateAutoRenew(ctx context.Context, userID string, autoRenew bool) error
}

// MembershipHandler 会员接口处理器。
type MembershipHandler struct {
	service MembershipService
}

// NewMembershipHandler 创建会员处理器。
func NewMembershipHandler(service MembershipService) *MembershipHandler {
	return &MembershipHandler{service: service}
}

// RegisterRoutes 注册会员路由。
func (h *MembershipHandler) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/plans", h.GetPlans)
	group.POST("/subscribe", h.Subscribe)
	group.GET("/status", h.GetMembershipStatus)
	group.POST("/renew", h.RenewMembership)
	group.POST("/cancel", h.CancelMembership)
	group.GET("/benefits", h.GetMemberBenefits)
	group.PUT("/auto-renew", h.UpdateAutoRenew)
}

type renewMembershipRequest struct {
	PlanCode string `json:"plan_code" binding:"required"`
}

type updateAutoRenewRequest struct {
	AutoRenew bool `json:"auto_renew"`
}

type subscribeRequest struct {
	OrderNo string `json:"order_no"`
}

func (h *MembershipHandler) GetPlans(c *gin.Context) {
	plans, err := h.service.GetPlans(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "membership_plans_failed",
			"message": "Failed to get membership plans",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    plans,
	})
}

func (h *MembershipHandler) Subscribe(c *gin.Context) {
	userID, ok := membershipUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User ID is required",
		})
		return
	}

	var req subscribeRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_request",
				"message": "Invalid request parameters",
				"details": err.Error(),
			})
			return
		}
	}
	if req.OrderNo == "" {
		req.OrderNo = c.Query("order_no")
	}
	if req.OrderNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "missing_order_no",
			"message": "order_no is required",
		})
		return
	}

	if err := h.service.Subscribe(c.Request.Context(), userID, req.OrderNo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "membership_subscribe_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "会员订阅成功",
	})
}

func (h *MembershipHandler) GetMembershipStatus(c *gin.Context) {
	userID, ok := membershipUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User ID is required",
		})
		return
	}

	status, err := h.service.GetMembershipStatus(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "membership_status_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}

func (h *MembershipHandler) RenewMembership(c *gin.Context) {
	userID, ok := membershipUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User ID is required",
		})
		return
	}

	var req renewMembershipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	orderNo, err := h.service.RenewMembership(c.Request.Context(), userID, req.PlanCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "membership_renew_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"order_no": orderNo,
		},
	})
}

func (h *MembershipHandler) CancelMembership(c *gin.Context) {
	userID, ok := membershipUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User ID is required",
		})
		return
	}

	if err := h.service.CancelMembership(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "membership_cancel_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "会员已取消自动续费",
	})
}

func (h *MembershipHandler) GetMemberBenefits(c *gin.Context) {
	userID, ok := membershipUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User ID is required",
		})
		return
	}

	benefits, err := h.service.GetMemberBenefits(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "membership_benefits_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    benefits,
	})
}

func (h *MembershipHandler) UpdateAutoRenew(c *gin.Context) {
	userID, ok := membershipUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User ID is required",
		})
		return
	}

	var req updateAutoRenewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	if err := h.service.UpdateAutoRenew(c.Request.Context(), userID, req.AutoRenew); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "membership_auto_renew_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "自动续费设置已更新",
		"data": gin.H{
			"auto_renew": req.AutoRenew,
		},
	})
}

func membershipUserID(c *gin.Context) (string, bool) {
	if userID := c.GetString("user_id"); userID != "" {
		return userID, true
	}
	if userID := c.GetHeader("X-User-ID"); userID != "" {
		return userID, true
	}
	if userID := c.Query("user_id"); userID != "" {
		return userID, true
	}
	return "", false
}
