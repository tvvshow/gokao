package config

// RedisConfig 通用 Redis 连接配置。
type RedisConfig struct {
	RedisURL      string `json:"redis_url"`
	RedisPassword string `json:"redis_password"`
	RedisDB       int    `json:"redis_db"`
}

// LoadRedis 从 env 装填 RedisConfig。
//
// 参数：
//   - defaultURL：默认连接字符串，省略时使用 "localhost:6379"
//   - defaultDB ：默认逻辑库编号；不同服务可分配不同 db 以隔离 key 命名空间
func LoadRedis(defaultURL string, defaultDB int) RedisConfig {
	if defaultURL == "" {
		defaultURL = "localhost:6379"
	}
	return RedisConfig{
		RedisURL:      GetEnv("REDIS_URL", defaultURL),
		RedisPassword: GetEnv("REDIS_PASSWORD", ""),
		RedisDB:       GetEnvAsInt("REDIS_DB", defaultDB),
	}
}
