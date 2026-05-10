package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tvvshow/gokao/services/payment-service/internal/adapters"
	"github.com/tvvshow/gokao/services/payment-service/internal/models"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// PaymentService 支付服务
type PaymentService struct {
	db             *sql.DB
	redis          *redis.Client
	adapterFactory adapters.PaymentAdapterFactory
}

// NewPaymentService 创建支付服务
func NewPaymentService(db *sql.DB, redis *redis.Client, adapterFactory adapters.PaymentAdapterFactory) *PaymentService {
	return &PaymentService{
		db:             db,
		redis:          redis,
		adapterFactory: adapterFactory,
	}
}

// CreatePayment 创建支付
func (s *PaymentService) CreatePayment(ctx context.Context, req *adapters.PaymentRequest) (*adapters.PaymentResponse, error) {
	// 获取支付适配器
	adapter, err := s.adapterFactory.GetAdapter(req.Metadata["channel"].(string))
	if err != nil {
		return nil, fmt.Errorf("failed to get payment adapter: %w", err)
	}

	// 创建支付订单记录
	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	expireTime := time.Now().Add(req.ExpireTime)
	order := &models.PaymentOrder{
		OrderNo:     req.OrderNo,
		UserID:      userUUID,
		Amount:      req.Amount,
		Currency:    "CNY",
		Subject:     req.Subject,
		Description: req.Description,
		Status:      models.PaymentStatusPending,
		Channel:     req.PaymentMethod,
		ClientIP:    req.ClientIP,
		NotifyURL:   req.NotifyURL,
		ReturnURL:   req.ReturnURL,
		ExpiredAt:   &expireTime,
		Metadata:    models.PaymentJSONB(req.Metadata),
	}

	// 保存订单到数据库
	if err := s.savePaymentOrder(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to save payment order: %w", err)
	}

	// 调用支付适配器创建支付
	resp, err := adapter.CreatePayment(ctx, req)
	if err != nil {
		// 更新订单状态为失败
		if updateErr := s.updateOrderStatus(ctx, req.OrderNo, "failed"); updateErr != nil {
			// 记录日志但不返回错误
		}
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// 缓存支付信息
	s.cachePaymentInfo(ctx, req.OrderNo, resp)

	return resp, nil
}

// VerifyCallback 验证支付回调
func (s *PaymentService) VerifyCallback(ctx context.Context, channel string, data []byte, signature string) (*adapters.PaymentCallback, error) {
	// 获取支付适配器
	adapter, err := s.adapterFactory.GetAdapter(channel)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment adapter: %w", err)
	}

	// 记录回调日志
	s.logCallback(ctx, channel, string(data), signature)

	// 验证回调签名
	callback, err := adapter.VerifyCallback(ctx, data, signature)
	if err != nil {
		return nil, fmt.Errorf("failed to verify callback: %w", err)
	}

	// 处理支付成功回调
	if callback.Status == "success" {
		if err := s.handlePaymentSuccess(ctx, callback); err != nil {
			return nil, fmt.Errorf("failed to handle payment success: %w", err)
		}
	}

	return callback, nil
}

// QueryPayment 查询支付状态
func (s *PaymentService) QueryPayment(ctx context.Context, orderNo string) (*adapters.QueryResponse, error) {
	// 从数据库获取订单信息
	order, err := s.getPaymentOrder(ctx, orderNo)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment order: %w", err)
	}

	// 如果订单已支付，直接返回
	if order.Status == models.PaymentStatusPaid {
		return &adapters.QueryResponse{
			OrderNo:        order.OrderNo,
			ChannelTradeNo: order.ChannelTradeNo,
			Amount:         order.Amount,
			Status:         order.Status,
			PaidAt:         order.PaidAt,
		}, nil
	}

	// 获取支付适配器
	adapter, err := s.adapterFactory.GetAdapter(order.Channel)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment adapter: %w", err)
	}

	// 查询支付渠道状态
	resp, err := adapter.QueryPayment(ctx, &adapters.QueryRequest{
		OrderNo:        orderNo,
		ChannelTradeNo: order.ChannelTradeNo,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query payment: %w", err)
	}

	// 如果状态发生变化，更新数据库
	if resp.Status != order.Status {
		if err := s.updateOrderStatus(ctx, orderNo, resp.Status); err != nil {
			// 记录日志但不返回错误
		}
	}

	return resp, nil
}

