// Package config 包装 user-service 自有配置。
//
// 通用字段（Server / Database / Redis / Audit）通过嵌入 pkg/config 的子结构复用，
// 仅保留 user-service 特有字段（JWT、Bcrypt、登录限流、设备认证）。
//
// 嵌入的字段通过 Go 的字段提升机制对外可见，所以现有 cfg.Port / cfg.DatabaseURL / cfg.EnableAudit
// 等访问继续生效，不需要改调用点。
package config

import (
	"time"

	sharedcfg "github.com/tvvshow/gokao/pkg/config"
)

// Config 应用配置结构。
type Config struct {
	sharedcfg.ServerConfig
	sharedcfg.DatabaseConfig
	sharedcfg.RedisConfig
	sharedcfg.AuditConfig

	// 认证 / 会话
	JWTSecret         string        `json:"jwt_secret"`
	JWTExpiration     time.Duration `json:"jwt_expiration"`
	RefreshExpiration time.Duration `json:"refresh_expiration"`

	// 密码 + 登录保护
	BcryptCost       int           `json:"bcrypt_cost"`
	MaxLoginAttempts int           `json:"max_login_attempts"`
	LockoutDuration  time.Duration `json:"lockout_duration"`

	// 设备认证
	DeviceAuthURL string `json:"device_auth_url"`
}

// Load 加载配置。
func Load() *Config {
	return &Config{
		ServerConfig:   sharedcfg.LoadServer("8083", "ENABLE_SWAGGER"),
		DatabaseConfig: sharedcfg.LoadDatabase("postgres://postgres:password@localhost:5432/gaokao_users?sslmode=disable"),
		RedisConfig:    sharedcfg.LoadRedis("", 0),
		AuditConfig:    sharedcfg.LoadAudit(),

		JWTSecret:         sharedcfg.GetEnv("JWT_SECRET", ""),
		JWTExpiration:     sharedcfg.GetEnvAsDuration("JWT_EXPIRATION", "15m"),
		RefreshExpiration: sharedcfg.GetEnvAsDuration("REFRESH_EXPIRATION", "7d"),

		BcryptCost:       sharedcfg.GetEnvAsInt("BCRYPT_COST", 12),
		MaxLoginAttempts: sharedcfg.GetEnvAsInt("MAX_LOGIN_ATTEMPTS", 5),
		LockoutDuration:  sharedcfg.GetEnvAsDuration("LOCKOUT_DURATION", "15m"),

		DeviceAuthURL: sharedcfg.GetEnv("DEVICE_AUTH_URL", "http://localhost:8085"),
	}
}
