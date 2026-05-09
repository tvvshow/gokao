package config

// ServerConfig 通用服务进程配置。各服务的 Config 通过嵌入此结构复用 Port/Environment/EnableSwagger 字段。
type ServerConfig struct {
	Port          string `json:"port"`
	Environment   string `json:"environment"`
	EnableSwagger bool   `json:"enable_swagger"`
}

// LoadServer 从 env 装填 ServerConfig。
//
// 参数：
//   - defaultPort：服务自有的默认端口（如 user-service=8083、data-service=8082）
//   - swaggerKey：开关 swagger 的 env key；空字符串则用 "ENABLE_SWAGGER"
func LoadServer(defaultPort, swaggerKey string) ServerConfig {
	if swaggerKey == "" {
		swaggerKey = "ENABLE_SWAGGER"
	}
	return ServerConfig{
		Port:          DefaultString(NormalizePort(FirstNonEmpty("SERVER_PORT", "PORT")), defaultPort),
		Environment:   DefaultString(FirstNonEmpty("SERVER_MODE", "GIN_MODE"), "debug"),
		EnableSwagger: GetEnvAsBool(swaggerKey, true),
	}
}
