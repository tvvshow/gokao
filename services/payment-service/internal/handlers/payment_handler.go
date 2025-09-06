package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
<<<<<<< HEAD
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"github.com/gaokaohub/gaokao/services/payment-service/internal/models"
	"github.com/gaokaohub/gaokao/services/payment-service/internal/service"
=======
	"github.com/sirupsen/logrus"

	"github.com/gaokao/payment-service/internal/models"
	"github.com/gaokao/payment-service/internal/service"
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
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
<<<<<<< HEAD
	UserID        string                 `json:"user_id" binding:"required"`
	Amount        float64                `json:"amount" binding:"required,gt=0"`
	Currency      string                 `json:"currency" binding:"required"`
	Description   string                 `json:"description" binding:"required"`
	PaymentMethod string                 `json:"payment_method" binding:"required,oneof=wechat_pay alipay alipay_qr"`
	Extra         map[string]interface{} `json:"extra"`
	ReturnURL     string                 `json:"return_url"`
=======
	UserID      int64   `json:"user_id" binding:"required"`
	OrderID     string  `json:"order_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	ProductType string  `json:"product_type" binding:"required"`
	ProductID   string  `json:"product_id" binding:"required"`
	Channel     string  `json:"channel" binding:"required,oneof=alipay wechat unionpay"`
	Subject     string  `json:"subject" binding:"required"`
	Body        string  `json:"body"`
	ReturnURL   string  `json:"return_url"`
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
}

// CreatePaymentResponse 创建支付响应
type CreatePaymentResponse struct {
<<<<<<< HEAD
	ID         string    `json:"id"`
	OrderNo    string    `json:"order_no"`
	Amount     float64   `json:"amount"`
	Currency   string    `json:"currency"`
	Subject    string    `json:"subject"`
	Channel    string    `json:"channel"`
	Status     string    `json:"status"`
	PaymentURL string    `json:"payment_url,omitempty"`
	QRCode     string    `json:"qr_code,omitempty"`
	ExpiredAt  *time.Time `json:"expired_at"`
	CreatedAt  time.Time `json:"created_at"`
=======
	PaymentID   string `json:"payment_id"`
	PaymentURL  string `json:"payment_url"`
	QRCode      string `json:"qr_code,omitempty"`
	PaymentData string `json:"payment_data,omitempty"`
	ExpireTime  int64  `json:"expire_time"`
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
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

<<<<<<< HEAD
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
=======
	// 创建支付记录
	payment := &models.Payment{
		UserID:      req.UserID,
		OrderID:     req.OrderID,
		Amount:      req.Amount,
		ProductType: req.ProductType,
		ProductID:   req.ProductID,
		Channel:     req.Channel,
		Subject:     req.Subject,
		Body:        req.Body,
		Status:      models.PaymentStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 调用支付服务
	paymentURL, qrCode, paymentData, err := h.paymentService.CreatePayment(c.Request.Context(), payment, req.ReturnURL)
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
	if err != nil {
		h.logger.WithError(err).Error("Failed to create payment")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "payment_creation_failed",
			"message": "Failed to create payment",
		})
		return
	}

	response := CreatePaymentResponse{
<<<<<<< HEAD
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
=======
		PaymentID:   payment.PaymentID,
		PaymentURL:  paymentURL,
		QRCode:      qrCode,
		PaymentData: paymentData,
		ExpireTime:  time.Now().Add(30 * time.Minute).Unix(),
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
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

<<<<<<< HEAD
	payment, err := h.paymentService.QueryPayment(c.Request.Context(), paymentID)
=======
	payment, err := h.paymentService.GetPaymentByID(c.Request.Context(), paymentID)
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
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

	var req PaymentCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid payment callback request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request parameters",
		})
		return
	}

<<<<<<< HEAD
	// 处理支付回调
	result, err := h.paymentService.HandleCallback(c.Request.Context(), channel, c.Request)
	if err != nil {
		h.logger.WithError(err).Error("Failed to handle payment callback")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "callback_processing_failed",
			"message": "Failed to process payment callback",
=======
	// 验证签名和回调数据
	isValid, err := h.paymentService.ValidateCallback(c.Request.Context(), channel, req.PaymentID, req.TradeNo, req.Amount, req.Sign, req.SignType)
	if err != nil || !isValid {
		h.logger.WithError(err).Warn("Invalid payment callback signature")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_signature",
			"message": "Invalid callback signature",
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
		})
		return
	}

<<<<<<< HEAD
	c.JSON(http.StatusOK, result)
=======
	// 处理支付结果
	err = h.paymentService.HandlePaymentResult(c.Request.Context(), req.PaymentID, req.Status, req.TradeNo, req.Amount)
	if err != nil {
		h.logger.WithError(err).Error("Failed to handle payment result")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "payment_processing_failed",
			"message": "Failed to process payment result",
		})
		return
	}
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Payment processed successfully",
	})
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

<<<<<<< HEAD
	// 创建退款请求
	refundReq := &models.RefundRequest{
		OrderNo:  req.PaymentID,
		RefundID: fmt.Sprintf("RF%d", time.Now().UnixNano()),
		Amount:   decimal.NewFromFloat(req.Amount),
		Reason:   req.Reason,
	}

	// 调用退款服务
	refundResp, err := h.paymentService.RefundPayment(c.Request.Context(), refundReq)
=======
	// 调用退款服务
	refundID, err := h.paymentService.CreateRefund(c.Request.Context(), req.PaymentID, req.Amount, req.Reason, req.NotifyURL)
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
	if err != nil {
		h.logger.WithError(err).Error("Failed to create refund")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "refund_creation_failed",
			"message": "Failed to create refund",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
<<<<<<< HEAD
		"refund_id": refundResp.RefundID,
		"status":    refundResp.Status,
=======
		"refund_id": refundID,
		"status":    "processing",
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
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

<<<<<<< HEAD
	// 退款查询功能需要在服务层实现
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "not_implemented",
		"message": "Refund query not implemented",
	})
=======
	refund, err := h.paymentService.GetRefundByID(c.Request.Context(), refundID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to query refund")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "refund_query_failed",
			"message": "Failed to query refund",
		})
		return
	}

	if refund == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "refund_not_found",
			"message": "Refund not found",
		})
		return
	}

	c.JSON(http.StatusOK, refund)
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
}

// ListPaymentsRequest 列出支付记录请求
type ListPaymentsRequest struct {
	UserID  int64  `form:"user_id"`
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

	// 构建查询条件
<<<<<<< HEAD
	userID := uuid.MustParse(fmt.Sprintf("%d", req.UserID))
	filter := &models.PaymentFilter{
		UserID:   &userID,
		Status:   &req.Status,
		Channel:  &req.Channel,
		Page:     req.Page,
		PageSize: req.Limit,
=======
	filter := &models.PaymentFilter{
		UserID:  req.UserID,
		Status:  req.Status,
		Channel: req.Channel,
		Page:    req.Page,
		Limit:   req.Limit,
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
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
<<<<<<< HEAD
		"pages":    (total + int64(req.Limit) - 1) / int64(req.Limit),
=======
		"pages":    (total + req.Limit - 1) / req.Limit,
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
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

<<<<<<< HEAD
	// 关闭支付功能需要在服务层实现
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "not_implemented",
		"message": "Close payment not implemented",
=======
	err := h.paymentService.ClosePayment(c.Request.Context(), paymentID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to close payment")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "payment_close_failed",
			"message": "Failed to close payment",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Payment closed successfully",
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
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
		"channel":    channel,
		"test_data":  string(jsonData),
		"timestamp":  time.Now().Unix(),
		"message":    "Webhook test data generated",
		"instructions": "Use this data to test your payment callback endpoint",
	})
}