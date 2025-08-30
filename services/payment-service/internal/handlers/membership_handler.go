package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gaokaohub/payment-service/internal/middleware"
	"github.com/gaokaohub/payment-service/internal/services"
)

// MembershipHandler 会员处理器
type MembershipHandler struct {
	membershipService *services.MembershipService
}

// NewMembershipHandler 创建会员处理器
func NewMembershipHandler(membershipService *services.MembershipService) *MembershipHandler {
	return &MembershipHandler{
		membershipService: membershipService,
	}
}

// GetPlans 获取会员套餐
// @Summary 获取会员套餐列表
// @Description 获取所有可用的会员套餐
// @Tags membership
// @Accept json
// @Produce json
// @Success 200 {array} models.MembershipPlan
// @Failure 500 {object} ErrorResponse
// @Router /membership/plans [get]
func (h *MembershipHandler) GetPlans(c *gin.Context) {
	plans, err := h.membershipService.GetPlans(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
			Code:  "GET_PLANS_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, plans)
}

// SubscribeRequest 订阅请求
type SubscribeRequest struct {
	OrderNo string `json:"order_no" binding:"required"`
}

// Subscribe 订阅会员
// @Summary 订阅会员
// @Description 通过已支付订单激活会员
// @Tags membership
// @Accept json
// @Produce json
// @Param request body SubscribeRequest true "订阅请求"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /membership/subscribe [post]
func (h *MembershipHandler) Subscribe(c *gin.Context) {
	var req SubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	// 获取用户ID
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "user not authenticated",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	// 订阅会员
	err := h.membershipService.Subscribe(c.Request.Context(), userID, req.OrderNo)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
			Code:  "SUBSCRIBE_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "membership subscribed successfully",
	})
}

// GetMembershipStatus 获取会员状态
// @Summary 获取会员状态
// @Description 获取当前用户的会员状态和权益信息
// @Tags membership
// @Accept json
// @Produce json
// @Success 200 {object} models.MembershipStatusResponse
// @Failure 500 {object} ErrorResponse
// @Router /membership/status [get]
func (h *MembershipHandler) GetMembershipStatus(c *gin.Context) {
	// 获取用户ID
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "user not authenticated",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	// 获取会员状态
	status, err := h.membershipService.GetMembershipStatus(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
			Code:  "GET_STATUS_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// RenewRequest 续费请求
type RenewRequest struct {
	PlanCode string `json:"plan_code" binding:"required"`
}

// RenewMembership 续费会员
// @Summary 续费会员
// @Description 续费会员服务
// @Tags membership
// @Accept json
// @Produce json
// @Param request body RenewRequest true "续费请求"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /membership/renew [post]
func (h *MembershipHandler) RenewMembership(c *gin.Context) {
	var req RenewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	// 获取用户ID
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "user not authenticated",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	// 续费会员
	orderNo, err := h.membershipService.RenewMembership(c.Request.Context(), userID, req.PlanCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
			Code:  "RENEW_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, map[string]string{
		"order_no": orderNo,
		"message":  "membership renewal order created",
	})
}

// CancelMembership 取消会员
// @Summary 取消会员
// @Description 取消当前会员服务
// @Tags membership
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /membership/cancel [post]
func (h *MembershipHandler) CancelMembership(c *gin.Context) {
	// 获取用户ID
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "user not authenticated",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	// 取消会员
	err := h.membershipService.CancelMembership(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
			Code:  "CANCEL_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "membership canceled successfully",
	})
}

// GetMemberBenefits 获取会员权益
// @Summary 获取会员权益
// @Description 获取当前用户的会员权益详情
// @Tags membership
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} ErrorResponse
// @Router /membership/benefits [get]
func (h *MembershipHandler) GetMemberBenefits(c *gin.Context) {
	// 获取用户ID
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "user not authenticated",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	// 获取会员权益
	benefits, err := h.membershipService.GetMemberBenefits(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
			Code:  "GET_BENEFITS_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, benefits)
}

// ConsumeQueryRequest 消费查询请求
type ConsumeQueryRequest struct {
	Type string `json:"type" binding:"required,oneof=query download"`
}

// ConsumeQuota 消费配额
// @Summary 消费配额
// @Description 消费用户的查询或下载配额
// @Tags membership
// @Accept json
// @Produce json
// @Param request body ConsumeQueryRequest true "消费请求"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /membership/consume [post]
func (h *MembershipHandler) ConsumeQuota(c *gin.Context) {
	var req ConsumeQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	// 获取用户ID
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "user not authenticated",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	var err error
	switch req.Type {
	case "query":
		err = h.membershipService.ConsumeQuery(c.Request.Context(), userID)
	case "download":
		err = h.membershipService.ConsumeDownload(c.Request.Context(), userID)
	default:
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "invalid consumption type",
			Code:  "INVALID_TYPE",
		})
		return
	}

	if err != nil {
		if err.Error() == "VIP membership required" {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error: err.Error(),
				Code:  "VIP_REQUIRED",
			})
			return
		}

		if err.Error() == "query limit exceeded" || err.Error() == "download limit exceeded" {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error: err.Error(),
				Code:  "QUOTA_EXCEEDED",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
			Code:  "CONSUME_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: req.Type + " quota consumed successfully",
	})
}