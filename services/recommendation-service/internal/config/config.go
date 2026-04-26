package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Config 应用配置
type Config struct {
	Server      *ServerConfig      `json:"server"`
	CPP         *CPPConfig         `json:"cpp"`
	Redis       *RedisConfig       `json:"redis"`
	Log         *LogConfig         `json:"log"`
	DataService *DataServiceConfig `json:"data_service"`
	LLM         *LLMConfig         `json:"llm"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string `json:"port"`
	Mode string `json:"mode"`
}

// CPPConfig C++配置
type CPPConfig struct {
	ConfigPath      string `json:"config_path"`
	LibraryPath     string `json:"library_path"`
	MaxWorkers      int    `json:"max_workers"`
	CacheEnabled    bool   `json:"cache_enabled"`
	CacheSize       int    `json:"cache_size"`
	CacheTTLMinutes int    `json:"cache_ttl_minutes"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Enabled  bool   `json:"enabled"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `json:"level"`
	File       string `json:"file"`
	MaxSize    int    `json:"max_size"`
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age"`
}

// DataServiceConfig 数据服务配置
type DataServiceConfig struct {
	URL          string        `json:"url"`
	APIKey       string        `json:"api_key"`
	SyncInterval time.Duration `json:"sync_interval"`
}

// LLMConfig 大模型配置（兼容 OpenAI Chat Completions 风格接口）
type LLMConfig struct {
	Enabled         bool          `json:"enabled"`
	Provider        string        `json:"provider"`
	BaseURL         string        `json:"base_url"`
	APIKey          string        `json:"api_key"`
	Model           string        `json:"model"`
	Timeout         time.Duration `json:"timeout"`
	MaxTokens       int           `json:"max_tokens"`
	Temperature     float64       `json:"temperature"`
	FallbackEnabled bool          `json:"fallback_enabled"`
	SystemPrompt    string        `json:"system_prompt"`
}

var (
	instance *Config
	once     sync.Once
)

// Load 加载配置
func Load() (*Config, error) {
	var err error
	once.Do(func() {
		instance, err = loadConfig()
	})
	return instance, err
}

// loadConfig 加载配置文件
func loadConfig() (*Config, error) {
	config := defaultConfig()

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config/config.json"
	}

	if _, err := os.Stat(configPath); err == nil {
		file, err := os.Open(configPath)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		decoder := json.NewDecoder(file)
		if err := decoder.Decode(config); err != nil {
			return nil, err
		}
	}

	loadFromEnv(config)
	ensureDirs(config)

	return config, nil
}

func defaultConfig() *Config {
	return &Config{
		Server: &ServerConfig{
			Port: "10083",
			Mode: gin.ReleaseMode,
		},
		CPP: &CPPConfig{
			ConfigPath:      "./config/hybrid_config.json",
			LibraryPath:     "/usr/local/lib/libvolunteer_matcher.so",
			MaxWorkers:      4,
			CacheEnabled:    true,
			CacheSize:       1000,
			CacheTTLMinutes: 30,
		},
		Redis: &RedisConfig{
			Enabled:  false,
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
		},
		Log: &LogConfig{
			Level:      "info",
			File:       "logs/recommendation-service.log",
			MaxSize:    100,
			MaxBackups: 10,
			MaxAge:     30,
		},
		DataService: &DataServiceConfig{
			URL:          "http://localhost:10081",
			APIKey:       "",
			SyncInterval: 30 * time.Minute,
		},
		LLM: &LLMConfig{
			Enabled:         false,
			Provider:        "openai-compatible",
			BaseURL:         "https://api.openai.com/v1",
			APIKey:          "",
			Model:           "gpt-4o-mini",
			Timeout:         15 * time.Second,
			MaxTokens:       800,
			Temperature:     0.3,
			FallbackEnabled: true,
			SystemPrompt:    "你是一名高考志愿分析助手。请基于学生分数、地区偏好、风险偏好和推荐结果，输出简洁、专业、可执行的中文分析。",
		},
	}
}

