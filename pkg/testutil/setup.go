package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestConfig 测试配置
type TestConfig struct {
	DatabaseURL     string
	RedisURL        string
	MigrationPath   string
	TestTimeout     time.Duration
	LogLevel        string
	EnableMock      bool
	CleanupDatabase bool
}

// DefaultTestConfig 默认测试配置
func DefaultTestConfig() *TestConfig {
	return &TestConfig{
		DatabaseURL:     "postgres://postgres:postgres@localhost:5432/gaokao_test?sslmode=disable",
		RedisURL:        "localhost:6379",
		MigrationPath:   "../migrations",
		TestTimeout:     30 * time.Second,
		LogLevel:        "error",
		EnableMock:      true,
		CleanupDatabase: true,
	}
}

// TestEnvironment 测试环境
type TestEnvironment struct {
	Config     *TestConfig
	DB         *gorm.DB
	Redis      *redis.Client
	SQLMock    sqlmock.Sqlmock
	CleanupFunc func()
	ctx        context.Context
}

// SetupTestEnvironment 设置测试环境
func SetupTestEnvironment(t *testing.T, config *TestConfig) *TestEnvironment {
	t.Helper()

	if config == nil {
		config = DefaultTestConfig()
	}

	env := &TestEnvironment{
		Config: config,
		ctx:    context.Background(),
	}

	// 设置清理函数
	env.CleanupFunc = func() {
		if env.DB != nil {
			sqlDB, err := env.DB.DB()
			if err == nil {
				sqlDB.Close()
			}
		}
		if env.Redis != nil {
			env.Redis.Close()
		}
	}

	// 设置数据库连接
	if config.EnableMock {
		env.setupMockDatabase(t)
	} else {
		env.setupRealDatabase(t)
	}

	// 设置Redis连接
	env.setupRedis(t)

	return env
}

// setupMockDatabase 设置模拟数据库
func (env *TestEnvironment) setupMockDatabase(t *testing.T) {
	t.Helper()

	var db *sql.DB
	var err error
	var mock sqlmock.Sqlmock

	db, mock, err = sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	env.DB = gormDB
	env.SQLMock = mock
}

// setupRealDatabase 设置真实数据库
func (env *TestEnvironment) setupRealDatabase(t *testing.T) {
	t.Helper()

	gormDB, err := gorm.Open(postgres.Open(env.Config.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	require.NoError(t, err)

	// 测试数据库连接
	sqlDB, err := gormDB.DB()
	require.NoError(t, err)

	err = sqlDB.Ping()
	require.NoError(t, err)

	env.DB = gormDB

	// 运行迁移（如果配置了清理）
	if env.Config.CleanupDatabase {
		env.cleanupDatabase(t)
		env.runMigrations(t)
	}
}

// setupRedis 设置Redis连接
func (env *TestEnvironment) setupRedis(t *testing.T) {
	t.Helper()

	client := redis.NewClient(&redis.Options{
		Addr:     env.Config.RedisURL,
		Password: "", // 无密码
		DB:       1,   // 使用测试数据库
	})

	// 测试连接
	_, err := client.Ping(env.ctx).Result()
	if err != nil {
		t.Logf("Redis not available, skipping: %v", err)
		return
	}

	// 清空测试数据库
	if env.Config.CleanupDatabase {
		client.FlushDB(env.ctx)
	}

	env.Redis = client
}

// cleanupDatabase 清理数据库
func (env *TestEnvironment) cleanupDatabase(t *testing.T) {
	t.Helper()

	if env.DB == nil {
		return
	}

	// 获取所有表
	var tables []string
	err := env.DB.Raw(
		"SELECT tablename FROM pg_tables WHERE schemaname = 'public'").
		Pluck("tablename", &tables).Error
	if err != nil {
		t.Logf("Failed to get tables: %v", err)
		return
	}

	// 禁用外键约束
	err = env.DB.Exec("SET CONSTRAINTS ALL DEFERRED").Error
	if err != nil {
		t.Logf("Failed to defer constraints: %v", err)
	}

	// 删除所有表数据
	for _, table := range tables {
		if table == "schema_migrations" {
			continue // 保留迁移表
		}
		err = env.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)).Error
		if err != nil {
			t.Logf("Failed to truncate table %s: %v", table, err)
		}
	}

	// 重新启用外键约束
	err = env.DB.Exec("SET CONSTRAINTS ALL IMMEDIATE").Error
	if err != nil {
		t.Logf("Failed to enable constraints: %v", err)
	}
}

