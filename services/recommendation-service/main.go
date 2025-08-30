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
	"github.com/oktetopython/gaokao/recommendation-service/internal/config"
	"github.com/oktetopython/gaokao/recommendation-service/internal/handlers"
	"github.com/oktetopython/gaokao/recommendation-service/internal/services"
	"github.com/oktetopython/gaokao/recommendation-service/pkg/cppbridge"
)

// @title 高考志愿填报推荐服务 API
// @version 1.0
// @description 混合推荐引擎API服务，融合传统算法和AI推荐
// @host localhost:8083
// @BasePath /api/v1
func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化C++桥接器
	bridge, err := cppbridge.NewHybridRecommendationBridge(cfg.CPP.ConfigPath)
	if err != nil {
		log.Fatalf("Failed to initialize C++ bridge: %v", err)
	}
	defer bridge.Close()

	// 设置Gin模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由器
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// CORS中间件
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "recommendation-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// API路由组
	v1 := router.Group("/api/v1")
	{
		// 创建服务
		analyticsService := services.NewAnalyticsService(bridge)
		
		// 创建处理器
		recommendationHandler := handlers.NewSimpleRecommendationHandler(bridge)
		hybridHandler := handlers.NewSimpleHybridHandler(bridge)
		analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)

		// 推荐相关路由
		recommendations := v1.Group("/recommendations")
		{
			recommendations.POST("/generate", recommendationHandler.GenerateRecommendations)
			recommendations.POST("/batch", recommendationHandler.BatchGenerateRecommendations)
			recommendations.GET("/explain/:id", recommendationHandler.ExplainRecommendation)
			recommendations.POST("/optimize", recommendationHandler.OptimizeRecommendations)
		}

		// 混合推荐路由
		hybrid := v1.Group("/hybrid")
		{
			hybrid.POST("/plan", hybridHandler.GenerateHybridPlan)
			hybrid.PUT("/weights", hybridHandler.UpdateFusionWeights)
			hybrid.GET("/config", hybridHandler.GetHybridConfig)
			hybrid.POST("/compare", hybridHandler.CompareRecommendations)
		}

		// 分析和统计路由
		analytics := v1.Group("/analytics")
		{
			analytics.GET("/performance", analyticsHandler.GetPerformanceMetrics)
			analytics.GET("/fusion-stats", analyticsHandler.GetFusionStatistics)
			analytics.POST("/quality-report", analyticsHandler.GenerateQualityReport)
			analytics.GET("/trends", analyticsHandler.GetRecommendationTrends)
		}

		// 系统管理路由
		system := v1.Group("/system")
		{
			system.GET("/status", recommendationHandler.GetSystemStatus)
			system.POST("/cache/clear", recommendationHandler.ClearCache)
			system.PUT("/model/update", recommendationHandler.UpdateModel)
		}
	}

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    cfg.Server.Port,
		Handler: router,
	}

	// 优雅启动
	go func() {
		log.Printf("Starting recommendation service on %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}