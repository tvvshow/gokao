package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// JSONB 自定义JSONB类型
type JSONB map[string]interface{}

// Value 实现driver.Valuer接口
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现sql.Scanner接口
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSONB", value)
	}

	if len(bytes) == 0 {
		*j = nil
		return nil
	}

	return json.Unmarshal(bytes, j)
}

// PaymentOrder 支付订单
type PaymentOrder struct {
	ID               int64           `json:"id" db:"id"`
	OrderNo          string          `json:"order_no" db:"order_no"`
	UserID           string          `json:"user_id" db:"user_id"`
	Amount           decimal.Decimal `json:"amount" db:"amount"`
	Currency         string          `json:"currency" db:"currency"`
	Subject          string          `json:"subject" db:"subject"`
	Description      string          `json:"description" db:"description"`
	Status           string          `json:"status" db:"status"`
	PaymentChannel   string          `json:"payment_channel" db:"payment_channel"`
	ChannelTradeNo   string          `json:"channel_trade_no" db:"channel_trade_no"`
	ClientIP         string          `json:"client_ip" db:"client_ip"`
	NotifyURL        string          `json:"notify_url" db:"notify_url"`
	ReturnURL        string          `json:"return_url" db:"return_url"`
	ExpireTime       *time.Time      `json:"expire_time" db:"expire_time"`
	PaidAt           *time.Time      `json:"paid_at" db:"paid_at"`
	CreatedAt        time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at" db:"updated_at"`
	Metadata         JSONB           `json:"metadata" db:"metadata"`
}

// RefundRecord 退款记录
type RefundRecord struct {
	ID               int64           `json:"id" db:"id"`
	RefundNo         string          `json:"refund_no" db:"refund_no"`
	OrderNo          string          `json:"order_no" db:"order_no"`
	ChannelRefundNo  string          `json:"channel_refund_no" db:"channel_refund_no"`
	Amount           decimal.Decimal `json:"amount" db:"amount"`
	Reason           string          `json:"reason" db:"reason"`
	Status           string          `json:"status" db:"status"`
	RefundedAt       *time.Time      `json:"refunded_at" db:"refunded_at"`
	CreatedAt        time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at" db:"updated_at"`
}