// CreateRefund 创建退款
func (s *PaymentService) CreateRefund(ctx context.Context, req *adapters.RefundRequest) (*adapters.RefundResponse, error) {
	// 检查订单是否可以退款
	order, err := s.getPaymentOrder(ctx, req.OrderNo)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment order: %w", err)
	}

	if !order.CanRefund() {
		return nil, fmt.Errorf("order cannot be refunded, status: %s", order.Status)
	}

	// 获取支付适配器
	adapter, err := s.adapterFactory.GetAdapter(order.Channel)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment adapter: %w", err)
	}

	// 创建退款记录
	refund := &models.PaymentRefund{
		RefundNo:     req.RefundNo,
		OrderID:      order.ID,
		OrderNo:      req.OrderNo,
		Amount:       order.Amount,
		RefundAmount: req.RefundAmount,
		Reason:       req.Reason,
		Status:       models.RefundStatusProcessing,
		Channel:      order.Channel,
	}

	if err := s.saveRefundRecord(ctx, refund); err != nil {
		return nil, fmt.Errorf("failed to save refund record: %w", err)
	}

	// 调用支付适配器创建退款
	resp, err := adapter.CreateRefund(ctx, req)
	if err != nil {
		// 更新退款状态为失败
		s.updateRefundStatus(ctx, req.RefundNo, models.RefundStatusFailed)
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}

	// 更新退款状态
	if err := s.updateRefundStatus(ctx, req.RefundNo, resp.Status); err != nil {
		// 记录日志但不返回错误
	}

	return resp, nil
}

// CloseOrder 关闭订单
func (s *PaymentService) CloseOrder(ctx context.Context, orderNo string) error {
	// 获取订单信息
	order, err := s.getPaymentOrder(ctx, orderNo)
	if err != nil {
		return fmt.Errorf("failed to get payment order: %w", err)
	}

	// 只有待支付的订单可以关闭
	if order.Status != models.PaymentStatusPending {
		return fmt.Errorf("order cannot be closed, status: %s", order.Status)
	}

	// 获取支付适配器
	adapter, err := s.adapterFactory.GetAdapter(order.Channel)
	if err != nil {
		return fmt.Errorf("failed to get payment adapter: %w", err)
	}

	// 调用支付适配器关闭订单
	if err := adapter.CloseOrder(ctx, orderNo); err != nil {
		return fmt.Errorf("failed to close order: %w", err)
	}

	// 更新订单状态
	return s.updateOrderStatus(ctx, orderNo, models.PaymentStatusCanceled)
}

// GetSupportedChannels 获取支持的支付渠道
func (s *PaymentService) GetSupportedChannels() []string {
	return []string{
		models.ChannelAlipay,
		models.ChannelWechat,
		models.ChannelUnionpay,
		models.ChannelQQ,
	}
}

