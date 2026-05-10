// Package config 包装 data-service 自有配置。
//
// 通用字段（Server / Database / Redis / Audit）通过嵌入 pkg/config 的子结构复用，
// data-service 特有字段：Elasticsearch、缓存策略、分页与查询超时。
package config

import (
	"time"

	sharedcfg "github.com/tvvshow/gokao/pkg/config"
)

// Config 数据服务配置结构。
type Config struct {
	sharedcfg.ServerConfig
	sharedcfg.DatabaseConfig
	sharedcfg.RedisConfig
	sharedcfg.AuditConfig

	// Elasticsearch
	ElasticsearchURL      string `json:"elasticsearch_url"`
	ElasticsearchUsername string `json:"elasticsearch_username"`
	ElasticsearchPassword string `json:"elasticsearch_password"`

	// 缓存策略
	CacheEnabled    bool          `json:"cache_enabled"`
	CacheDefaultTTL time.Duration `json:"cache_default_ttl"`

	// 分页 / 查询
	MaxPageSize     int           `json:"max_page_size"`
	DefaultPageSize int           `json:"default_page_size"`
	QueryTimeout    time.Duration `json:"query_timeout"`
}

func Load() *Config {
	return &Config{
		ServerConfig:   sharedcfg.LoadServer("8082", "ENABLE_SWAGGER"),
		DatabaseConfig: sharedcfg.LoadDatabase("postgres://postgres:password@localhost:5432/gaokao_data?sslmode=disable"),
		RedisConfig:    sharedcfg.LoadRedis("", 1), // data-service 用 db=1 与其它服务隔离
		AuditConfig:    sharedcfg.LoadAudit(),

		ElasticsearchURL:      sharedcfg.GetEnv("ELASTICSEARCH_URL", "http://localhost:9200"),
		ElasticsearchUsername: sharedcfg.GetEnv("ELASTICSEARCH_USERNAME", ""),
		ElasticsearchPassword: sharedcfg.GetEnv("ELASTICSEARCH_PASSWORD", ""),

		CacheEnabled:    sharedcfg.GetEnvAsBool("CACHE_ENABLED", true),
		CacheDefaultTTL: sharedcfg.GetEnvAsDuration("CACHE_DEFAULT_TTL", "5m"),

		MaxPageSize:     sharedcfg.GetEnvAsInt("MAX_PAGE_SIZE", 100),
		DefaultPageSize: sharedcfg.GetEnvAsInt("DEFAULT_PAGE_SIZE", 20),
		QueryTimeout:    sharedcfg.GetEnvAsDuration("QUERY_TIMEOUT", "30s"),
	}
}
