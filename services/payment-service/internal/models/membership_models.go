package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// JSONB payment-service 统一 JSONB 自定义类型，覆盖 payment + membership
// 全套 jsonb 列。原来 payment_models.go 重复定义过一个 PaymentJSONB（带
// 空字节边界保护），本实现已合并该保护后删除 PaymentJSONB。
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

	// 显式处理空字节：某些驱动会把 NULL 物化为 []byte{}，直接 Unmarshal 会
	// 报 "unexpected end of JSON input"。视为 nil 更安全。
	if len(bytes) == 0 {
		*j = nil
		return nil
	}

	return json.Unmarshal(bytes, j)
}

// MembershipPlan 会员套餐模型
type MembershipPlan struct {
	ID           int       `json:"id" db:"id"`
	PlanCode     string    `json:"plan_code" db:"plan_code"`
	Name         string    `json:"name" db:"name"`
	Description  string    `json:"description" db:"description"`
	Price        float64   `json:"price" db:"price"`
	DurationDays int       `json:"duration_days" db:"duration_days"`
	Features     JSONB     `json:"features" db:"features"`
	MaxQueries   int       `json:"max_queries" db:"max_queries"`
	MaxDownloads int       `json:"max_downloads" db:"max_downloads"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// UserMembership 用户会员模型
type UserMembership struct {
	ID            int             `json:"id" db:"id"`
	UserID        string          `json:"user_id" db:"user_id"`
	PlanCode      string          `json:"plan_code" db:"plan_code"`
	OrderNo       string          `json:"order_no" db:"order_no"`
	StartTime     time.Time       `json:"start_time" db:"start_time"`
	EndTime       time.Time       `json:"end_time" db:"end_time"`
	Status        string          `json:"status" db:"status"`
	AutoRenew     bool            `json:"auto_renew" db:"auto_renew"`
	UsedQueries   int             `json:"used_queries" db:"used_queries"`
	UsedDownloads int             `json:"used_downloads" db:"used_downloads"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
	Plan          *MembershipPlan `json:"plan,omitempty"`
}

// IsActive 检查会员是否有效
func (um *UserMembership) IsActive() bool {
	return um.Status == MembershipStatusActive && time.Now().Before(um.EndTime)
}

