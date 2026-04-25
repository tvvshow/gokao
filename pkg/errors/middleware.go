package errors

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ErrorHandlerMiddleware Gin中间件，统一错误处理
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString("request_id") == "" {
			c.Set("request_id", c.GetHeader("X-Request-ID"))
		}

		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			var errorResp *ErrorResponse
			if customErr, ok := err.(*ErrorResponse); ok {
				errorResp = customErr
			} else {
				errorResp = InternalServerError("服务器内部错误")
			}

			requestID := c.GetString("request_id")
			if requestID != "" {
				errorResp.RequestID = requestID
			}

			logError(c, errorResp, err)

			statusCode := ErrorMapping[errorResp.Code]
			if statusCode == 0 {
				statusCode = http.StatusInternalServerError
			}
			c.JSON(statusCode, errorResp)
			c.Abort()
		}
	}
}

func logError(c *gin.Context, errResp *ErrorResponse, originalErr error) {
	statusCode := ErrorMapping[errResp.Code]
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}

	if statusCode >= 500 {
		buf := make([]byte, 4096)
		n := runtime.Stack(buf, false)
		stackTrace := string(buf[:n])

		logrus.WithFields(logrus.Fields{
			"time":          time.Now().Format(time.RFC3339),
			"method":        c.Request.Method,
			"path":          c.Request.URL.Path,
			"status":        statusCode,
			"client_ip":     c.ClientIP(),
			"user_agent":    c.Request.UserAgent(),
			"request_id":    c.GetString("request_id"),
			"error_code":    errResp.Code,
			"error_message": errResp.Message,
			"stack_trace":   stackTrace,
			"original_error": originalErr.Error(),
		}).Error("Server error occurred")
	} else {
		logrus.WithFields(logrus.Fields{
			"time":          time.Now().Format(time.RFC3339),
			"method":        c.Request.Method,
			"path":          c.Request.URL.Path,
			"status":        statusCode,
			"client_ip":     c.ClientIP(),
			"user_agent":    c.Request.UserAgent(),
			"request_id":    c.GetString("request_id"),
			"error_code":    errResp.Code,
			"error_message": errResp.Message,
		}).Warn("Client error occurred")
	}
}

// RecoveryMiddleware 恢复中间件，处理panic
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				buf := make([]byte, 4096)
				n := runtime.Stack(buf, false)
				stackTrace := string(buf[:n])

				errorResp := InternalServerError("服务器内部错误")

				requestID := c.GetString("request_id")
				if requestID != "" {
					errorResp.RequestID = requestID
				}

				logrus.WithFields(logrus.Fields{
					"panic":      err,
					"stack":      stackTrace,
					"request_id": requestID,
					"path":       c.Request.URL.Path,
					"method":     c.Request.Method,
				}).Error("PANIC recovered")

				c.JSON(http.StatusInternalServerError, errorResp)
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

	if requestID := c.GetString("request_id"); requestID != "" {
		response["request_id"] = requestID
	}

	c.JSON(http.StatusOK, response)
}
