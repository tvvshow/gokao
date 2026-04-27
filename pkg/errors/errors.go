package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ErrorCode 错误码类型
type ErrorCode string

// 通用错误码
const (
	// 系统错误
	ErrInternalServer     ErrorCode = "internal_server_error"
	ErrServiceUnavailable ErrorCode = "service_unavailable"
	ErrDatabaseError      ErrorCode = "database_error"
	ErrCacheError         ErrorCode = "cache_error"
	ErrConfigError        ErrorCode = "config_error"

	// 客户端错误
	ErrInvalidRequest     ErrorCode = "invalid_request"
	ErrValidationFailed   ErrorCode = "validation_failed"
	ErrUnauthorized       ErrorCode = "unauthorized"
	ErrForbidden          ErrorCode = "forbidden"
	ErrNotFound           ErrorCode = "not_found"
	ErrMethodNotAllowed   ErrorCode = "method_not_allowed"
	ErrRequestTimeout     ErrorCode = "request_timeout"
	ErrTooManyRequests    ErrorCode = "too_many_requests"
	ErrConflict           ErrorCode = "conflict"
	ErrPreconditionFailed ErrorCode = "precondition_failed"

	// 业务错误
	ErrPaymentFailed     ErrorCode = "payment_failed"
	ErrResourceExhausted ErrorCode = "resource_exhausted"
	ErrQuotaExceeded     ErrorCode = "quota_exceeded"
	ErrFeatureDisabled   ErrorCode = "feature_disabled"
	ErrLicenseExpired    ErrorCode = "license_expired"
)

// ErrorResponse 统一错误响应结构
type ErrorResponse struct {
	Code             ErrorCode    `json:"error"`
	Message          string       `json:"message"`
	Details          interface{}  `json:"details,omitempty"`
	RequestID        string       `json:"request_id,omitempty"`
	Timestamp        string       `json:"timestamp"`
	DocumentationURL string       `json:"documentation_url,omitempty"`
	RetryAfter       int          `json:"retry_after,omitempty"`
	Errors           []FieldError `json:"errors,omitempty"`
}

// FieldError 字段级错误
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// Error 实现error接口
func (e ErrorResponse) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// ErrorMapping 错误码到HTTP状态码的映射
var ErrorMapping = map[ErrorCode]int{
	// 系统错误
	ErrInternalServer:     http.StatusInternalServerError,
	ErrServiceUnavailable: http.StatusServiceUnavailable,
	ErrDatabaseError:      http.StatusInternalServerError,
	ErrCacheError:         http.StatusInternalServerError,
	ErrConfigError:        http.StatusInternalServerError,

	// 客户端错误
	ErrInvalidRequest:     http.StatusBadRequest,
	ErrValidationFailed:   http.StatusBadRequest,
	ErrUnauthorized:       http.StatusUnauthorized,
	ErrForbidden:          http.StatusForbidden,
	ErrNotFound:           http.StatusNotFound,
	ErrMethodNotAllowed:   http.StatusMethodNotAllowed,
	ErrRequestTimeout:     http.StatusRequestTimeout,
	ErrTooManyRequests:    http.StatusTooManyRequests,
	ErrConflict:           http.StatusConflict,
	ErrPreconditionFailed: http.StatusPreconditionFailed,

	// 业务错误
	ErrPaymentFailed:     http.StatusPaymentRequired,
	ErrResourceExhausted: http.StatusTooManyRequests,
	ErrQuotaExceeded:     http.StatusTooManyRequests,
	ErrFeatureDisabled:   http.StatusForbidden,
	ErrLicenseExpired:    http.StatusForbidden,
}

// ErrorMessages 错误码到默认消息的映射
var ErrorMessages = map[ErrorCode]string{
	ErrInternalServer:     "Internal server error occurred",
	ErrServiceUnavailable: "Service is temporarily unavailable",
	ErrDatabaseError:      "Database operation failed",
	ErrCacheError:         "Cache operation failed",
	ErrConfigError:        "Configuration error occurred",

	ErrInvalidRequest:     "Invalid request parameters",
	ErrValidationFailed:   "Request validation failed",
	ErrUnauthorized:       "Authentication required",
	ErrForbidden:          "Access forbidden",
	ErrNotFound:           "Resource not found",
	ErrMethodNotAllowed:   "Method not allowed",
	ErrRequestTimeout:     "Request timeout",
	ErrTooManyRequests:    "Too many requests",
	ErrConflict:           "Resource conflict",
	ErrPreconditionFailed: "Precondition failed",

	ErrPaymentFailed:     "Payment processing failed",
	ErrResourceExhausted: "Resource exhausted",
	ErrQuotaExceeded:     "Quota exceeded",
	ErrFeatureDisabled:   "Feature is disabled",
	ErrLicenseExpired:    "License has expired",
}

