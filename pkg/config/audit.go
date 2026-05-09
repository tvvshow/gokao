package config

// AuditConfig 通用审计日志配置。
type AuditConfig struct {
	EnableAudit   bool   `json:"enable_audit"`
	AuditLogLevel string `json:"audit_log_level"`
}

// LoadAudit 从 env 装填 AuditConfig。
func LoadAudit() AuditConfig {
	return AuditConfig{
		EnableAudit:   GetEnvAsBool("ENABLE_AUDIT", true),
		AuditLogLevel: GetEnv("AUDIT_LOG_LEVEL", "info"),
	}
}
