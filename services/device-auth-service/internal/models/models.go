package models

import (
	"time"

	"github.com/google/uuid"
)

// Device 设备模型
type Device struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;"`
	DeviceID      string    `gorm:"uniqueIndex;not null"`
	DeviceType    string    `gorm:"not null"`
	Fingerprint   string    `gorm:"type:text"`
	LastSeen      time.Time `gorm:"not null"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// License 许可证模型
type License struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;"`
	DeviceID    string    `gorm:"not null;index"`
	LicenseData string    `gorm:"type:text;not null"`
	ExpiresAt   time.Time `gorm:"not null"`
	Revoked     bool      `gorm:"default:false"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}