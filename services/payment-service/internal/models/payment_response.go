package models

import (
	"time"
	
	"github.com/shopspring/decimal"
)

// PaymentResponse 支付响应
type PaymentResponse struct {
	OrderID     string          `json:"order_id"`
	PaymentURL  string          `json:"payment_url"`
	Amount      decimal.Decimal `json:"amount"`
	Currency    string          `json:"currency"`
	Status      string          `json:"status"`
	ExpiresAt   time.Time       `json:"expires_at"`
	QRCode      string          `json:"qr_code,omitempty"`
	TradeNumber string          `json:"trade_number,omitempty"`
}