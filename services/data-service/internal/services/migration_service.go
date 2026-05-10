package services

import (
	"fmt"
	"github.com/tvvshow/gokao/services/data-service/internal/database"

	"gorm.io/gorm"
)

type migrationDefinition struct {
	Version     string
	Name        string
	SQL         string
	RollbackSQL string
}

// MigrationService 数据库迁移服务
type MigrationService struct {
	db       *database.DB
	migrator *database.Migrator
}

// NewMigrationService 创建新的迁移服务
func NewMigrationService(db *database.DB) *MigrationService {
	return &MigrationService{
		db:       db,
		migrator: database.NewMigrator(db.PostgreSQL),
	}
}

func defaultMigrationDefinitions() []migrationDefinition {
	return []migrationDefinition{
		{
			Version: "001",
			Name:    "Create universities table",
			SQL: `CREATE TABLE IF NOT EXISTS universities (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
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
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				university_id UUID REFERENCES universities(id),
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
			Name:    "Create admission_data table",
			SQL: `CREATE TABLE IF NOT EXISTS admission_data (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
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
			RollbackSQL: "DROP TABLE IF EXISTS admission_data;",
		},
		{
			Version: "004",
			Name:    "Add indexes for performance",
			SQL: `CREATE INDEX IF NOT EXISTS idx_universities_province ON universities(province);
			      CREATE INDEX IF NOT EXISTS idx_universities_level ON universities(level);
			      CREATE INDEX IF NOT EXISTS idx_majors_university_id ON majors(university_id);
			      CREATE INDEX IF NOT EXISTS idx_majors_category ON majors(category);
			      CREATE INDEX IF NOT EXISTS idx_admission_data_year ON admission_data(year);
			      CREATE INDEX IF NOT EXISTS idx_admission_data_province ON admission_data(province);
			      CREATE INDEX IF NOT EXISTS idx_admission_data_university_id ON admission_data(university_id);
			      CREATE INDEX IF NOT EXISTS idx_admission_data_major_id ON admission_data(major_id);`,
			RollbackSQL: `DROP INDEX IF EXISTS idx_universities_province;
			              DROP INDEX IF EXISTS idx_universities_level;
			              DROP INDEX IF EXISTS idx_majors_university_id;
			              DROP INDEX IF EXISTS idx_majors_category;
			              DROP INDEX IF EXISTS idx_admission_data_year;
			              DROP INDEX IF EXISTS idx_admission_data_province;
			              DROP INDEX IF EXISTS idx_admission_data_university_id;
			              DROP INDEX IF EXISTS idx_admission_data_major_id;`,
		},
	}
}

func rollbackStatements() map[string]string {
	statements := make(map[string]string, len(defaultMigrationDefinitions()))
	for _, migration := range defaultMigrationDefinitions() {
		statements[migration.Version] = migration.RollbackSQL
	}
	return statements
}

// ApplyAllMigrations 应用所有待处理的迁移
func (s *MigrationService) ApplyAllMigrations() error {
	// 确保迁移表存在
	if err := s.migrator.SetupMigrationTable(); err != nil {
		return fmt.Errorf("设置迁移表失败: %w", err)
	}

	// 应用所有迁移
	for _, m := range defaultMigrationDefinitions() {
		if err := s.migrator.ApplyMigration(m.Version, m.Name, m.SQL); err != nil {
			return fmt.Errorf("应用迁移 %s 失败: %w", m.Version, err)
		}
	}

	return nil
}

// GetMigrationStatus 获取迁移状态
func (s *MigrationService) GetMigrationStatus() ([]database.Migration, error) {
	// 确保迁移表存在
	if err := s.migrator.SetupMigrationTable(); err != nil {
		return nil, fmt.Errorf("设置迁移表失败: %w", err)
	}

	// 获取迁移状态
	status, err := s.migrator.GetMigrationStatus()
	if err != nil {
		return nil, fmt.Errorf("获取迁移状态失败: %w", err)
	}

	return status, nil
}

// RollbackMigration 回滚指定版本的迁移
func (s *MigrationService) RollbackMigration(version string) error {
	// 确保迁移表存在
	if err := s.migrator.SetupMigrationTable(); err != nil {
		return fmt.Errorf("设置迁移表失败: %w", err)
	}

	// 查找回滚SQL
	rollbackSQL, exists := rollbackStatements()[version]
	if !exists {
		return fmt.Errorf("未找到版本 %s 的回滚SQL", version)
	}

	// 执行回滚
	if err := s.migrator.RollbackMigration(version, rollbackSQL); err != nil {
		return fmt.Errorf("回滚迁移 %s 失败: %w", version, err)
	}

	return nil
}

// GetDB 返回数据库连接
func (s *MigrationService) GetDB() *gorm.DB {
	return s.db.PostgreSQL
}
