package handlers

import (
	"time"
)

// APIResponse 统一API响应结构
type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Timestamp int64       `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// ErrorInfo 错误信息结构
type ErrorInfo struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// PaginationInfo 分页信息结构
type PaginationInfo struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data interface{}) *APIResponse {
	return &APIResponse{
		Success:   true,
		Message:   "操作成功",
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
}

// NewSuccessResponseWithMessage 创建带消息的成功响应
func NewSuccessResponseWithMessage(data interface{}, message string) *APIResponse {
	return &APIResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(message string) *APIResponse {
	return &APIResponse{
		Success:   false,
		Message:   message,
		Error: &ErrorInfo{
			Code:    "GENERAL_ERROR",
			Message: message,
		},
		Timestamp: time.Now().Unix(),
	}
}

// NewErrorResponseWithCode 创建带错误码的错误响应
func NewErrorResponseWithCode(code, message string) *APIResponse {
	return &APIResponse{
		Success:   false,
		Message:   message,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
		Timestamp: time.Now().Unix(),
	}
}

// NewValidationErrorResponse 创建验证错误响应
func NewValidationErrorResponse(details interface{}) *APIResponse {
	return &APIResponse{
		Success: false,
		Message: "请求参数验证失败",
		Error: &ErrorInfo{
			Code:    "VALIDATION_ERROR",
			Message: "请求参数验证失败",
			Details: details,
		},
		Timestamp: time.Now().Unix(),
	}
}

// WithRequestID 设置请求ID
func (r *APIResponse) WithRequestID(requestID string) *APIResponse {
	r.RequestID = requestID
	return r
}