// savePaymentOrder 保存支付订单
func (s *PaymentService) savePaymentOrder(ctx context.Context, order *models.PaymentOrder) error {
	query := `
		INSERT INTO payment_orders (
			order_no, user_id, amount, currency, subject, description, 
			status, payment_channel, client_ip, notify_url, return_url, 
			expire_time, created_at, updated_at, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err := s.db.ExecContext(ctx, query,
		order.OrderNo, order.UserID, order.Amount, order.Currency,
		order.Subject, order.Description, order.Status, order.Channel,
		order.ClientIP, order.NotifyURL, order.ReturnURL, order.ExpiredAt,
		order.CreatedAt, order.UpdatedAt, order.Metadata,
	)

	return err
}

// getPaymentOrder 获取支付订单
// getPaymentOrder 获取支付订单 - 优化版本，添加缓存机制
func (s *PaymentService) getPaymentOrder(ctx context.Context, orderNo string) (*models.PaymentOrder, error) {
	// 先尝试从缓存获取
	cacheKey := fmt.Sprintf("payment_order:%s", orderNo)
	if s.redis != nil {
		cachedData, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var order models.PaymentOrder
			if json.Unmarshal([]byte(cachedData), &order) == nil {
				return &order, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	query := `
		SELECT id, order_no, user_id, amount, currency, subject, description,
		       status, payment_channel, channel_trade_no, client_ip, notify_url,
		       return_url, expire_time, paid_at, created_at, updated_at, metadata
		FROM payment_orders
		WHERE order_no = $1
	`

	order := &models.PaymentOrder{}
	err := s.db.QueryRowContext(ctx, query, orderNo).Scan(
		&order.ID, &order.OrderNo, &order.UserID, &order.Amount,
		&order.Currency, &order.Subject, &order.Description, &order.Status,
		&order.Channel, &order.ChannelTradeNo, &order.ClientIP,
		&order.NotifyURL, &order.ReturnURL, &order.ExpiredAt,
		&order.PaidAt, &order.CreatedAt, &order.UpdatedAt, &order.Metadata,
	)

	if err != nil {
		return nil, err
	}

	// 缓存查询结果（已支付订单缓存更长时间）
	if s.redis != nil {
		data, _ := json.Marshal(order)
		cacheTTL := time.Minute * 5 // 默认5分钟
		if order.Status == models.PaymentStatusPaid {
			cacheTTL = time.Hour * 2 // 已支付订单缓存2小时
		}
		s.redis.Set(ctx, cacheKey, data, cacheTTL)
	}

	return order, nil
}

// updateOrderStatus 更新订单状态 - 优化版本，添加缓存失效机制
func (s *PaymentService) updateOrderStatus(ctx context.Context, orderNo, status string) error {
	query := `
		UPDATE payment_orders 
		SET status = $1, updated_at = $2
		WHERE order_no = $3
	`

	_, err := s.db.ExecContext(ctx, query, status, time.Now(), orderNo)
	if err != nil {
		return err
	}

	// 清除相关缓存
	if s.redis != nil {
		cacheKeys := []string{
			fmt.Sprintf("payment_order:%s", orderNo),
			fmt.Sprintf("payment:%s", orderNo),
		}
		for _, key := range cacheKeys {
			s.redis.Del(ctx, key)
		}
	}

	return nil
}

// saveRefundRecord 保存退款记录
func (s *PaymentService) saveRefundRecord(ctx context.Context, refund *models.PaymentRefund) error {
	query := `
		INSERT INTO refund_records (
			refund_no, order_no, amount, reason, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := s.db.ExecContext(ctx, query,
		refund.RefundNo, refund.OrderNo, refund.Amount, refund.Reason,
		refund.Status, refund.CreatedAt, refund.UpdatedAt,
	)

	return err
}

// updateRefundStatus 更新退款状态
func (s *PaymentService) updateRefundStatus(ctx context.Context, refundNo, status string) error {
	query := `
		UPDATE refund_records 
		SET status = $1, updated_at = $2
		WHERE refund_no = $3
	`

	_, err := s.db.ExecContext(ctx, query, status, time.Now(), refundNo)
	return err
}

// logCallback 记录回调日志
func (s *PaymentService) logCallback(ctx context.Context, channel, data, signature string) error {
	query := `
		INSERT INTO payment_callbacks (
			order_no, channel, callback_data, signature, created_at
		) VALUES ($1, $2, $3, $4, $5)
	`

	// 从回调数据中提取订单号（这里简化处理）
	orderNo := ""
	if data != "" {
		// 实际实现中需要根据不同渠道解析订单号
		orderNo = "unknown"
	}

	_, err := s.db.ExecContext(ctx, query, orderNo, channel, data, signature, time.Now())
	return err
}

// handlePaymentSuccess 处理支付成功
func (s *PaymentService) handlePaymentSuccess(ctx context.Context, callback *adapters.PaymentCallback) error {
	// 更新订单状态和支付信息
	query := `
		UPDATE payment_orders 
		SET status = $1, channel_trade_no = $2, paid_at = $3, updated_at = $4
		WHERE order_no = $5
	`

	_, err := s.db.ExecContext(ctx, query,
		models.PaymentStatusPaid, callback.ChannelTradeNo, callback.PaidAt, time.Now(), callback.OrderNo,
	)

	if err != nil {
		return err
	}

	// 清除缓存
	s.redis.Del(ctx, fmt.Sprintf("payment:%s", callback.OrderNo))

	return nil
}

// cachePaymentInfo 缓存支付信息
func (s *PaymentService) cachePaymentInfo(ctx context.Context, orderNo string, resp *adapters.PaymentResponse) {
	data, err := json.Marshal(resp)
	if err != nil {
		return
	}

	key := fmt.Sprintf("payment:%s", orderNo)
	s.redis.Set(ctx, key, data, time.Hour)
}

// timePtr 时间指针辅助函数
func timePtr(t time.Time) *time.Time {
	return &t
}
