package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheckResponse 健康检查响应
type HealthCheckResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// HealthCheckHandler 健康检查处理器
func HealthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, HealthCheckResponse{
		Status:  "ok",
		Message: "Device Auth Service is running",
	})
}