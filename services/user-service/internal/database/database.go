package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"time"

	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	shareddb "github.com/tvvshow/gokao/pkg/database"
	"github.com/tvvshow/gokao/services/user-service/internal/config"
)

// embedMigrations 把 migrations/*.sql 嵌入二进制，避免运行时依赖文件系统路径。
//
//go:embed migrations/*.sql
var embedMigrations embed.FS

// Initialize 初始化数据库连接：连接 + 池配置走共享层，schema/seed 走 goose 版本化迁移。
//
// 与旧 AutoMigrate + seedDefaultData 的差异：
//   - schema 变更走版本化 SQL（migrations/0000N_*.sql），可回滚；
//   - goose 在 goose_db_version 表追踪版本，二进制启动时自动 Up；
//   - GORM 仅承担运行时 ORM，结构演进权归迁移文件，杜绝 model/schema 漂移；
//   - 默认权限/角色/角色-权限映射收敛到 00002_seed.sql，单条 INSERT...SELECT
//     + ON CONFLICT DO NOTHING 替代旧 N 行 First/Create 循环。
func Initialize(cfg *config.Config) (*gorm.DB, error) {
	db, err := shareddb.OpenGorm(cfg.DatabaseConfig, shareddb.GormOpenOptions{
		Production: cfg.Environment == "production",
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("postgres ping: %w", err)
	}

	if err := migrate(sqlDB); err != nil {
		return nil, fmt.Errorf("goose migrate: %w", err)
	}

	return db, nil
}

// InitializeRedis 初始化 Redis 连接（含 ping 验证）。
func InitializeRedis(cfg *config.Config) (*redis.Client, error) {
	return shareddb.OpenRedis(cfg.RedisConfig, 0)
}

// migrate 应用所有未运行的 goose 迁移到目标 DB。
func migrate(sqlDB *sql.DB) error {
	log := logrus.StandardLogger()
	log.Info("开始数据库迁移（goose）...")

	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}
	goose.SetLogger(gooseLogger{log})

	if err := goose.Up(sqlDB, "migrations"); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	log.Info("数据库迁移完成")
	return nil
}

// gooseLogger 把 goose 内部日志桥接到 logrus，统一日志格式。
type gooseLogger struct{ l *logrus.Logger }

func (g gooseLogger) Fatalf(format string, v ...interface{}) { g.l.Fatalf(format, v...) }
func (g gooseLogger) Printf(format string, v ...interface{}) { g.l.Infof(format, v...) }
