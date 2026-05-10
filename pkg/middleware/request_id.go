package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

// RequestID 返回一个 Gin 中间件，确保每个请求都有 X-Request-ID。
// 如果上游已提供则透传，否则生成 16 字节随机 ID。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = generateID()
		}
		c.Set("request_id", id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

// TraceID 返回一个 Gin 中间件，确保每个请求都有 X-Trace-ID。
// 如果上游已提供则透传，否则生成 16 字节随机 ID。
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Trace-ID")
		if id == "" {
			id = generateID()
		}
		c.Set("trace_id", id)
		c.Header("X-Trace-ID", id)
		c.Next()
	}
}

func generateID() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}
