package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RequestLogMiddleware 请求日志中间件
func RequestLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 记录请求日志
		logrus.WithFields(logrus.Fields{
			"method":  c.Request.Method,
			"uri":     c.Request.RequestURI,
			"client":  c.ClientIP(),
			"status":  c.Writer.Status(),
			"latency": time.Since(startTime),
		}).Info("HTTP request")
	}
}