package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Username   string     `gorm:"uniqueIndex;not null;size:50" json:"username" validate:"required,min=3,max=50"`
	Email      string     `gorm:"uniqueIndex;not null;size:100" json:"email" validate:"required,email"`
	Phone      string     `gorm:"uniqueIndex;size:20" json:"phone,omitempty" validate:"omitempty,min=11,max=20"`
	Password   string     `gorm:"not null;size:255" json:"-" validate:"required,min=6"`
	Nickname   string     `gorm:"size:50" json:"nickname,omitempty"`
	Avatar     string     `gorm:"size:255" json:"avatar,omitempty"`
	Gender     string     `gorm:"size:10" json:"gender,omitempty" validate:"omitempty,oneof=male female other"`
	Birthday   *time.Time `json:"birthday,omitempty"`
	Province   string     `gorm:"size:50;index" json:"province,omitempty"`
	City       string     `gorm:"size:50" json:"city,omitempty"`
	School     string     `gorm:"size:100" json:"school,omitempty"`
	Grade      string     `gorm:"size:20" json:"grade,omitempty"`
	Status     string     `gorm:"default:'active';size:20;index" json:"status" validate:"oneof=active inactive suspended"`
	IsVerified bool       `gorm:"default:false;index" json:"is_verified"`

	// 会员相关字段
	MembershipLevel  string     `gorm:"default:'free';size:20;index" json:"membership_level" validate:"oneof=free basic premium enterprise"`
	MembershipExpiry *time.Time `gorm:"index" json:"membership_expiry,omitempty"`
	MaxDevices       int        `gorm:"default:1" json:"max_devices"`
	TrialUsed        bool       `gorm:"default:false" json:"trial_used"`
	TrialExpiry      *time.Time `json:"trial_expiry,omitempty"`

	// 登录相关字段
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	LastLoginIP string     `gorm:"size:45" json:"last_login_ip,omitempty"`
	LoginCount  uint64     `gorm:"default:0" json:"login_count"`

	// 审计字段
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	Roles              []Role              `gorm:"many2many:user_roles;" json:"roles,omitempty"`
	DeviceFingerprints []DeviceFingerprint `gorm:"foreignKey:UserID" json:"device_fingerprints,omitempty"`
	MembershipOrders   []MembershipOrder   `gorm:"foreignKey:UserID" json:"membership_orders,omitempty"`
	UserSessions       []UserSession       `gorm:"foreignKey:UserID" json:"user_sessions,omitempty"`
}

// APIResponse 通用API响应
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// JWTClaims JWT声明
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// Role 角色模型
type Role struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null;size:50" json:"name" validate:"required,min=2,max=50"`
	Description string    `gorm:"size:255" json:"description,omitempty"`
	IsSystem    bool      `gorm:"default:false" json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联关系
	Users       []User       `gorm:"many2many:user_roles;" json:"users,omitempty"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}

// Permission 权限模型
type Permission struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null;size:100" json:"name" validate:"required,min=2,max=100"`
	Description string    `gorm:"size:255" json:"description,omitempty"`
	Resource    string    `gorm:"size:50" json:"resource,omitempty"`
	Action      string    `gorm:"size:50" json:"action,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联关系
	Roles []Role `gorm:"many2many:role_permissions;" json:"roles,omitempty"`
}

// UserRole 用户角色关联表
type UserRole struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	RoleID    uint      `gorm:"primaryKey" json:"role_id"`
	CreatedAt time.Time `json:"created_at"`

	// 关联关系
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Role Role `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

