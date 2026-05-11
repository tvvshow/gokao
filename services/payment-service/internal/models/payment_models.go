package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// PaymentOrder 支付订单模型
type PaymentOrder struct {
	ID             uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	OrderNo        string          `json:"order_no" gorm:"uniqueIndex;not null"`
	UserID         uuid.UUID       `json:"user_id" gorm:"type:uuid;not null;index"`
	Amount         decimal.Decimal `json:"amount" gorm:"type:decimal(10,2);not null"`
	Currency       string          `json:"currency" gorm:"default:'CNY'"`
	Subject        string          `json:"subject" gorm:"not null"`
	Description    string          `json:"description"`
	Channel        string          `json:"channel" gorm:"not null"` // wechat, alipay, unionpay
	ChannelTradeNo string          `json:"channel_trade_no" gorm:"index"`
	Status         string          `json:"status" gorm:"default:'pending'"`
	PaidAt         *time.Time      `json:"paid_at"`
	ExpiredAt      *time.Time      `json:"expired_at"`
	NotifyURL      string          `json:"notify_url"`
	ReturnURL      string          `json:"return_url"`
	ClientIP       string          `json:"client_ip"`
	Metadata       JSONB           `json:"metadata" gorm:"type:jsonb"`
	PaymentURL     string          `json:"payment_url"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	DeletedAt      gorm.DeletedAt  `json:"deleted_at" gorm:"index"`
}

// BeforeCreate GORM钩子
func (po *PaymentOrder) BeforeCreate(tx *gorm.DB) error {
	if po.ID == uuid.Nil {
		po.ID = uuid.New()
	}
	return nil
}

// PaymentRefund 退款记录模型
type PaymentRefund struct {
	ID              uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	RefundNo        string          `json:"refund_no" gorm:"uniqueIndex;not null"`
	OrderID         uuid.UUID       `json:"order_id" gorm:"type:uuid;not null;index"`
	OrderNo         string          `json:"order_no" gorm:"not null;index"`
	Amount          decimal.Decimal `json:"amount" gorm:"type:decimal(10,2);not null"`
	RefundAmount    decimal.Decimal `json:"refund_amount" gorm:"type:decimal(10,2);not null"`
	Reason          string          `json:"reason"`
	Channel         string          `json:"channel" gorm:"not null"`
	ChannelRefundNo string          `json:"channel_refund_no" gorm:"index"`
	Status          string          `json:"status" gorm:"default:'processing'"`
	RefundedAt      *time.Time      `json:"refunded_at"`
	NotifyURL       string          `json:"notify_url"`
	Metadata        JSONB           `json:"metadata" gorm:"type:jsonb"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	DeletedAt       gorm.DeletedAt  `json:"deleted_at" gorm:"index"`
}

// BeforeCreate GORM钩子
func (pr *PaymentRefund) BeforeCreate(tx *gorm.DB) error {
	if pr.ID == uuid.Nil {
		pr.ID = uuid.New()
	}
	return nil
}

