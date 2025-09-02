package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Config 数据库配置
type Config struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// LoadConfig 从环境变量加载数据库配置
func LoadConfig() *Config {
	return &Config{
		Host:            getEnv("DB_HOST", "localhost"),
		Port:            getEnvAsInt("DB_PORT", 5432),
		User:            getEnv("DB_USER", "postgres"),
		Password:        getEnv("DB_PASSWORD", ""),
		DBName:          getEnv("DB_NAME", "gaokao_db"),
		SSLMode:         getEnv("DB_SSL_MODE", "disable"),
		MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: time.Duration(getEnvAsInt("DB_CONN_MAX_LIFETIME", 300)) * time.Second,
		ConnMaxIdleTime: time.Duration(getEnvAsInt("DB_CONN_MAX_IDLE_TIME", 60)) * time.Second,
	}
}

// NewConnection 创建新的数据库连接
func NewConnection(config *Config) (*sqlx.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DBName,
		config.SSLMode,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 配置连接池
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// NewConnectionWithDSN 使用DSN创建数据库连接
func NewConnectionWithDSN(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 默认连接池配置
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// HealthCheck 数据库健康检查
func HealthCheck(db *sqlx.DB) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// 检查连接池状态
	stats := db.Stats()
	if stats.OpenConnections == 0 {
		return fmt.Errorf("no open database connections")
	}

	return nil
}

// Transaction 事务处理辅助函数
func Transaction(db *sqlx.DB, fn func(*sqlx.Tx) error) error {
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(tx)
	return err
}

// 辅助函数
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
