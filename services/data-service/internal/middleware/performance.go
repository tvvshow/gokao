package middleware

import (
	"github.com/tvvshow/gokao/services/data-service/internal/services"
	"time"

	"github.com/gin-gonic/gin"
)

// PerformanceMonitoring 性能监控中间件
func PerformanceMonitoring(perfService *services.PerformanceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录开始时间
		start := time.Now()

		// 处理请求
		c.Next()

		// 计算处理时间
		duration := time.Since(start)

		// 确定请求是否成功
		success := c.Writer.Status() < 400

		// 获取endpoint路径
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}

		// 记录性能指标
		perfService.RecordRequest(endpoint, duration, success)

		// 在响应头中添加处理时间
		c.Header("X-Response-Time", duration.String())
	}
}
