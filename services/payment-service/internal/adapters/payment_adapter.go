package adapters

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// PaymentRequest 支付请求参数
type PaymentRequest struct {
	OrderNo     string          `json:"order_no"`
	Amount      decimal.Decimal `json:"amount"`
	Subject     string          `json:"subject"`
	Description string          `json:"description"`
	NotifyURL   string          `json:"notify_url"`
	ReturnURL   string          `json:"return_url"`
	UserID      string          `json:"user_id"`
	ClientIP    string          `json:"client_ip"`
	ExpireTime  time.Duration   `json:"expire_time"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// PaymentResponse 支付响应
type PaymentResponse struct {
	OrderNo        string    `json:"order_no"`
	ChannelTradeNo string    `json:"channel_trade_no"`
	PayURL         string    `json:"pay_url"`
	QRCode         string    `json:"qr_code"`
	FormData       string    `json:"form_data"`
	ExpiredAt      time.Time `json:"expired_at"`
}

// PaymentCallback 支付回调数据
type PaymentCallback struct {
	OrderNo        string          `json:"order_no"`
	ChannelTradeNo string          `json:"channel_trade_no"`
	Amount         decimal.Decimal `json:"amount"`
	ActualAmount   decimal.Decimal `json:"actual_amount"`
	Status         string          `json:"status"`
	PaidAt         time.Time       `json:"paid_at"`
	RawData        string          `json:"raw_data"`
	Signature      string          `json:"signature"`
}

// RefundRequest 退款请求
type RefundRequest struct {
	OrderNo        string          `json:"order_no"`
	RefundNo       string          `json:"refund_no"`
	ChannelTradeNo string          `json:"channel_trade_no"`
	Amount         decimal.Decimal `json:"amount"`
	Reason         string          `json:"reason"`
	NotifyURL      string          `json:"notify_url"`
}

// RefundResponse 退款响应
type RefundResponse struct {
	RefundNo       string          `json:"refund_no"`
	ChannelRefundNo string         `json:"channel_refund_no"`
	Amount         decimal.Decimal `json:"amount"`
	Status         string          `json:"status"`
	RefundedAt     time.Time       `json:"refunded_at"`
}

// QueryRequest 查询请求
type QueryRequest struct {
	OrderNo        string `json:"order_no"`
	ChannelTradeNo string `json:"channel_trade_no"`
}

// QueryResponse 查询响应
type QueryResponse struct {
	OrderNo        string          `json:"order_no"`
	ChannelTradeNo string          `json:"channel_trade_no"`
	Amount         decimal.Decimal `json:"amount"`
	Status         string          `json:"status"`
	PaidAt         *time.Time      `json:"paid_at"`
	RefundedAmount decimal.Decimal `json:"refunded_amount"`
}

// PaymentAdapter 支付适配器接口
type PaymentAdapter interface {
	// GetName 获取支付渠道名称
	GetName() string

	// CreatePayment 创建支付
	CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error)

	// VerifyCallback 验证回调签名
	VerifyCallback(ctx context.Context, data []byte, signature string) (*PaymentCallback, error)

	// QueryPayment 查询支付状态
	QueryPayment(ctx context.Context, req *QueryRequest) (*QueryResponse, error)

	// CreateRefund 创建退款
	CreateRefund(ctx context.Context, req *RefundRequest) (*RefundResponse, error)

	// QueryRefund 查询退款状态
	QueryRefund(ctx context.Context, refundNo string) (*RefundResponse, error)

	// CloseOrder 关闭订单
	CloseOrder(ctx context.Context, orderNo string) error
}

// PaymentAdapterFactory 支付适配器工厂
type PaymentAdapterFactory interface {
	GetAdapter(channel string) (PaymentAdapter, error)
}

// AdapterConfig 适配器配置
type AdapterConfig struct {
	AppID      string `json:"app_id"`
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	NotifyURL  string `json:"notify_url"`
	ReturnURL  string `json:"return_url"`
	SignType   string `json:"sign_type"`
	Sandbox    bool   `json:"sandbox"`
}

// PaymentError 支付错误
type PaymentError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details"`
}

func (e *PaymentError) Error() string {
	return e.Message
}

// NewPaymentError 创建支付错误
func NewPaymentError(code, message, details string) error {
	return &PaymentError{
		Code:    code,
		Message: message,
		Details: details,
	}
}