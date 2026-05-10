//go:build legacy
// +build legacy

package handler

import (
	"net/http"
	"strconv"
	"time"
	"io"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/tvvshow/gokao/services/payment-service/internal/wechat"
	"github.com/tvvshow/gokao/services/payment-service/internal/models"
	"github.com/tvvshow/gokao/services/payment-service/internal/services"
)

// WeChatPayHandler 微信支付处理器
type WeChatPayHandler struct {
	wechatClient   *wechat.WeChatPayClient
	paymentService *services.PaymentService
}

// NewWeChatPayHandler 创建微信支付处理器
func NewWeChatPayHandler(client *wechat.WeChatPayClient, paymentService *services.PaymentService) *WeChatPayHandler {
	return &WeChatPayHandler{
		wechatClient:   client,
		paymentService: paymentService,
	}
}

// CreatePaymentRequest 创建支付请求
type CreatePaymentRequest struct {
	UserID      string  `json:"user_id" binding:"required"`
	ProductType string  `json:"product_type" binding:"required"` // membership, premium_feature
	ProductID   string  `json:"product_id" binding:"required"`   // plan_id, feature_id
	Amount      float64 `json:"amount" binding:"required"`
	Currency    string  `json:"currency"`
	ClientIP    string  `json:"client_ip"`
	PaymentType string  `json:"payment_type"` // native, jsapi, h5
	OpenID      string  `json:"open_id"`      // JSAPI支付需要
}

// CreatePayment 创建支付订单
// @Summary 创建微信支付订单
// @Description 创建微信支付订单，支持扫码支付和JSAPI支付
// @Tags 微信支付
// @Accept json
// @Produce json
// @Param request body CreatePaymentRequest true "支付请求"
// @Success 200 {object} models.APIResponse{data=wechat.PaymentResult}
// @Failure 400 {object} models.APIResponse
// @Router /api/v1/payment/wechat/create [post]
func (h *WeChatPayHandler) CreatePayment(c *gin.Context) {
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 设置默认值
	if req.Currency == "" {
		req.Currency = "CNY"
	}
	if req.ClientIP == "" {
		req.ClientIP = c.ClientIP()
	}
	if req.PaymentType == "" {
		req.PaymentType = "native"
	}

	// 生成订单ID
	orderID := h.generateOrderID()

	// 创建数据库订单记录
	order, err := h.paymentService.CreateOrder(&models.PaymentOrder{
		OrderID:     orderID,
		UserID:      req.UserID,
		ProductType: req.ProductType,
		ProductID:   req.ProductID,
		Amount:      int64(req.Amount * 100), // 转换为分
		Currency:    req.Currency,
		Status:      models.OrderStatusPending,
		PaymentType: "wechat",
		ClientIP:    req.ClientIP,
		ExpireTime:  time.Now().Add(30 * time.Minute), // 30分钟过期
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "创建订单失败: " + err.Error(),
		})
		return
	}

	// 构建微信支付订单
	wechatOrder := &wechat.PaymentOrder{
		OrderID:    orderID,
		UserID:     req.UserID,
		Amount:     order.Amount,
		Currency:   req.Currency,
		Subject:    h.getProductSubject(req.ProductType, req.ProductID),
		Body:       h.getProductBody(req.ProductType, req.ProductID),
		ClientIP:   req.ClientIP,
		TimeExpire: order.ExpireTime,
		CreatedAt:  time.Now(),
	}

	var result *wechat.PaymentResult
	var wechatErr error

	// 根据支付类型创建不同的订单
	switch req.PaymentType {
	case "jsapi":
		if req.OpenID == "" {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Message: "JSAPI支付需要提供openid",
			})
			return
		}
		result, wechatErr = h.wechatClient.CreateJSAPIOrder(wechatOrder, req.OpenID)
	case "native":
		result, wechatErr = h.wechatClient.CreateOrder(wechatOrder)
	default:
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "不支持的支付类型: " + req.PaymentType,
		})
		return
	}

	if wechatErr != nil {
		// 更新订单状态为失败
		h.paymentService.UpdateOrderStatus(orderID, models.OrderStatusFailed, wechatErr.Error())
		
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "创建微信支付订单失败: " + wechatErr.Error(),
		})
		return
	}

	if !result.Success {
		// 更新订单状态为失败
		h.paymentService.UpdateOrderStatus(orderID, models.OrderStatusFailed, result.Message)
		
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "微信支付创建失败: " + result.Message,
		})
		return
	}

	// 更新订单状态为等待支付
	h.paymentService.UpdateOrderStatus(orderID, models.OrderStatusWaitingPayment, "")

	// 保存微信支付相关信息
	h.paymentService.UpdateOrderWeChatInfo(orderID, result.PrepayID, result.QRCode)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "支付订单创建成功",
		Data:    result,
	})
}

// QueryPayment 查询支付状态
// @Summary 查询微信支付状态
// @Description 查询微信支付订单状态
// @Tags 微信支付
// @Accept json
// @Produce json
// @Param order_id path string true "订单ID"
// @Success 200 {object} models.APIResponse{data=wechat.PaymentResult}
// @Failure 400 {object} models.APIResponse
// @Router /api/v1/payment/wechat/query/{order_id} [get]
func (h *WeChatPayHandler) QueryPayment(c *gin.Context) {
	orderID := c.Param("order_id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "订单ID不能为空",
		})
		return
	}

	// 查询微信支付状态
	result, err := h.wechatClient.QueryOrder(orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "查询支付状态失败: " + err.Error(),
		})
		return
	}

	// 更新本地订单状态
	if result.Success {
		err = h.paymentService.UpdateOrderStatus(orderID, models.OrderStatusPaid, "")
		if err != nil {
			// 记录日志但不影响返回结果
			fmt.Printf("更新订单状态失败: %v\n", err)
		}
		
		// 激活会员权限
		go h.activateMembership(orderID)
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "查询成功",
		Data:    result,
	})
}