// MembershipPlan 会员套餐
type MembershipPlan struct {
	ID           int64     `json:"id" db:"id"`
	PlanCode     string    `json:"plan_code" db:"plan_code"`
	Name         string    `json:"name" db:"name"`
	Description  string    `json:"description" db:"description"`
	Price        decimal.Decimal `json:"price" db:"price"`
	DurationDays int       `json:"duration_days" db:"duration_days"`
	Features     JSONB     `json:"features" db:"features"`
	MaxQueries   int       `json:"max_queries" db:"max_queries"`
	MaxDownloads int       `json:"max_downloads" db:"max_downloads"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// UserMembership 用户会员
type UserMembership struct {
	ID            int64     `json:"id" db:"id"`
	UserID        string    `json:"user_id" db:"user_id"`
	PlanCode      string    `json:"plan_code" db:"plan_code"`
	OrderNo       string    `json:"order_no" db:"order_no"`
	StartTime     time.Time `json:"start_time" db:"start_time"`
	EndTime       time.Time `json:"end_time" db:"end_time"`
	Status        string    `json:"status" db:"status"`
	AutoRenew     bool      `json:"auto_renew" db:"auto_renew"`
	UsedQueries   int       `json:"used_queries" db:"used_queries"`
	UsedDownloads int       `json:"used_downloads" db:"used_downloads"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`

	// 关联数据
	Plan *MembershipPlan `json:"plan,omitempty"`
}

// PaymentCallback 支付回调
type PaymentCallback struct {
	ID           int64     `json:"id" db:"id"`
	OrderNo      string    `json:"order_no" db:"order_no"`
	Channel      string    `json:"channel" db:"channel"`
	CallbackData string    `json:"callback_data" db:"callback_data"`
	Signature    string    `json:"signature" db:"signature"`
	Verified     bool      `json:"verified" db:"verified"`
	Processed    bool      `json:"processed" db:"processed"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// LicenseInfo 许可证信息
type LicenseInfo struct {
	ID                int64     `json:"id" db:"id"`
	UserID            string    `json:"user_id" db:"user_id"`
	DeviceID          string    `json:"device_id" db:"device_id"`
	DeviceFingerprint string    `json:"device_fingerprint" db:"device_fingerprint"`
	LicenseKey        string    `json:"license_key" db:"license_key"`
	EncryptedData     string    `json:"encrypted_data" db:"encrypted_data"`
	ExpiresAt         time.Time `json:"expires_at" db:"expires_at"`
	Status            string    `json:"status" db:"status"`
	BindCount         int       `json:"bind_count" db:"bind_count"`
	MaxBindCount      int       `json:"max_bind_count" db:"max_bind_count"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	PlanCode        string            `json:"plan_code" binding:"required"`
	PaymentChannel  string            `json:"payment_channel" binding:"required"`
	ReturnURL       string            `json:"return_url"`
	AutoRenew       bool              `json:"auto_renew"`
	DeviceID        string            `json:"device_id"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// CreateOrderResponse 创建订单响应
type CreateOrderResponse struct {
	OrderNo        string    `json:"order_no"`
	Amount         decimal.Decimal `json:"amount"`
	PayURL         string    `json:"pay_url"`
	QRCode         string    `json:"qr_code"`
	FormData       string    `json:"form_data"`
	ExpiredAt      time.Time `json:"expired_at"`
}

// OrderListRequest 订单列表请求
type OrderListRequest struct {
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
	Status   string `form:"status"`
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
}

// OrderListResponse 订单列表响应
type OrderListResponse struct {
	Total  int64           `json:"total"`
	Orders []*PaymentOrder `json:"orders"`
}

// MembershipStatusResponse 会员状态响应
type MembershipStatusResponse struct {
	IsVIP          bool               `json:"is_vip"`
	PlanCode       string             `json:"plan_code"`
	PlanName       string             `json:"plan_name"`
	StartTime      *time.Time         `json:"start_time"`
	EndTime        *time.Time         `json:"end_time"`
	RemainingDays  int                `json:"remaining_days"`
	UsedQueries    int                `json:"used_queries"`
	MaxQueries     int                `json:"max_queries"`
	UsedDownloads  int                `json:"used_downloads"`
	MaxDownloads   int                `json:"max_downloads"`
	Features       map[string]interface{} `json:"features"`
	AutoRenew      bool               `json:"auto_renew"`
}

// PaymentStatusConstants 支付状态常量
const (
	PaymentStatusPending   = "pending"   // 待支付
	PaymentStatusPaid      = "paid"      // 已支付
	PaymentStatusCanceled  = "canceled"  // 已取消
	PaymentStatusExpired   = "expired"   // 已过期
	PaymentStatusRefunded  = "refunded"  // 已退款
	PaymentStatusRefunding = "refunding" // 退款中
)

// RefundStatusConstants 退款状态常量
const (
	RefundStatusPending   = "pending"   // 退款中
	RefundStatusSuccess   = "success"   // 退款成功
	RefundStatusFailed    = "failed"    // 退款失败
)

// MembershipStatusConstants 会员状态常量
const (
	MembershipStatusActive   = "active"   // 有效
	MembershipStatusExpired  = "expired"  // 已过期
	MembershipStatusCanceled = "canceled" // 已取消
)

// LicenseStatusConstants 许可证状态常量
const (
	LicenseStatusActive   = "active"   // 有效
	LicenseStatusExpired  = "expired"  // 已过期
	LicenseStatusRevoked  = "revoked"  // 已撤销
	LicenseStatusSuspended = "suspended" // 已暂停
)

// PaymentChannelConstants 支付渠道常量
const (
	PaymentChannelAlipay   = "alipay"   // 支付宝
	PaymentChannelWechat   = "wechat"   // 微信支付
	PaymentChannelUnionPay = "unionpay" // 银联支付
)

// IsExpired 检查订单是否已过期
func (o *PaymentOrder) IsExpired() bool {
	if o.ExpireTime == nil {
		return false
	}
	return time.Now().After(*o.ExpireTime)
}

// CanRefund 检查订单是否可以退款
func (o *PaymentOrder) CanRefund() bool {
	return o.Status == PaymentStatusPaid
}

// IsActive 检查会员是否有效
func (m *UserMembership) IsActive() bool {
	now := time.Now()
	return m.Status == MembershipStatusActive && 
		   now.After(m.StartTime) && 
		   now.Before(m.EndTime)
}

// RemainingDays 计算剩余天数
func (m *UserMembership) RemainingDays() int {
	if !m.IsActive() {
		return 0
	}
	
	remaining := time.Until(m.EndTime)
	days := int(remaining.Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

// CanUseFeature 检查是否可以使用某个功能
func (m *UserMembership) CanUseFeature(feature string) bool {
	if !m.IsActive() {
		return false
	}
	
	if m.Plan == nil || m.Plan.Features == nil {
		return false
	}
	
	if value, exists := m.Plan.Features[feature]; exists {
		if boolValue, ok := value.(bool); ok {
			return boolValue
		}
	}
	
	return false
}

// CanQuery 检查是否还可以查询
func (m *UserMembership) CanQuery() bool {
	if !m.IsActive() {
		return false
	}
	
	// -1 表示无限制
	if m.Plan != nil && m.Plan.MaxQueries == -1 {
		return true
	}
	
	return m.UsedQueries < m.Plan.MaxQueries
}

// CanDownload 检查是否还可以下载
func (m *UserMembership) CanDownload() bool {
	if !m.IsActive() {
		return false
	}
	
	// -1 表示无限制
	if m.Plan != nil && m.Plan.MaxDownloads == -1 {
		return true
	}
	
	return m.UsedDownloads < m.Plan.MaxDownloads
}