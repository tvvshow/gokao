package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"time"
)

// SecurityConfig 安全配置
type SecurityConfig struct {
	// JWT配置
	JWTSecret             string `json:"jwt_secret"`
	JWTExpireHours        int    `json:"jwt_expire_hours"`
	JWTRefreshExpireHours int    `json:"jwt_refresh_expire_hours"`

	// 加密配置
	AESEncryptionKey  string `json:"aes_encryption_key"`
	RSAPrivateKeyPath string `json:"rsa_private_key_path"`
	RSAPublicKeyPath  string `json:"rsa_public_key_path"`

	// 会话配置
	SessionSecret string `json:"session_secret"`
	CSRFSecret    string `json:"csrf_secret"`

	// 密码策略
	PasswordMinLength      int  `json:"password_min_length"`
	PasswordRequireSpecial bool `json:"password_require_special"`
	PasswordRequireNumber  bool `json:"password_require_number"`
	PasswordRequireUpper   bool `json:"password_require_upper"`
	PasswordRequireLower   bool `json:"password_require_lower"`

	// 账户安全
	MaxLoginAttempts int           `json:"max_login_attempts"`
	LockoutDuration  time.Duration `json:"lockout_duration"`
	SessionTimeout   time.Duration `json:"session_timeout"`

	// 速率限制
	RateLimitEnabled   bool `json:"rate_limit_enabled"`
	RateLimitPerMinute int  `json:"rate_limit_per_minute"`
	RateLimitBurst     int  `json:"rate_limit_burst"`

	// 安全头
	SecurityHeadersEnabled bool   `json:"security_headers_enabled"`
	HSTSMaxAge             int    `json:"hsts_max_age"`
	CSPPolicy              string `json:"csp_policy"`

	// 审计日志
	AuditLogEnabled   bool `json:"audit_log_enabled"`
	AuditLogRetention int  `json:"audit_log_retention_days"`
}

// LoadSecurityConfig 加载安全配置
func LoadSecurityConfig() (*SecurityConfig, error) {
	config := &SecurityConfig{
		// JWT默认配置
		JWTExpireHours:        getSecurityEnvAsInt("JWT_EXPIRE_HOURS", 24),
		JWTRefreshExpireHours: getSecurityEnvAsInt("JWT_REFRESH_EXPIRE_HOURS", 168),

		// 密码策略默认配置
		PasswordMinLength:      getSecurityEnvAsInt("PASSWORD_MIN_LENGTH", 8),
		PasswordRequireSpecial: getSecurityEnvAsBool("PASSWORD_REQUIRE_SPECIAL", true),
		PasswordRequireNumber:  getSecurityEnvAsBool("PASSWORD_REQUIRE_NUMBER", true),
		PasswordRequireUpper:   getSecurityEnvAsBool("PASSWORD_REQUIRE_UPPER", true),
		PasswordRequireLower:   getSecurityEnvAsBool("PASSWORD_REQUIRE_LOWER", true),

		// 账户安全默认配置
		MaxLoginAttempts: getSecurityEnvAsInt("MAX_LOGIN_ATTEMPTS", 5),
		LockoutDuration:  time.Duration(getSecurityEnvAsInt("LOCKOUT_DURATION_MINUTES", 30)) * time.Minute,
		SessionTimeout:   time.Duration(getSecurityEnvAsInt("SESSION_TIMEOUT_HOURS", 24)) * time.Hour,

		// 速率限制默认配置
		RateLimitEnabled:   getSecurityEnvAsBool("RATE_LIMIT_ENABLED", true),
		RateLimitPerMinute: getSecurityEnvAsInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 60),
		RateLimitBurst:     getSecurityEnvAsInt("RATE_LIMIT_BURST", 10),

		// 安全头默认配置
		SecurityHeadersEnabled: getSecurityEnvAsBool("SECURITY_HEADERS_ENABLED", true),
		HSTSMaxAge:             getSecurityEnvAsInt("SECURITY_HSTS_MAX_AGE", 31536000),
		CSPPolicy:              getSecurityEnv("SECURITY_CSP_POLICY", "default-src 'self'; script-src 'self' 'unsafe-inline'"),

		// 审计日志默认配置
		AuditLogEnabled:   getSecurityEnvAsBool("AUDIT_LOG_ENABLED", true),
		AuditLogRetention: getSecurityEnvAsInt("AUDIT_LOG_RETENTION_DAYS", 90),
	}

	// 加载必需的密钥
	if err := config.loadSecrets(); err != nil {
		return nil, fmt.Errorf("failed to load security secrets: %w", err)
	}

	// 验证配置
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid security configuration: %w", err)
	}

	return config, nil
}