// PaymentNotify 支付回调处理
// @Summary 微信支付回调
// @Description 处理微信支付回调通知
// @Tags 微信支付
// @Accept xml
// @Produce xml
// @Success 200 {string} string "success"
// @Router /api/v1/payment/wechat/notify [post]
func (h *WeChatPayHandler) PaymentNotify(c *gin.Context) {
	// 读取回调数据
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.XML(http.StatusBadRequest, gin.H{
			"return_code": "FAIL",
			"return_msg":  "读取数据失败",
		})
		return
	}

	// 处理微信支付回调
	notify, err := h.wechatClient.ProcessNotify(body)
	if err != nil {
		c.XML(http.StatusBadRequest, gin.H{
			"return_code": "FAIL",
			"return_msg":  "处理回调失败: " + err.Error(),
		})
		return
	}

	// 验证订单存在
	order, err := h.paymentService.GetOrderByID(notify.OutTradeNo)
	if err != nil {
		c.XML(http.StatusBadRequest, gin.H{
			"return_code": "FAIL",
			"return_msg":  "订单不存在",
		})
		return
	}

	// 验证金额
	totalFee, _ := strconv.ParseInt(notify.TotalFee, 10, 64)
	if totalFee != order.Amount {
		c.XML(http.StatusBadRequest, gin.H{
			"return_code": "FAIL",
			"return_msg":  "金额不匹配",
		})
		return
	}

	// 更新订单状态
	if notify.TradeState == "SUCCESS" {
		err = h.paymentService.UpdateOrderStatus(notify.OutTradeNo, models.OrderStatusPaid, "")
		if err == nil {
			// 异步激活会员权限
			go h.activateMembership(notify.OutTradeNo)
		}
	} else {
		h.paymentService.UpdateOrderStatus(notify.OutTradeNo, models.OrderStatusFailed, "支付失败")
	}

	// 返回成功响应
	c.XML(http.StatusOK, gin.H{
		"return_code": "SUCCESS",
		"return_msg":  "OK",
	})
}

// RefundPayment 退款处理
// @Summary 微信支付退款
// @Description 处理微信支付退款
// @Tags 微信支付
// @Accept json
// @Produce json
// @Param request body wechat.RefundRequest true "退款请求"
// @Success 200 {object} models.APIResponse{data=wechat.RefundResult}
// @Router /api/v1/payment/wechat/refund [post]
func (h *WeChatPayHandler) RefundPayment(c *gin.Context) {
	var req wechat.RefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 验证订单状态
	order, err := h.paymentService.GetOrderByID(req.OrderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "订单不存在",
		})
		return
	}

	if order.Status != models.OrderStatusPaid {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "订单状态不支持退款",
		})
		return
	}

	// 执行退款
	result, err := h.wechatClient.RefundOrder(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "退款失败: " + err.Error(),
		})
		return
	}

	if result.Success {
		// 更新订单状态
		h.paymentService.UpdateOrderStatus(req.OrderID, models.OrderStatusRefunded, "")
		
		// 异步处理会员权限回收
		go h.revokeMembership(req.OrderID)
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "退款处理成功",
		Data:    result,
	})
}

// 生成订单ID
func (h *WeChatPayHandler) generateOrderID() string {
	return fmt.Sprintf("WX%d%06d", time.Now().Unix(), time.Now().Nanosecond()%1000000)
}

// 获取产品标题
func (h *WeChatPayHandler) getProductSubject(productType, productID string) string {
	switch productType {
	case "membership":
		return "高考志愿填报助手-会员服务"
	case "premium_feature":
		return "高考志愿填报助手-高级功能"
	default:
		return "高考志愿填报助手-付费服务"
	}
}

// 获取产品描述
func (h *WeChatPayHandler) getProductBody(productType, productID string) string {
	switch productType {
	case "membership":
		switch productID {
		case "basic_monthly":
			return "基础会员-月度订阅"
		case "basic_yearly":
			return "基础会员-年度订阅"
		case "premium_monthly":
			return "高级会员-月度订阅"
		case "premium_yearly":
			return "高级会员-年度订阅"
		default:
			return "会员服务订阅"
		}
	case "premium_feature":
		return "高级功能解锁"
	default:
		return "付费服务"
	}
}

// 激活会员权限
func (h *WeChatPayHandler) activateMembership(orderID string) {
	order, err := h.paymentService.GetOrderByID(orderID)
	if err != nil {
		fmt.Printf("获取订单失败: %v\n", err)
		return
	}

	err = h.paymentService.ActivateMembership(order.UserID, order.ProductID)
	if err != nil {
		fmt.Printf("激活会员权限失败: %v\n", err)
	}
}

// 回收会员权限
func (h *WeChatPayHandler) revokeMembership(orderID string) {
	order, err := h.paymentService.GetOrderByID(orderID)
	if err != nil {
		fmt.Printf("获取订单失败: %v\n", err)
		return
	}

	err = h.paymentService.RevokeMembership(order.UserID, order.ProductID)
	if err != nil {
		fmt.Printf("回收会员权限失败: %v\n", err)
	}
}
