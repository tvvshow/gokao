package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Migration represents a database migration
type Migration struct {
	ID      uint   `gorm:"primaryKey"`
	Version string `gorm:"uniqueIndex;size:50"`
	Name    string `gorm:"size:200"`
	Applied bool   `gorm:"default:true"`
}

func (Migration) TableName() string {
	return "migrations"
}

// MigrationTool 数据库迁移工具
type MigrationTool struct {
	db *gorm.DB
}

// NewMigrationTool 创建新的迁移工具实例
func NewMigrationTool(dsn string) (*MigrationTool, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &MigrationTool{db: db}, nil
}

// SetupMigrationTable 创建迁移表
func (mt *MigrationTool) SetupMigrationTable() error {
	if err := mt.db.AutoMigrate(&Migration{}); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}
	return nil
}

// ApplyMigration 应用单个迁移
func (mt *MigrationTool) ApplyMigration(version, name, sql string) error {
	// 检查是否已应用
	var existing Migration
	if err := mt.db.Where("version = ?", version).First(&existing).Error; err == nil {
		if existing.Applied {
			log.Printf("Migration %s already applied, skipping", version)
			return nil
		}
	}

	// 开始事务
	tx := mt.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return err
	}

	// 执行迁移SQL
	if sql != "" {
		if err := tx.Exec(sql).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration SQL: %w", err)
		}
	}

	// 记录迁移
	migration := Migration{
		Version: version,
		Name:    name,
		Applied: true,
	}

	if err := tx.Where("version = ?", version).First(&Migration{}).Error; err != nil {
		// 插入新记录
		if err := tx.Create(&migration).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration: %w", err)
		}
	} else {
		// 更新现有记录
		if err := tx.Model(&Migration{}).Where("version = ?", version).Updates(map[string]interface{}{
			"applied": true,
			"name":    name,
		}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update migration record: %w", err)
		}
	}

	// 提交事务
	return tx.Commit().Error
}

// RollbackMigration 回滚单个迁移
func (mt *MigrationTool) RollbackMigration(version, rollbackSQL string) error {
	// 开始事务
	tx := mt.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return err
	}

	// 执行回滚SQL
	if rollbackSQL != "" {
		if err := tx.Exec(rollbackSQL).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute rollback SQL: %w", err)
		}
	}

	// 更新迁移记录
	if err := tx.Model(&Migration{}).Where("version = ?", version).Updates(map[string]interface{}{
		"applied": false,
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update migration record: %w", err)
	}

	// 提交事务
	return tx.Commit().Error
}

// GetMigrationStatus 获取迁移状态
func (mt *MigrationTool) GetMigrationStatus() ([]Migration, error) {
	var migrations []Migration
	if err := mt.db.Find(&migrations).Error; err != nil {
		return nil, fmt.Errorf("failed to get migrations: %w", err)
	}
	return migrations, nil
}

// Close 关闭数据库连接
func (mt *MigrationTool) Close() error {
	sqlDB, err := mt.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func main() {
	// 优先读取环境变量，避免硬编码凭据。
	dsn := os.Getenv("DATA_SERVICE_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		log.Fatal("missing database dsn: set DATA_SERVICE_DATABASE_URL or DATABASE_URL")
	}

	// 创建迁移工具
	mt, err := NewMigrationTool(dsn)
	if err != nil {
		log.Fatalf("Failed to create migration tool: %v", err)
	}
	defer mt.Close()

	// 创建迁移表
	if err := mt.SetupMigrationTable(); err != nil {
		log.Fatalf("Failed to setup migration table: %v", err)
	}

	// 定义迁移
	migrations := []struct {
		Version     string
		Name        string
		SQL         string
		RollbackSQL string
	}{
		{
			Version: "001",
			Name:    "Create universities table",
			SQL: `CREATE TABLE IF NOT EXISTS universities (
				id UUID PRIMARY KEY,
				code VARCHAR(20) UNIQUE,
				name VARCHAR(200) NOT NULL,
				province VARCHAR(50),
				city VARCHAR(50),
				level VARCHAR(50),
				type VARCHAR(100),
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);`,
			RollbackSQL: "DROP TABLE IF EXISTS universities;",
		},
		{
			Version: "002",
			Name:    "Create majors table",
			SQL: `CREATE TABLE IF NOT EXISTS majors (
				id UUID PRIMARY KEY,
				code VARCHAR(20) UNIQUE,
				name VARCHAR(200) NOT NULL,
				category VARCHAR(100),
				sub_category VARCHAR(100),
				education_level VARCHAR(50),
				description TEXT,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);`,
			RollbackSQL: "DROP TABLE IF EXISTS majors;",
		},
		{
			Version: "003",
			Name:    "Create admissions table",
			SQL: `CREATE TABLE IF NOT EXISTS admissions (
				id UUID PRIMARY KEY,
				year INTEGER NOT NULL,
				province VARCHAR(100) NOT NULL,
				university_id UUID REFERENCES universities(id),
				major_id UUID REFERENCES majors(id),
				batch VARCHAR(100),
				min_score INTEGER,
				max_score INTEGER,
				avg_score INTEGER,
				admit_count INTEGER,
				subject_type VARCHAR(50),
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);`,
			RollbackSQL: "DROP TABLE IF EXISTS admissions;",
		},
	}

	// 应用迁移
	for _, m := range migrations {
		fmt.Printf("Applying migration %s: %s\n", m.Version, m.Name)
		if err := mt.ApplyMigration(m.Version, m.Name, m.SQL); err != nil {
			log.Printf("Failed to apply migration %s: %v", m.Version, err)
		} else {
			fmt.Printf("Successfully applied migration %s\n", m.Version)
		}
	}

	// 显示迁移状态
	fmt.Println("\nMigration Status:")
	migrationsStatus, err := mt.GetMigrationStatus()
	if err != nil {
		log.Printf("Failed to get migration status: %v", err)
	} else {
		for _, m := range migrationsStatus {
			status := "Applied"
			if !m.Applied {
				status = "Pending"
			}
			fmt.Printf("- %s (%s): %s\n", m.Version, m.Name, status)
		}
	}
}
