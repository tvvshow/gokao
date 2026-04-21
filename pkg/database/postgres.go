package database

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
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

	// 配置连接池 - 优化版本
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

	// 检查连接池使用率
	if stats.MaxOpenConnections > 0 {
		usagePercentage := float64(stats.InUse) / float64(stats.MaxOpenConnections) * 100
		if usagePercentage > 80 {
			return fmt.Errorf("database connection pool usage high: %.1f%%", usagePercentage)
		}
	}

	return nil
}

// GetConnectionPoolMetrics 获取连接池性能指标
func GetConnectionPoolMetrics(db *sqlx.DB) map[string]interface{} {
	if db == nil {
		return map[string]interface{}{
			"error": "database connection is nil",
		}
	}

	stats := db.Stats()

	// 计算使用率指标
	var usagePercentage float64
	if stats.MaxOpenConnections > 0 {
		usagePercentage = float64(stats.InUse) / float64(stats.MaxOpenConnections) * 100
	}

	var idlePercentage float64
	if stats.MaxOpenConnections > 0 {
		idlePercentage = float64(stats.Idle) / float64(stats.MaxOpenConnections) * 100
	}

	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"usage_percentage":     fmt.Sprintf("%.1f%%", usagePercentage),
		"idle_percentage":      fmt.Sprintf("%.1f%%", idlePercentage),
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
		"timestamp":            time.Now().Format(time.RFC3339),
	}
}

// MonitorConnectionPool 监控连接池状态并记录性能指标
func MonitorConnectionPool(db *sqlx.DB, logger *logrus.Logger) {
	if db == nil || logger == nil {
		return
	}

	metrics := GetConnectionPoolMetrics(db)

	// 记录性能指标
	logger.WithFields(logrus.Fields{
		"component": "database",
		"metrics":   metrics,
	}).Info("Database connection pool metrics")

	// 检查连接池健康状态
	if err := HealthCheck(db); err != nil {
		logger.WithFields(logrus.Fields{
			"component": "database",
			"error":     err.Error(),
		}).Warn("Database connection pool health check failed")
	}
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
