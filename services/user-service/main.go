package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"user-service/internal/config"
	"user-service/internal/database"
	"user-service/internal/handlers"
	"user-service/internal/middleware"
	"user-service/internal/services"
)

// @title GaokaoHub User Service API
// @version 1.0
// @description 用户服务API，提供注册、登录、权限管理等功能
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8081
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found")
	}

	// 初始化配置
	cfg := config.Load()

	// 设置日志级别
	if cfg.Environment == "production" {
		logrus.SetLevel(logrus.InfoLevel)
		gin.SetMode(gin.ReleaseMode)
	} else {
		logrus.SetLevel(logrus.DebugLevel)
		gin.SetMode(gin.DebugMode)
	}

	// 初始化数据库
	db, err := database.Initialize(cfg)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// 初始化Redis
	redisClient, err := database.InitializeRedis(cfg)
	if err != nil {
		log.Fatal("Failed to initialize Redis:", err)
	}

	// 初始化服务
	userService := services.NewUserService(db, redisClient, cfg)
	authService := services.NewAuthService(db, redisClient, cfg, userService)
	roleService := services.NewRoleService(db, redisClient, cfg)

	// 初始化处理器
	userHandler := handlers.NewUserHandler(userService, roleService)
	authHandler := handlers.NewAuthHandler(authService, userService)
	roleHandler := handlers.NewRoleHandler(roleService)

	// 创建Gin路由器
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())

	// Swagger文档路由
	if cfg.EnableSwagger {
		if cfg.Environment == "production" {
			logrus.Warn("⚠️  Swagger UI is enabled in production environment. For security, set ENABLE_SWAGGER=false")
		}
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		logrus.Info("📚 Swagger UI available at: /swagger/index.html")
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "user-service"})
	})

	// API路由组
	v1 := r.Group("/api/v1")
	{
		// 认证相关路由（无需JWT）
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", middleware.JWTAuth(cfg.JWTSecret), authHandler.Logout)
		}

		// 用户相关路由（需要JWT）
		users := v1.Group("/users")
		users.Use(middleware.JWTAuth(cfg.JWTSecret))
		{
			users.GET("/profile", userHandler.GetProfile)
			users.PUT("/profile", userHandler.UpdateProfile)
			users.POST("/change-password", userHandler.ChangePassword)
			users.GET("/", middleware.RequireRole("admin"), userHandler.ListUsers)
			users.GET("/:id", middleware.RequireRole("admin"), userHandler.GetUser)
			users.PUT("/:id", middleware.RequireRole("admin"), userHandler.UpdateUser)
			users.DELETE("/:id", middleware.RequireRole("admin"), userHandler.DeleteUser)
		}

		// 角色权限相关路由（需要管理员权限）
		roles := v1.Group("/roles")
		roles.Use(middleware.JWTAuth(cfg.JWTSecret))
		roles.Use(middleware.RequireRole("admin"))
		{
			roles.GET("/", roleHandler.ListRoles)
			roles.POST("/", roleHandler.CreateRole)
			roles.GET("/:id", roleHandler.GetRole)
			roles.PUT("/:id", roleHandler.UpdateRole)
			roles.DELETE("/:id", roleHandler.DeleteRole)
			roles.POST("/:id/permissions", roleHandler.AssignPermissions)
			roles.DELETE("/:id/permissions/:permissionId", roleHandler.RevokePermission)
		}
	}

	// 启动服务器
	port := cfg.Port
	if port == "" {
		port = "8081"
	}

	logrus.Infof("🚀 User Service starting on port %s", port)
	logrus.Infof("📖 Environment: %s", cfg.Environment)
	logrus.Infof("🔒 JWT Secret configured: %t", cfg.JWTSecret != "")
	logrus.Infof("📊 Database: %s", cfg.DatabaseURL)

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
