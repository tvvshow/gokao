package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config 数据服务配置结构
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

	ElasticsearchURL      string `json:"elasticsearch_url"`
	ElasticsearchUsername string `json:"elasticsearch_username"`
	ElasticsearchPassword string `json:"elasticsearch_password"`

	EnableSwagger bool `json:"enable_swagger"`

	CacheEnabled    bool          `json:"cache_enabled"`
	CacheDefaultTTL time.Duration `json:"cache_default_ttl"`

	MaxPageSize     int           `json:"max_page_size"`
	DefaultPageSize int           `json:"default_page_size"`
	QueryTimeout    time.Duration `json:"query_timeout"`

	EnableAudit   bool   `json:"enable_audit"`
	AuditLogLevel string `json:"audit_log_level"`
}

func Load() *Config {
	return &Config{
		Port:        defaultString(normalizePortValue(firstNonEmptyEnv("SERVER_PORT", "PORT")), "8082"),
		Environment: defaultString(firstNonEmptyEnv("SERVER_MODE", "GIN_MODE"), "debug"),

		DatabaseURL:            getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/gaokao_data?sslmode=disable"),
		MaxOpenConns:           getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:           getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime:        getEnvAsInt("DB_CONN_MAX_LIFETIME", 1800),
		ConnMaxIdleTime:        getEnvAsInt("DB_CONN_MAX_IDLE_TIME", 900),
		RedisURL:               getEnv("REDIS_URL", "localhost:6379"),
		RedisPassword:          getEnv("REDIS_PASSWORD", ""),
		RedisDB:                getEnvAsInt("REDIS_DB", 1),
		ElasticsearchURL:       getEnv("ELASTICSEARCH_URL", "http://localhost:9200"),
		ElasticsearchUsername:  getEnv("ELASTICSEARCH_USERNAME", ""),
		ElasticsearchPassword:  getEnv("ELASTICSEARCH_PASSWORD", ""),
		EnableSwagger:          getEnvAsBool("ENABLE_SWAGGER", true),
		CacheEnabled:           getEnvAsBool("CACHE_ENABLED", true),
		CacheDefaultTTL:        getEnvAsDuration("CACHE_DEFAULT_TTL", "5m"),
		MaxPageSize:            getEnvAsInt("MAX_PAGE_SIZE", 100),
		DefaultPageSize:        getEnvAsInt("DEFAULT_PAGE_SIZE", 20),
		QueryTimeout:           getEnvAsDuration("QUERY_TIMEOUT", "30s"),
		EnableAudit:            getEnvAsBool("ENABLE_AUDIT", true),
		AuditLogLevel:          getEnv("AUDIT_LOG_LEVEL", "info"),
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

func normalizePortValue(port string) string {
	port = strings.TrimSpace(port)
	if strings.HasPrefix(port, ":") {
		return port[1:]
	}
	return port
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
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
	return 30 * time.Second
}
