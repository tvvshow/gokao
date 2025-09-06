package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/gaokao/monitoring-service/internal/alerts"
	"github.com/gaokao/monitoring-service/internal/metrics"
)

func main() {
	// 初始化日志
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 初始化Redis客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 初始化指标收集器
	metricsCollector := metrics.NewMetricsCollector()

	// 初始化告警管理器
	alertManager := alerts.NewAlertManager(redisClient, logger, metricsCollector)

	// 加载默认告警规则
	alertManager.LoadDefaultRules()

	// 启动指标收集
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go metricsCollector.StartMetricsCollection(ctx, 10*time.Second)

	// 启动HTTP服务
	r := gin.Default()

	// 注册指标端点
	r.GET("/metrics", func(c *gin.Context) {
		// 这里应该集成Prometheus的HTTP处理程序
		c.JSON(http.StatusOK, gin.H{"message": "Metrics endpoint"})
	})

	// 注册告警相关端点
	r.POST("/alerts", func(c *gin.Context) {
		var alert alerts.Alert
		if err := c.ShouldBindJSON(&alert); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := alertManager.TriggerAlert(ctx, &alert); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Alert triggered successfully"})
	})

	r.GET("/alerts", func(c *gin.Context) {
		alerts, err := alertManager.GetAlerts(ctx, 10, 0)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, alerts)
	})

	// 启动服务器
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// 在goroutine中启动服务器
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	logger.Info("Server started on :8080")

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// 上下文用于通知服务器它有5秒的时间来完成当前请求
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown:", zap.Error(err))
	}

	logger.Info("Server exiting")
}