package wechat

import (
	"time"
)

// WeChatPayConfig 微信支付配置
type WeChatPayConfig struct {
	AppID       string `json:"app_id" yaml:"app_id"`
	MchID       string `json:"mch_id" yaml:"mch_id"`
	APIKey      string `json:"api_key" yaml:"api_key"`
	CertPath    string `json:"cert_path" yaml:"cert_path"`
	KeyPath     string `json:"key_path" yaml:"key_path"`
	NotifyURL   string `json:"notify_url" yaml:"notify_url"`
	ReturnURL   string `json:"return_url" yaml:"return_url"`
	Sandbox     bool   `json:"sandbox" yaml:"sandbox"`
}

// PaymentOrder 支付订单结构
type PaymentOrder struct {
	OrderID     string    `json:"order_id"`
	UserID      string    `json:"user_id"`
	Amount      int64     `json:"amount"`       // 金额（分）
	Currency    string    `json:"currency"`     // 货币类型
	Subject     string    `json:"subject"`      // 商品描述
	Body        string    `json:"body"`         // 商品详情
	ClientIP    string    `json:"client_ip"`    // 客户端IP
	TimeExpire  time.Time `json:"time_expire"`  // 超时时间
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PaymentResult 支付结果
type PaymentResult struct {
	Success     bool              `json:"success"`
	TradeNo     string            `json:"trade_no"`      // 微信交易号
	OutTradeNo  string            `json:"out_trade_no"`  // 商户订单号
	Amount      int64             `json:"amount"`
	PayTime     time.Time         `json:"pay_time"`
	PaymentType string            `json:"payment_type"`  // 支付方式
	PrepayID    string            `json:"prepay_id"`     // 预支付ID
	QRCode      string            `json:"qr_code"`       // 二维码
	JsAPIData   map[string]string `json:"jsapi_data"`    // JSAPI支付数据
	Message     string            `json:"message"`
}

// NotifyRequest 支付回调请求
type NotifyRequest struct {
	AppID         string `xml:"appid"`
	MchID         string `xml:"mch_id"`
	NonceStr      string `xml:"nonce_str"`
	Sign          string `xml:"sign"`
	SignType      string `xml:"sign_type"`
	OpenID        string `xml:"openid"`
	TradeType     string `xml:"trade_type"`
	TradeState    string `xml:"trade_state"`
	BankType      string `xml:"bank_type"`
	TotalFee      string `xml:"total_fee"`
	FeeType       string `xml:"fee_type"`
	CashFee       string `xml:"cash_fee"`
	TransactionID string `xml:"transaction_id"`
	OutTradeNo    string `xml:"out_trade_no"`
	TimeEnd       string `xml:"time_end"`
}

// RefundRequest 退款请求
type RefundRequest struct {
	OrderID     string `json:"order_id"`
	RefundID    string `json:"refund_id"`
	TotalAmount int64  `json:"total_amount"`  // 原订单金额
	RefundAmount int64  `json:"refund_amount"` // 退款金额
	Reason      string `json:"reason"`        // 退款原因
}

// RefundResult 退款结果
type RefundResult struct {
	Success      bool      `json:"success"`
	RefundID     string    `json:"refund_id"`
	OutRefundNo  string    `json:"out_refund_no"`
	RefundAmount int64     `json:"refund_amount"`
	RefundTime   time.Time `json:"refund_time"`
	Message      string    `json:"message"`
}