// RolePermission 角色权限关联表
type RolePermission struct {
	RoleID       uint      `gorm:"primaryKey" json:"role_id"`
	PermissionID uint      `gorm:"primaryKey" json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`

	// 关联关系
	Role       Role       `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Permission Permission `gorm:"foreignKey:PermissionID" json:"permission,omitempty"`
}

// LoginAttempt 登录尝试记录
type LoginAttempt struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"index;not null;size:50" json:"username"`
	IP        string    `gorm:"index;not null;size:45" json:"ip"`
	UserAgent string    `gorm:"size:500" json:"user_agent,omitempty"`
	Success   bool      `gorm:"index" json:"success"`
	Reason    string    `gorm:"size:255" json:"reason,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// RefreshToken 刷新令牌
type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Token     string    `gorm:"uniqueIndex;not null;size:255" json:"token"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	IsRevoked bool      `gorm:"default:false;index" json:"is_revoked"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联关系
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// AuditLog 审计日志
type AuditLog struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID     *uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`
	Action     string     `gorm:"not null;size:100;index" json:"action"`
	Resource   string     `gorm:"size:100;index" json:"resource,omitempty"`
	ResourceID string     `gorm:"size:100;index" json:"resource_id,omitempty"`
	Details    string     `gorm:"type:text" json:"details,omitempty"`
	IP         string     `gorm:"size:45;index" json:"ip,omitempty"`
	UserAgent  string     `gorm:"size:500" json:"user_agent,omitempty"`
	Status     string     `gorm:"size:20;index" json:"status"`
	CreatedAt  time.Time  `json:"created_at"`

	// 关联关系
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// DeviceFingerprint 设备指纹模型
type DeviceFingerprint struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID           uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	DeviceID         string         `gorm:"uniqueIndex;not null;size:255" json:"device_id"`
	DeviceName       string         `gorm:"size:100" json:"device_name,omitempty"`
	DeviceType       string         `gorm:"size:20;index" json:"device_type,omitempty" validate:"omitempty,oneof=mobile tablet desktop"`
	Platform         string         `gorm:"size:50" json:"platform,omitempty"`
	Browser          string         `gorm:"size:100" json:"browser,omitempty"`
	BrowserVersion   string         `gorm:"size:50" json:"browser_version,omitempty"`
	OS               string         `gorm:"size:50" json:"os,omitempty"`
	OSVersion        string         `gorm:"size:50" json:"os_version,omitempty"`
	ScreenResolution string         `gorm:"size:20" json:"screen_resolution,omitempty"`
	Timezone         string         `gorm:"size:50" json:"timezone,omitempty"`
	Language         string         `gorm:"size:10" json:"language,omitempty"`
	UserAgent        string         `gorm:"size:500" json:"user_agent,omitempty"`
	IPAddress        string         `gorm:"size:45;index" json:"ip_address,omitempty"`
	Location         string         `gorm:"size:100" json:"location,omitempty"`
	IsActive         bool           `gorm:"default:true;index" json:"is_active"`
	IsTrusted        bool           `gorm:"default:false" json:"is_trusted"`
	LastSeenAt       *time.Time     `gorm:"index" json:"last_seen_at,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// MembershipOrder 会员订单模型
type MembershipOrder struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID          uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	OrderNo         string         `gorm:"uniqueIndex;not null;size:50" json:"order_no"`
	ProductName     string         `gorm:"not null;size:100" json:"product_name"`
	MembershipLevel string         `gorm:"not null;size:20;index" json:"membership_level" validate:"required,oneof=basic premium enterprise"`
	Duration        int            `gorm:"not null" json:"duration"`        // 会员时长（天）
	OriginalPrice   int64          `gorm:"not null" json:"original_price"`  // 原价（分）
	DiscountPrice   int64          `gorm:"default:0" json:"discount_price"` // 优惠金额（分）
	FinalPrice      int64          `gorm:"not null" json:"final_price"`     // 实付金额（分）
	Currency        string         `gorm:"default:'CNY';size:10" json:"currency"`
	PaymentMethod   string         `gorm:"size:50;index" json:"payment_method,omitempty"`
	PaymentProvider string         `gorm:"size:50" json:"payment_provider,omitempty"`
	PaymentID       string         `gorm:"size:100;index" json:"payment_id,omitempty"`
	DiscountCode    string         `gorm:"size:50" json:"discount_code,omitempty"`
	Status          string         `gorm:"not null;size:20;index" json:"status" validate:"required,oneof=pending paid cancelled refunded expired"`
	PaidAt          *time.Time     `gorm:"index" json:"paid_at,omitempty"`
	ExpiredAt       *time.Time     `gorm:"index" json:"expired_at,omitempty"`
	RefundedAt      *time.Time     `json:"refunded_at,omitempty"`
	RefundAmount    int64          `gorm:"default:0" json:"refund_amount"`
	RefundReason    string         `gorm:"size:255" json:"refund_reason,omitempty"`
	Notes           string         `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// UserSession 用户会话模型
type UserSession struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID           uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	DeviceID         string         `gorm:"size:255;index" json:"device_id,omitempty"`
	SessionToken     string         `gorm:"uniqueIndex;not null;size:255" json:"session_token"`
	RefreshToken     string         `gorm:"uniqueIndex;size:255" json:"refresh_token,omitempty"`
	IPAddress        string         `gorm:"size:45;index" json:"ip_address,omitempty"`
	UserAgent        string         `gorm:"size:500" json:"user_agent,omitempty"`
	Location         string         `gorm:"size:100" json:"location,omitempty"`
	IsActive         bool           `gorm:"default:true;index" json:"is_active"`
	ExpiresAt        time.Time      `gorm:"not null;index" json:"expires_at"`
	RefreshExpiresAt *time.Time     `gorm:"index" json:"refresh_expires_at,omitempty"`
	LastActivityAt   time.Time      `gorm:"index" json:"last_activity_at"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 设置表名
func (User) TableName() string {
	return "users"
}

func (Role) TableName() string {
	return "roles"
}

func (Permission) TableName() string {
	return "permissions"
}

func (UserRole) TableName() string {
	return "user_roles"
}

func (RolePermission) TableName() string {
	return "role_permissions"
}

func (LoginAttempt) TableName() string {
	return "login_attempts"
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

func (DeviceFingerprint) TableName() string {
	return "device_fingerprints"
}

func (MembershipOrder) TableName() string {
	return "membership_orders"
}

func (UserSession) TableName() string {
	return "user_sessions"
}

// BeforeCreate GORM钩子：创建前
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (rt *RefreshToken) BeforeCreate(tx *gorm.DB) error {
	if rt.ID == uuid.Nil {
		rt.ID = uuid.New()
	}
	return nil
}

func (al *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if al.ID == uuid.Nil {
		al.ID = uuid.New()
	}
	return nil
}

func (df *DeviceFingerprint) BeforeCreate(tx *gorm.DB) error {
	if df.ID == uuid.Nil {
		df.ID = uuid.New()
	}
	return nil
}

func (mo *MembershipOrder) BeforeCreate(tx *gorm.DB) error {
	if mo.ID == uuid.Nil {
		mo.ID = uuid.New()
	}
	return nil
}

// 用户状态常量
const (
	UserStatusActive    = "active"
	UserStatusInactive  = "inactive"
	UserStatusSuspended = "suspended"
	UserStatusDeleted   = "deleted"
)

// 用户角色常量
const (
	UserRoleStudent = "student"
	UserRoleParent  = "parent"
	UserRoleTeacher = "teacher"
	UserRoleAdmin   = "admin"
)

func (us *UserSession) BeforeCreate(tx *gorm.DB) error {
	if us.ID == uuid.Nil {
		us.ID = uuid.New()
	}
	return nil
}
