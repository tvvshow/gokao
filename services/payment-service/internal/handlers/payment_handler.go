package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/gaokaohub/payment-service/internal/adapters"
	"github.com/gaokaohub/payment-service/internal/middleware"
	"github.com/gaokaohub/payment-service/internal/services"
)

// PaymentHandler 支付处理器
type PaymentHandler struct {
	paymentService *services.PaymentService
}

// NewPaymentHandler 创建支付处理器
func NewPaymentHandler(paymentService *services.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

// CreatePaymentRequest 创建支付请求
type CreatePaymentRequest struct {
	OrderNo     string                 `json:"order_no" binding:"required"`
	Amount      string                 `json:"amount" binding:"required"`
	Subject     string                 `json:"subject" binding:"required"`
	Description string                 `json:"description"`
	Channel     string                 `json:"channel" binding:"required,oneof=alipay wechat unionpay"`
	ReturnURL   string                 `json:"return_url"`
	ClientIP    string                 `json:"client_ip"`
	ExpireHours int                    `json:"expire_hours"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// CreatePayment 创建支付
// @Summary 创建支付
// @Description 创建支付订单并返回支付链接
// @Tags payment
// @Accept json
// @Produce json
// @Param request body CreatePaymentRequest true "支付请求"
// @Success 200 {object} adapters.PaymentResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /payments/create [post]
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	// 解析金额
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "invalid amount format",
			Code:  "INVALID_AMOUNT",
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

	// 设置默认值
	if req.ExpireHours == 0 {
		req.ExpireHours = 1 // 默认1小时过期
	}
	if req.ClientIP == "" {
		req.ClientIP = c.ClientIP()
	}
	if req.Metadata == nil {
		req.Metadata = make(map[string]interface{})
	}
	req.Metadata["channel"] = req.Channel

	// 构造支付请求
	paymentReq := &adapters.PaymentRequest{
		OrderNo:     req.OrderNo,
		Amount:      amount,
		Subject:     req.Subject,
		Description: req.Description,
		NotifyURL:   fmt.Sprintf("/api/v1/payments/callback/%s", req.Channel),
		ReturnURL:   req.ReturnURL,
		UserID:      userID,
		ClientIP:    req.ClientIP,
		ExpireTime:  time.Duration(req.ExpireHours) * time.Hour,
		Metadata:    req.Metadata,
	}

	// 调用支付服务
	resp, err := h.paymentService.CreatePayment(c.Request.Context(), paymentReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
			Code:  "PAYMENT_CREATE_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// HandleCallback 处理支付回调
// @Summary 处理支付回调
// @Description 接收支付渠道的回调通知
// @Tags payment
// @Accept json
// @Produce json
// @Param channel path string true "支付渠道"
// @Success 200 {string} string "success"
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /payments/callback/{channel} [post]
func (h *PaymentHandler) HandleCallback(c *gin.Context) {
	channel := c.Param("channel")
	if channel == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "missing payment channel",
			Code:  "MISSING_CHANNEL",
		})
		return
	}

	// 读取回调数据
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "failed to read callback data",
			Code:  "INVALID_CALLBACK_DATA",
		})
		return
	}

	// 获取签名
	signature := c.GetHeader("X-Signature")
	if signature == "" {
		// 某些支付渠道可能使用不同的签名头
		signature = c.GetHeader("Signature")
	}

	// 验证回调
	callback, err := h.paymentService.VerifyCallback(c.Request.Context(), channel, body, signature)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
			Code:  "CALLBACK_VERIFICATION_FAILED",
		})
		return
	}

	// 返回成功响应（各个支付渠道要求不同的响应格式）
	switch channel {
	case "alipay":
		c.String(http.StatusOK, "success")
	case "wechat":
		c.XML(http.StatusOK, gin.H{
			"return_code": "SUCCESS",
			"return_msg":  "OK",
		})
	case "unionpay":
		c.String(http.StatusOK, "success")
	default:
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data":   callback,
		})
	}
}

// QueryPayment 查询支付状态
// @Summary 查询支付状态
// @Description 查询指定订单的支付状态
// @Tags payment
// @Accept json
// @Produce json
// @Param orderNo path string true "订单号"
// @Success 200 {object} adapters.QueryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /payments/query/{orderNo} [get]
func (h *PaymentHandler) QueryPayment(c *gin.Context) {
	orderNo := c.Param("orderNo")
	if orderNo == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "missing order number",
			Code:  "MISSING_ORDER_NO",
		})
		return
	}

	// 查询支付状态
	resp, err := h.paymentService.QueryPayment(c.Request.Context(), orderNo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
			Code:  "QUERY_PAYMENT_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// CreateRefundRequest 创建退款请求
type CreateRefundRequest struct {
	OrderNo  string `json:"order_no" binding:"required"`
	Amount   string `json:"amount" binding:"required"`
	Reason   string `json:"reason" binding:"required"`
	RefundNo string `json:"refund_no"`
}

// CreateRefund 创建退款
// @Summary 创建退款
// @Description 对已支付订单申请退款
// @Tags payment
// @Accept json
// @Produce json
// @Param request body CreateRefundRequest true "退款请求"
// @Success 200 {object} adapters.RefundResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /payments/refund [post]
func (h *PaymentHandler) CreateRefund(c *gin.Context) {
	var req CreateRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	// 解析退款金额
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "invalid amount format",
			Code:  "INVALID_AMOUNT",
		})
		return
	}

	// 生成退款单号
	if req.RefundNo == "" {
		req.RefundNo = generateRefundNo()
	}

	// 构造退款请求
	refundReq := &adapters.RefundRequest{
		OrderNo:   req.OrderNo,
		RefundNo:  req.RefundNo,
		Amount:    amount,
		Reason:    req.Reason,
		NotifyURL: "/api/v1/payments/refund/callback",
	}

	// 调用支付服务
	resp, err := h.paymentService.CreateRefund(c.Request.Context(), refundReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
			Code:  "REFUND_CREATE_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// CloseOrder 关闭订单
// @Summary 关闭订单
// @Description 关闭未支付的订单
// @Tags payment
// @Accept json
// @Produce json
// @Param orderNo path string true "订单号"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /payments/close/{orderNo} [post]
func (h *PaymentHandler) CloseOrder(c *gin.Context) {
	orderNo := c.Param("orderNo")
	if orderNo == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "missing order number",
			Code:  "MISSING_ORDER_NO",
		})
		return
	}

	// 关闭订单
	if err := h.paymentService.CloseOrder(c.Request.Context(), orderNo); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
			Code:  "CLOSE_ORDER_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "order closed successfully",
	})
}

// GetSupportedChannels 获取支持的支付渠道
// @Summary 获取支持的支付渠道
// @Description 获取系统支持的所有支付渠道列表
// @Tags payment
// @Accept json
// @Produce json
// @Success 200 {object} ChannelsResponse
// @Router /payments/channels [get]
func (h *PaymentHandler) GetSupportedChannels(c *gin.Context) {
	channels := h.paymentService.GetSupportedChannels()

	channelInfo := make([]map[string]interface{}, 0, len(channels))
	for _, channel := range channels {
		info := map[string]interface{}{
			"code": channel,
			"name": getChannelName(channel),
		}
		channelInfo = append(channelInfo, info)
	}

	c.JSON(http.StatusOK, ChannelsResponse{
		Channels: channelInfo,
	})
}

// 响应结构体

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

// SuccessResponse 成功响应
type SuccessResponse struct {
	Message string `json:"message"`
}

// ChannelsResponse 支付渠道响应
type ChannelsResponse struct {
	Channels []map[string]interface{} `json:"channels"`
}

// 辅助函数

// generateRefundNo 生成退款单号
func generateRefundNo() string {
	now := time.Now()
	return fmt.Sprintf("REFUND%s%s",
		now.Format("20060102150405"),
		uuid.New().String()[:8],
	)
}

// getChannelName 获取支付渠道名称
func getChannelName(code string) string {
	names := map[string]string{
		"alipay":   "支付宝",
		"wechat":   "微信支付",
		"unionpay": "银联支付",
	}

	if name, exists := names[code]; exists {
		return name
	}
	return code
}