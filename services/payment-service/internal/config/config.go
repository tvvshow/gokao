package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

// Config 应用配置
type Config struct {
	Server    ServerConfig    `json:"server"`
	Database  DatabaseConfig  `json:"database"`
	Redis     RedisConfig     `json:"redis"`
	JWT       JWTConfig       `json:"jwt"`
	Payment   PaymentConfig   `json:"payment"`
	RateLimit RateLimitConfig `json:"rate_limit"`
	License   LicenseConfig   `json:"license"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         int    `json:"port"`
	Mode         string `json:"mode"`
	ReadTimeout  int    `json:"read_timeout"`
	WriteTimeout int    `json:"write_timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string `json:"host"`
	Port            int    `json:"port"`
	User            string `json:"user"`
	Password        string `json:"password"`
	Database        string `json:"database"`
	SSLMode         string `json:"ssl_mode"`
	MaxOpenConns    int    `json:"max_open_conns"`
	MaxIdleConns    int    `json:"max_idle_conns"`
	ConnMaxLifetime int    `json:"conn_max_lifetime"`
	ConnMaxIdleTime int    `json:"conn_max_idle_time"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret      string `json:"secret"`
	ExpireHours int    `json:"expire_hours"`
}

// PaymentConfig 支付配置
type PaymentConfig struct {
	Alipay   AlipayConfig   `json:"alipay"`
	WeChat   WeChatConfig   `json:"wechat"`
	UnionPay UnionPayConfig `json:"unionpay"`
}

// AlipayConfig 支付宝配置
type AlipayConfig struct {
	AppID      string `json:"app_id"`
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	NotifyURL  string `json:"notify_url"`
	ReturnURL  string `json:"return_url"`
	SignType   string `json:"sign_type"`
	Sandbox    bool   `json:"sandbox"`
	EncryptKey string `json:"encrypt_key"`
}

// WeChatConfig 微信支付配置
type WeChatConfig struct {
	AppID     string `json:"app_id"`
	MchID     string `json:"mch_id"`
	APIKey    string `json:"api_key"`
	CertPath  string `json:"cert_path"`
	KeyPath   string `json:"key_path"`
	NotifyURL string `json:"notify_url"`
	Sandbox   bool   `json:"sandbox"`
}

// UnionPayConfig 银联支付配置
type UnionPayConfig struct {
	MerchantID string `json:"merchant_id"`
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	NotifyURL  string `json:"notify_url"`
	ReturnURL  string `json:"return_url"`
	Sandbox    bool   `json:"sandbox"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	RPS    int  `json:"rps"`
	Burst  int  `json:"burst"`
	Enable bool `json:"enable"`
}

// LicenseConfig 许可证配置
type LicenseConfig struct {
	LibraryPath string `json:"library_path"`
	PublicKey   string `json:"public_key"`
	PrivateKey  string `json:"private_key"`
	EnableCheck bool   `json:"enable_check"`
}

// Load 加载配置
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:         getEnvAsIntAlias([]string{"SERVER_PORT", "PORT"}, 8085),
			Mode:         getEnvAlias([]string{"SERVER_MODE", "GIN_MODE"}, "debug"),
			ReadTimeout:  getEnvAsInt("READ_TIMEOUT", 10),
			WriteTimeout: getEnvAsInt("WRITE_TIMEOUT", 10),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "password"),
			Database:        getEnv("DB_NAME", "gaokao_users"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsInt("DB_CONN_MAX_LIFETIME", 1800),
			ConnMaxIdleTime: getEnvAsInt("DB_CONN_MAX_IDLE_TIME", 900),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 2),
		},
		JWT: JWTConfig{
			Secret:      getEnv("JWT_SECRET", ""),
			ExpireHours: getEnvAsInt("JWT_EXPIRE_HOURS", 24),
		},
		Payment: PaymentConfig{
			Alipay: AlipayConfig{
				AppID:      getEnv("ALIPAY_APP_ID", ""),
				PrivateKey: getEnv("ALIPAY_PRIVATE_KEY", ""),
				PublicKey:  getEnv("ALIPAY_PUBLIC_KEY", ""),
				NotifyURL:  getEnv("ALIPAY_NOTIFY_URL", ""),
				ReturnURL:  getEnv("ALIPAY_RETURN_URL", ""),
				SignType:   getEnv("ALIPAY_SIGN_TYPE", "RSA2"),
				Sandbox:    getEnvAsBool("ALIPAY_SANDBOX", true),
			},
			WeChat: WeChatConfig{
				AppID:     getEnv("WECHAT_APP_ID", ""),
				MchID:     getEnv("WECHAT_MCH_ID", ""),
				APIKey:    getEnv("WECHAT_API_KEY", ""),
				CertPath:  getEnv("WECHAT_CERT_PATH", ""),
				KeyPath:   getEnv("WECHAT_KEY_PATH", ""),
				NotifyURL: getEnv("WECHAT_NOTIFY_URL", ""),
				Sandbox:   getEnvAsBool("WECHAT_SANDBOX", true),
			},
			UnionPay: UnionPayConfig{
				MerchantID: getEnv("UNIONPAY_MERCHANT_ID", ""),
				PrivateKey: getEnv("UNIONPAY_PRIVATE_KEY", ""),
				PublicKey:  getEnv("UNIONPAY_PUBLIC_KEY", ""),
				NotifyURL:  getEnv("UNIONPAY_NOTIFY_URL", ""),
				ReturnURL:  getEnv("UNIONPAY_RETURN_URL", ""),
				Sandbox:    getEnvAsBool("UNIONPAY_SANDBOX", true),
			},
		},
		RateLimit: RateLimitConfig{
			RPS:    getEnvAsInt("RATE_LIMIT_RPS", 100),
			Burst:  getEnvAsInt("RATE_LIMIT_BURST", 200),
			Enable: getEnvAsBool("RATE_LIMIT_ENABLE", true),
		},
		License: LicenseConfig{
			LibraryPath: getEnv("LICENSE_LIBRARY_PATH", "./cpp-modules/license/liblicense.so"),
			PublicKey:   getEnv("LICENSE_PUBLIC_KEY", ""),
			PrivateKey:  getEnv("LICENSE_PRIVATE_KEY", ""),
			EnableCheck: getEnvAsBool("LICENSE_ENABLE_CHECK", true),
		},
	}

	// 如果存在配置文件，则加载配置文件
	if configFile := getEnv("CONFIG_FILE", ""); configFile != "" {
		if err := loadFromFile(cfg, configFile); err != nil {
			return nil, fmt.Errorf("failed to load config from file: %w", err)
		}
	}

	return cfg, nil
}

// loadFromFile 从文件加载配置
func loadFromFile(cfg *Config, filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, cfg)
}

func getEnvAlias(keys []string, defaultValue string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return defaultValue
}

func getEnvAsIntAlias(keys []string, defaultValue int) int {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			if parsed, err := strconv.Atoi(value); err == nil {
				return parsed
			}
		}
	}
	return defaultValue
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取环境变量作为整数
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}

	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}

	return defaultValue
}

// getEnvAsBool 获取环境变量作为布尔值
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}

	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}

	return defaultValue
}
