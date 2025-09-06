package logger

import (
	"context"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// Logger 统一日志接口
type Logger struct {
	*logrus.Logger
	serviceName string
}

// Config 日志配置
type Config struct {
	Level       string `json:"level"`
	Format      string `json:"format"` // json, text
	ServiceName string `json:"service_name"`
	Output      string `json:"output"` // stdout, stderr, file
	FilePath    string `json:"file_path,omitempty"`
}

// NewLogger 创建新的日志实例
func NewLogger(config *Config) *Logger {
	log := logrus.New()

	// 设置日志级别
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	log.SetLevel(level)

	// 设置日志格式
	switch strings.ToLower(config.Format) {
	case "json":
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		})
	default:
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	}

	// 设置输出
	switch strings.ToLower(config.Output) {
	case "stderr":
		log.SetOutput(os.Stderr)
	case "file":
		if config.FilePath != "" {
			file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err == nil {
				log.SetOutput(file)
			}
		}
	default:
		log.SetOutput(os.Stdout)
	}

	return &Logger{
		Logger:      log,
		serviceName: config.ServiceName,
	}
}

// NewDefaultLogger 创建默认日志实例
func NewDefaultLogger(serviceName string) *Logger {
	config := &Config{
		Level:       getEnv("LOG_LEVEL", "info"),
		Format:      getEnv("LOG_FORMAT", "json"),
		ServiceName: serviceName,
		Output:      getEnv("LOG_OUTPUT", "stdout"),
	}
	return NewLogger(config)
}

// WithContext 添加上下文信息
func (l *Logger) WithContext(ctx context.Context) *logrus.Entry {
	entry := l.WithField("service", l.serviceName)

	// 添加请求ID
	if requestID := ctx.Value("request_id"); requestID != nil {
		entry = entry.WithField("request_id", requestID)
	}

	// 添加用户ID
	if userID := ctx.Value("user_id"); userID != nil {
		entry = entry.WithField("user_id", userID)
	}

	// 添加追踪ID
	if traceID := ctx.Value("trace_id"); traceID != nil {
		entry = entry.WithField("trace_id", traceID)
	}

	return entry
}

// WithRequestID 添加请求ID
func (l *Logger) WithRequestID(requestID string) *logrus.Entry {
	return l.WithFields(logrus.Fields{
		"service":    l.serviceName,
		"request_id": requestID,
	})
}

// WithUserID 添加用户ID
func (l *Logger) WithUserID(userID string) *logrus.Entry {
	return l.WithFields(logrus.Fields{
		"service": l.serviceName,
		"user_id": userID,
	})
}

// WithError 添加错误信息
func (l *Logger) WithError(err error) *logrus.Entry {
	return l.WithFields(logrus.Fields{
		"service": l.serviceName,
		"error":   err.Error(),
	})
}

// WithFields 添加多个字段
func (l *Logger) WithFields(fields logrus.Fields) *logrus.Entry {
	if fields == nil {
		fields = make(logrus.Fields)
	}
	fields["service"] = l.serviceName
	return l.Logger.WithFields(fields)
}

// API请求日志
func (l *Logger) LogAPIRequest(method, path, userID, requestID string, statusCode int, duration int64) {
	l.WithFields(logrus.Fields{
		"type":        "api_request",
		"method":      method,
		"path":        path,
		"user_id":     userID,
		"request_id":  requestID,
		"status_code": statusCode,
		"duration_ms": duration,
	}).Info("API request completed")
}

// 数据库操作日志
func (l *Logger) LogDBOperation(operation, table, userID, requestID string, duration int64, err error) {
	fields := logrus.Fields{
		"type":        "db_operation",
		"operation":   operation,
		"table":       table,
		"user_id":     userID,
		"request_id":  requestID,
		"duration_ms": duration,
	}

	if err != nil {
		fields["error"] = err.Error()
		l.WithFields(fields).Error("Database operation failed")
	} else {
		l.WithFields(fields).Debug("Database operation completed")
	}
}

// 业务事件日志
func (l *Logger) LogBusinessEvent(event, userID, requestID string, data map[string]interface{}) {
	fields := logrus.Fields{
		"type":       "business_event",
		"event":      event,
		"user_id":    userID,
		"request_id": requestID,
	}

	for k, v := range data {
		fields[k] = v
	}

	l.WithFields(fields).Info("Business event occurred")
}

// 安全事件日志
func (l *Logger) LogSecurityEvent(event, userID, ip, requestID string, severity string) {
	l.WithFields(logrus.Fields{
		"type":       "security_event",
		"event":      event,
		"user_id":    userID,
		"ip":         ip,
		"request_id": requestID,
		"severity":   severity,
	}).Warn("Security event detected")
}

// 性能监控日志
func (l *Logger) LogPerformance(component, operation string, duration int64, success bool) {
	fields := logrus.Fields{
		"type":        "performance",
		"component":   component,
		"operation":   operation,
		"duration_ms": duration,
		"success":     success,
	}

	if success {
		l.WithFields(fields).Debug("Performance metric recorded")
	} else {
		l.WithFields(fields).Warn("Performance issue detected")
	}
}

// 辅助函数
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
