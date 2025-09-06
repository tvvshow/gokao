package config

import (
	"os"
	"strconv"
	"time"
)

// Config 应用配置结构
type Config struct {
	// 服务配置
	Port        string `json:"port"`
	Environment string `json:"environment"`

	// 数据库配置
	DatabaseURL        string `json:"database_url"`
	MaxOpenConns       int    `json:"max_open_conns"`
	MaxIdleConns       int    `json:"max_idle_conns"`
	ConnMaxLifetime   int    `json:"conn_max_lifetime"`
	ConnMaxIdleTime   int    `json:"conn_max_idle_time"`

	// Redis配置
	RedisURL      string `json:"redis_url"`
	RedisPassword string `json:"redis_password"`
	RedisDB       int    `json:"redis_db"`

	// JWT配置
	JWTSecret         string        `json:"jwt_secret"`
	JWTExpiration     time.Duration `json:"jwt_expiration"`
	RefreshExpiration time.Duration `json:"refresh_expiration"`

	// Swagger配置
	EnableSwagger bool `json:"enable_swagger"`

	// 安全配置
	BcryptCost       int           `json:"bcrypt_cost"`
	MaxLoginAttempts int           `json:"max_login_attempts"`
	LockoutDuration  time.Duration `json:"lockout_duration"`

	// 审计配置
	EnableAudit   bool   `json:"enable_audit"`
	AuditLogLevel string `json:"audit_log_level"`

	// 设备认证服务配置
	DeviceAuthURL string `json:"device_auth_url"`
}

// Load 加载配置
func Load() *Config {
	return &Config{
		// 服务配置
		Port:        getEnv("PORT", "10081"),
		Environment: getEnv("GIN_MODE", "debug"),

		// 数据库配置
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/gaokao_users?sslmode=disable"),
		MaxOpenConns:     getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:     getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime:  getEnvAsInt("DB_CONN_MAX_LIFETIME", 1800), // 30分钟
		ConnMaxIdleTime: getEnvAsInt("DB_CONN_MAX_IDLE_TIME", 900),  // 15分钟

		// Redis配置
		RedisURL:      getEnv("REDIS_URL", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),

		// JWT配置
		JWTSecret:         getEnv("JWT_SECRET", ""),
		JWTExpiration:     getEnvAsDuration("JWT_EXPIRATION", "15m"),
		RefreshExpiration: getEnvAsDuration("REFRESH_EXPIRATION", "7d"),

		// Swagger配置
		EnableSwagger: getEnvAsBool("ENABLE_SWAGGER", true),

		// 安全配置
		BcryptCost:       getEnvAsInt("BCRYPT_COST", 12),
		MaxLoginAttempts: getEnvAsInt("MAX_LOGIN_ATTEMPTS", 5),
		LockoutDuration:  getEnvAsDuration("LOCKOUT_DURATION", "15m"),

		// 审计配置
		EnableAudit:   getEnvAsBool("ENABLE_AUDIT", true),
		AuditLogLevel: getEnv("AUDIT_LOG_LEVEL", "info"),

		// 设备认证服务配置
		DeviceAuthURL: getEnv("DEVICE_AUTH_URL", "http://localhost:8085"),
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取环境变量并转换为整数
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool 获取环境变量并转换为布尔值
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getEnvAsDuration 获取环境变量并转换为时间间隔
func getEnvAsDuration(key, defaultValue string) time.Duration {
	value := getEnv(key, defaultValue)
	if duration, err := time.ParseDuration(value); err == nil {
		return duration
	}
	// 如果解析失败，尝试解析为秒数
	if seconds, err := strconv.Atoi(value); err == nil {
		return time.Duration(seconds) * time.Second
	}
	// 最后的默认值
	if duration, err := time.ParseDuration(defaultValue); err == nil {
		return duration
	}
	return 15 * time.Minute
}