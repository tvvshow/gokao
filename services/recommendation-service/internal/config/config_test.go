package config

import (
	"os"
	"sync"
	"testing"
	"time"
)

func resetConfigSingleton() {
	once = sync.Once{}
	instance = nil
}

func TestLoad(t *testing.T) {
	// 保存原始环境变量
	originalConfigPath := os.Getenv("CONFIG_PATH")
	defer func() {
		if originalConfigPath != "" {
			os.Setenv("CONFIG_PATH", originalConfigPath)
		} else {
			os.Unsetenv("CONFIG_PATH")
		}
	}()

	// 测试默认配置加载
	config, err := Load()
	if err != nil {
		t.Fatalf("加载默认配置失败: %v", err)
	}

	// 验证默认配置
	if config == nil {
		t.Fatal("配置为空")
	}

	if config.Server == nil {
		t.Fatal("服务器配置为空")
	}

	if config.Server.Port == "" {
		t.Error("服务器端口为空")
	}

	if config.CPP == nil {
		t.Fatal("C++配置为空")
	}

	if config.Redis == nil {
		t.Fatal("Redis配置为空")
	}

	if config.Log == nil {
		t.Fatal("日志配置为空")
	}

	// 验证默认值
	expectedPort := "8084"
	if config.Server.Port != expectedPort {
		t.Errorf("期望端口 %s，实际 %s", expectedPort, config.Server.Port)
	}

	if config.CPP.MaxWorkers <= 0 {
		t.Error("C++最大工作线程数应大于0")
	}

	if config.CPP.CacheSize <= 0 {
		t.Error("C++缓存大小应大于0")
	}
}

func TestLoadFromEnv(t *testing.T) {
	// 设置环境变量
	os.Setenv("SERVER_PORT", ":9999")
	os.Setenv("SERVER_MODE", "debug")
	os.Setenv("REDIS_ENABLED", "true")
	os.Setenv("REDIS_HOST", "test-redis")
	os.Setenv("LOG_LEVEL", "debug")

	defer func() {
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("SERVER_MODE")
		os.Unsetenv("REDIS_ENABLED")
		os.Unsetenv("REDIS_HOST")
		os.Unsetenv("LOG_LEVEL")
	}()

	// 重新加载配置（清除单例）
	resetConfigSingleton()

	config, err := Load()
	if err != nil {
		t.Fatalf("从环境变量加载配置失败: %v", err)
	}

	// 验证环境变量设置
	if config.Server.Port != "9999" {
		t.Errorf("期望端口 9999，实际 %s", config.Server.Port)
	}

	if config.Server.Mode != "debug" {
		t.Errorf("期望模式 debug，实际 %s", config.Server.Mode)
	}

	if !config.Redis.Enabled {
		t.Error("期望Redis启用")
	}

	if config.Redis.Host != "test-redis" {
		t.Errorf("期望Redis主机 test-redis，实际 %s", config.Redis.Host)
	}

	if config.Log.Level != "debug" {
		t.Errorf("期望日志级别 debug，实际 %s", config.Log.Level)
	}

	// 清除单例以不影响其他测试
	resetConfigSingleton()
}

func TestLoadFromLegacyPortEnvAliases(t *testing.T) {
	os.Setenv("PORT", ":8084")
	os.Setenv("GIN_MODE", "release")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("GIN_MODE")
		resetConfigSingleton()
	}()

	resetConfigSingleton()
	config, err := Load()
	if err != nil {
		t.Fatalf("从 legacy 环境变量加载配置失败: %v", err)
	}
	if config.Server.Port != "8084" {
		t.Fatalf("expected normalized port 8084, got %s", config.Server.Port)
	}
	if config.Server.Mode != "release" {
		t.Fatalf("expected mode release, got %s", config.Server.Mode)
	}
}

func TestGetInstance(t *testing.T) {
	// 先加载配置
	_, err := Load()
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 获取实例
	config := GetInstance()
	if config == nil {
		t.Fatal("获取配置实例失败")
	}

	// 再次获取应该是同一个实例
	config2 := GetInstance()
	if config != config2 {
		t.Error("配置实例不是单例")
	}
}

func TestReload(t *testing.T) {
	// 加载初始配置
	originalConfig, err := Load()
	if err != nil {
		t.Fatalf("加载初始配置失败: %v", err)
	}

	originalPort := originalConfig.Server.Port

	// 设置新的环境变量
	os.Setenv("SERVER_PORT", ":8888")
	defer os.Unsetenv("SERVER_PORT")

	// 重新加载配置
	err = Reload()
	if err != nil {
		t.Fatalf("重新加载配置失败: %v", err)
	}

	// 验证配置已更新
	newConfig := GetInstance()
	if newConfig.Server.Port == originalPort {
		t.Error("配置重新加载后端口没有更新")
	}

	if newConfig.Server.Port != "8888" {
		t.Errorf("期望新端口 8888，实际 %s", newConfig.Server.Port)
	}

	// 清除环境变量和单例
	os.Unsetenv("SERVER_PORT")
	resetConfigSingleton()
}

