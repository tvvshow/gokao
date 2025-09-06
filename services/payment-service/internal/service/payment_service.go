package service

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"github.com/gaokaohub/gaokao/services/payment-service/internal/adapters"
	"github.com/gaokaohub/gaokao/services/payment-service/internal/models"
	"github.com/gaokaohub/gaokao/services/payment-service/internal/repository"
)

// PaymentService 支付服务
type PaymentService struct {
	repo             repository.PaymentRepository
	wechatAdapter    adapters.PaymentAdapter
	alipayAdapter    adapters.PaymentAdapter
	logger           *logrus.Logger
}

// NewPaymentService 创建支付服务
func NewPaymentService(repo repository.PaymentRepository, logger *logrus.Logger) (*PaymentService, error) {
	// 初始化微信支付适配器
	wechatAdapter, _ := adapters.NewWechatPayAdapter(adapters.WechatPayConfig{})

	// 初始化支付宝适配器
	alipayAdapter, _ := adapters.NewAlipayAdapter(adapters.AlipayConfig{})

	return &PaymentService{
		repo:          repo,
		wechatAdapter: wechatAdapter,
		alipayAdapter: alipayAdapter,
		logger:        logger,
	}, nil
}

// CreatePayment 创建支付订单
func (s *PaymentService) CreatePayment(ctx context.Context, req *models.CreatePaymentRequest) (*models.PaymentOrderResponse, error) {
	// 验证请求
	if err := s.validateCreatePaymentRequest(req); err != nil {
		return nil, fmt.Errorf("invalid payment request: %w", err)
	}

	// 创建订单记录
	order := &models.PaymentOrder{
		OrderNo:     generateOrderID(),
		UserID:      uuid.MustParse(req.UserID),
		Amount:      req.Amount,
		Currency:    req.Currency,
		Subject:     req.Description,
		Description: req.Description,
		Channel:     req.PaymentMethod,
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    models.PaymentJSONB(req.Extra),
	}

	// 保存到数据库
	if err := s.repo.CreatePayment(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// 根据支付方式调用相应的适配器
	var paymentResp *adapters.PaymentResponse
	var err error

	switch req.PaymentMethod {
	case "wechat_pay":
		if s.wechatAdapter == nil {
			return nil, fmt.Errorf("WeChat Pay is not available")
		}
		// 构造支付请求
		paymentReq := &adapters.PaymentRequest{
			OrderNo:     order.OrderNo,
			Amount:      order.Amount,
			Subject:     order.Subject,
			Description: order.Description,
			NotifyURL:   "", // 从配置中获取
			ReturnURL:   "", // 从配置中获取
			UserID:      req.UserID,
			ExpireTime:  30 * time.Minute,
		}
		paymentResp, err = s.wechatAdapter.CreatePayment(ctx, paymentReq)
	case "alipay", "alipay_qr":
		if s.alipayAdapter == nil {
			return nil, fmt.Errorf("Alipay is not available")
		}
		// 构造支付请求
		paymentReq := &adapters.PaymentRequest{
			OrderNo:     order.OrderNo,
			Amount:      order.Amount,
			Subject:     order.Subject,
			Description: order.Description,
			NotifyURL:   "", // 从配置中获取
			ReturnURL:   "", // 从配置中获取
			UserID:      req.UserID,
			ExpireTime:  30 * time.Minute,
		}
		paymentResp, err = s.alipayAdapter.CreatePayment(ctx, paymentReq)
	default:
		return nil, fmt.Errorf("unsupported payment method: %s", req.PaymentMethod)
	}

	if err != nil {
		// 更新订单状态为失败
		order.Status = "failed"
		order.UpdatedAt = time.Now()
		s.repo.UpdatePaymentStatus(ctx, order.ID.String(), "failed", "")
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// 转换响应格式
	resp := &models.PaymentOrderResponse{
		ID:         order.ID,
		OrderNo:    paymentResp.OrderNo,
		Amount:     order.Amount,
		Currency:   order.Currency,
		Subject:    order.Subject,
		Channel:    order.Channel,
		Status:     order.Status,
		PaymentURL: paymentResp.PayURL,
		QRCode:     paymentResp.QRCode,
		ExpiredAt:  &paymentResp.ExpiredAt,
		CreatedAt:  order.CreatedAt,
	}

	// 更新订单信息
	order.PaymentURL = paymentResp.PayURL
	if !paymentResp.ExpiredAt.IsZero() {
		order.ExpiredAt = &paymentResp.ExpiredAt
	}
	order.UpdatedAt = time.Now()
	if err := s.repo.UpdatePaymentStatus(ctx, order.ID.String(), order.Status, order.ChannelTradeNo); err != nil {
		s.logger.WithError(err).Error("Failed to update order after payment creation")
	}

	s.logger.WithFields(logrus.Fields{
		"order_id":       order.OrderNo,
		"user_id":        order.UserID,
		"payment_method": order.Channel,
		"amount":         order.Amount,
	}).Info("Payment order created successfully")

	return resp, nil
}

// QueryPayment 查询支付状态
func (s *PaymentService) QueryPayment(ctx context.Context, orderID string) (*models.PaymentStatus, error) {
	// 从数据库获取订单信息
	order, err := s.repo.GetPaymentByOrderID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// 如果订单已经是最终状态，直接返回
	if order.Status == "completed" || order.Status == "failed" || order.Status == "cancelled" {
		return &models.PaymentStatus{
			OrderNo:       order.OrderNo,
			Status:        order.Status,
			PaymentMethod: order.Channel,
			TransactionID: order.ChannelTradeNo,
			PaidAt:        order.PaidAt,
			Amount:        order.Amount,
			Currency:      order.Currency,
		}, nil
	}

	// 调用相应的支付适配器查询状态
	var queryResp *adapters.QueryResponse

	switch order.Channel {
	case "wechat_pay":
		if s.wechatAdapter == nil {
			return nil, fmt.Errorf("WeChat Pay is not available")
		}
		queryReq := &adapters.QueryRequest{
			OrderNo: orderID,
		}
		queryResp, err = s.wechatAdapter.QueryPayment(ctx, queryReq)
	case "alipay", "alipay_qr":
		if s.alipayAdapter == nil {
			return nil, fmt.Errorf("Alipay is not available")
		}
		queryReq := &adapters.QueryRequest{
			OrderNo: orderID,
		}
		queryResp, err = s.alipayAdapter.QueryPayment(ctx, queryReq)
	default:
		return nil, fmt.Errorf("unsupported payment method: %s", order.Channel)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query payment status: %w", err)
	}

	// 转换响应格式
	paymentStatus := &models.PaymentStatus{
		OrderNo:       queryResp.OrderNo,
		Status:        queryResp.Status,
		PaymentMethod: order.Channel,
		TransactionID: queryResp.ChannelTradeNo,
		PaidAt:        queryResp.PaidAt,
		Amount:        queryResp.Amount,
		Currency:      order.Currency,
	}

	// 更新数据库中的订单状态
	if queryResp.Status != order.Status {
		order.Status = queryResp.Status
		order.ChannelTradeNo = queryResp.ChannelTradeNo
		if queryResp.PaidAt != nil {
			order.PaidAt = queryResp.PaidAt
		}
		order.UpdatedAt = time.Now()
		
		if err := s.repo.UpdatePaymentStatus(ctx, order.ID.String(), order.Status, order.ChannelTradeNo); err != nil {
			s.logger.WithError(err).Error("Failed to update order status")
		}

		s.logger.WithFields(logrus.Fields{
			"order_id":       orderID,
			"old_status":     order.Status,
			"new_status":     queryResp.Status,
			"transaction_id": queryResp.ChannelTradeNo,
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
	order, err := s.repo.GetPaymentByOrderID(ctx, req.OrderNo)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// 验证订单状态
	if order.Status != "completed" {
		return nil, fmt.Errorf("order is not in completed status, cannot refund")
	}

	// 验证退款金额
	if req.Amount.GreaterThan(order.Amount) {
		return nil, fmt.Errorf("refund amount cannot exceed order amount")
	}

	// 调用相应的支付适配器进行退款
	var refundResp *adapters.RefundResponse

	switch order.Channel {
	case "wechat_pay":
		if s.wechatAdapter == nil {
			return nil, fmt.Errorf("WeChat Pay is not available")
		}
		// 构造退款请求
		refundReq := &adapters.RefundRequest{
			OrderNo:      req.OrderNo,
			RefundNo:     req.RefundID,
			Amount:       req.Amount,
			RefundAmount: req.Amount,
			TotalAmount:  order.Amount,
			Reason:       req.Reason,
			NotifyURL:    "", // 从配置中获取
		}
		refundResp, err = s.wechatAdapter.CreateRefund(ctx, refundReq)
	case "alipay", "alipay_qr":
		if s.alipayAdapter == nil {
			return nil, fmt.Errorf("Alipay is not available")
		}
		// 构造退款请求
		refundReq := &adapters.RefundRequest{
			OrderNo:      req.OrderNo,
			RefundNo:     req.RefundID,
			Amount:       req.Amount,
			RefundAmount: req.Amount,
			TotalAmount:  order.Amount,
			Reason:       req.Reason,
			NotifyURL:    "", // 从配置中获取
		}
		refundResp, err = s.alipayAdapter.CreateRefund(ctx, refundReq)
	default:
		return nil, fmt.Errorf("unsupported payment method: %s", order.Channel)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to process refund: %w", err)
	}

	// 转换响应格式
	resp := &models.RefundResponse{
		RefundID:      refundResp.RefundNo,
		OrderNo:       req.OrderNo,
		Status:        refundResp.Status,
		Amount:        refundResp.Amount,
		Currency:      "CNY",
		RefundedAt:    refundResp.RefundedAt,
		PaymentMethod: order.Channel,
		Extra:         models.PaymentJSONB{},
	}

	// 创建退款记录
	refund := &models.RefundRecord{
		RefundNo:        req.RefundID,
		OrderNo:         req.OrderNo,
		ChannelRefundNo: refundResp.ChannelRefundNo,
		Amount:          req.Amount,
		Reason:          req.Reason,
		Status:          refundResp.Status,
		RefundedAt:      &refundResp.RefundedAt,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.repo.CreateRefund(ctx, refund); err != nil {
		s.logger.WithError(err).Error("Failed to create refund record")
	}

	// 更新订单状态
	if refundResp.Status == "success" && req.Amount.Equal(order.Amount) {
		order.Status = "refunded"
		order.UpdatedAt = time.Now()
		if err := s.repo.UpdatePaymentStatus(ctx, order.ID.String(), order.Status, order.ChannelTradeNo); err != nil {
			s.logger.WithError(err).Error("Failed to update order status after refund")
		}
	}

	s.logger.WithFields(logrus.Fields{
		"refund_id": req.RefundID,
		"order_id":  req.OrderNo,
		"amount":    req.Amount,
		"status":    refundResp.Status,
	}).Info("Refund processed successfully")

	return resp, nil
}

// HandleCallback 处理支付回调（带事务和行级锁）
func (s *PaymentService) HandleCallback(ctx context.Context, paymentMethod string, request *http.Request) (*models.CallbackResult, error) {
	// 读取请求体
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	// 根据支付方式调用相应的适配器处理回调
	var callback *adapters.PaymentCallback

	switch paymentMethod {
	case "wechat_pay":
		if s.wechatAdapter == nil {
			return nil, fmt.Errorf("WeChat Pay is not available")
		}
		// 微信支付回调验证签名
		callback, err = s.wechatAdapter.VerifyCallback(ctx, body, "")
	case "alipay":
		if s.alipayAdapter == nil {
			return nil, fmt.Errorf("Alipay is not available")
		}
		// 支付宝回调验证签名
		callback, err = s.alipayAdapter.VerifyCallback(ctx, body, "")
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
		return &models.CallbackResult{
			OrderNo:       callback.OrderNo,
			Status:        callback.Status,
			PaymentMethod: paymentMethod,
			TransactionID: callback.ChannelTradeNo,
			Amount:        callback.Amount,
			Currency:      "CNY",
			PaidAt:        &callback.PaidAt,
			Extra:         models.PaymentJSONB{},
		}, nil
	}
	defer tx.Rollback()

	// 使用行级锁获取订单信息，防止并发更新
	order, err := s.repo.WithTx(tx).GetPaymentByIDWithLock(ctx, callback.OrderNo)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get order with lock for callback")
		return &models.CallbackResult{
			OrderNo:       callback.OrderNo,
			Status:        callback.Status,
			PaymentMethod: paymentMethod,
			TransactionID: callback.ChannelTradeNo,
			Amount:        callback.Amount,
			Currency:      "CNY",
			PaidAt:        &callback.PaidAt,
			Extra:         models.PaymentJSONB{},
		}, nil
	}

	// 检查订单是否已经是最终状态，避免重复处理
	if order.Status == "completed" || order.Status == "failed" || order.Status == "cancelled" {
		s.logger.WithFields(logrus.Fields{
			"order_id": callback.OrderNo,
			"status":   order.Status,
		}).Warn("Order already in final state, skipping callback processing")
		tx.Rollback()
		return &models.CallbackResult{
			OrderNo:       callback.OrderNo,
			Status:        callback.Status,
			PaymentMethod: paymentMethod,
			TransactionID: callback.ChannelTradeNo,
			Amount:        callback.Amount,
			Currency:      "CNY",
			PaidAt:        &callback.PaidAt,
			Extra:         models.PaymentJSONB{},
		}, nil
	}

	// 更新订单信息
	order.Status = callback.Status
	order.ChannelTradeNo = callback.ChannelTradeNo
	if !callback.PaidAt.IsZero() {
		order.PaidAt = &callback.PaidAt
	}
	order.UpdatedAt = time.Now()

	if err := s.repo.WithTx(tx).UpdatePaymentStatus(ctx, order.ID.String(), order.Status, order.ChannelTradeNo); err != nil {
		s.logger.WithError(err).Error("Failed to update order from callback")
		tx.Rollback()
		return &models.CallbackResult{
			OrderNo:       callback.OrderNo,
			Status:        callback.Status,
			PaymentMethod: paymentMethod,
			TransactionID: callback.ChannelTradeNo,
			Amount:        callback.Amount,
			Currency:      "CNY",
			PaidAt:        &callback.PaidAt,
			Extra:         models.PaymentJSONB{},
		}, nil
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		s.logger.WithError(err).Error("Failed to commit transaction for callback")
		return &models.CallbackResult{
			OrderNo:       callback.OrderNo,
			Status:        callback.Status,
			PaymentMethod: paymentMethod,
			TransactionID: callback.ChannelTradeNo,
			Amount:        callback.Amount,
			Currency:      "CNY",
			PaidAt:        &callback.PaidAt,
			Extra:         models.PaymentJSONB{},
		}, nil
	}

	result := &models.CallbackResult{
		OrderNo:       callback.OrderNo,
		Status:        callback.Status,
		PaymentMethod: paymentMethod,
		TransactionID: callback.ChannelTradeNo,
		Amount:        callback.Amount,
		Currency:      "CNY",
		PaidAt:        &callback.PaidAt,
		Extra:         models.PaymentJSONB{},
	}

	s.logger.WithFields(logrus.Fields{
		"order_id":       callback.OrderNo,
		"payment_method": paymentMethod,
		"status":         callback.Status,
		"transaction_id": callback.ChannelTradeNo,
	}).Info("Payment callback processed successfully with transaction")

	return result, nil
}

// 辅助函数

// validateCreatePaymentRequest 验证创建支付请求
func (s *PaymentService) validateCreatePaymentRequest(req *models.CreatePaymentRequest) error {
	if req.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if req.Amount.LessThanOrEqual(decimal.NewFromInt(0)) {
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
	if req.OrderNo == "" {
		return fmt.Errorf("order_id is required")
	}
	if req.RefundID == "" {
		return fmt.Errorf("refund_id is required")
	}
	if req.Amount.LessThanOrEqual(decimal.NewFromInt(0)) {
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

// ListPayments 获取支付记录列表
func (s *PaymentService) ListPayments(ctx context.Context, filter *models.PaymentFilter) ([]*models.PaymentOrder, int64, error) {
	return s.repo.ListPayments(ctx, filter)
}

// GetPaymentStatistics 获取支付统计信息
func (s *PaymentService) GetPaymentStatistics(ctx context.Context, startDate, endDate time.Time, channel string) (*models.PaymentStatistics, error) {
	return s.repo.GetPaymentStatistics(ctx, startDate, endDate, channel)
}