// loadFromEnv 从环境变量加载配置
func loadFromEnv(config *Config) {
	if config.Server == nil {
		config.Server = defaultConfig().Server
	}
	if config.CPP == nil {
		config.CPP = defaultConfig().CPP
	}
	if config.Redis == nil {
		config.Redis = defaultConfig().Redis
	}
	if config.Log == nil {
		config.Log = defaultConfig().Log
	}
	if config.DataService == nil {
		config.DataService = defaultConfig().DataService
	}
	if config.LLM == nil {
		config.LLM = defaultConfig().LLM
	}

	if port := os.Getenv("SERVER_PORT"); port != "" {
		config.Server.Port = port
	}
	if mode := os.Getenv("SERVER_MODE"); mode != "" {
		config.Server.Mode = mode
	}

	if path := os.Getenv("CPP_CONFIG_PATH"); path != "" {
		config.CPP.ConfigPath = path
	}
	if path := os.Getenv("CPP_LIBRARY_PATH"); path != "" {
		config.CPP.LibraryPath = path
	}
	if workers, ok := getEnvInt("CPP_MAX_WORKERS"); ok {
		config.CPP.MaxWorkers = workers
	}
	if enabled, ok := getEnvBool("CPP_CACHE_ENABLED"); ok {
		config.CPP.CacheEnabled = enabled
	}
	if size, ok := getEnvInt("CPP_CACHE_SIZE"); ok {
		config.CPP.CacheSize = size
	}
	if ttl, ok := getEnvInt("CPP_CACHE_TTL_MINUTES"); ok {
		config.CPP.CacheTTLMinutes = ttl
	}

	if enabled, ok := getEnvBool("REDIS_ENABLED"); ok {
		config.Redis.Enabled = enabled
	}
	if host := os.Getenv("REDIS_HOST"); host != "" {
		config.Redis.Host = host
	}
	if port, ok := getEnvInt("REDIS_PORT"); ok {
		config.Redis.Port = port
	}
	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		config.Redis.Password = password
	}
	if db, ok := getEnvInt("REDIS_DB"); ok {
		config.Redis.DB = db
	}

	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.Log.Level = level
	}
	if file := os.Getenv("LOG_FILE"); file != "" {
		config.Log.File = file
	}
	if size, ok := getEnvInt("LOG_MAX_SIZE"); ok {
		config.Log.MaxSize = size
	}
	if backups, ok := getEnvInt("LOG_MAX_BACKUPS"); ok {
		config.Log.MaxBackups = backups
	}
	if age, ok := getEnvInt("LOG_MAX_AGE"); ok {
		config.Log.MaxAge = age
	}

	if url := os.Getenv("DATA_SERVICE_URL"); url != "" {
		config.DataService.URL = url
	}
	if apiKey := os.Getenv("DATA_SERVICE_API_KEY"); apiKey != "" {
		config.DataService.APIKey = apiKey
	}
	if interval, ok := getEnvDuration("DATA_SERVICE_SYNC_INTERVAL"); ok {
		config.DataService.SyncInterval = interval
	}

	if enabled, ok := getEnvBool("LLM_ENABLED"); ok {
		config.LLM.Enabled = enabled
	}
	if provider := os.Getenv("LLM_PROVIDER"); provider != "" {
		config.LLM.Provider = provider
	}
	if baseURL := os.Getenv("LLM_BASE_URL"); baseURL != "" {
		config.LLM.BaseURL = baseURL
	}
	if apiKey := os.Getenv("LLM_API_KEY"); apiKey != "" {
		config.LLM.APIKey = apiKey
	}
	if model := os.Getenv("LLM_MODEL"); model != "" {
		config.LLM.Model = model
	}
	if timeout, ok := getEnvDuration("LLM_TIMEOUT"); ok {
		config.LLM.Timeout = timeout
	}
	if maxTokens, ok := getEnvInt("LLM_MAX_TOKENS"); ok {
		config.LLM.MaxTokens = maxTokens
	}
	if temperature, ok := getEnvFloat("LLM_TEMPERATURE"); ok {
		config.LLM.Temperature = temperature
	}
	if enabled, ok := getEnvBool("LLM_FALLBACK_ENABLED"); ok {
		config.LLM.FallbackEnabled = enabled
	}
	if systemPrompt := os.Getenv("LLM_SYSTEM_PROMPT"); systemPrompt != "" {
		config.LLM.SystemPrompt = systemPrompt
	}
}

func getEnvInt(key string) (int, bool) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return 0, false
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, false
	}
	return parsed, true
}

func getEnvBool(key string) (bool, bool) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return false, false
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, false
	}
	return parsed, true
}

func getEnvFloat(key string) (float64, bool) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return 0, false
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, false
	}
	return parsed, true
}

func getEnvDuration(key string) (time.Duration, bool) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return 0, false
	}
	if parsed, err := time.ParseDuration(value); err == nil {
		return parsed, true
	}
	if seconds, err := strconv.Atoi(value); err == nil {
		return time.Duration(seconds) * time.Second, true
	}
	return 0, false
}

// ensureDirs 确保必要的目录存在
func ensureDirs(config *Config) {
	logDir := filepath.Dir(config.Log.File)
	if logDir != "." {
		_ = os.MkdirAll(logDir, 0o755)
	}

	configDir := filepath.Dir(config.CPP.ConfigPath)
	if configDir != "." {
		_ = os.MkdirAll(configDir, 0o755)
	}
}

// GetInstance 获取配置实例
func GetInstance() *Config {
	return instance
}

// Reload 重新加载配置
func Reload() error {
	newConfig, err := loadConfig()
	if err != nil {
		return err
	}

	instance = newConfig
	return nil
}