// NewError 创建新的错误响应
func NewError(code ErrorCode, message string, details interface{}) *ErrorResponse {
	if message == "" {
		message = ErrorMessages[code]
	}

	return &ErrorResponse{
		Code:      code,
		Message:   message,
		Details:   details,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

// WithRequestID 设置请求ID
func (e *ErrorResponse) WithRequestID(requestID string) *ErrorResponse {
	e.RequestID = requestID
	return e
}

// WithRetryAfter 设置重试时间
func (e *ErrorResponse) WithRetryAfter(seconds int) *ErrorResponse {
	e.RetryAfter = seconds
	return e
}

// WithDocumentation 设置文档链接
func (e *ErrorResponse) WithDocumentation(url string) *ErrorResponse {
	e.DocumentationURL = url
	return e
}

// WithFieldErrors 设置字段错误
func (e *ErrorResponse) WithFieldErrors(errors []FieldError) *ErrorResponse {
	e.Errors = errors
	return e
}

// ErrorHandler 统一错误处理中间件
func ErrorHandler(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查是否有错误
		if len(c.Errors) > 0 {
			// 获取最后一个错误
			lastErr := c.Errors.Last()
			if lastErr != nil {
				HandleError(c, lastErr.Err, logger)
			}
		}
	}
}

// HandleError 处理错误并返回统一格式的响应
func HandleError(c *gin.Context, err error, logger *logrus.Logger) {
	if logger == nil {
		logger = logrus.StandardLogger()
	}

	var errorResp *ErrorResponse

	// 如果是自定义错误响应，直接使用
	if resp, ok := err.(*ErrorResponse); ok {
		errorResp = resp
	} else {
		// 根据错误类型创建相应的错误响应
		switch {
		case strings.Contains(err.Error(), "validation"):
			errorResp = NewError(ErrValidationFailed, err.Error(), nil)
		case strings.Contains(err.Error(), "unauthorized"):
			errorResp = NewError(ErrUnauthorized, err.Error(), nil)
		case strings.Contains(err.Error(), "forbidden"):
			errorResp = NewError(ErrForbidden, err.Error(), nil)
		case strings.Contains(err.Error(), "not found"):
			errorResp = NewError(ErrNotFound, err.Error(), nil)
		case strings.Contains(err.Error(), "timeout"):
			errorResp = NewError(ErrRequestTimeout, err.Error(), nil)
		default:
			errorResp = NewError(ErrInternalServer, "Internal server error", nil)
		}
	}

	if errorResp.RequestID == "" {
		errorResp.RequestID = c.GetString("request_id")
	}
	if errorResp.Timestamp == "" {
		errorResp.Timestamp = time.Now().Format(time.RFC3339)
	}

	// 设置HTTP状态码
	statusCode := ErrorMapping[errorResp.Code]
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}

	// 记录错误日志
	if statusCode >= 500 {
		logger.WithFields(logrus.Fields{
			"error":      err.Error(),
			"request_id": c.GetString("request_id"),
			"path":       c.Request.URL.Path,
			"method":     c.Request.Method,
		}).Error("Server error occurred")
	} else {
		logger.WithFields(logrus.Fields{
			"error":      err.Error(),
			"request_id": c.GetString("request_id"),
			"path":       c.Request.URL.Path,
			"method":     c.Request.Method,
		}).Warn("Client error occurred")
	}

	// 设置响应头
	c.Header("Content-Type", "application/json; charset=utf-8")
	if errorResp.RetryAfter > 0 {
		c.Header("Retry-After", fmt.Sprintf("%d", errorResp.RetryAfter))
	}

	// 返回JSON响应
	c.JSON(statusCode, errorResp)
}

// ValidationError 创建验证错误
func ValidationError(fieldErrors []FieldError) *ErrorResponse {
	return NewError(ErrValidationFailed, "Validation failed", nil).WithFieldErrors(fieldErrors)
}

// DatabaseError 创建数据库错误
func DatabaseError(err error) *ErrorResponse {
	return NewError(ErrDatabaseError, "Database operation failed", err.Error())
}

// NotFoundError 创建资源未找到错误
func NotFoundError(resource string) *ErrorResponse {
	return NewError(ErrNotFound, fmt.Sprintf("%s not found", resource), nil)
}

// UnauthorizedError 创建未授权错误
func UnauthorizedError(message string) *ErrorResponse {
	if message == "" {
		message = "Authentication required"
	}
	return NewError(ErrUnauthorized, message, nil)
}

// ForbiddenError 创建禁止访问错误
func ForbiddenError(message string) *ErrorResponse {
	if message == "" {
		message = "Access forbidden"
	}
	return NewError(ErrForbidden, message, nil)
}

// TooManyRequestsError 创建请求过多错误
func TooManyRequestsError(retryAfter int) *ErrorResponse {
	return NewError(ErrTooManyRequests, "Too many requests", nil).WithRetryAfter(retryAfter)
}

// ConflictError 创建冲突错误
func ConflictError(message string) *ErrorResponse {
	if message == "" {
		message = "Resource conflict"
	}
	return NewError(ErrConflict, message, nil)
}

// InternalServerError 创建内部服务器错误
func InternalServerError(message string) *ErrorResponse {
	if message == "" {
		message = "Internal server error"
	}
	return NewError(ErrInternalServer, message, nil)
}

// ServiceUnavailableError 创建服务不可用错误
func ServiceUnavailableError(message string, retryAfter int) *ErrorResponse {
	if message == "" {
		message = "Service temporarily unavailable"
	}
	return NewError(ErrServiceUnavailable, message, nil).WithRetryAfter(retryAfter)
}

// JSON 将错误响应转换为JSON字符串
func (e *ErrorResponse) JSON() string {
	data, _ := json.Marshal(e)
	return string(data)
}

// AbortWithError 中止请求并返回错误
func AbortWithError(c *gin.Context, err *ErrorResponse) {
	c.Abort()
	HandleError(c, err, nil)
}

// IsError 检查错误是否为特定错误码
func IsError(err error, code ErrorCode) bool {
	if resp, ok := err.(*ErrorResponse); ok {
		return resp.Code == code
	}
	return false
}
