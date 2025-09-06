package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gaokaohub/gaokao/services/payment-service/internal/models"
)

// PaymentRepository 支付数据访问接口
type PaymentRepository interface {
	CreatePayment(ctx context.Context, payment *models.PaymentOrder) error
	CreateRefund(ctx context.Context, refund *models.RefundRecord) error
	GetPaymentByID(ctx context.Context, paymentID string) (*models.PaymentOrder, error)
	GetPaymentByOrderID(ctx context.Context, orderID string) (*models.PaymentOrder, error)
	GetPaymentByIDWithLock(ctx context.Context, paymentID string) (*models.PaymentOrder, error)
	UpdatePaymentStatus(ctx context.Context, paymentID, status, tradeNo string) error
	UpdatePaymentAmount(ctx context.Context, paymentID string, amount float64) error
	ListPayments(ctx context.Context, filter *models.PaymentFilter) ([]*models.PaymentOrder, int64, error)
	ClosePayment(ctx context.Context, paymentID string) error
	GetPaymentStatistics(ctx context.Context, startDate, endDate time.Time, channel string) (*models.PaymentStatistics, error)
	
	// 事务相关方法
	BeginTx(ctx context.Context) (*sql.Tx, error)
	WithTx(tx *sql.Tx) PaymentRepository
}

// paymentRepository 支付数据访问实现
type paymentRepository struct {
	db *sql.DB
	tx *sql.Tx // 当前事务，如果存在
}

// NewPaymentRepository 创建支付数据访问实例
func NewPaymentRepository(db *sql.DB) PaymentRepository {
	return &paymentRepository{
		db: db,
	}
}

// BeginTx 开始事务
func (r *paymentRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	if r.tx != nil {
		return nil, fmt.Errorf("nested transactions not supported")
	}
	
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	return tx, nil
}

// WithTx 使用指定的事务
func (r *paymentRepository) WithTx(tx *sql.Tx) PaymentRepository {
	return &paymentRepository{
		db: r.db,
		tx: tx,
	}
}

// CreatePayment 创建支付记录
func (r *paymentRepository) CreatePayment(ctx context.Context, payment *models.PaymentOrder) error {
	query := `
		INSERT INTO payment_orders (
			order_no, user_id, amount, currency, subject, description,
			channel, channel_trade_no, status, expired_at,
			notify_url, return_url, client_ip, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id
	`
	
	payment.CreatedAt = time.Now()
	payment.UpdatedAt = time.Now()
	
	var db *sql.DB
	if r.tx != nil {
		db = nil // 使用事务
	} else {
		db = r.db
	}
	
	var id int64
	err := db.QueryRowContext(ctx, query,
		payment.OrderNo,
		payment.UserID,
		payment.Amount,
		payment.Currency,
		payment.Subject,
		payment.Description,
		payment.Channel,
		payment.ChannelTradeNo,
		payment.Status,
		payment.ExpiredAt,
		payment.NotifyURL,
		payment.ReturnURL,
		payment.ClientIP,
		payment.Metadata,
		payment.CreatedAt,
		payment.UpdatedAt,
	).Scan(&id)
	
	if err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}
	
	// 注意：由于数据库使用SERIAL类型而模型使用UUID类型，这里我们不设置payment.ID
	return nil
}

// CreateRefund 创建退款记录
func (r *paymentRepository) CreateRefund(ctx context.Context, refund *models.RefundRecord) error {
	query := `
		INSERT INTO refund_records (
			refund_no, order_no, channel_refund_no, amount, reason, status, refunded_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`
	
	refund.CreatedAt = time.Now()
	refund.UpdatedAt = time.Now()
	
	var db *sql.DB
	if r.tx != nil {
		db = nil // 使用事务
	} else {
		db = r.db
	}
	
	var id int64
	err := db.QueryRowContext(ctx, query,
		refund.RefundNo,
		refund.OrderNo,
		refund.ChannelRefundNo,
		refund.Amount,
		refund.Reason,
		refund.Status,
		refund.RefundedAt,
		refund.CreatedAt,
		refund.UpdatedAt,
	).Scan(&id)
	
	if err != nil {
		return fmt.Errorf("failed to create refund: %w", err)
	}
	
	refund.ID = id
	return nil
}

