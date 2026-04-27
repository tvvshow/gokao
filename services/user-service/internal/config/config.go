package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config 应用配置结构
type Config struct {
	Port        string `json:"port"`
	Environment string `json:"environment"`

	DatabaseURL     string `json:"database_url"`
	MaxOpenConns    int    `json:"max_open_conns"`
	MaxIdleConns    int    `json:"max_idle_conns"`
	ConnMaxLifetime int    `json:"conn_max_lifetime"`
	ConnMaxIdleTime int    `json:"conn_max_idle_time"`

	RedisURL      string `json:"redis_url"`
	RedisPassword string `json:"redis_password"`
	RedisDB       int    `json:"redis_db"`

	JWTSecret         string        `json:"jwt_secret"`
	JWTExpiration     time.Duration `json:"jwt_expiration"`
	RefreshExpiration time.Duration `json:"refresh_expiration"`

	EnableSwagger bool `json:"enable_swagger"`

	BcryptCost       int           `json:"bcrypt_cost"`
	MaxLoginAttempts int           `json:"max_login_attempts"`
	LockoutDuration  time.Duration `json:"lockout_duration"`

	EnableAudit   bool   `json:"enable_audit"`
	AuditLogLevel string `json:"audit_log_level"`

	DeviceAuthURL string `json:"device_auth_url"`
}

// Load 加载配置
func Load() *Config {
	return &Config{
		Port:        defaultString(normalizePortValue(firstNonEmptyEnv("SERVER_PORT", "PORT")), "8083"),
		Environment: defaultString(firstNonEmptyEnv("SERVER_MODE", "GIN_MODE"), "debug"),

		DatabaseURL:     getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/gaokao_users?sslmode=disable"),
		MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: getEnvAsInt("DB_CONN_MAX_LIFETIME", 1800),
		ConnMaxIdleTime: getEnvAsInt("DB_CONN_MAX_IDLE_TIME", 900),

		RedisURL:      getEnv("REDIS_URL", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),

		JWTSecret:         getEnv("JWT_SECRET", ""),
		JWTExpiration:     getEnvAsDuration("JWT_EXPIRATION", "15m"),
		RefreshExpiration: getEnvAsDuration("REFRESH_EXPIRATION", "7d"),

		EnableSwagger: getEnvAsBool("ENABLE_SWAGGER", true),

		BcryptCost:       getEnvAsInt("BCRYPT_COST", 12),
		MaxLoginAttempts: getEnvAsInt("MAX_LOGIN_ATTEMPTS", 5),
		LockoutDuration:  getEnvAsDuration("LOCKOUT_DURATION", "15m"),

		EnableAudit:   getEnvAsBool("ENABLE_AUDIT", true),
		AuditLogLevel: getEnv("AUDIT_LOG_LEVEL", "info"),

		DeviceAuthURL: getEnv("DEVICE_AUTH_URL", "http://localhost:8085"),
	}
}

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func normalizePortValue(port string) string {
	port = strings.TrimSpace(port)
	if strings.HasPrefix(port, ":") {
		return port[1:]
	}
	return port
}

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

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key, defaultValue string) time.Duration {
	value := getEnv(key, defaultValue)
	if duration, err := time.ParseDuration(value); err == nil {
		return duration
	}
	if seconds, err := strconv.Atoi(value); err == nil {
		return time.Duration(seconds) * time.Second
	}
	if duration, err := time.ParseDuration(defaultValue); err == nil {
		return duration
	}
	return 15 * time.Minute
}
