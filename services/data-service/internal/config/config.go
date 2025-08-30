package config

import (
	"os"
	"strconv"
	"time"
)

// Config 数据服务配置结构
type Config struct {
	// 服务配置
	Port        string `json:"port"`
	Environment string `json:"environment"`

	// 数据库配置
	DatabaseURL string `json:"database_url"`

	// Redis配置
	RedisURL      string `json:"redis_url"`
	RedisPassword string `json:"redis_password"`
	RedisDB       int    `json:"redis_db"`

	// Elasticsearch配置
	ElasticsearchURL      string `json:"elasticsearch_url"`
	ElasticsearchUsername string `json:"elasticsearch_username"`
	ElasticsearchPassword string `json:"elasticsearch_password"`

	// Swagger配置
	EnableSwagger bool `json:"enable_swagger"`

	// 缓存配置
	CacheEnabled    bool          `json:"cache_enabled"`
	CacheDefaultTTL time.Duration `json:"cache_default_ttl"`

	// 性能配置
	MaxPageSize     int `json:"max_page_size"`
	DefaultPageSize int `json:"default_page_size"`
	QueryTimeout    time.Duration `json:"query_timeout"`

	// C++算法引擎配置
	AlgorithmEngineEnabled bool   `json:"algorithm_engine_enabled"`
	AlgorithmEngineAddr    string `json:"algorithm_engine_addr"`

	// 审计配置
	EnableAudit   bool   `json:"enable_audit"`
	AuditLogLevel string `json:"audit_log_level"`
}

// Load 加载配置
func Load() *Config {
	return &Config{
		// 服务配置
		Port:        getEnv("PORT", "8082"),
		Environment: getEnv("GIN_MODE", "debug"),

		// 数据库配置
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/gaokao_data?sslmode=disable"),

		// Redis配置
		RedisURL:      getEnv("REDIS_URL", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 1),

		// Elasticsearch配置
		ElasticsearchURL:      getEnv("ELASTICSEARCH_URL", "http://localhost:9200"),
		ElasticsearchUsername: getEnv("ELASTICSEARCH_USERNAME", ""),
		ElasticsearchPassword: getEnv("ELASTICSEARCH_PASSWORD", ""),

		// Swagger配置
		EnableSwagger: getEnvAsBool("ENABLE_SWAGGER", true),

		// 缓存配置
		CacheEnabled:    getEnvAsBool("CACHE_ENABLED", true),
		CacheDefaultTTL: getEnvAsDuration("CACHE_DEFAULT_TTL", "5m"),

		// 性能配置
		MaxPageSize:     getEnvAsInt("MAX_PAGE_SIZE", 100),
		DefaultPageSize: getEnvAsInt("DEFAULT_PAGE_SIZE", 20),
		QueryTimeout:    getEnvAsDuration("QUERY_TIMEOUT", "30s"),

		// C++算法引擎配置
		AlgorithmEngineEnabled: getEnvAsBool("ALGORITHM_ENGINE_ENABLED", true),
		AlgorithmEngineAddr:    getEnv("ALGORITHM_ENGINE_ADDR", "localhost:50051"),

		// 审计配置
		EnableAudit:   getEnvAsBool("ENABLE_AUDIT", true),
		AuditLogLevel: getEnv("AUDIT_LOG_LEVEL", "info"),
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
	return 30 * time.Second
}