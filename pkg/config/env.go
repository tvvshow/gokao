// Package config 提供跨服务复用的配置原语：
//   - env.go：从环境变量读取并做类型解析的小工具
//   - server.go / database.go / redis.go / audit.go：服务通用配置子结构
//
// 设计原则：每个 service 通过结构嵌入（embedding）共用子结构，避免再各自重复
// 拷贝同一份 Port/DatabaseURL/Redis*/Audit* 字段及加载逻辑。
package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// FirstNonEmpty 顺序读取 env keys，返回第一个非空（trim 后）值。
func FirstNonEmpty(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

// DefaultString 当 value 为空时返回 fallback。
func DefaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

// NormalizePort 接受 "8080" 或 ":8080" 两种写法，统一返回不带冒号的端口。
func NormalizePort(port string) string {
	port = strings.TrimSpace(port)
	if strings.HasPrefix(port, ":") {
		return port[1:]
	}
	return port
}

// GetEnv 读取 env，未设置则返回 defaultValue。空字符串视作未设置。
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvAsInt 读取 env 并解析为 int，失败时返回 defaultValue。
func GetEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetEnvAsBool 读取 env 并解析为 bool，失败时返回 defaultValue。
func GetEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// GetEnvAsDuration 读取 env 并按 time.ParseDuration 解析；如果是纯整数视为秒数；
// 全部失败时按 defaultValue 解析；defaultValue 也无效则返回 15 分钟兜底。
func GetEnvAsDuration(key, defaultValue string) time.Duration {
	value := GetEnv(key, defaultValue)
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
