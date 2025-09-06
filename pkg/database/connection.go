package database

import "time"

// PoolConfig 数据库连接池统一配置
const (
	// DefaultMaxOpenConns 默认最大打开连接数
	DefaultMaxOpenConns = 25
	// DefaultMaxIdleConns 默认最大空闲连接数
	DefaultMaxIdleConns = 5
	// DefaultConnMaxLifetime 默认连接最大生存时间
	DefaultConnMaxLifetime = 30 * time.Minute
	// DefaultConnMaxIdleTime 默认连接最大空闲时间
	DefaultConnMaxIdleTime = 15 * time.Minute
	// DefaultDialTimeout 默认连接超时时间
	DefaultDialTimeout = 5 * time.Second
	// DefaultReadTimeout 默认读取超时时间
	DefaultReadTimeout = 3 * time.Second
	// DefaultWriteTimeout 默认写入超时时间
	DefaultWriteTimeout = 3 * time.Second
	// DefaultPoolSize 默认连接池大小
	DefaultPoolSize = 10
	// DefaultMinIdleConns 默认最小空闲连接数
	DefaultMinIdleConns = 5
)

// RedisPoolConfig Redis连接池统一配置
const (
	// RedisDefaultPoolSize Redis默认连接池大小
	RedisDefaultPoolSize = 10
	// RedisDefaultMinIdleConns Redis默认最小空闲连接数
	RedisDefaultMinIdleConns = 5
	// RedisDefaultDialTimeout Redis默认连接超时时间
	RedisDefaultDialTimeout = 5 * time.Second
	// RedisDefaultReadTimeout Redis默认读取超时时间
	RedisDefaultReadTimeout = 3 * time.Second
	// RedisDefaultWriteTimeout Redis默认写入超时时间
	RedisDefaultWriteTimeout = 3 * time.Second
)

// NewDefaultPoolConfig 创建默认的数据库连接池配置
func NewDefaultPoolConfig() *PoolConfig {
	return &PoolConfig{
		MaxOpenConns:    DefaultMaxOpenConns,
		MaxIdleConns:    DefaultMaxIdleConns,
		ConnMaxLifetime: DefaultConnMaxLifetime,
		ConnMaxIdleTime: DefaultConnMaxIdleTime,
	}
}

// PoolConfig 数据库连接池配置结构体
type PoolConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// RedisConfig Redis连接配置结构体
type RedisConfig struct {
	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// NewDefaultRedisConfig 创建默认的Redis连接配置
func NewDefaultRedisConfig() *RedisConfig {
	return &RedisConfig{
		PoolSize:     RedisDefaultPoolSize,
		MinIdleConns: RedisDefaultMinIdleConns,
		DialTimeout:  RedisDefaultDialTimeout,
		ReadTimeout:  RedisDefaultReadTimeout,
		WriteTimeout: RedisDefaultWriteTimeout,
	}
}