// loadSecrets 加载安全密钥
func (c *SecurityConfig) loadSecrets() error {
	// 加载JWT密钥
	jwtSecret := getSecurityEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		return fmt.Errorf("JWT_SECRET environment variable is required")
	}
	if len(jwtSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters long")
	}
	c.JWTSecret = jwtSecret

	// 加载AES加密密钥
	aesKey := getSecurityEnv("AES_ENCRYPTION_KEY", "")
	if aesKey == "" {
		return fmt.Errorf("AES_ENCRYPTION_KEY environment variable is required")
	}
	if len(aesKey) < 32 {
		return fmt.Errorf("AES_ENCRYPTION_KEY must be at least 32 characters long")
	}
	c.AESEncryptionKey = aesKey

	// 加载RSA密钥路径
	c.RSAPrivateKeyPath = getSecurityEnv("RSA_PRIVATE_KEY_PATH", "")
	c.RSAPublicKeyPath = getSecurityEnv("RSA_PUBLIC_KEY_PATH", "")

	// 加载会话密钥
	sessionSecret := getSecurityEnv("SESSION_SECRET", "")
	if sessionSecret == "" {
		// 生成临时会话密钥（开发环境）
		if getSecurityEnv("ENVIRONMENT", "development") == "development" {
			sessionSecret = generateRandomKey(64)
		} else {
			return fmt.Errorf("SESSION_SECRET environment variable is required in production")
		}
	}
	c.SessionSecret = sessionSecret

	// 加载CSRF密钥
	csrfSecret := getSecurityEnv("CSRF_SECRET", "")
	if csrfSecret == "" {
		// 生成临时CSRF密钥（开发环境）
		if getSecurityEnv("ENVIRONMENT", "development") == "development" {
			csrfSecret = generateRandomKey(32)
		} else {
			return fmt.Errorf("CSRF_SECRET environment variable is required in production")
		}
	}
	c.CSRFSecret = csrfSecret

	return nil
}

// validate 验证安全配置
func (c *SecurityConfig) validate() error {
	// 验证JWT配置
	if c.JWTExpireHours <= 0 || c.JWTExpireHours > 168 {
		return fmt.Errorf("JWT expire hours must be between 1 and 168")
	}

	if c.JWTRefreshExpireHours <= c.JWTExpireHours {
		return fmt.Errorf("JWT refresh expire hours must be greater than JWT expire hours")
	}

	// 验证密码策略
	if c.PasswordMinLength < 6 || c.PasswordMinLength > 128 {
		return fmt.Errorf("password minimum length must be between 6 and 128")
	}

	// 验证账户安全配置
	if c.MaxLoginAttempts <= 0 || c.MaxLoginAttempts > 20 {
		return fmt.Errorf("max login attempts must be between 1 and 20")
	}

	if c.LockoutDuration < time.Minute || c.LockoutDuration > 24*time.Hour {
		return fmt.Errorf("lockout duration must be between 1 minute and 24 hours")
	}

	// 验证速率限制配置
	if c.RateLimitEnabled {
		if c.RateLimitPerMinute <= 0 || c.RateLimitPerMinute > 10000 {
			return fmt.Errorf("rate limit per minute must be between 1 and 10000")
		}

		if c.RateLimitBurst <= 0 || c.RateLimitBurst > c.RateLimitPerMinute {
			return fmt.Errorf("rate limit burst must be between 1 and rate limit per minute")
		}
	}

	return nil
}

// IsProduction 检查是否为生产环境
func (c *SecurityConfig) IsProduction() bool {
	return getSecurityEnv("ENVIRONMENT", "development") == "production"
}

// GetJWTExpireDuration 获取JWT过期时间
func (c *SecurityConfig) GetJWTExpireDuration() time.Duration {
	return time.Duration(c.JWTExpireHours) * time.Hour
}

// GetJWTRefreshExpireDuration 获取JWT刷新过期时间
func (c *SecurityConfig) GetJWTRefreshExpireDuration() time.Duration {
	return time.Duration(c.JWTRefreshExpireHours) * time.Hour
}

// 工具函数

// getSecurityEnv 获取环境变量，如果不存在则返回默认值
func getSecurityEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getSecurityEnvAsInt 获取环境变量并转换为整数
func getSecurityEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getSecurityEnvAsBool 获取环境变量并转换为布尔值
func getSecurityEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// generateRandomKey 生成随机密钥
func generateRandomKey(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		panic(fmt.Sprintf("failed to generate random key: %v", err))
	}
	return base64.URLEncoding.EncodeToString(bytes)
}
