package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tvvshow/gokao/services/payment-service/internal/config"
	"github.com/tvvshow/gokao/services/payment-service/internal/database"
	"github.com/tvvshow/gokao/services/payment-service/internal/handlers"
	"github.com/tvvshow/gokao/services/payment-service/internal/middleware"
	"github.com/tvvshow/gokao/services/payment-service/internal/repository"
	"github.com/tvvshow/gokao/services/payment-service/internal/service"

	pkghealth "github.com/tvvshow/gokao/pkg/health"
	pkgmw "github.com/tvvshow/gokao/pkg/middleware"
	"github.com/tvvshow/gokao/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// @title 高考志愿填报助手 - 支付服务
// @version 1.0
// @description 支付服务API，支持微信支付、支付宝、银联等多种支付方式
// @termsOfService http://gaokaohub.com/terms/

// @contact.name API Support
// @contact.url http://gaokaohub.com/support
// @contact.email support@gaokaohub.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8085
// @BasePath /api/v1

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// 初始化数据库
	db, err := database.Initialize(cfg.Database)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// 初始化Redis
	redisClient, err := database.InitializeRedis(cfg.Redis)
	if err != nil {
		log.Fatal("Failed to initialize redis:", err)
	}
	defer redisClient.Close()

	// 初始化支付适配器工厂
	// adapterFactory := adapters.NewPaymentAdapterFactory(cfg.Payment)

	// 初始化服务层
	logger := logrus.New()
	repo := repository.NewPaymentRepository(db)
	paymentService, err := service.NewPaymentService(repo, logger)
	if err != nil {
		log.Fatal("Failed to initialize payment service:", err)
	}
	membershipService := service.NewMembershipService(db, redisClient)

	// 初始化处理器
	paymentHandler := handlers.NewPaymentHandler(paymentService, logger)
	membershipHandler := handlers.NewMembershipHandler(membershipService)

	// 设置Gin模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由器
	router := gin.New()

	// 添加中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())
	router.Use(pkgmw.SecurityHeaders())
	router.Use(middleware.RateLimit(cfg.RateLimit))

	// 健康检查（共享 pkg/health：DB(*sql.DB) + Redis(v8) 实测）
	healthChecker := pkghealth.NewHealthChecker()
	healthChecker.Register(&sqlDBHealthCheck{db: db})
	healthChecker.Register(&redisV8HealthCheck{client: redisClient})
	healthHTTP := healthChecker.HTTPHandler()
	healthGin := func(c *gin.Context) { healthHTTP(c.Writer, c.Request) }
	router.GET("/health", healthGin)
	router.GET("/ready", healthGin)
	router.GET("/healthz", healthGin)
	router.GET("/readyz", healthGin)

	// API路由组
	v1 := router.Group("/api/v1")

	// 基础路由
	v1.GET("/status", func(c *gin.Context) {
		response.OK(c, gin.H{
			"status":  "ok",
			"service": "payment-service",
			"version": "1.0.0",
		})
	})

	// 支付路由
	paymentGroup := v1.Group("/payments")
	{
		// 幂等中间件挂在写入类接口上，TTL 24h 覆盖客户端常见重试窗口；
		// 客户端可通过 X-Idempotency-Key 头声明键，重复键直接回放首次响应。
		idempotency := middleware.Idempotency(redisClient, 24*time.Hour)

		paymentGroup.POST("", idempotency, paymentHandler.CreatePayment)
		paymentGroup.GET("/:payment_id", paymentHandler.QueryPayment)
		paymentGroup.POST("/:payment_id/close", paymentHandler.ClosePayment)
		paymentGroup.GET("", paymentHandler.ListPayments)
		paymentGroup.GET("/statistics", paymentHandler.GetPaymentStatistics)
		paymentGroup.POST("/callback/:channel", paymentHandler.PaymentCallback)
		paymentGroup.GET("/webhook-test/:channel", paymentHandler.WebhookTest)

		paymentMembershipGroup := paymentGroup.Group("/membership")
		membershipHandler.RegisterRoutes(paymentMembershipGroup)
	}

	// 退款路由
	refundGroup := v1.Group("/refunds")
	{
		refundGroup.POST("", middleware.Idempotency(redisClient, 24*time.Hour), paymentHandler.Refund)
		refundGroup.GET("/:refund_id", paymentHandler.QueryRefund)
	}

	membershipGroup := v1.Group("/membership")
	membershipHandler.RegisterRoutes(membershipGroup)

	// 启动服务器
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// 优雅关闭
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	log.Printf("Payment service started on port %d", cfg.Server.Port)

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down payment service...")

	// 5秒超时的优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Payment service forced to shutdown:", err)
	}

	log.Println("Payment service exited")
}

// sqlDBHealthCheck 适配 *sql.DB 到 pkghealth.HealthCheck 接口（payment-service 用裸 sql.DB，非 gorm）。
type sqlDBHealthCheck struct {
	db *sql.DB
}

func (s *sqlDBHealthCheck) Name() string { return "database" }

func (s *sqlDBHealthCheck) Check(ctx context.Context) pkghealth.CheckResult {
	start := time.Now()
	if err := s.db.PingContext(ctx); err != nil {
		return pkghealth.CheckResult{
			Name:     s.Name(),
			Status:   pkghealth.StatusUnhealthy,
			Message:  "Database ping failed",
			Duration: time.Since(start),
			Error:    err.Error(),
		}
	}
	return pkghealth.CheckResult{
		Name:     s.Name(),
		Status:   pkghealth.StatusHealthy,
		Message:  "Database is healthy",
		Duration: time.Since(start),
	}
}

// redisV8HealthCheck 适配 go-redis/v8 客户端到 pkghealth.HealthCheck 接口
// （pkghealth 自带的 RedisHealthCheck 绑定 v9，与本服务依赖冲突）。
type redisV8HealthCheck struct {
	client *redis.Client
}

func (r *redisV8HealthCheck) Name() string { return "redis" }

func (r *redisV8HealthCheck) Check(ctx context.Context) pkghealth.CheckResult {
	start := time.Now()
	if err := r.client.Ping(ctx).Err(); err != nil {
		return pkghealth.CheckResult{
			Name:     r.Name(),
			Status:   pkghealth.StatusUnhealthy,
			Message:  "Redis ping failed",
			Duration: time.Since(start),
			Error:    err.Error(),
		}
	}
	return pkghealth.CheckResult{
		Name:     r.Name(),
		Status:   pkghealth.StatusHealthy,
		Message:  "Redis is healthy",
		Duration: time.Since(start),
	}
}
