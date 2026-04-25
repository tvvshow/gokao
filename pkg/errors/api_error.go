package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// APIError 统一的API错误响应格式
type APIError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	RequestID string    `json:"request_id,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

// Error 实现error接口
func (e *APIError) Error() string {
	return fmt.Sprintf("APIError: code=%d, message=%s", e.Code, e.Message)
}

// New 创建新的API错误
func New(code int, message string, details interface{}) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// WithRequestID 设置请求ID
func (e *APIError) WithRequestID(requestID string) *APIError {
	e.RequestID = requestID
	return e
}

// WriteJSON 将错误以JSON格式写入HTTP响应
func (e *APIError) WriteJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.Code)
	json.NewEncoder(w).Encode(e)
}

// BadRequest 创建400错误
func BadRequest(message string, details interface{}) *APIError {
	return New(http.StatusBadRequest, message, details)
}

// Unauthorized 创建401错误
func Unauthorized(message string, details interface{}) *APIError {
	return New(http.StatusUnauthorized, message, details)
}

// Forbidden 创建403错误
func Forbidden(message string, details interface{}) *APIError {
	return New(http.StatusForbidden, message, details)
}

// NotFound 创建404错误
func NotFound(message string, details interface{}) *APIError {
	return New(http.StatusNotFound, message, details)
}

// Conflict 创建409错误
func Conflict(message string, details interface{}) *APIError {
	return New(http.StatusConflict, message, details)
}

// Validation 创建422错误
func Validation(message string, details interface{}) *APIError {
	return New(http.StatusUnprocessableEntity, message, details)
}

// RateLimit 创建429错误
func RateLimit(message string, details interface{}) *APIError {
	return New(http.StatusTooManyRequests, message, details)
}

// Internal 创建500错误
func Internal(message string, details interface{}) *APIError {
	return New(http.StatusInternalServerError, message, details)
}

// ServiceUnavailable 创建503错误
func ServiceUnavailable(message string, details interface{}) *APIError {
	return New(http.StatusServiceUnavailable, message, details)
}

// FromError 从标准error创建API错误
func FromError(err error) *APIError {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr
	}
	return Internal("服务器内部错误", err.Error())
}
