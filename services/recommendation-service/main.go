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
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/oktetopython/gaokao/services/recommendation-service/internal/config"
	"github.com/oktetopython/gaokao/services/recommendation-service/internal/handlers"
	"github.com/oktetopython/gaokao/services/recommendation-service/internal/cache"
	"github.com/oktetopython/gaokao/services/recommendation-service/internal/services"
	"github.com/oktetopython/gaokao/services/recommendation-service/pkg/cppbridge"
)

// @title 高考志愿填报推荐服务 API
// @version 1.0
// @description 混合推荐引擎API服务，融合传统算法和AI推荐
// @host localhost:10083
// @BasePath /api/v1
func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		logrus.Warn("未找到.env文件，使用默认配置")
	}

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 设置日志级别
	if cfg.Server.Mode == "production" {
		logrus.SetLevel(logrus.InfoLevel)
		gin.SetMode(gin.ReleaseMode)
	} else {
		logrus.SetLevel(logrus.DebugLevel)
		gin.SetMode(gin.DebugMode)
	}

	logger := logrus.New()
	logger.Info("启动推荐服务...")

	// 初始化数据同步服务
	dataSyncService := services.NewDataSyncService(
		cfg.DataService.URL,
		cfg.DataService.APIKey,
		cfg.DataService.SyncInterval,
		logger,
	)

	// 初始化权重服务
	weightService := services.NewWeightService(logger)

	// 初始化推荐桥接器（优先使用增强版规则引擎）
	var bridge cppbridge.HybridRecommendationBridge
	var bridgeType string
	
	// 尝试使用增强版规则引擎
	enhancedBridge, err := cppbridge.NewEnhancedRuleRecommendationBridge(dataSyncService, weightService, logger)
	if err != nil {
		logger.Warnf("初始化增强版推荐引擎失败，使用简化版: %v", err)
		
		// 回退到简化版规则引擎
		bridge, err = cppbridge.NewSimpleRuleRecommendationBridge(cfg.CPP.ConfigPath)
		if err != nil {
			logger.Fatalf("初始化推荐桥接器失败: %v", err)
		}
		bridgeType = "simple_rule"
	} else {
		bridge = enhancedBridge
		bridgeType = "enhanced_rule"
		
		// 启动数据同步服务
		go dataSyncService.Start(context.Background())
	}
	defer bridge.Close()
	
	logger.Infof("使用%s推荐引擎", bridgeType)

	// 初始化缓存
	cacheService, err := cache.NewCache(cfg.Redis)
	if err != nil {
		logger.Warnf("缓存初始化失败，使用内存缓存: %v", err)
		cacheService = cache.NewMemoryCache()
	}
	defer cacheService.Close()

	// 初始化处理器
	recommendationHandler := handlers.NewSimpleRecommendationHandler(bridge, cacheService)
	weightHandler := handlers.NewWeightHandler(weightService, logger)

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
	api := router.Group("/api/v1")
	{
		// 推荐路由
		recommendations := api.Group("/recommendations")
		{
			recommendations.POST("/generate", recommendationHandler.GenerateRecommendations)
			recommendations.POST("/batch", recommendationHandler.BatchGenerateRecommendations)
			recommendations.POST("/explain", recommendationHandler.ExplainRecommendation)
			recommendations.POST("/optimize", recommendationHandler.OptimizeRecommendations)
			recommendations.DELETE("/cache", recommendationHandler.ClearCache)
		}

		// 权重配置路由
		weightHandler.RegisterRoutes(api)

		// 系统管理路由
		system := api.Group("/system")
		{
			system.GET("/status", recommendationHandler.GetSystemStatus)
			system.POST("/model", recommendationHandler.UpdateModel)
			system.GET("/data/stats", func(c *gin.Context) {
				c.JSON(http.StatusOK, dataSyncService.GetCacheStats())
			})
			system.GET("/weight/stats", func(c *gin.Context) {
				c.JSON(http.StatusOK, weightService.GetWeightStats())
			})
		}
	}

	// Swagger文档
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	logger.Info("Swagger文档已启用: http://localhost:10083/swagger/index.html")

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动服务器
	go func() {
		logger.Infof("推荐服务启动在端口: %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("启动服务器失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("正在关闭推荐服务...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Errorf("服务器关闭失败: %v", err)
	}

	logger.Info("推荐服务已关闭")
}