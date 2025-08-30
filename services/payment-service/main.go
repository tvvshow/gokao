package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gaokaohub/payment-service/internal/config"
	"github.com/gaokaohub/payment-service/internal/handlers"
	"github.com/gaokaohub/payment-service/internal/middleware"
	"github.com/gaokaohub/payment-service/internal/services"
	"github.com/gaokaohub/payment-service/internal/adapters"
	"github.com/gaokaohub/payment-service/internal/database"
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

// @host localhost:8084
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
	orderService := services.NewOrderService(db, redisClient)
	membershipService := services.NewMembershipService(db, redisClient)

	// 初始化处理器
	paymentHandler := handlers.NewPaymentHandler(paymentService)
	orderHandler := handlers.NewOrderHandler(orderService)
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
	router.Use(middleware.RateLimit(cfg.RateLimit))

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"service": "payment-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// API路由组
	v1 := router.Group("/api/v1")
	v1.Use(middleware.Auth(cfg.JWT))

	// 支付相关路由
	paymentGroup := v1.Group("/payments")
	{
		paymentGroup.POST("/create", paymentHandler.CreatePayment)
		paymentGroup.POST("/callback/:channel", paymentHandler.HandleCallback)
		paymentGroup.GET("/query/:orderNo", paymentHandler.QueryPayment)
		paymentGroup.POST("/refund", paymentHandler.CreateRefund)
		paymentGroup.POST("/close/:orderNo", paymentHandler.CloseOrder)
		paymentGroup.GET("/channels", paymentHandler.GetSupportedChannels)
	}

	// 订单相关路由
	orderGroup := v1.Group("/orders")
	{
		orderGroup.POST("/create", orderHandler.CreateOrder)
		orderGroup.GET("/list", orderHandler.GetOrders)
		orderGroup.GET("/:orderNo", orderHandler.GetOrder)
		orderGroup.PUT("/:orderNo/cancel", orderHandler.CancelOrder)
		orderGroup.GET("/:orderNo/invoice", orderHandler.GetInvoice)
	}

	// 会员相关路由
	memberGroup := v1.Group("/membership")
	{
		memberGroup.GET("/plans", membershipHandler.GetPlans)
		memberGroup.POST("/subscribe", membershipHandler.Subscribe)
		memberGroup.GET("/status", membershipHandler.GetMembershipStatus)
		memberGroup.POST("/renew", membershipHandler.RenewMembership)
		memberGroup.POST("/cancel", membershipHandler.CancelMembership)
		memberGroup.GET("/benefits", membershipHandler.GetMemberBenefits)
	}

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