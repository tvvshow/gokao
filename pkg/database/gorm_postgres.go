// gorm_postgres.go：基于 pkg/config.DatabaseConfig 的 GORM Postgres 初始化与 Redis 客户端构造。
//
// 各 service 的 internal/database/Initialize 应缩为：调 OpenGorm → 自身 AutoMigrate / Seed。
// Initialize 与 SeedDefaultData 是业务私有，不在共享层处理。
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	sharedcfg "github.com/oktetopython/gaokao/pkg/config"
)

// GormOpenOptions 构造 GORM 连接时的可调项。
type GormOpenOptions struct {
	// Production 为 true 时只记录错误日志；否则使用 Info 级别。
	Production bool
}

// OpenGorm 基于 pkg/config.DatabaseConfig 打开 GORM Postgres 连接并应用连接池。
// 业务侧（autoMigrate / seedDefaultData）由调用方在返回的 *gorm.DB 上自行处理。
func OpenGorm(cfg sharedcfg.DatabaseConfig, opts GormOpenOptions) (*gorm.DB, error) {
	logLevel := gormlogger.Info
	if opts.Production {
		logLevel = gormlogger.Error
	}

	db, err := gorm.Open(
		postgres.New(postgres.Config{DSN: cfg.DatabaseURL}),
		&gorm.Config{Logger: gormlogger.Default.LogMode(logLevel)},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Second)

	return db, nil
}

// OpenRedis 基于 pkg/config.RedisConfig 创建 redis.Client 并 ping 验证连通性。
//
// 默认应用 connection.go 内的 Redis 池/超时常量（DialTimeout=5s、Read/WriteTimeout=3s、
// PoolSize=10、MinIdleConns=5）。pingTimeout <= 0 时使用 5 秒兜底。
func OpenRedis(cfg sharedcfg.RedisConfig, pingTimeout time.Duration) (*redis.Client, error) {
	if pingTimeout <= 0 {
		pingTimeout = 5 * time.Second
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisURL,
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		DialTimeout:  RedisDefaultDialTimeout,
		ReadTimeout:  RedisDefaultReadTimeout,
		WriteTimeout: RedisDefaultWriteTimeout,
		PoolSize:     RedisDefaultPoolSize,
		MinIdleConns: RedisDefaultMinIdleConns,
	})

	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return rdb, nil
}