// GetPaymentByID 根据支付ID获取支付记录
func (r *paymentRepository) GetPaymentByID(ctx context.Context, paymentID string) (*models.PaymentOrder, error) {
	query := `
		SELECT id, payment_id, user_id, order_id, amount, product_type, product_id,
			   channel, subject, body, status, trade_no, notify_data,
			   created_at, updated_at, completed_at, closed_at
		FROM payments 
		WHERE payment_id = $1 AND deleted_at IS NULL
	`
	
	var db *sql.DB
	if r.tx != nil {
		db = nil // 使用事务
	} else {
		db = r.db
	}
	
	row := db.QueryRowContext(ctx, query, paymentID)
	
	var payment models.PaymentOrder
	var paidAt, expiredAt sql.NullTime
	var metadata []byte
	
	err := row.Scan(
		&payment.ID,
		&payment.OrderNo,
		&payment.UserID,
		&payment.Amount,
		&payment.Currency,
		&payment.Subject,
		&payment.Description,
		&payment.Channel,
		&payment.ChannelTradeNo,
		&payment.Status,
		&paidAt,
		&expiredAt,
		&payment.NotifyURL,
		&payment.ReturnURL,
		&payment.ClientIP,
		&metadata,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get payment by ID: %w", err)
	}
	
	if paidAt.Valid {
		payment.PaidAt = &paidAt.Time
	}
	if expiredAt.Valid {
		payment.ExpiredAt = &expiredAt.Time
	}
	if len(metadata) > 0 {
		if err := json.Unmarshal(metadata, &payment.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}
	
	return &payment, nil
}

// GetPaymentByIDWithLock 根据支付ID获取支付记录（带行级锁）
func (r *paymentRepository) GetPaymentByIDWithLock(ctx context.Context, paymentID string) (*models.PaymentOrder, error) {
	query := `
		SELECT id, order_no, user_id, amount, currency, subject, description,
		       channel, channel_trade_no, status, paid_at, expired_at,
		       notify_url, return_url, client_ip, metadata, created_at, updated_at
		FROM payment_orders 
		WHERE id = $1 AND deleted_at IS NULL
		FOR UPDATE
	`
	
	var db *sql.DB
	if r.tx != nil {
		db = nil // 使用事务
	} else {
		db = r.db
	}
	
	row := db.QueryRowContext(ctx, query, paymentID)
	
	var payment models.PaymentOrder
	var paidAt, expiredAt sql.NullTime
	var metadata []byte
	
	err := row.Scan(
		&payment.ID,
		&payment.OrderNo,
		&payment.UserID,
		&payment.Amount,
		&payment.Currency,
		&payment.Subject,
		&payment.Description,
		&payment.Channel,
		&payment.ChannelTradeNo,
		&payment.Status,
		&paidAt,
		&expiredAt,
		&payment.NotifyURL,
		&payment.ReturnURL,
		&payment.ClientIP,
		&metadata,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get payment by ID with lock: %w", err)
	}
	
	if paidAt.Valid {
		payment.PaidAt = &paidAt.Time
	}
	if expiredAt.Valid {
		payment.ExpiredAt = &expiredAt.Time
	}
	if len(metadata) > 0 {
		if err := json.Unmarshal(metadata, &payment.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}
	
	return &payment, nil
}

// GetPaymentByOrderID 根据订单ID获取支付记录
func (r *paymentRepository) GetPaymentByOrderID(ctx context.Context, orderID string) (*models.PaymentOrder, error) {
	query := `
		SELECT id, order_no, user_id, amount, currency, subject, description,
		       channel, channel_trade_no, status, paid_at, expired_at,
		       notify_url, return_url, client_ip, metadata, created_at, updated_at
		FROM payment_orders 
		WHERE order_no = $1 AND deleted_at IS NULL
	`
	
	var db *sql.DB
	if r.tx != nil {
		db = nil // 使用事务
	} else {
		db = r.db
	}
	
	row := db.QueryRowContext(ctx, query, orderID)
	
	var payment models.PaymentOrder
	var paidAt, expiredAt sql.NullTime
	var metadata []byte
	
	err := row.Scan(
		&payment.ID,
		&payment.OrderNo,
		&payment.UserID,
		&payment.Amount,
		&payment.Currency,
		&payment.Subject,
		&payment.Description,
		&payment.Channel,
		&payment.ChannelTradeNo,
		&payment.Status,
		&paidAt,
		&expiredAt,
		&payment.NotifyURL,
		&payment.ReturnURL,
		&payment.ClientIP,
		&metadata,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get payment by order ID: %w", err)
	}
	
	if paidAt.Valid {
		payment.PaidAt = &paidAt.Time
	}
	if expiredAt.Valid {
		payment.ExpiredAt = &expiredAt.Time
	}
	if len(metadata) > 0 {
		if err := json.Unmarshal(metadata, &payment.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}
	
	return &payment, nil
}

// UpdatePaymentStatus 更新支付状态（带事务支持和行锁）
func (r *paymentRepository) UpdatePaymentStatus(ctx context.Context, paymentID, status, tradeNo string) error {
	query := `
		UPDATE payments 
		SET status = $1, trade_no = $2, updated_at = $3,
			completed_at = CASE WHEN $1 = 'success' THEN $3 ELSE completed_at END
		WHERE payment_id = $4 AND deleted_at IS NULL
	`
	
	var db *sql.DB
	if r.tx != nil {
		db = nil // 使用事务
	} else {
		db = r.db
	}
	
	result, err := db.ExecContext(ctx, query, status, tradeNo, time.Now(), paymentID)
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("payment not found or already updated")
	}
	
	return nil
}

// UpdatePaymentAmount 更新支付金额
func (r *paymentRepository) UpdatePaymentAmount(ctx context.Context, paymentID string, amount float64) error {
	query := `
		UPDATE payments 
		SET amount = $1, updated_at = $2
		WHERE payment_id = $3 AND deleted_at IS NULL
	`
	
	var db *sql.DB
	if r.tx != nil {
		db = nil // 使用事务
	} else {
		db = r.db
	}
	
	result, err := db.ExecContext(ctx, query, amount, time.Now(), paymentID)
	if err != nil {
		return fmt.Errorf("failed to update payment amount: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("payment not found")
	}
	
	return nil
}

// ListPayments 获取支付记录列表
func (r *paymentRepository) ListPayments(ctx context.Context, filter *models.PaymentFilter) ([]*models.PaymentOrder, int64, error) {
	baseQuery := `
		SELECT id, order_no, user_id, amount, currency, subject, description,
		       channel, channel_trade_no, status, paid_at, expired_at,
		       notify_url, return_url, client_ip, metadata, created_at, updated_at
		FROM payment_orders 
		WHERE deleted_at IS NULL
	`
	
	countQuery := `SELECT COUNT(*) FROM payment_orders WHERE deleted_at IS NULL`
	
	var args []interface{}
	var whereClauses []string
	
	// 构建查询条件
	if filter.UserID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("user_id = $%d", len(args)+1))
		args = append(args, filter.UserID)
	}
	
	if filter.Status != nil && *filter.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", len(args)+1))
		args = append(args, *filter.Status)
	}
	
	if filter.Channel != nil && *filter.Channel != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("channel = $%d", len(args)+1))
		args = append(args, *filter.Channel)
	}
	
	if filter.StartTime != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("created_at >= $%d", len(args)+1))
		args = append(args, filter.StartTime)
	}
	
	if filter.EndTime != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("created_at <= $%d", len(args)+1))
		args = append(args, filter.EndTime)
	}
	
	// 添加WHERE条件
	if len(whereClauses) > 0 {
		baseQuery += " AND " + strings.Join(whereClauses, " AND ")
		countQuery += " AND " + strings.Join(whereClauses, " AND ")
	}
	
	// 获取总数
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count payments: %w", err)
	}
	
	// 添加排序和分页
	baseQuery += " ORDER BY created_at DESC"
	
	if filter.PageSize > 0 {
		baseQuery += fmt.Sprintf(" LIMIT $%d", len(args)+1)
		args = append(args, filter.PageSize)
		
		if filter.Page > 0 {
			baseQuery += fmt.Sprintf(" OFFSET $%d", len(args)+1)
			args = append(args, (filter.Page-1)*filter.PageSize)
		}
	}
	
	// 执行查询
	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query payments: %w", err)
	}
	defer rows.Close()
	
	var payments []*models.PaymentOrder
	for rows.Next() {
		var payment models.PaymentOrder
		var paidAt, expiredAt sql.NullTime
		var metadata []byte
		
		err := rows.Scan(
			&payment.ID,
			&payment.OrderNo,
			&payment.UserID,
			&payment.Amount,
			&payment.Currency,
			&payment.Subject,
			&payment.Description,
			&payment.Channel,
			&payment.ChannelTradeNo,
			&payment.Status,
			&paidAt,
			&expiredAt,
			&payment.NotifyURL,
			&payment.ReturnURL,
			&payment.ClientIP,
			&metadata,
			&payment.CreatedAt,
			&payment.UpdatedAt,
		)
		
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan payment: %w", err)
		}
		
		if paidAt.Valid {
			payment.PaidAt = &paidAt.Time
		}
		if expiredAt.Valid {
			payment.ExpiredAt = &expiredAt.Time
		}
		if len(metadata) > 0 {
			if err := json.Unmarshal(metadata, &payment.Metadata); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}
		
		payments = append(payments, &payment)
	}
	
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}
	
	return payments, total, nil
}

