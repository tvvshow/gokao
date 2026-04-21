package database

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Migrator 数据库迁移器
type Migrator struct {
	db *gorm.DB
}

// NewMigrator 创建新的迁移器实例
func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{db: db}
}

// SetupMigrationTable 创建迁移表
func (m *Migrator) SetupMigrationTable() error {
	// 创建迁移表，用于记录已应用的迁移
	if err := m.db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			version VARCHAR(50) UNIQUE NOT NULL,
			name VARCHAR(200) NOT NULL,
			applied BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`).Error; err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}
	return nil
}

// ApplyMigration 应用单个迁移
func (m *Migrator) ApplyMigration(version, name, sql string) error {
	// 检查是否已应用
	var count int64
	m.db.Model(&Migration{}).Where("version = ?", version).Count(&count)
	if count > 0 {
		// 检查是否已应用
		var existing Migration
		if err := m.db.Where("version = ?", version).First(&existing).Error; err == nil {
			if existing.Applied {
				return nil // 已应用，直接返回
			}
		}
	}

	// 开始事务
	tx := m.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return tx.Error
	}

	// 执行迁移SQL
	if sql != "" {
		if err := tx.Exec(sql).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration SQL: %w", err)
		}
	}

	// 使用原生SQL插入或更新
	if count > 0 {
		// 更新现有记录
		if err := tx.Exec(`
			UPDATE migrations 
			SET applied = ?, name = ?, updated_at = CURRENT_TIMESTAMP
			WHERE version = ?
		`, true, name, version).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update migration record: %w", err)
		}
	} else {
		// 插入新记录
		if err := tx.Exec(`
			INSERT INTO migrations (version, name, applied, created_at, updated_at)
			VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`, version, name, true).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert migration record: %w", err)
		}
	}

	// 提交事务
	return tx.Commit().Error
}

// RollbackMigration 回滚单个迁移
func (m *Migrator) RollbackMigration(version, rollbackSQL string) error {
	// 开始事务
	tx := m.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return tx.Error
	}

	// 执行回滚SQL
	if rollbackSQL != "" {
		if err := tx.Exec(rollbackSQL).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute rollback SQL: %w", err)
		}
	}

	// 更新迁移记录
	if err := tx.Exec(`
		UPDATE migrations 
		SET applied = ?, updated_at = CURRENT_TIMESTAMP
		WHERE version = ?
	`, false, version).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update migration record: %w", err)
	}

	// 提交事务
	return tx.Commit().Error
}

// GetMigrationStatus 获取迁移状态
func (m *Migrator) GetMigrationStatus() ([]Migration, error) {
	var migrations []Migration
	if err := m.db.Find(&migrations).Error; err != nil {
		return nil, fmt.Errorf("failed to get migrations: %w", err)
	}
	return migrations, nil
}

// Migration 迁移记录模型
type Migration struct {
	ID        uint      `gorm:"primaryKey"`
	Version   string    `gorm:"uniqueIndex;size:50"`
	Name      string    `gorm:"size:200"`
	Applied   bool      `gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}