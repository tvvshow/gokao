// response.go 兼容层。
//
// 历史上 data-service 在此自定义了 APIResponse / ErrorInfo / PaginationInfo 与全套
// New*Response 工厂；现在统一到 pkg/response，本文件保留为薄包装：
//   - 类型为 pkg/response 类型的别名，handler 现存对 *APIResponse 字段的直接访问继续生效；
//   - New* 工厂转发到 pkg/response 的构造函数；
//   - WithRequestID 受益方法保留；新代码建议直接用 response.OK(c, data) 等 helper。
package handlers

import (
	sharedresp "github.com/oktetopython/gaokao/pkg/response"
)

type (
	APIResponse    = sharedresp.APIResponse
	ErrorInfo      = sharedresp.ErrorInfo
	PaginationInfo = sharedresp.PaginationInfo
)

// NewSuccessResponse 创建成功响应（默认消息 "操作成功"）。
func NewSuccessResponse(data interface{}) *APIResponse {
	return sharedresp.Success(nil, data, "")
}

// NewSuccessResponseWithMessage 创建带自定义消息的成功响应。
func NewSuccessResponseWithMessage(data interface{}, message string) *APIResponse {
	return sharedresp.Success(nil, data, message)
}

// NewErrorResponse 通用错误响应；旧调用方未传 code，固定为 "GENERAL_ERROR"。
func NewErrorResponse(message string) *APIResponse {
	return sharedresp.Error(nil, "GENERAL_ERROR", message, nil)
}

// NewErrorResponseWithCode 带错误码的错误响应。
func NewErrorResponseWithCode(code, message string) *APIResponse {
	return sharedresp.Error(nil, code, message, nil)
}

// NewValidationErrorResponse 验证失败响应（VALIDATION_ERROR + 详情）。
func NewValidationErrorResponse(details interface{}) *APIResponse {
	return sharedresp.Error(nil, "VALIDATION_ERROR", "请求参数验证失败", details)
}