// RemainingDays 计算剩余天数
func (um *UserMembership) RemainingDays() int {
	if !um.IsActive() {
		return 0
	}

	remaining := um.EndTime.Sub(time.Now())
	days := int(remaining.Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

// MembershipStatusResponse 会员状态响应
type MembershipStatusResponse struct {
	IsVIP         bool                   `json:"is_vip"`
	PlanCode      string                 `json:"plan_code,omitempty"`
	PlanName      string                 `json:"plan_name,omitempty"`
	StartTime     *time.Time             `json:"start_time,omitempty"`
	EndTime       *time.Time             `json:"end_time,omitempty"`
	RemainingDays int                    `json:"remaining_days"`
	UsedQueries   int                    `json:"used_queries"`
	MaxQueries    int                    `json:"max_queries"`
	UsedDownloads int                    `json:"used_downloads"`
	MaxDownloads  int                    `json:"max_downloads"`
	Features      map[string]interface{} `json:"features"`
	AutoRenew     bool                   `json:"auto_renew"`
}

// MembershipSubscribeRequest 会员订阅请求
type MembershipSubscribeRequest struct {
	PlanCode  string `json:"plan_code" validate:"required"`
	AutoRenew bool   `json:"auto_renew"`
}

// MembershipRenewRequest 会员续费请求
type MembershipRenewRequest struct {
	PlanCode string `json:"plan_code" validate:"required"`
}

// MembershipUpdateRequest 会员更新请求
type MembershipUpdateRequest struct {
	AutoRenew *bool `json:"auto_renew,omitempty"`
}

// 会员状态常量
const (
	MembershipStatusActive    = "active"
	MembershipStatusExpired   = "expired"
	MembershipStatusCanceled  = "canceled"
	MembershipStatusSuspended = "suspended"
)

// 会员套餐代码常量
const (
	PlanCodeBasic      = "basic"
	PlanCodeStandard   = "standard"
	PlanCodePremium    = "premium"
	PlanCodeEnterprise = "enterprise"
)

// 功能权限常量
const (
	FeatureBasicQuery      = "basic_query"
	FeatureAdvancedQuery   = "advanced_query"
	FeatureAIRecommend     = "ai_recommend"
	FeatureExpertConsult   = "expert_consult"
	FeatureDataExport      = "data_export"
	FeatureCustomReport    = "custom_report"
	FeaturePrioritySupport = "priority_support"
	FeatureAPIAccess       = "api_access"
)

// GetDefaultFeatures 获取套餐默认功能
func GetDefaultFeatures(planCode string) map[string]interface{} {
	switch planCode {
	case PlanCodeBasic:
		return map[string]interface{}{
			FeatureBasicQuery: true,
		}
	case PlanCodeStandard:
		return map[string]interface{}{
			FeatureBasicQuery:    true,
			FeatureAdvancedQuery: true,
			FeatureDataExport:    true,
		}
	case PlanCodePremium:
		return map[string]interface{}{
			FeatureBasicQuery:    true,
			FeatureAdvancedQuery: true,
			FeatureAIRecommend:   true,
			FeatureDataExport:    true,
			FeatureCustomReport:  true,
		}
	case PlanCodeEnterprise:
		return map[string]interface{}{
			FeatureBasicQuery:      true,
			FeatureAdvancedQuery:   true,
			FeatureAIRecommend:     true,
			FeatureExpertConsult:   true,
			FeatureDataExport:      true,
			FeatureCustomReport:    true,
			FeaturePrioritySupport: true,
			FeatureAPIAccess:       true,
		}
	default:
		return map[string]interface{}{
			FeatureBasicQuery: true,
		}
	}
}

// GetPlanLimits 获取套餐限制
func GetPlanLimits(planCode string) (maxQueries, maxDownloads int) {
	switch planCode {
	case PlanCodeBasic:
		return 100, 10
	case PlanCodeStandard:
		return 500, 50
	case PlanCodePremium:
		return 2000, 200
	case PlanCodeEnterprise:
		return -1, -1 // 无限制
	default:
		return 10, 0 // 免费用户限制
	}
}

// ValidatePlanCode 验证套餐代码
func ValidatePlanCode(planCode string) bool {
	validPlans := []string{
		PlanCodeBasic,
		PlanCodeStandard,
		PlanCodePremium,
		PlanCodeEnterprise,
	}

	for _, valid := range validPlans {
		if planCode == valid {
			return true
		}
	}
	return false
}

// MembershipBenefits 会员权益信息
type MembershipBenefits struct {
	Features      map[string]interface{} `json:"features"`
	QueryLimit    LimitInfo              `json:"query_limit"`
	DownloadLimit LimitInfo              `json:"download_limit"`
	ExpireInfo    *ExpireInfo            `json:"expire_info,omitempty"`
}

// LimitInfo 限制信息
type LimitInfo struct {
	Used      int  `json:"used"`
	Max       int  `json:"max"`
	Unlimited bool `json:"unlimited"`
}

// ExpireInfo 到期信息
type ExpireInfo struct {
	EndTime       *time.Time `json:"end_time"`
	RemainingDays int        `json:"remaining_days"`
	AutoRenew     bool       `json:"auto_renew"`
}

// MembershipHistory 会员历史记录
type MembershipHistory struct {
	ID        int       `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	PlanCode  string    `json:"plan_code" db:"plan_code"`
	OrderNo   string    `json:"order_no" db:"order_no"`
	Action    string    `json:"action" db:"action"` // subscribe, renew, cancel, expire
	StartTime time.Time `json:"start_time" db:"start_time"`
	EndTime   time.Time `json:"end_time" db:"end_time"`
	Amount    float64   `json:"amount" db:"amount"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// 会员历史动作常量
const (
	MembershipActionSubscribe = "subscribe"
	MembershipActionRenew     = "renew"
	MembershipActionCancel    = "cancel"
	MembershipActionExpire    = "expire"
	MembershipActionUpgrade   = "upgrade"
	MembershipActionDowngrade = "downgrade"
)
