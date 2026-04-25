package main

import (
	"context"
	"data-service/internal/config"
	"data-service/internal/database"
	"data-service/internal/handlers"
	"data-service/internal/middleware"
	"data-service/internal/services"
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
)

// @title 高考志愿填报系统 - 数据服务API
// @version 1.0
// @description 高考志愿填报系统的数据查询服务API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:10082
// @BasePath /

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		logrus.Warn("未找到.env文件，使用默认配置")
	}

	// 加载配置
	cfg := config.Load()

	// 设置日志级别
	logrus.SetLevel(logrus.InfoLevel)
	if cfg.Environment == "debug" {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// 设置Gin模式
	gin.SetMode(cfg.Environment)

	logger := logrus.New()
	logger.Info("启动数据服务...")

	// 初始化数据库连接
	db, err := database.NewDB(cfg, logger)
	if err != nil {
		logger.Fatalf("初始化数据库失败: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Errorf("关闭数据库连接失败: %v", err)
		}
	}()

	// 初始化服务
	universityService := services.NewUniversityService(db, logger)
	majorService := services.NewMajorService(db, logger)
	admissionService := services.NewAdmissionService(db, logger)
	searchService := services.NewSearchService(db, logger)
	algorithmService := services.NewAlgorithmService(db, logger)
	cacheService := services.NewCacheService(db, logger)
	performanceService := services.NewPerformanceService(db, logger)
	migrationService := services.NewMigrationService(db)

	// 初始化处理器
	universityHandler := handlers.NewUniversityHandler(universityService, logger)
	majorHandler := handlers.NewMajorHandler(majorService, logger)
	admissionHandler := handlers.NewAdmissionHandler(admissionService, logger)
	searchHandler := handlers.NewSearchHandler(searchService, logger)
	algorithmHandler := handlers.NewAlgorithmHandler(algorithmService, logger)
	performanceHandler := handlers.NewPerformanceHandler(performanceService, cacheService, db, logger)
	migrationHandler := handlers.NewMigrationHandler(db, migrationService, logger)
	dataHandler := handlers.NewDataHandler(db, logger)

	// 创建Gin引擎
	router := gin.New()

	// 注册中间件
	router.Use(middleware.Logger(logger))
	router.Use(middleware.Recovery(logger))
	router.Use(middleware.Security())
	router.Use(middleware.RateLimit(logger))
	router.Use(middleware.ValidatePageSize(cfg.MaxPageSize))
	router.Use(middleware.PerformanceMonitoring(performanceService))

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		status := db.Health(c.Request.Context())
		if status["postgresql"] && status["redis"] {
			c.JSON(http.StatusOK, gin.H{
				"status":    "healthy",
				"timestamp": time.Now().Unix(),
				"services":  status,
			})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":    "unhealthy",
				"timestamp": time.Now().Unix(),
				"services":  status,
			})
		}
	})

	// API路由组 - 统一使用/api/v1前缀
	apiV1 := router.Group("/api/v1")
	{
		// 院校路由
		universities := apiV1.Group("/universities")
		{
			universities.GET("", universityHandler.ListUniversities)
			universities.GET("/search", universityHandler.SearchUniversities)
			universities.GET("/statistics", universityHandler.GetUniversityStatistics)
			universities.GET("/provinces", universityHandler.GetUniversityProvinces)
			universities.GET("/types", universityHandler.GetUniversityTypes)
			universities.GET("/levels", universityHandler.GetUniversityLevels)
			universities.GET("/:id", universityHandler.GetUniversityByID)
			universities.GET("/code/:code", universityHandler.GetUniversityByCode)
		}

		// 专业路由
		majors := apiV1.Group("/majors")
		{
			majors.GET("", majorHandler.ListMajors)
			majors.GET("/search", majorHandler.SearchMajors)
			majors.GET("/statistics", majorHandler.GetMajorStatistics)
			majors.GET("/categories", majorHandler.GetMajorCategories)
			majors.GET("/disciplines", majorHandler.GetMajorDisciplines)
			majors.GET("/degree-types", majorHandler.GetDegreeTypes)
			majors.GET("/:id", majorHandler.GetMajorByID)
		}

		// 录取数据路由
		admission := apiV1.Group("/admission")
		{
			admission.GET("/data", admissionHandler.ListAdmissionData)
			admission.GET("/analyze", admissionHandler.AnalyzeAdmissionData)
			admission.POST("/predict", admissionHandler.PredictAdmission)
			admission.GET("/statistics", admissionHandler.GetAdmissionStatistics)
			admission.GET("/batches", admissionHandler.GetBatches)
			admission.GET("/categories", admissionHandler.GetCategories)
			admission.GET("/difficulties", admissionHandler.GetDifficulties)
		}

		// 搜索路由
		search := apiV1.Group("/search")
		{
			search.GET("", searchHandler.GlobalSearch)
			search.GET("/autocomplete", searchHandler.AutoComplete)
			search.GET("/hot", searchHandler.GetHotSearches)
			search.GET("/suggestions", searchHandler.GetSearchSuggestions)
		}

		// 算法路由
		algorithm := apiV1.Group("/algorithm")
		{
			algorithm.POST("/match", algorithmHandler.MatchVolunteers)
			algorithm.GET("/risk-tolerance", algorithmHandler.GetRiskToleranceOptions)
			algorithm.GET("/recommend-types", algorithmHandler.GetRecommendTypes)
		}

		// 性能监控路由
		performance := apiV1.Group("/performance")
		{
			performance.GET("/metrics", performanceHandler.GetMetrics)
			performance.GET("/summary", performanceHandler.GetSummary)
			performance.POST("/reset", performanceHandler.ResetMetrics)
			performance.GET("/cache-stats", performanceHandler.GetCacheStats)
			performance.GET("/db-pool-stats", performanceHandler.GetDBPoolStats)
			performance.POST("/clear-cache", performanceHandler.ClearCache)
			performance.POST("/refresh-cache", performanceHandler.RefreshCache)
			performance.POST("/warmup-cache", performanceHandler.WarmupCache)
		}

		// 数据库迁移路由
		migrations := apiV1.Group("/migrations")
		{
			migrations.POST("/apply", migrationHandler.ApplyMigrations)
			migrations.GET("/status", migrationHandler.GetMigrationStatus)
			migrations.POST("/rollback/:version", migrationHandler.RollbackMigration)
		}

		// 数据处理路由
		data := apiV1.Group("/data")
		{
			data.POST("/process", dataHandler.ProcessData)
			data.POST("/import", dataHandler.ImportData)
		}
	}

	// Swagger文档
	if cfg.EnableSwagger {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		logger.Info("Swagger文档已启用: http://localhost:10082/swagger/index.html")
	}

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动性能监控
	perfCtx, perfCancel := context.WithCancel(context.Background())
	defer perfCancel()
	go performanceService.StartPeriodicCollection(perfCtx, 30*time.Second)

	// 缓存预热
	if err := cacheService.WarmupCache(context.Background()); err != nil {
		logger.Warnf("缓存预热失败: %v", err)
	}

	// 启动服务器
	go func() {
		logger.Infof("数据服务启动在端口: %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("启动服务器失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("正在关闭数据服务...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Errorf("服务器关闭失败: %v", err)
	}

	logger.Info("数据服务已关闭")
}
