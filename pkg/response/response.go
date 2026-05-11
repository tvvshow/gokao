// Package response 提供跨服务统一的 HTTP JSON 响应类型与构造助手。
//
// 设计动机：data-service / recommendation-service 等都各自定义了 APIResponse / ErrorResponse
// 结构体（19 个文件级别重复），并复制粘贴成功 / 失败工厂函数。集中到 pkg/response 后：
//   - 字段统一：Success / Message / Data / Error / Timestamp / RequestID
//   - 错误信息统一：ErrorInfo { Code, Message, Details }
//   - 调用统一：response.OK(c, data) / response.Created(c, data) / response.BadRequest(c, code, msg)
//
// 服务可以渐进式接入：旧自定义类型保留也不会冲突。
package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// APIResponse 统一 API 响应结构。Data 与 Error 二选一。
type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Timestamp int64       `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// ErrorInfo 错误信息结构。
type ErrorInfo struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// PaginationInfo 分页信息结构（可附在 Data 内或单独返回）。
type PaginationInfo struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// 上下文中存储 request id 的 key（与各服务 middleware 约定一致）。
const RequestIDKey = "X-Request-ID"

// requestID 安全读取 *gin.Context 中的请求 ID；不存在时返回空字符串。
func requestID(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if v, ok := c.Get(RequestIDKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	if h := c.GetHeader(RequestIDKey); h != "" {
		return h
	}
	return ""
}

// Success 构造成功响应（不写出，由调用方决定 status code）。
func Success(c *gin.Context, data interface{}, message string) *APIResponse {
	if message == "" {
		message = "操作成功"
	}
	return &APIResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
		RequestID: requestID(c),
	}
}

// Error 构造失败响应（不写出）。
func Error(c *gin.Context, code, message string, details interface{}) *APIResponse {
	return &APIResponse{
		Success: false,
		Message: message,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now().Unix(),
		RequestID: requestID(c),
	}
}

// OK 写 200 + 成功 payload。
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Success(c, data, ""))
}

// OKWithMessage 写 200 + 自定义 message。
func OKWithMessage(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusOK, Success(c, data, message))
}

// Created 写 201，默认 message "创建成功"。
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Success(c, data, "创建成功"))
}

// CreatedWithMessage 写 201 + 自定义 message。
func CreatedWithMessage(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusCreated, Success(c, data, message))
}

// NoContent 写 204；body 通常被忽略，但保留 RequestID 便于追踪。
func NoContent(c *gin.Context) {
	c.JSON(http.StatusNoContent, Success(c, nil, ""))
}

// BadRequest 写 400。
func BadRequest(c *gin.Context, code, message string, details interface{}) {
	c.JSON(http.StatusBadRequest, Error(c, code, message, details))
}

// Unauthorized 写 401。
func Unauthorized(c *gin.Context, code, message string) {
	c.JSON(http.StatusUnauthorized, Error(c, code, message, nil))
}

// Forbidden 写 403。
func Forbidden(c *gin.Context, code, message string) {
	c.JSON(http.StatusForbidden, Error(c, code, message, nil))
}

// NotFound 写 404。
func NotFound(c *gin.Context, code, message string) {
	c.JSON(http.StatusNotFound, Error(c, code, message, nil))
}

// RequestTimeout 写 408。
func RequestTimeout(c *gin.Context, code, message string) {
	c.JSON(http.StatusRequestTimeout, Error(c, code, message, nil))
}

// Conflict 写 409。
func Conflict(c *gin.Context, code, message string, details interface{}) {
	c.JSON(http.StatusConflict, Error(c, code, message, details))
}

// Gone 写 410。
func Gone(c *gin.Context, code, message string) {
	c.JSON(http.StatusGone, Error(c, code, message, nil))
}

// Locked 写 423，常用于账户/资源锁定。
func Locked(c *gin.Context, code, message string) {
	c.JSON(http.StatusLocked, Error(c, code, message, nil))
}

// UnprocessableEntity 写 422，常用于业务级别的参数失败。
func UnprocessableEntity(c *gin.Context, code, message string, details interface{}) {
	c.JSON(http.StatusUnprocessableEntity, Error(c, code, message, details))
}

// TooManyRequests 写 429，常用于限流。
func TooManyRequests(c *gin.Context, code, message string, details interface{}) {
	c.JSON(http.StatusTooManyRequests, Error(c, code, message, details))
}

// InternalError 写 500，避免在生产暴露 details；details 仅在 dev/debug 路径塞错误对象。
func InternalError(c *gin.Context, code, message string, details interface{}) {
	c.JSON(http.StatusInternalServerError, Error(c, code, message, details))
}

// NotImplemented 写 501。
func NotImplemented(c *gin.Context, code, message string) {
	c.JSON(http.StatusNotImplemented, Error(c, code, message, nil))
}

// ServiceUnavailable 写 503，常用于熔断或依赖不可用。
func ServiceUnavailable(c *gin.Context, code, message string, details interface{}) {
	c.JSON(http.StatusServiceUnavailable, Error(c, code, message, details))
}

// GatewayTimeout 写 504。
func GatewayTimeout(c *gin.Context, code, message string) {
	c.JSON(http.StatusGatewayTimeout, Error(c, code, message, nil))
}

// WriteError 写任意 status code 的错误响应（用于不在常用集合内的状态码）。
func WriteError(c *gin.Context, status int, code, message string, details interface{}) {
	c.JSON(status, Error(c, code, message, details))
}

// AbortWithError 用 c.AbortWithStatusJSON，确保中间件链在错误时不再继续。
func AbortWithError(c *gin.Context, status int, code, message string, details interface{}) {
	c.AbortWithStatusJSON(status, Error(c, code, message, details))
}