func TestCPPConfigValidation(t *testing.T) {
	config, err := Load()
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 验证C++配置的关键参数
	if config.CPP.ConfigPath == "" {
		t.Error("C++配置路径不能为空")
	}

	if config.CPP.LibraryPath == "" {
		t.Error("C++库路径不能为空")
	}

	if config.CPP.MaxWorkers <= 0 {
		t.Error("最大工作线程数必须大于0")
	}

	if config.CPP.CacheSize <= 0 {
		t.Error("缓存大小必须大于0")
	}

	if config.CPP.CacheTTLMinutes <= 0 {
		t.Error("缓存TTL必须大于0")
	}
}

func TestRedisConfigValidation(t *testing.T) {
	config, err := Load()
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// Redis配置验证
	if config.Redis.Host == "" {
		t.Error("Redis主机不能为空")
	}

	if config.Redis.Port <= 0 || config.Redis.Port > 65535 {
		t.Error("Redis端口必须在有效范围内")
	}

	if config.Redis.DB < 0 {
		t.Error("Redis数据库索引不能为负数")
	}
}

func TestLLMConfigDefaultsAndEnv(t *testing.T) {
	os.Setenv("LLM_ENABLED", "true")
	os.Setenv("LLM_BASE_URL", "http://localhost:11434/v1")
	os.Setenv("LLM_API_KEY", "test-key")
	os.Setenv("LLM_MODEL", "gpt-4o-mini")
	os.Setenv("LLM_TIMEOUT", "20s")
	os.Setenv("LLM_MAX_TOKENS", "1024")
	os.Setenv("LLM_TEMPERATURE", "0.6")
	defer func() {
		os.Unsetenv("LLM_ENABLED")
		os.Unsetenv("LLM_BASE_URL")
		os.Unsetenv("LLM_API_KEY")
		os.Unsetenv("LLM_MODEL")
		os.Unsetenv("LLM_TIMEOUT")
		os.Unsetenv("LLM_MAX_TOKENS")
		os.Unsetenv("LLM_TEMPERATURE")
		resetConfigSingleton()
	}()

	resetConfigSingleton()
	config, err := Load()
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	if config.LLM == nil {
		t.Fatal("LLM配置为空")
	}
	if !config.LLM.Enabled {
		t.Fatal("期望 LLM 已启用")
	}
	if config.LLM.BaseURL != "http://localhost:11434/v1" {
		t.Fatalf("unexpected base url: %s", config.LLM.BaseURL)
	}
	if config.LLM.APIKey != "test-key" {
		t.Fatalf("unexpected api key: %s", config.LLM.APIKey)
	}
	if config.LLM.Timeout != 20*time.Second {
		t.Fatalf("unexpected timeout: %s", config.LLM.Timeout)
	}
	if config.LLM.MaxTokens != 1024 {
		t.Fatalf("unexpected max tokens: %d", config.LLM.MaxTokens)
	}
	if config.LLM.Temperature != 0.6 {
		t.Fatalf("unexpected temperature: %v", config.LLM.Temperature)
	}
}

func TestCacheWarmConfigDefaultsAndEnv(t *testing.T) {
	os.Setenv("CACHE_WARM_ENABLED", "false")
	os.Setenv("CACHE_WARM_ASYNC", "false")
	os.Setenv("CACHE_WARM_TIMEOUT", "5s")
	defer func() {
		os.Unsetenv("CACHE_WARM_ENABLED")
		os.Unsetenv("CACHE_WARM_ASYNC")
		os.Unsetenv("CACHE_WARM_TIMEOUT")
		resetConfigSingleton()
	}()

	resetConfigSingleton()
	config, err := Load()
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	if config.CacheWarm == nil {
		t.Fatal("缓存预热配置为空")
	}
	if config.CacheWarm.Enabled {
		t.Fatal("期望缓存预热已禁用")
	}
	if config.CacheWarm.Async {
		t.Fatal("期望缓存预热同步执行")
	}
	if config.CacheWarm.RequestTimeout != 5*time.Second {
		t.Fatalf("unexpected warm timeout: %s", config.CacheWarm.RequestTimeout)
	}
}

func TestLogConfigValidation(t *testing.T) {
	config, err := Load()
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 日志配置验证
	validLevels := map[string]bool{
		"debug": true, "info": true, "warn": true,
		"warning": true, "error": true, "fatal": true, "panic": true,
	}

	if !validLevels[config.Log.Level] {
		t.Errorf("无效的日志级别: %s", config.Log.Level)
	}

	if config.Log.File == "" {
		t.Error("日志文件路径不能为空")
	}

	if config.Log.MaxSize <= 0 {
		t.Error("日志文件最大大小必须大于0")
	}

	if config.Log.MaxBackups < 0 {
		t.Error("日志文件备份数量不能为负数")
	}

	if config.Log.MaxAge < 0 {
		t.Error("日志文件保留天数不能为负数")
	}
}

// 基准测试
func BenchmarkLoad(b *testing.B) {
	// 清除单例
	resetConfigSingleton()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Load()
		if err != nil {
			b.Fatalf("加载配置失败: %v", err)
		}
		// 清除单例以便下次重新加载
		once = sync.Once{}
		instance = nil
	}
}

func BenchmarkGetInstance(b *testing.B) {
	// 先加载一次
	_, err := Load()
	if err != nil {
		b.Fatalf("加载配置失败: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config := GetInstance()
		if config == nil {
			b.Fatal("获取配置实例失败")
		}
	}
}