// runMigrations 运行数据库迁移
func (env *TestEnvironment) runMigrations(t *testing.T) {
	t.Helper()

	if env.Config.MigrationPath == "" {
		t.Log("No migration path specified, skipping migrations")
		return
	}

	// 检查迁移目录是否存在
	if _, err := os.Stat(env.Config.MigrationPath); os.IsNotExist(err) {
		t.Logf("Migration path %s does not exist, skipping migrations", env.Config.MigrationPath)
		return
	}

	// 这里应该实现具体的迁移逻辑
	// 可以使用gorm的AutoMigrate或者执行SQL文件
	t.Log("Migrations would be run here")
}

// Teardown 清理测试环境
func (env *TestEnvironment) Teardown() {
	if env.CleanupFunc != nil {
		env.CleanupFunc()
	}
}

// WithContext 使用指定的上下文
func (env *TestEnvironment) WithContext(ctx context.Context) *TestEnvironment {
	env.ctx = ctx
	return env
}

// GetContext 获取当前上下文
func (env *TestEnvironment) GetContext() context.Context {
	if env.ctx == nil {
		env.ctx = context.Background()
	}
	return env.ctx
}

// MockExpectationsWereMet 检查mock期望是否满足
func (env *TestEnvironment) MockExpectationsWereMet(t *testing.T) {
	t.Helper()

	if env.SQLMock != nil {
		err := env.SQLMock.ExpectationsWereMet()
		require.NoError(t, err)
	}
}

// Test utilities for specific scenarios

// CreateTestUser 创建测试用户
func (env *TestEnvironment) CreateTestUser(t *testing.T, user interface{}) interface{} {
	t.Helper()

	if env.DB == nil {
		t.Fatal("Database not initialized")
	}

	err := env.DB.Create(user).Error
	require.NoError(t, err)

	return user
}

// CreateTestUniversity 创建测试大学数据
func (env *TestEnvironment) CreateTestUniversity(t *testing.T, university interface{}) interface{} {
	t.Helper()

	if env.DB == nil {
		t.Fatal("Database not initialized")
	}

	err := env.DB.Create(university).Error
	require.NoError(t, err)

	return university
}

// CreateTestMajor 创建测试专业数据
func (env *TestEnvironment) CreateTestMajor(t *testing.T, major interface{}) interface{} {
	t.Helper()

	if env.DB == nil {
		t.Fatal("Database not initialized")
	}

	err := env.DB.Create(major).Error
	require.NoError(t, err)

	return major
}

// AssertRedisKeyExists 断言Redis key存在
func (env *TestEnvironment) AssertRedisKeyExists(t *testing.T, key string) {
	t.Helper()

	if env.Redis == nil {
		t.Skip("Redis not available")
		return
	}

	exists, err := env.Redis.Exists(env.ctx, key).Result()
	require.NoError(t, err)
	require.True(t, exists > 0, "Redis key %s should exist", key)
}

// LoadTestData 加载测试数据
func (env *TestEnvironment) LoadTestData(t *testing.T, dataFile string) {
	t.Helper()

	// 这里可以实现从JSON/YAML文件加载测试数据的逻辑
	t.Logf("Would load test data from %s", dataFile)
}

// GetTestTimeout 获取测试超时时间
func (env *TestEnvironment) GetTestTimeout() time.Duration {
	if env.Config != nil && env.Config.TestTimeout > 0 {
		return env.Config.TestTimeout
	}
	return 30 * time.Second
}