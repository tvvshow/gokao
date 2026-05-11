package handlers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"github.com/tvvshow/gokao/pkg/response"
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
		response.BadRequest(c, "invalid_request", "Invalid request parameters", err.Error())
		return
	}

	paymentReq := &models.CreatePaymentRequest{
		UserID:        req.UserID,
		Amount:        decimal.NewFromFloat(req.Amount),
		Currency:      req.Currency,
		Description:   req.Description,
		PaymentMethod: req.PaymentMethod,
		Extra:         req.Extra,
	}

	paymentResp, err := h.paymentService.CreatePayment(c.Request.Context(), paymentReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create payment")
		response.InternalError(c, "payment_creation_failed", "Failed to create payment", nil)
		return
	}

	resp := CreatePaymentResponse{
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

	response.Created(c, resp)
}

// QueryPaymentRequest 查询支付请求
type QueryPaymentRequest struct {
	PaymentID string `form:"payment_id" binding:"required"`
}

// QueryPayment 查询支付状态
func (h *PaymentHandler) QueryPayment(c *gin.Context) {
	paymentID := c.Param("payment_id")
	if paymentID == "" {
		response.BadRequest(c, "invalid_payment_id", "Payment ID is required", nil)
		return
	}

	payment, err := h.paymentService.QueryPayment(c.Request.Context(), paymentID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to query payment")
		response.InternalError(c, "payment_query_failed", "Failed to query payment", nil)
		return
	}

	if payment == nil {
		response.NotFound(c, "payment_not_found", "Payment not found")
		return
	}

	response.OK(c, payment)
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
		response.BadRequest(c, "invalid_channel", "Payment channel is required", nil)
		return
	}

	// HandleCallback 直接从 request.Body 读取原始字节做签名验证，
	// 不能先用 ShouldBindJSON 消费 body，否则后续 ReadAll 读到空
	result, err := h.paymentService.HandleCallback(c.Request.Context(), channel, c.Request)
	if err != nil {
		h.logger.WithError(err).Error("Failed to handle payment callback")
		response.InternalError(c, "callback_processing_failed", "Failed to process payment callback", nil)
		return
	}

	response.OK(c, result)
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
		response.BadRequest(c, "invalid_request", "Invalid request parameters", nil)
		return
	}

	refundReq := &models.RefundRequest{
		OrderNo:  req.PaymentID,
		RefundID: fmt.Sprintf("RF%d", time.Now().UnixNano()),
		Amount:   decimal.NewFromFloat(req.Amount),
		Reason:   req.Reason,
	}

	refundResp, err := h.paymentService.RefundPayment(c.Request.Context(), refundReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create refund")
		response.InternalError(c, "refund_creation_failed", "Failed to create refund", nil)
		return
	}

	response.CreatedWithMessage(c, gin.H{
		"refund_id": refundResp.RefundID,
		"status":    refundResp.Status,
	}, "Refund request submitted")
}

// QueryRefundRequest 查询退款请求
type QueryRefundRequest struct {
	RefundID string `form:"refund_id" binding:"required"`
}

// QueryRefund 查询退款状态
func (h *PaymentHandler) QueryRefund(c *gin.Context) {
	refundID := c.Param("refund_id")
	if refundID == "" {
		response.BadRequest(c, "invalid_refund_id", "Refund ID is required", nil)
		return
	}

	// 退款查询功能需要在服务层实现
	response.NotImplemented(c, "not_implemented", "Refund query not implemented")
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
		response.BadRequest(c, "invalid_request", "Invalid request parameters", nil)
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 20
	}

	// 优先使用网关透传用户ID，其次兼容 query.user_id
	userIDRaw := c.GetString("user_id")
	if userIDRaw == "" {
		userIDRaw = c.GetHeader("X-User-ID")
	}
	if userIDRaw == "" {
		userIDRaw = req.UserID
	}
	if userIDRaw == "" {
		response.Unauthorized(c, "unauthorized", "User ID is required")
		return
	}
	userID, err := uuid.Parse(userIDRaw)
	if err != nil {
		response.BadRequest(c, "invalid_user_id", "Invalid user ID format", nil)
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
		response.InternalError(c, "payment_list_failed", "Failed to list payments", nil)
		return
	}

	response.OK(c, gin.H{
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
		response.BadRequest(c, "invalid_payment_id", "Payment ID is required", nil)
		return
	}

	// 关闭支付功能需要在服务层实现
	response.NotImplemented(c, "not_implemented", "Close payment not implemented")
}

// GetPaymentStatistics 获取支付统计
func (h *PaymentHandler) GetPaymentStatistics(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	channel := c.Query("channel")

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			response.BadRequest(c, "invalid_start_date", "Start date format should be YYYY-MM-DD", nil)
			return
		}
	} else {
		startDate = time.Now().AddDate(0, -1, 0)
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			response.BadRequest(c, "invalid_end_date", "End date format should be YYYY-MM-DD", nil)
			return
		}
	} else {
		endDate = time.Now()
	}

	stats, err := h.paymentService.GetPaymentStatistics(c.Request.Context(), startDate, endDate, channel)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get payment statistics")
		response.InternalError(c, "statistics_failed", "Failed to get payment statistics", nil)
		return
	}

	response.OK(c, stats)
}

// WebhookTest 支付webhook测试
func (h *PaymentHandler) WebhookTest(c *gin.Context) {
	channel := c.Param("channel")
	if channel == "" {
		response.BadRequest(c, "invalid_channel", "Payment channel is required", nil)
		return
	}

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
		response.InternalError(c, "test_data_failed", "Failed to create test data", nil)
		return
	}

	response.OKWithMessage(c, gin.H{
		"channel":      channel,
		"test_data":    string(jsonData),
		"timestamp":    time.Now().Unix(),
		"instructions": "Use this data to test your payment callback endpoint",
	}, "Webhook test data generated")
}