// PaymentNotify 支付通知记录
type PaymentNotify struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	OrderNo   string         `json:"order_no" gorm:"not null;index"`
	Channel   string         `json:"channel" gorm:"not null"`
	Type      string         `json:"type" gorm:"not null"` // payment, refund
	Status    string         `json:"status" gorm:"not null"`
	RawData   string         `json:"raw_data" gorm:"type:text"`
	Signature string         `json:"signature"`
	Verified  bool           `json:"verified" gorm:"default:false"`
	Processed bool           `json:"processed" gorm:"default:false"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// BeforeCreate GORM钩子
func (pn *PaymentNotify) BeforeCreate(tx *gorm.DB) error {
	if pn.ID == uuid.Nil {
		pn.ID = uuid.New()
	}
	return nil
}

// 支付状态常量
const (
	PaymentStatusPending  = "pending"  // 待支付
	PaymentStatusPaid     = "paid"     // 已支付
	PaymentStatusFailed   = "failed"   // 支付失败
	PaymentStatusCanceled = "canceled" // 已取消
	PaymentStatusRefunded = "refunded" // 已退款
	PaymentStatusExpired  = "expired"  // 已过期
)

// 退款状态常量
const (
	RefundStatusProcessing = "processing" // 处理中
	RefundStatusSuccess    = "success"    // 退款成功
	RefundStatusFailed     = "failed"     // 退款失败
	RefundStatusCanceled   = "canceled"   // 已取消
)

// 支付渠道常量
const (
	ChannelWechat   = "wechat"
	ChannelAlipay   = "alipay"
	ChannelUnionpay = "unionpay"
	ChannelQQ       = "qq"
)

// 通知类型常量
const (
	NotifyTypePayment = "payment"
	NotifyTypeRefund  = "refund"
)

// IsPaid 检查订单是否已支付
func (po *PaymentOrder) IsPaid() bool {
	return po.Status == PaymentStatusPaid
}

// IsExpired 检查订单是否已过期
func (po *PaymentOrder) IsExpired() bool {
	if po.ExpiredAt == nil {
		return false
	}
	return time.Now().After(*po.ExpiredAt)
}

// CanRefund 检查订单是否可以退款
func (po *PaymentOrder) CanRefund() bool {
	return po.Status == PaymentStatusPaid
}

// IsRefundSuccess 检查退款是否成功
func (pr *PaymentRefund) IsRefundSuccess() bool {
	return pr.Status == RefundStatusSuccess
}

// IsProcessing 检查退款是否处理中
func (pr *PaymentRefund) IsProcessing() bool {
	return pr.Status == RefundStatusProcessing
}

// PaymentOrderCreateRequest 创建支付订单请求
type PaymentOrderCreateRequest struct {
	Amount        decimal.Decimal        `json:"amount" validate:"required,gt=0"`
	Subject       string                 `json:"subject" validate:"required"`
	Description   string                 `json:"description"`
	Channel       string                 `json:"channel" validate:"required,oneof=wechat alipay unionpay qq"`
	NotifyURL     string                 `json:"notify_url"`
	ReturnURL     string                 `json:"return_url"`
	ExpireMinutes int                    `json:"expire_minutes" validate:"min=1,max=1440"` // 1分钟到24小时
	Metadata      map[string]interface{} `json:"metadata"`
}

// PaymentOrderResponse 支付订单响应
type PaymentOrderResponse struct {
	ID         uuid.UUID       `json:"id"`
	OrderNo    string          `json:"order_no"`
	Amount     decimal.Decimal `json:"amount"`
	Currency   string          `json:"currency"`
	Subject    string          `json:"subject"`
	Channel    string          `json:"channel"`
	Status     string          `json:"status"`
	PaymentURL string          `json:"payment_url,omitempty"`
	QRCode     string          `json:"qr_code,omitempty"`
	ExpiredAt  *time.Time      `json:"expired_at"`
	CreatedAt  time.Time       `json:"created_at"`
}

// PaymentRefundCreateRequest 创建退款请求
type PaymentRefundCreateRequest struct {
	OrderNo      string          `json:"order_no" validate:"required"`
	RefundAmount decimal.Decimal `json:"refund_amount" validate:"required,gt=0"`
	Reason       string          `json:"reason" validate:"required"`
	NotifyURL    string          `json:"notify_url"`
}

// PaymentRefundResponse 退款响应
type PaymentRefundResponse struct {
	ID           uuid.UUID       `json:"id"`
	RefundNo     string          `json:"refund_no"`
	OrderNo      string          `json:"order_no"`
	Amount       decimal.Decimal `json:"amount"`
	RefundAmount decimal.Decimal `json:"refund_amount"`
	Reason       string          `json:"reason"`
	Status       string          `json:"status"`
	RefundedAt   *time.Time      `json:"refunded_at"`
	CreatedAt    time.Time       `json:"created_at"`
}

// PaymentNotifyResponse 支付通知响应
type PaymentNotifyResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// PaymentStatus 支付状态查询响应
type PaymentStatus struct {
	OrderNo       string          `json:"order_no"`
	Status        string          `json:"status"`
	PaymentMethod string          `json:"payment_method"`
	TransactionID string          `json:"transaction_id"`
	PaidAt        *time.Time      `json:"paid_at"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	Extra         JSONB           `json:"extra"`
}

// RefundRequest 退款请求
type RefundRequest struct {
	OrderNo  string          `json:"order_no"`
	RefundID string          `json:"refund_id"`
	Amount   decimal.Decimal `json:"amount"`
	Reason   string          `json:"reason"`
}

// RefundResponse 退款响应
type RefundResponse struct {
	RefundID      string          `json:"refund_id"`
	OrderNo       string          `json:"order_no"`
	Status        string          `json:"status"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	RefundedAt    time.Time       `json:"refunded_at"`
	PaymentMethod string          `json:"payment_method"`
	Extra         JSONB           `json:"extra"`
}

// CallbackResult 回调处理结果
type CallbackResult struct {
	OrderNo       string          `json:"order_no"`
	Status        string          `json:"status"`
	PaymentMethod string          `json:"payment_method"`
	TransactionID string          `json:"transaction_id"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	PaidAt        *time.Time      `json:"paid_at"`
	Extra         JSONB           `json:"extra"`
}

// CreatePaymentRequest 创建支付请求
type CreatePaymentRequest struct {
	UserID        string                 `json:"user_id"`
	Amount        decimal.Decimal        `json:"amount"`
	Currency      string                 `json:"currency"`
	Description   string                 `json:"description"`
	PaymentMethod string                 `json:"payment_method"`
	Extra         map[string]interface{} `json:"extra"`
}

// APIResponse 通用API响应
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ValidateChannel 验证支付渠道
func ValidateChannel(channel string) bool {
	validChannels := []string{
		ChannelWechat,
		ChannelAlipay,
		ChannelUnionpay,
		ChannelQQ,
	}

	for _, validChannel := range validChannels {
		if channel == validChannel {
			return true
		}
	}
	return false
}

// ValidatePaymentStatus 验证支付状态
func ValidatePaymentStatus(status string) bool {
	validStatuses := []string{
		PaymentStatusPending,
		PaymentStatusPaid,
		PaymentStatusFailed,
		PaymentStatusCanceled,
		PaymentStatusRefunded,
		PaymentStatusExpired,
	}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// ValidateRefundStatus 验证退款状态
func ValidateRefundStatus(status string) bool {
	validStatuses := []string{
		RefundStatusProcessing,
		RefundStatusSuccess,
		RefundStatusFailed,
		RefundStatusCanceled,
	}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}
