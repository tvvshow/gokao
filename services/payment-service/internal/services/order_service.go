package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/gaokaohub/payment-service/internal/models"
)

// OrderService 订单服务
type OrderService struct {
	db    *sql.DB
	redis *redis.Client
}

// NewOrderService 创建订单服务
func NewOrderService(db *sql.DB, redis *redis.Client) *OrderService {
	return &OrderService{
		db:    db,
		redis: redis,
	}
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(ctx context.Context, userID string, req *models.CreateOrderRequest) (*models.CreateOrderResponse, error) {
	// 获取会员套餐信息
	plan, err := s.getMembershipPlan(ctx, req.PlanCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get membership plan: %w", err)
	}

	if !plan.IsActive {
		return nil, fmt.Errorf("membership plan is not active")
	}

	// 生成订单号
	orderNo := generateOrderNo()

	// 计算过期时间（30分钟）
	expireTime := time.Now().Add(30 * time.Minute)

	// 创建订单
	order := &models.PaymentOrder{
		OrderNo:        orderNo,
		UserID:         userID,
		Amount:         plan.Price,
		Currency:       "CNY",
		Subject:        fmt.Sprintf("会员套餐-%s", plan.Name),
		Description:    plan.Description,
		Status:         models.PaymentStatusPending,
		PaymentChannel: req.PaymentChannel,
		ExpireTime:     &expireTime,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Metadata: models.JSONB{
			"plan_code":  req.PlanCode,
			"auto_renew": req.AutoRenew,
			"device_id":  req.DeviceID,
			"custom":     req.Metadata,
		},
	}

	// 保存订单
	if err := s.saveOrder(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	return &models.CreateOrderResponse{
		OrderNo:   orderNo,
		Amount:    plan.Price,
		ExpiredAt: expireTime,
	}, nil
}

// GetOrders 获取订单列表
func (s *OrderService) GetOrders(ctx context.Context, userID string, req *models.OrderListRequest) (*models.OrderListResponse, error) {
	// 构建查询条件
	whereClause := "WHERE user_id = $1"
	args := []interface{}{userID}
	argIndex := 2

	if req.Status != "" {
		whereClause += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, req.Status)
		argIndex++
	}

	if req.StartTime != "" {
		whereClause += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, req.StartTime)
		argIndex++
	}

	if req.EndTime != "" {
		whereClause += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, req.EndTime)
		argIndex++
	}

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM payment_orders %s", whereClause)
	var total int64
	err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count orders: %w", err)
	}

	// 查询订单列表
	offset := (req.Page - 1) * req.PageSize
	listQuery := fmt.Sprintf(`
		SELECT id, order_no, user_id, amount, currency, subject, description,
		       status, payment_channel, channel_trade_no, client_ip, notify_url,
		       return_url, expire_time, paid_at, created_at, updated_at, metadata
		FROM payment_orders %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, req.PageSize, offset)

	rows, err := s.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	var orders []*models.PaymentOrder
	for rows.Next() {
		order := &models.PaymentOrder{}
		err := rows.Scan(
			&order.ID, &order.OrderNo, &order.UserID, &order.Amount,
			&order.Currency, &order.Subject, &order.Description, &order.Status,
			&order.PaymentChannel, &order.ChannelTradeNo, &order.ClientIP,
			&order.NotifyURL, &order.ReturnURL, &order.ExpireTime,
			&order.PaidAt, &order.CreatedAt, &order.UpdatedAt, &order.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	return &models.OrderListResponse{
		Total:  total,
		Orders: orders,
	}, nil
}

// GetOrder 获取单个订单
func (s *OrderService) GetOrder(ctx context.Context, userID, orderNo string) (*models.PaymentOrder, error) {
	query := `
		SELECT id, order_no, user_id, amount, currency, subject, description,
		       status, payment_channel, channel_trade_no, client_ip, notify_url,
		       return_url, expire_time, paid_at, created_at, updated_at, metadata
		FROM payment_orders
		WHERE order_no = $1 AND user_id = $2
	`

	order := &models.PaymentOrder{}
	err := s.db.QueryRowContext(ctx, query, orderNo, userID).Scan(
		&order.ID, &order.OrderNo, &order.UserID, &order.Amount,
		&order.Currency, &order.Subject, &order.Description, &order.Status,
		&order.PaymentChannel, &order.ChannelTradeNo, &order.ClientIP,
		&order.NotifyURL, &order.ReturnURL, &order.ExpireTime,
		&order.PaidAt, &order.CreatedAt, &order.UpdatedAt, &order.Metadata,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return order, nil
}

// CancelOrder 取消订单
func (s *OrderService) CancelOrder(ctx context.Context, userID, orderNo string) error {
	// 检查订单是否存在且属于该用户
	order, err := s.GetOrder(ctx, userID, orderNo)
	if err != nil {
		return err
	}

	// 只有待支付的订单可以取消
	if order.Status != models.PaymentStatusPending {
		return fmt.Errorf("order cannot be canceled, status: %s", order.Status)
	}

	// 更新订单状态
	query := `
		UPDATE payment_orders 
		SET status = $1, updated_at = $2
		WHERE order_no = $3 AND user_id = $4
	`

	_, err = s.db.ExecContext(ctx, query, models.PaymentStatusCanceled, time.Now(), orderNo, userID)
	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	return nil
}

// GetInvoice 获取发票信息
func (s *OrderService) GetInvoice(ctx context.Context, userID, orderNo string) (map[string]interface{}, error) {
	// 检查订单是否存在且属于该用户
	order, err := s.GetOrder(ctx, userID, orderNo)
	if err != nil {
		return nil, err
	}

	// 只有已支付的订单可以开发票
	if order.Status != models.PaymentStatusPaid {
		return nil, fmt.Errorf("order is not paid, cannot generate invoice")
	}

	// 生成发票信息
	invoice := map[string]interface{}{
		"order_no":     order.OrderNo,
		"amount":       order.Amount,
		"subject":      order.Subject,
		"description":  order.Description,
		"paid_at":      order.PaidAt,
		"invoice_no":   generateInvoiceNo(),
		"invoice_date": time.Now(),
		"tax_rate":     decimal.NewFromFloat(0.06), // 6%税率
		"tax_amount":   order.Amount.Mul(decimal.NewFromFloat(0.06)),
	}

	return invoice, nil
}

// getMembershipPlan 获取会员套餐
func (s *OrderService) getMembershipPlan(ctx context.Context, planCode string) (*models.MembershipPlan, error) {
	query := `
		SELECT id, plan_code, name, description, price, duration_days,
		       features, max_queries, max_downloads, is_active,
		       created_at, updated_at
		FROM membership_plans
		WHERE plan_code = $1
	`

	plan := &models.MembershipPlan{}
	err := s.db.QueryRowContext(ctx, query, planCode).Scan(
		&plan.ID, &plan.PlanCode, &plan.Name, &plan.Description,
		&plan.Price, &plan.DurationDays, &plan.Features,
		&plan.MaxQueries, &plan.MaxDownloads, &plan.IsActive,
		&plan.CreatedAt, &plan.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("membership plan not found")
		}
		return nil, err
	}

	return plan, nil
}

// saveOrder 保存订单
func (s *OrderService) saveOrder(ctx context.Context, order *models.PaymentOrder) error {
	query := `
		INSERT INTO payment_orders (
			order_no, user_id, amount, currency, subject, description,
			status, payment_channel, expire_time, created_at, updated_at, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := s.db.ExecContext(ctx, query,
		order.OrderNo, order.UserID, order.Amount, order.Currency,
		order.Subject, order.Description, order.Status, order.PaymentChannel,
		order.ExpireTime, order.CreatedAt, order.UpdatedAt, order.Metadata,
	)

	return err
}

// generateOrderNo 生成订单号
func generateOrderNo() string {
	now := time.Now()
	return fmt.Sprintf("ORDER%s%s", 
		now.Format("20060102150405"),
		uuid.New().String()[:8],
	)
}

// generateInvoiceNo 生成发票号
func generateInvoiceNo() string {
	now := time.Now()
	return fmt.Sprintf("INV%s%s",
		now.Format("20060102"),
		uuid.New().String()[:8],
	)
}