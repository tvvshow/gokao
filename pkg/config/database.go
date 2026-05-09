package config

// DatabaseConfig 通用数据库连接 + 连接池配置。
type DatabaseConfig struct {
	DatabaseURL     string `json:"database_url"`
	MaxOpenConns    int    `json:"max_open_conns"`
	MaxIdleConns    int    `json:"max_idle_conns"`
	ConnMaxLifetime int    `json:"conn_max_lifetime"`
	ConnMaxIdleTime int    `json:"conn_max_idle_time"`
}

// LoadDatabase 从 env 装填 DatabaseConfig；defaultURL 为服务自有的默认 DSN。
func LoadDatabase(defaultURL string) DatabaseConfig {
	return DatabaseConfig{
		DatabaseURL:     GetEnv("DATABASE_URL", defaultURL),
		MaxOpenConns:    GetEnvAsInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    GetEnvAsInt("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: GetEnvAsInt("DB_CONN_MAX_LIFETIME", 1800),
		ConnMaxIdleTime: GetEnvAsInt("DB_CONN_MAX_IDLE_TIME", 900),
	}
}