// ClosePayment 关闭支付订单
func (r *paymentRepository) ClosePayment(ctx context.Context, paymentID string) error {
	query := `
		UPDATE payments 
		SET status = 'closed', closed_at = $1, updated_at = $1
		WHERE payment_id = $2 AND status = 'pending' AND deleted_at IS NULL
	`
	
	var db *sql.DB
	if r.tx != nil {
		db = nil // 使用事务
	} else {
		db = r.db
	}
	
	result, err := db.ExecContext(ctx, query, time.Now(), paymentID)
	if err != nil {
		return fmt.Errorf("failed to close payment: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("payment not found or not in pending status")
	}
	
	return nil
}

// GetPaymentStatistics 获取支付统计信息
func (r *paymentRepository) GetPaymentStatistics(ctx context.Context, startDate, endDate time.Time, channel string) (*models.PaymentStatistics, error) {
	query := `
		SELECT 
			COUNT(*) as total_count,
			COUNT(CASE WHEN status = 'paid' THEN 1 END) as success_count,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_count,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_count,
			COUNT(CASE WHEN status = 'canceled' THEN 1 END) as canceled_count,
			COUNT(CASE WHEN status = 'refunded' THEN 1 END) as refunded_count,
			COUNT(CASE WHEN status = 'expired' THEN 1 END) as expired_count,
			COALESCE(SUM(CASE WHEN status = 'paid' THEN amount ELSE 0 END), 0) as total_amount,
			COALESCE(AVG(CASE WHEN status = 'paid' THEN amount ELSE NULL END), 0) as avg_amount,
			COUNT(DISTINCT user_id) as unique_users,
			COUNT(DISTINCT channel) as unique_channels
		FROM payment_orders 
		WHERE created_at BETWEEN $1 AND $2 AND deleted_at IS NULL
	`
	
	var args []interface{}
	args = append(args, startDate, endDate)
	
	if channel != "" {
		query += " AND channel = $3"
		args = append(args, channel)
	}
	
	var stats models.PaymentStatistics
	
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&stats.TotalCount,
		&stats.SuccessCount,
		&stats.FailedCount,
		&stats.PendingCount,
		&stats.CanceledCount,
		&stats.RefundedCount,
		&stats.ExpiredCount,
		&stats.TotalAmount,
		&stats.AvgAmount,
		&stats.UniqueUsers,
		&stats.UniqueChannels,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get payment statistics: %w", err)
	}
	
	return &stats, nil
}