package errors

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ErrorHandlerMiddleware Gin中间件，统一错误处理
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 在处理请求之前设置请求ID（如果尚未设置）
		if c.GetString("request_id") == "" {
			c.Set("request_id", c.GetHeader("X-Request-ID"))
		}

		// 继续处理请求
		c.Next()

		// 检查是否有错误
		if len(c.Errors) > 0 {
			// 获取最后一个错误
			err := c.Errors.Last().Err

			// 转换为APIError
			var apiErr *APIError
			if customErr, ok := err.(*APIError); ok {
				apiErr = customErr
			} else {
				// 未知错误，转换为500错误
				apiErr = Internal("服务器内部错误", err.Error())
			}

			// 设置请求ID
			requestID := c.GetString("request_id")
			if requestID != "" {
				apiErr.RequestID = requestID
			}

			// 记录错误日志
			logError(c, apiErr, err)

			// 返回统一的错误响应
			c.JSON(apiErr.Code, apiErr)
			c.Abort()
		}
	}
}

// logError 记录错误日志
func logError(c *gin.Context, apiErr *APIError, originalErr error) {
	// 构建日志信息
	logFields := map[string]interface{}{
		"time":         time.Now().Format(time.RFC3339),
		"method":       c.Request.Method,
		"path":         c.Request.URL.Path,
		"status":       apiErr.Code,
		"client_ip":    c.ClientIP(),
		"user_agent":   c.Request.UserAgent(),
		"request_id":   c.GetString("request_id"),
		"error_code":   apiErr.Code,
		"error_message": apiErr.Message,
	}

	// 如果是服务器错误，记录堆栈信息
	if apiErr.Code >= 500 {
		// 获取调用堆栈
		buf := make([]byte, 4096)
		n := runtime.Stack(buf, false)
		stackTrace := string(buf[:n])

		logFields["stack_trace"] = stackTrace
		logFields["original_error"] = originalErr.Error()

		// 记录错误日志
		log.Printf("SERVER_ERROR: %+v", logFields)
	} else {
		// 记录客户端错误
		log.Printf("CLIENT_ERROR: %+v", logFields)
	}
}

// RecoveryMiddleware 恢复中间件，处理panic
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录panic信息
				log.Printf("PANIC: %v", err)

				// 获取堆栈信息
				buf := make([]byte, 4096)
				n := runtime.Stack(buf, false)
				stackTrace := string(buf[:n])

				// 创建500错误
				apiErr := Internal("服务器内部错误", map[string]interface{}{
					"panic": err,
					"stack": stackTrace,
				})

				// 设置请求ID
				requestID := c.GetString("request_id")
				if requestID != "" {
					apiErr.RequestID = requestID
				}

				// 返回错误响应
				c.JSON(apiErr.Code, apiErr)
				c.Abort()
			}
		}()

		c.Next()
	}
}

// SuccessResponse 统一的成功响应格式
func SuccessResponse(c *gin.Context, data interface{}) {
	response := map[string]interface{}{
		"code":    http.StatusOK,
		"message": "success",
		"data":    data,
	}

	// 添加请求ID
	if requestID := c.GetString("request_id"); requestID != "" {
		response["request_id"] = requestID
	}

	c.JSON(http.StatusOK, response)
}

// SuccessResponseWithMessage 带自定义消息的成功响应
func SuccessResponseWithMessage(c *gin.Context, message string, data interface{}) {
	response := map[string]interface{}{
		"code":    http.StatusOK,
		"message": message,
		"data":    data,
	}

	// 添加请求ID
	if requestID := c.GetString("request_id"); requestID != "" {
		response["request_id"] = requestID
	}

	c.JSON(http.StatusOK, response)
}