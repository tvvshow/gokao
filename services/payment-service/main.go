package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gaokaohub/payment-service/internal/adapters"
	"github.com/gaokaohub/payment-service/internal/config"
	"github.com/gaokaohub/payment-service/internal/database"
	"github.com/gaokaohub/payment-service/internal/handlers"
	"github.com/gaokaohub/payment-service/internal/middleware"
	"github.com/gaokaohub/payment-service/internal/services"
	"github.com/gin-gonic/gin"
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

// @host localhost:10084
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
	adapterFactory := adapters.NewPaymentAdapterFactory(cfg.Payment)

	// 初始化服务层
	paymentService := services.NewPaymentService(db, redisClient, adapterFactory)
	membershipService := services.NewMembershipService(db, redisClient)

	// 初始化处理器
	healthHandler := handlers.NewHealthHandler()
	paymentHandler := handlers.NewPaymentHandler(paymentService, log.Default())
	membershipHandler := handlers.NewMembershipHandler(membershipService, log.Default())

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
	router.Use(middleware.RateLimit(cfg.RateLimit))

	// 健康检查
	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Ready)

	// API路由组
	v1 := router.Group("/api/v1")

	// 基础路由
	v1.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "payment-service",
			"version": "1.0.0",
		})
	})

	// 支付路由
	paymentGroup := v1.Group("/payments")
	{
		paymentGroup.POST("", paymentHandler.CreatePayment)
		paymentGroup.GET("/:payment_id", paymentHandler.QueryPayment)
		paymentGroup.POST("/:payment_id/close", paymentHandler.ClosePayment)
		paymentGroup.GET("", paymentHandler.ListPayments)
		paymentGroup.GET("/statistics", paymentHandler.GetPaymentStatistics)
		paymentGroup.POST("/callback/:channel", paymentHandler.PaymentCallback)
		paymentGroup.GET("/webhook-test/:channel", paymentHandler.WebhookTest)
	}

	// 退款路由
	refundGroup := v1.Group("/refunds")
	{
		refundGroup.POST("", paymentHandler.Refund)
		refundGroup.GET("/:refund_id", paymentHandler.QueryRefund)
	}

	// 会员路由
	membershipGroup := v1.Group("/memberships")
	{
		membershipGroup.POST("", membershipHandler.CreateMembership)
		membershipGroup.GET("/:user_id", membershipHandler.GetMembership)
		membershipGroup.PUT("/:user_id", membershipHandler.UpdateMembership)
		membershipGroup.DELETE("/:user_id", membershipHandler.DeleteMembership)
		membershipGroup.GET("", membershipHandler.ListMemberships)
	}

	// 启动服务器
	server := &http.Server{
		Addr:         ":10084",
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
