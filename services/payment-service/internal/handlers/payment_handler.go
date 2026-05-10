package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"github.com/tvvshow/gokao/services/payment-service/internal/models"
	"github.com/tvvshow/gokao/services/payment-service/internal/service"
)

// PaymentHandler 支付处理器
type PaymentHandler struct {
	paymentService *service.PaymentService
	logger         *logrus.Logger
}

// NewPaymentHandler 创建支付处理器
func NewPaymentHandler(paymentService *service.PaymentService, logger *logrus.Logger) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
		logger:         logger,
	}
}

// CreatePaymentRequest 创建支付请求
type CreatePaymentRequest struct {
	UserID        string                 `json:"user_id" binding:"required"`
	Amount        float64                `json:"amount" binding:"required,gt=0"`
	Currency      string                 `json:"currency" binding:"required"`
	Description   string                 `json:"description" binding:"required"`
	PaymentMethod string                 `json:"payment_method" binding:"required,oneof=wechat_pay alipay alipay_qr"`
	Extra         map[string]interface{} `json:"extra"`
	ReturnURL     string                 `json:"return_url"`
}

// CreatePaymentResponse 创建支付响应
type CreatePaymentResponse struct {
	ID         string     `json:"id"`
	OrderNo    string     `json:"order_no"`
	Amount     float64    `json:"amount"`
	Currency   string     `json:"currency"`
	Subject    string     `json:"subject"`
	Channel    string     `json:"channel"`
	Status     string     `json:"status"`
	PaymentURL string     `json:"payment_url,omitempty"`
	QRCode     string     `json:"qr_code,omitempty"`
	ExpiredAt  *time.Time `json:"expired_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

// CreatePayment 创建支付订单
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid create payment request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	// 创建支付请求
	paymentReq := &models.CreatePaymentRequest{
		UserID:        req.UserID,
		Amount:        decimal.NewFromFloat(req.Amount),
		Currency:      req.Currency,
		Description:   req.Description,
		PaymentMethod: req.PaymentMethod,
		Extra:         req.Extra,
	}

	// 调用支付服务
	paymentResp, err := h.paymentService.CreatePayment(c.Request.Context(), paymentReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create payment")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "payment_creation_failed",
			"message": "Failed to create payment",
		})
		return
	}

	response := CreatePaymentResponse{
		ID:         paymentResp.ID.String(),
		OrderNo:    paymentResp.OrderNo,
		Amount:     paymentResp.Amount.InexactFloat64(),
		Currency:   paymentResp.Currency,
		Subject:    paymentResp.Subject,
		Channel:    paymentResp.Channel,
		Status:     paymentResp.Status,
		PaymentURL: paymentResp.PaymentURL,
		QRCode:     paymentResp.QRCode,
		ExpiredAt:  paymentResp.ExpiredAt,
		CreatedAt:  paymentResp.CreatedAt,
	}

	c.JSON(http.StatusCreated, response)
}

// QueryPaymentRequest 查询支付请求
type QueryPaymentRequest struct {
	PaymentID string `form:"payment_id" binding:"required"`
}

// QueryPayment 查询支付状态
func (h *PaymentHandler) QueryPayment(c *gin.Context) {
	paymentID := c.Param("payment_id")
	if paymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_payment_id",
			"message": "Payment ID is required",
		})
		return
	}

	payment, err := h.paymentService.QueryPayment(c.Request.Context(), paymentID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to query payment")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "payment_query_failed",
			"message": "Failed to query payment",
		})
		return
	}

	if payment == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "payment_not_found",
			"message": "Payment not found",
		})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// PaymentCallbackRequest 支付回调请求
type PaymentCallbackRequest struct {
	PaymentID string `json:"payment_id" binding:"required"`
	Status    string `json:"status" binding:"required,oneof=success failed cancelled"`
	Amount    string `json:"amount"`
	Channel   string `json:"channel"`
	TradeNo   string `json:"trade_no"`
	NotifyID  string `json:"notify_id"`
	Sign      string `json:"sign"`
	SignType  string `json:"sign_type"`
}

// PaymentCallback 支付回调处理
func (h *PaymentHandler) PaymentCallback(c *gin.Context) {
	channel := c.Param("channel")
	if channel == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_channel",
			"message": "Payment channel is required",
		})
		return
	}

	// 处理支付回调（HandleCallback 直接从 request.Body 读取原始字节做签名验证，
	// 不能先用 ShouldBindJSON 消费 body，否则后续 ReadAll 读到空）
	result, err := h.paymentService.HandleCallback(c.Request.Context(), channel, c.Request)
	if err != nil {
		h.logger.WithError(err).Error("Failed to handle payment callback")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "callback_processing_failed",
			"message": "Failed to process payment callback",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// RefundRequest 退款请求
type RefundRequest struct {
	PaymentID string  `json:"payment_id" binding:"required"`
	Amount    float64 `json:"amount" binding:"required,gt=0"`
	Reason    string  `json:"reason" binding:"required"`
	NotifyURL string  `json:"notify_url"`
}

// Refund 发起退款
func (h *PaymentHandler) Refund(c *gin.Context) {
	var req RefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid refund request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request parameters",
		})
		return
	}

	// 创建退款请求
	refundReq := &models.RefundRequest{
		OrderNo:  req.PaymentID,
		RefundID: fmt.Sprintf("RF%d", time.Now().UnixNano()),
		Amount:   decimal.NewFromFloat(req.Amount),
		Reason:   req.Reason,
	}

	// 调用退款服务
	refundResp, err := h.paymentService.RefundPayment(c.Request.Context(), refundReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create refund")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "refund_creation_failed",
			"message": "Failed to create refund",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"refund_id": refundResp.RefundID,
		"status":    refundResp.Status,
		"message":   "Refund request submitted",
	})
}

// QueryRefundRequest 查询退款请求
type QueryRefundRequest struct {
	RefundID string `form:"refund_id" binding:"required"`
}

// QueryRefund 查询退款状态
func (h *PaymentHandler) QueryRefund(c *gin.Context) {
	refundID := c.Param("refund_id")
	if refundID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_refund_id",
			"message": "Refund ID is required",
		})
		return
	}

	// 退款查询功能需要在服务层实现
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "not_implemented",
		"message": "Refund query not implemented",
	})
}

// ListPaymentsRequest 列出支付记录请求
type ListPaymentsRequest struct {
	UserID  string `form:"user_id"`
	Status  string `form:"status"`
	Channel string `form:"channel"`
	Page    int    `form:"page,default=1"`
	Limit   int    `form:"limit,default=20"`
}

// ListPayments 列出支付记录
func (h *PaymentHandler) ListPayments(c *gin.Context) {
	var req ListPaymentsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid list payments request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request parameters",
		})
		return
	}

	// 验证分页参数
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 20
	}

	// 构建查询条件：优先使用网关透传用户ID，其次兼容 query.user_id。
	userIDRaw := c.GetString("user_id")
	if userIDRaw == "" {
		userIDRaw = c.GetHeader("X-User-ID")
	}
	if userIDRaw == "" {
		userIDRaw = req.UserID
	}
	if userIDRaw == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User ID is required",
		})
		return
	}
	userID, err := uuid.Parse(userIDRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_user_id",
			"message": "Invalid user ID format",
		})
		return
	}

	filter := &models.PaymentFilter{
		UserID:   &userID,
		Status:   &req.Status,
		Channel:  &req.Channel,
		Page:     req.Page,
		PageSize: req.Limit,
	}

	payments, total, err := h.paymentService.ListPayments(c.Request.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list payments")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "payment_list_failed",
			"message": "Failed to list payments",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"payments": payments,
		"total":    total,
		"page":     req.Page,
		"limit":    req.Limit,
		"pages":    (total + int64(req.Limit) - 1) / int64(req.Limit),
	})
}

// ClosePayment 关闭支付订单
func (h *PaymentHandler) ClosePayment(c *gin.Context) {
	paymentID := c.Param("payment_id")
	if paymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_payment_id",
			"message": "Payment ID is required",
		})
		return
	}

	// 关闭支付功能需要在服务层实现
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "not_implemented",
		"message": "Close payment not implemented",
	})
}

// GetPaymentStatistics 获取支付统计
func (h *PaymentHandler) GetPaymentStatistics(c *gin.Context) {
	// 获取查询参数
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	channel := c.Query("channel")

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_start_date",
				"message": "Start date format should be YYYY-MM-DD",
			})
			return
		}
	} else {
		startDate = time.Now().AddDate(0, -1, 0) // 默认最近一个月
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_end_date",
				"message": "End date format should be YYYY-MM-DD",
			})
			return
		}
	} else {
		endDate = time.Now()
	}

	// 获取统计信息
	stats, err := h.paymentService.GetPaymentStatistics(c.Request.Context(), startDate, endDate, channel)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get payment statistics")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "statistics_failed",
			"message": "Failed to get payment statistics",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// WebhookTest 支付webhook测试
func (h *PaymentHandler) WebhookTest(c *gin.Context) {
	channel := c.Param("channel")
	if channel == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_channel",
			"message": "Payment channel is required",
		})
		return
	}

	// 模拟支付回调数据
	testData := map[string]interface{}{
		"payment_id": "test_" + strconv.FormatInt(time.Now().Unix(), 10),
		"status":     "success",
		"amount":     "100.00",
		"trade_no":   "TEST" + strconv.FormatInt(time.Now().Unix(), 10),
		"notify_id":  "test_notify_" + strconv.FormatInt(time.Now().Unix(), 10),
		"sign":       "test_signature",
		"sign_type":  "RSA2",
	}

	jsonData, err := json.Marshal(testData)
	if err != nil {
		h.logger.WithError(err).Error("Failed to marshal test data")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "test_data_failed",
			"message": "Failed to create test data",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"channel":      channel,
		"test_data":    string(jsonData),
		"timestamp":    time.Now().Unix(),
		"message":      "Webhook test data generated",
		"instructions": "Use this data to test your payment callback endpoint",
	})
}
