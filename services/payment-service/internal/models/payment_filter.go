package models

import (
	"time"

	"github.com/google/uuid"
)

// PaymentFilter 支付订单查询过滤器
type PaymentFilter struct {
	UserID     *uuid.UUID `json:"user_id"`
	OrderNo    *string    `json:"order_no"`
	Channel    *string    `json:"channel"`
	Status     *string    `json:"status"`
	StartTime  *time.Time `json:"start_time"`
	EndTime    *time.Time `json:"end_time"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	OrderBy    string     `json:"order_by"`
	Descending bool       `json:"descending"`
}

// PaymentStatistics 支付统计信息
type PaymentStatistics struct {
	TotalAmount     float64 `json:"total_amount"`
	TotalCount      int64   `json:"total_count"`
	SuccessCount    int64   `json:"success_count"`
	FailedCount     int64   `json:"failed_count"`
	CanceledCount   int64   `json:"canceled_count"`
	RefundedCount   int64   `json:"refunded_count"`
	ExpiredCount    int64   `json:"expired_count"`
	PendingCount    int64   `json:"pending_count"`
	AvgAmount       float64 `json:"avg_amount"`
	UniqueUsers     int64   `json:"unique_users"`
	UniqueChannels  int64   `json:"unique_channels"`
	ChannelStats    []ChannelStat `json:"channel_stats"`
	DateStats       []DateStat    `json:"date_stats"`
}

// ChannelStat 渠道统计
type ChannelStat struct {
	Channel      string  `json:"channel"`
	TotalAmount  float64 `json:"total_amount"`
	TotalCount   int64   `json:"total_count"`
	SuccessCount int64   `json:"success_count"`
}

// DateStat 日期统计
type DateStat struct {
	Date         string  `json:"date"`
	TotalAmount  float64 `json:"total_amount"`
	TotalCount   int64   `json:"total_count"`
	SuccessCount int64   `json:"success_count"`
}