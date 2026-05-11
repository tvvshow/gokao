package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	sharedcfg "github.com/tvvshow/gokao/pkg/config"
	shareddb "github.com/tvvshow/gokao/pkg/database"
	"github.com/tvvshow/gokao/pkg/response"
	"github.com/tvvshow/gokao/services/monitoring-service/internal/alerts"
	"github.com/tvvshow/gokao/services/monitoring-service/internal/metrics"
)

func main() {
	// 初始化日志
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 通过 pkg/config + pkg/database 收敛 Redis 初始化（含 ping 校验、统一池参数）。
	// docker-compose 历来用 REDIS_ADDR，保留作为 REDIS_URL 之后的回退。
	redisURL := sharedcfg.FirstNonEmpty("REDIS_URL", "REDIS_ADDR")
	if redisURL == "" {
		redisURL = "redis:6379"
	}
	redisCfg := sharedcfg.RedisConfig{
		RedisURL:      redisURL,
		RedisPassword: sharedcfg.GetEnv("REDIS_PASSWORD", ""),
		RedisDB:       sharedcfg.GetEnvAsInt("REDIS_DB", 0),
	}
	redisClient, err := shareddb.OpenRedis(redisCfg, 5*time.Second)
	if err != nil {
		log.Fatalf("failed to init Redis: %v", err)
	}
	defer redisClient.Close()

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
		metricsText := formatPrometheusMetrics(metricsCollector.GetAllMetrics())
		c.Data(http.StatusOK, "text/plain; version=0.0.4; charset=utf-8", []byte(metricsText))
	})

	// 注册告警相关端点
	r.POST("/alerts", func(c *gin.Context) {
		var alert alerts.Alert
		if err := c.ShouldBindJSON(&alert); err != nil {
			response.BadRequest(c, "invalid_request", err.Error(), nil)
			return
		}

		if err := alertManager.TriggerAlert(ctx, &alert); err != nil {
			response.InternalError(c, "alert_trigger_failed", err.Error(), nil)
			return
		}

		response.OKWithMessage(c, nil, "Alert triggered successfully")
	})

	r.GET("/alerts", func(c *gin.Context) {
		alertList, err := alertManager.GetAlerts(ctx, 10, 0)
		if err != nil {
			response.InternalError(c, "alert_fetch_failed", err.Error(), nil)
			return
		}

		response.OK(c, alertList)
	})

	// 启动服务器
	srv := &http.Server{
		Addr:    ":8086",
		Handler: r,
	}

	// 在goroutine中启动服务器
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	logger.Info("Server started on :8086")

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

func formatPrometheusMetrics(values map[string]float64) string {
	var b strings.Builder
	for name, value := range values {
		safeName := sanitizeMetricName(name)
		b.WriteString("# TYPE " + safeName + " gauge\n")
		b.WriteString(fmt.Sprintf("%s %v\n", safeName, value))
	}
	return b.String()
}

func sanitizeMetricName(name string) string {
	replacer := strings.NewReplacer(" ", "_", "-", "_", ".", "_", "/", "_")
	return replacer.Replace(name)
}
