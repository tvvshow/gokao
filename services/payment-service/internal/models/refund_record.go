package models

import (
	"time"
	
	"github.com/shopspring/decimal"
)

// RefundRecord 退款记录
type RefundRecord struct {
	ID              int64           `json:"id"`
	RefundNo        string          `json:"refund_no"`
	OrderNo         string          `json:"order_no"`
	ChannelRefundNo string          `json:"channel_refund_no"`
	Amount          decimal.Decimal `json:"amount"`
	Reason          string          `json:"reason"`
	Status          string          `json:"status"`
	RefundedAt      *time.Time      `json:"refunded_at"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}