package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ConnectionPoolManager 统一的数据库连接池管理器
type ConnectionPoolManager struct {
	postgresConnections map[string]*gorm.DB
	redisConnections    map[string]*redis.Client
	mu                  sync.RWMutex
	config              *PoolConfig
}

// NewConnectionPoolManager 创建新的连接池管理器
func NewConnectionPoolManager(config *PoolConfig) *ConnectionPoolManager {
	if config == nil {
		config = NewDefaultPoolConfig()
	}

	return &ConnectionPoolManager{
		postgresConnections: make(map[string]*gorm.DB),
		redisConnections:    make(map[string]*redis.Client),
		config:              config,
	}
}

// GetPostgresConnection 获取PostgreSQL连接（单例模式）
func (m *ConnectionPoolManager) GetPostgresConnection(connectionName, dsn string) (*gorm.DB, error) {
	m.mu.RLock()
	if conn, exists := m.postgresConnections[connectionName]; exists {
		m.mu.RUnlock()
		return conn, nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	// 双重检查
	if conn, exists := m.postgresConnections[connectionName]; exists {
		return conn, nil
	}

	// 创建新连接
	conn, err := createPostgresConnection(dsn, m.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create PostgreSQL connection %s: %w", connectionName, err)
	}

	m.postgresConnections[connectionName] = conn
	return conn, nil
}

// GetRedisConnection 获取Redis连接（单例模式）
func (m *ConnectionPoolManager) GetRedisConnection(connectionName, addr, password string, db int) (*redis.Client, error) {
	m.mu.RLock()
	if conn, exists := m.redisConnections[connectionName]; exists {
		m.mu.RUnlock()
		return conn, nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	// 双重检查
	if conn, exists := m.redisConnections[connectionName]; exists {
		return conn, nil
	}

	// 创建新连接
	conn := createRedisConnection(addr, password, db, m.config)
	m.redisConnections[connectionName] = conn
	return conn, nil
}

// CloseAll 关闭所有连接
func (m *ConnectionPoolManager) CloseAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error

	// 关闭PostgreSQL连接
	for name, conn := range m.postgresConnections {
		if sqlDB, err := conn.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				errs = append(errs, fmt.Errorf("failed to close PostgreSQL connection %s: %w", name, err))
			}
		}
	}

	// 关闭Redis连接
	for name, conn := range m.redisConnections {
		if err := conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close Redis connection %s: %w", name, err))
		}
	}

	// 清空连接映射
	m.postgresConnections = make(map[string]*gorm.DB)
	m.redisConnections = make(map[string]*redis.Client)

	if len(errs) > 0 {
		return fmt.Errorf("errors closing connections: %v", errs)
	}

	return nil
}

// GetStats 获取连接池统计信息
func (m *ConnectionPoolManager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]interface{})
	
	// PostgreSQL连接统计
	postgresStats := make(map[string]interface{})
	for name, conn := range m.postgresConnections {
		if sqlDB, err := conn.DB(); err == nil {
			postgresStats[name] = map[string]interface{}{
				"max_open_conns": sqlDB.Stats().MaxOpenConnections,
				"open_conns":     sqlDB.Stats().OpenConnections,
				"in_use":         sqlDB.Stats().InUse,
				"idle":           sqlDB.Stats().Idle,
			}
		}
	}
	stats["postgres"] = postgresStats

	// Redis连接统计
	redisStats := make(map[string]interface{})
	for name, conn := range m.redisConnections {
		redisStats[name] = map[string]interface{}{
			"pool_size": conn.Options().PoolSize,
			"idle_conns": conn.PoolStats().IdleConns,
		}
	}
	stats["redis"] = redisStats

	return stats
}

// createPostgresConnection 创建PostgreSQL连接
func createPostgresConnection(dsn string, config *PoolConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open postgres with gorm failed: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("extract sql.DB from gorm failed: %w", err)
	}

	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping postgres failed: %w", err)
	}

	return db, nil
}

// createRedisConnection 创建Redis连接
func createRedisConnection(addr, password string, db int, config *PoolConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     config.MaxOpenConns,
		MinIdleConns: config.MaxIdleConns,
		MaxIdleConns: config.MaxIdleConns,
		ConnMaxLifetime: config.ConnMaxLifetime,
		ConnMaxIdleTime: config.ConnMaxIdleTime,
	})
}

// HealthCheck 健康检查
func (m *ConnectionPoolManager) HealthCheck(ctx context.Context) map[string]bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	results := make(map[string]bool)

	// 检查PostgreSQL连接
	for name, conn := range m.postgresConnections {
		if sqlDB, err := conn.DB(); err == nil {
			err := sqlDB.PingContext(ctx)
			results["postgres_"+name] = err == nil
		} else {
			results["postgres_"+name] = false
		}
	}

	// 检查Redis连接
	for name, conn := range m.redisConnections {
		err := conn.Ping(ctx).Err()
		results["redis_"+name] = err == nil
	}

	return results
}
