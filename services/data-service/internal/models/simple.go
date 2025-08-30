package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SimpleTest 简单测试模型
type SimpleTest struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Name      string    `gorm:"size:100" json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate GORM钩子：创建前生成UUID
func (st *SimpleTest) BeforeCreate(tx *gorm.DB) error {
	if st.ID == uuid.Nil {
		st.ID = uuid.New()
	}
	return nil
}