package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"payment-service/internal/adapters"
	"payment-service/internal/models"
	"payment-service/internal/repository"
)

// PaymentService 支付服务
type PaymentService struct {
	repo           repository.PaymentRepository
	wechatAdapter  *adapters.WechatPayAdapter
	alipayAdapter  *adapters.AlipayAdapter
	logger         *logrus.Logger
}

// NewPaymentService 创建支付服务
func NewPaymentService(repo repository.PaymentRepository, logger *logrus.Logger) (*PaymentService, error) {
	// 初始化微信支付适配器
	wechatAdapter, err := adapters.NewWechatPayAdapter()
	if err != nil {
		logger.WithError(err).Warn("Failed to initialize WeChat Pay adapter")
		wechatAdapter = nil
	}

	// 初始化支付宝适配器
	alipayAdapter, err := adapters.NewAlipayAdapter()
	if err != nil {
		logger.WithError(err).Warn("Failed to initialize Alipay adapter")
		alipayAdapter = nil
	}

	return &PaymentService{
		repo:          repo,
		wechatAdapter: wechatAdapter,
		alipayAdapter: alipayAdapter,
		logger:        logger,
	}, nil
}

// CreatePayment 创建支付订单
func (s *PaymentService) CreatePayment(ctx context.Context, req *models.CreatePaymentRequest) (*models.PaymentResponse, error) {
	// 验证请求
	if err := s.validateCreatePaymentRequest(req); err != nil {
		return nil, fmt.Errorf("invalid payment request: %w", err)
	}

	// 创建订单记录
	order := &models.PaymentOrder{
		OrderID:       generateOrderID(),
		UserID:        req.UserID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Description:   req.Description,
		PaymentMethod: req.PaymentMethod,
		Status:        "pending",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Extra:         req.Extra,
	}

	// 保存到数据库
	if err := s.repo.CreateOrder(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// 根据支付方式调用相应的适配器
	var paymentResp *models.PaymentResponse
	var err error

	switch req.PaymentMethod {
	case "wechat_pay":
		if s.wechatAdapter == nil {
			return nil, fmt.Errorf("WeChat Pay is not available")
		}
		paymentResp, err = s.wechatAdapter.CreateOrder(ctx, order)
	case "alipay":
		if s.alipayAdapter == nil {
			return nil, fmt.Errorf("Alipay is not available")
		}
		paymentResp, err = s.alipayAdapter.CreateOrder(ctx, order)
	case "alipay_qr":
		if s.alipayAdapter == nil {
			return nil, fmt.Errorf("Alipay is not available")
		}
		paymentResp, err = s.alipayAdapter.CreateQROrder(ctx, order)
	default:
		return nil, fmt.Errorf("unsupported payment method: %s", req.PaymentMethod)
	}

	if err != nil {
		// 更新订单状态为失败
		order.Status = "failed"
		order.UpdatedAt = time.Now()
		s.repo.UpdateOrder(ctx, order)
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// 更新订单信息
	order.PaymentURL = paymentResp.PaymentURL
	order.ExpiresAt = &paymentResp.ExpiresAt
	order.UpdatedAt = time.Now()
	if err := s.repo.UpdateOrder(ctx, order); err != nil {
		s.logger.WithError(err).Error("Failed to update order after payment creation")
	}

	s.logger.WithFields(logrus.Fields{
		"order_id":       order.OrderID,
		"user_id":        order.UserID,
		"payment_method": order.PaymentMethod,
		"amount":         order.Amount,
	}).Info("Payment order created successfully")

	return paymentResp, nil
}

// QueryPayment 查询支付状态
func (s *PaymentService) QueryPayment(ctx context.Context, orderID string) (*models.PaymentStatus, error) {
	// 从数据库获取订单信息
	order, err := s.repo.GetOrder(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// 如果订单已经是最终状态，直接返回
	if order.Status == "completed" || order.Status == "failed" || order.Status == "cancelled" {
		return &models.PaymentStatus{
			OrderID:       order.OrderID,
			Status:        order.Status,
			PaymentMethod: order.PaymentMethod,
			TransactionID: order.TransactionID,
			PaidAt:        order.PaidAt,
			Amount:        order.Amount,
			Currency:      order.Currency,
		}, nil
	}

	// 调用相应的支付适配器查询状态
	var paymentStatus *models.PaymentStatus

	switch order.PaymentMethod {
	case "wechat_pay":
		if s.wechatAdapter == nil {
			return nil, fmt.Errorf("WeChat Pay is not available")
		}
		paymentStatus, err = s.wechatAdapter.QueryOrder(ctx, orderID)
	case "alipay", "alipay_qr":
		if s.alipayAdapter == nil {
			return nil, fmt.Errorf("Alipay is not available")
		}
		paymentStatus, err = s.alipayAdapter.QueryOrder(ctx, orderID)
	default:
		return nil, fmt.Errorf("unsupported payment method: %s", order.PaymentMethod)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query payment status: %w", err)
	}

	// 更新数据库中的订单状态
	if paymentStatus.Status != order.Status {
		order.Status = paymentStatus.Status
		order.TransactionID = paymentStatus.TransactionID
		order.PaidAt = paymentStatus.PaidAt
		order.UpdatedAt = time.Now()
		
		if err := s.repo.UpdateOrder(ctx, order); err != nil {
			s.logger.WithError(err).Error("Failed to update order status")
		}

		s.logger.WithFields(logrus.Fields{
			"order_id":       orderID,
			"old_status":     order.Status,
			"new_status":     paymentStatus.Status,
			"transaction_id": paymentStatus.TransactionID,
		}).Info("Payment status updated")
	}

	return paymentStatus, nil
}

// RefundPayment 退款
func (s *PaymentService) RefundPayment(ctx context.Context, req *models.RefundRequest) (*models.RefundResponse, error) {
	// 验证请求
	if err := s.validateRefundRequest(req); err != nil {
		return nil, fmt.Errorf("invalid refund request: %w", err)
	}

	// 获取原订单信息
	order, err := s.repo.GetOrder(ctx, req.OrderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// 验证订单状态
	if order.Status != "completed" {
		return nil, fmt.Errorf("order is not in completed status, cannot refund")
	}

	// 验证退款金额
	if req.Amount > order.Amount {
		return nil, fmt.Errorf("refund amount cannot exceed order amount")
	}

	// 调用相应的支付适配器进行退款
	var refundResp *models.RefundResponse

	switch order.PaymentMethod {
	case "wechat_pay":
		if s.wechatAdapter == nil {
			return nil, fmt.Errorf("WeChat Pay is not available")
		}
		refundResp, err = s.wechatAdapter.RefundOrder(ctx, req)
	case "alipay", "alipay_qr":
		if s.alipayAdapter == nil {
			return nil, fmt.Errorf("Alipay is not available")
		}
		refundResp, err = s.alipayAdapter.RefundOrder(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported payment method: %s", order.PaymentMethod)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to process refund: %w", err)
	}

	// 创建退款记录
	refund := &models.RefundRecord{
		RefundID:      req.RefundID,
		OrderID:       req.OrderID,
		Amount:        req.Amount,
		Reason:        req.Reason,
		Status:        refundResp.Status,
		RefundedAt:    &refundResp.RefundedAt,
		PaymentMethod: order.PaymentMethod,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.CreateRefund(ctx, refund); err != nil {
		s.logger.WithError(err).Error("Failed to create refund record")
	}

	// 更新订单状态
	if refundResp.Status == "completed" && req.Amount == order.Amount {
		order.Status = "refunded"
		order.UpdatedAt = time.Now()
		if err := s.repo.UpdateOrder(ctx, order); err != nil {
			s.logger.WithError(err).Error("Failed to update order status after refund")
		}
	}

	s.logger.WithFields(logrus.Fields{
		"refund_id": req.RefundID,
		"order_id":  req.OrderID,
		"amount":    req.Amount,
		"status":    refundResp.Status,
	}).Info("Refund processed successfully")

	return refundResp, nil
}

// HandleCallback 处理支付回调（带事务和行级锁）
func (s *PaymentService) HandleCallback(ctx context.Context, paymentMethod string, request *http.Request) (*models.CallbackResult, error) {
	var result *models.CallbackResult
	var err error

	// 根据支付方式调用相应的适配器处理回调
	switch paymentMethod {
	case "wechat_pay":
		if s.wechatAdapter == nil {
			return nil, fmt.Errorf("WeChat Pay is not available")
		}
		result, err = s.wechatAdapter.HandleCallback(ctx, request)
	case "alipay":
		if s.alipayAdapter == nil {
			return nil, fmt.Errorf("Alipay is not available")
		}
		result, err = s.alipayAdapter.HandleCallback(ctx, request)
	default:
		return nil, fmt.Errorf("unsupported payment method: %s", paymentMethod)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to handle callback: %w", err)
	}

	// 开始事务处理回调，确保并发安全
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to begin transaction for callback")
		return result, nil
	}
	defer tx.Rollback()

	// 使用行级锁获取订单信息，防止并发更新
	order, err := s.repo.WithTx(tx).GetOrderWithLock(ctx, result.OrderID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get order with lock for callback")
		return result, nil
	}

	// 检查订单是否已经是最终状态，避免重复处理
	if order.Status == "completed" || order.Status == "failed" || order.Status == "cancelled" {
		s.logger.WithFields(logrus.Fields{
			"order_id": result.OrderID,
			"status":   order.Status,
		}).Warn("Order already in final state, skipping callback processing")
		tx.Rollback()
		return result, nil
	}

	// 更新订单信息
	order.Status = result.Status
	order.TransactionID = result.TransactionID
	order.PaidAt = result.PaidAt
	order.UpdatedAt = time.Now()

	if err := s.repo.WithTx(tx).UpdateOrder(ctx, order); err != nil {
		s.logger.WithError(err).Error("Failed to update order from callback")
		tx.Rollback()
		return result, nil
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		s.logger.WithError(err).Error("Failed to commit transaction for callback")
		return result, nil
	}

	s.logger.WithFields(logrus.Fields{
		"order_id":       result.OrderID,
		"payment_method": paymentMethod,
		"status":         result.Status,
		"transaction_id": result.TransactionID,
	}).Info("Payment callback processed successfully with transaction")

	return result, nil
}

// 辅助函数

// validateCreatePaymentRequest 验证创建支付请求
func (s *PaymentService) validateCreatePaymentRequest(req *models.CreatePaymentRequest) error {
	if req.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if req.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}
	if req.Currency == "" {
		req.Currency = "CNY"
	}
	if req.Description == "" {
		return fmt.Errorf("description is required")
	}
	if req.PaymentMethod == "" {
		return fmt.Errorf("payment_method is required")
	}
	
	// 验证支付方式
	supportedMethods := []string{"wechat_pay", "alipay", "alipay_qr"}
	for _, method := range supportedMethods {
		if req.PaymentMethod == method {
			return nil
		}
	}
	
	return fmt.Errorf("unsupported payment method: %s", req.PaymentMethod)
}

// validateRefundRequest 验证退款请求
func (s *PaymentService) validateRefundRequest(req *models.RefundRequest) error {
	if req.OrderID == "" {
		return fmt.Errorf("order_id is required")
	}
	if req.RefundID == "" {
		return fmt.Errorf("refund_id is required")
	}
	if req.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}
	if req.Reason == "" {
		return fmt.Errorf("reason is required")
	}
	
	return nil
}

// generateOrderID 生成订单ID
func generateOrderID() string {
	return fmt.Sprintf("GK%d", time.Now().UnixNano())
}
