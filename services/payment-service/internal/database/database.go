package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/tvvshow/gokao/services/payment-service/internal/config"
)

// embedMigrations 嵌入 migrations/*.sql 到二进制，避免运行时依赖文件系统路径。
//go:embed migrations/*.sql
var embedMigrations embed.FS

// Initialize 初始化数据库连接 + 跑迁移到最新版本。
//
// 与旧版 createTables() 的差异：
//   - schema 变更走版本化 SQL 文件（goose），可向下回滚；
//   - 二进制启动时自动 Up 到最新 — 与 docker compose depends_on healthy 协同保证次序。
func Initialize(cfg config.DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
	db.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	return db, nil
}

// runMigrations 应用所有未运行的迁移到目标 DB。
// goose 把 schema 版本追踪存在 `goose_db_version` 表内，第一次运行自动建。
func runMigrations(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}
	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}

// InitializeRedis 初始化Redis连接
func InitializeRedis(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := client.Ping(client.Context()).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return client, nil
}
