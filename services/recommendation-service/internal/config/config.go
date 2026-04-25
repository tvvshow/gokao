package config

import (
	"encoding/json"
	"os"
	"path/filepath"
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
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string `json:"port"`
	Mode string `json:"mode"`
}

// CPPConfig C++配置
type CPPConfig struct {
	ConfigPath       string `json:"config_path"`
	LibraryPath      string `json:"library_path"`
	MaxWorkers       int    `json:"max_workers"`
	CacheEnabled     bool   `json:"cache_enabled"`
	CacheSize        int    `json:"cache_size"`
	CacheTTLMinutes  int    `json:"cache_ttl_minutes"`
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
	// 默认配置
	config := &Config{
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
			MaxSize:    100, // MB
			MaxBackups: 10,
			MaxAge:     30, // days
		},
	}

	// 尝试从环境变量加载配置
	loadFromEnv(config)

	// 尝试从配置文件加载
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

	// 确保必要的目录存在
	ensureDirs(config)

	return config, nil
}

// loadFromEnv 从环境变量加载配置
func loadFromEnv(config *Config) {
	// Server配置
	if port := os.Getenv("SERVER_PORT"); port != "" {
		config.Server.Port = port
	}
	if mode := os.Getenv("SERVER_MODE"); mode != "" {
		config.Server.Mode = mode
	}

	// CPP配置
	if path := os.Getenv("CPP_CONFIG_PATH"); path != "" {
		config.CPP.ConfigPath = path
	}
	if path := os.Getenv("CPP_LIBRARY_PATH"); path != "" {
		config.CPP.LibraryPath = path
	}
	if workers := os.Getenv("CPP_MAX_WORKERS"); workers != "" {
		// 环境变量处理逻辑...
	}

	// Redis配置
	if enabled := os.Getenv("REDIS_ENABLED"); enabled != "" {
		config.Redis.Enabled = enabled == "true"
	}
	if host := os.Getenv("REDIS_HOST"); host != "" {
		config.Redis.Host = host
	}
	if port := os.Getenv("REDIS_PORT"); port != "" {
		// 环境变量处理逻辑...
	}

	// 日志配置
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.Log.Level = level
	}
	if file := os.Getenv("LOG_FILE"); file != "" {
		config.Log.File = file
	}
}

// ensureDirs 确保必要的目录存在
func ensureDirs(config *Config) {
	// 创建日志目录
	logDir := filepath.Dir(config.Log.File)
	if logDir != "." {
		os.MkdirAll(logDir, 0755)
	}

	// 创建配置目录
	configDir := filepath.Dir(config.CPP.ConfigPath)
	if configDir != "." {
		os.MkdirAll(configDir, 0755)
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
