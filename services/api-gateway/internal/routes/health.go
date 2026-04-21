package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Version   string            `json:"version"`
	Services  map[string]string `json:"services"`
	Uptime    string            `json:"uptime"`
}

// ServiceStatus 服务状态
type ServiceStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	URL    string `json:"url"`
}

var startTime = time.Now()

// SetupHealthRoutes 设置健康检查路由
func SetupHealthRoutes(router *gin.Engine) {
	health := router.Group("/health")
	{
		health.GET("", getHealth)
		health.GET("/ready", getReadiness)
		health.GET("/live", getLiveness)
	}
}

// getHealth 获取健康状态
func getHealth(c *gin.Context) {
	services := checkServices()
	
	status := "healthy"
	for _, serviceStatus := range services {
		if serviceStatus != "healthy" {
			status = "unhealthy"
			break
		}
	}

	response := HealthResponse{
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   getVersion(),
		Services:  services,
		Uptime:    time.Since(startTime).String(),
	}

	if status == "healthy" {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusServiceUnavailable, response)
	}
}

// getReadiness 就绪检查
func getReadiness(c *gin.Context) {
	services := checkServices()
	
	ready := true
	for _, serviceStatus := range services {
		if serviceStatus != "healthy" {
			ready = false
			break
		}
	}

	if ready {
		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not_ready",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"services": services,
		})
	}
}

// getLiveness 存活检查
func getLiveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"uptime": time.Since(startTime).String(),
	})
}

// checkServices 检查服务状态
func checkServices() map[string]string {
	services := map[string]string{
		"user-service":           checkServiceHealth("http://user-service:8081/health"),
		"data-service":           checkServiceHealth("http://data-service:8082/health"),
		"payment-service":        checkServiceHealth("http://payment-service:8083/health"),
		"recommendation-service": checkServiceHealth("http://recommendation-service:8083/health"),
	}
	
	return services
}

// checkServiceHealth 检查单个服务健康状态
func checkServiceHealth(url string) string {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	
	resp, err := client.Get(url)
	if err != nil {
		return "unhealthy"
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusOK {
		return "healthy"
	}
	
	return "unhealthy"
}

// getVersion 获取版本信息
func getVersion() string {
	// 这里可以从环境变量或构建信息中获取版本
	return "1.0.0"
}
