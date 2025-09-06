package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"device-auth-service/internal/config"
	"device-auth-service/internal/handlers"
	"device-auth-service/internal/middleware"
	"device-auth-service/internal/models"
	"device-auth-service/internal/services"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化日志
	initLogger(cfg)

	// 初始化数据库
	db, err := initDatabase(cfg)
	if err != nil {
		logrus.Fatalf("Failed to initialize database: %v", err)
	}

	// 初始化Redis
	redisClient := initRedis(cfg)

	// 初始化服务
	deviceAuthService := services.NewDeviceAuthService(db, redisClient)

	// 初始化Gin引擎
	r := gin.New()
	r.Use(middleware.RequestLogMiddleware())
	r.Use(gin.Recovery())

	// 注册路由
	registerRoutes(r, deviceAuthService)

	// 启动服务器
	serverAddr := ":" + cfg.Server.Port
	logrus.Infof("Starting server on %s", serverAddr)
	go func() {
		if err := r.Run(serverAddr); err != nil {
			logrus.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down server...")
}

func initLogger(cfg *config.Config) {
	// 设置日志级别
	level, err := logrus.ParseLevel(cfg.Log.Level)
	if err != nil {
		logrus.Fatalf("Failed to parse log level: %v", err)
	}
	logrus.SetLevel(level)

	// 设置日志格式
	if cfg.Log.Format == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{})
	}

	// 设置日志输出
	if cfg.Log.Output == "stdout" {
		logrus.SetOutput(os.Stdout)
	}
}

func initDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := "host=" + cfg.Database.Host +
		" user=" + cfg.Database.User +
		" password=" + cfg.Database.Password +
		" dbname=" + cfg.Database.Name +
		" port=" + cfg.Database.Port +
		" sslmode=" + cfg.Database.SSLMode

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// 自动迁移数据库模型
	err = db.AutoMigrate(
		&models.Device{},
		&models.License{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func initRedis(cfg *config.Config) *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// 测试Redis连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		logrus.Fatalf("Failed to connect to Redis: %v", err)
	}

	return redisClient
}

func registerRoutes(r *gin.Engine, deviceAuthService *services.DeviceAuthService) {
	// 创建设备认证处理器
	deviceAuthHandler := handlers.NewDeviceAuthHandler(deviceAuthService)

	// 健康检查接口
	r.GET("/health", deviceAuthHandler.HealthCheckHandler)

	// 设备认证相关接口
	v1 := r.Group("/api/v1")
	{
		// 设备指纹相关接口
		v1.POST("/device/fingerprint", deviceAuthHandler.CollectDeviceFingerprintHandler)
		v1.POST("/device/register", deviceAuthHandler.RegisterDeviceHandler)

		// 许可证相关接口
		v1.POST("/license/validate", deviceAuthHandler.ValidateLicenseHandler)
		v1.POST("/license/generate", deviceAuthHandler.GenerateLicenseHandler)
	}
}