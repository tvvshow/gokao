package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	pkgmodels "github.com/tvvshow/gokao/pkg/models"
)

// SimpleTest 简单测试模型
type SimpleTest struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Name      string    `gorm:"size:100" json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate GORM钩子：创建前生成UUID（走 pkg/models helper）
func (st *SimpleTest) BeforeCreate(tx *gorm.DB) error {
	pkgmodels.AssignNewUUIDIfZero(&st.ID)
	return nil
}