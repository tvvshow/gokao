package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestUniversity 测试用高校模型
type TestUniversity struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Code         string    `gorm:"uniqueIndex;not null;size:20" json:"code"`
	Name         string    `gorm:"index;not null;size:200" json:"name"`
	Province     string    `gorm:"size:50;index" json:"province"`
	Type         string    `gorm:"size:50;index" json:"type"`
	Level        string    `gorm:"size:50;index" json:"level"`
	NationalRank int       `gorm:"index" json:"national_rank"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (TestUniversity) TableName() string {
	return "universities"
}

// APIResponse 统一API响应格式
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Total   int64       `json:"total,omitempty"`
}

func main() {
	fmt.Println("🚀 启动快速前后端数据交互测试")

	// 1. 初始化SQLite数据库
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	// 2. 自动迁移
	err = db.AutoMigrate(&TestUniversity{})
	if err != nil {
		log.Fatal("数据库迁移失败:", err)
	}

	// 3. 插入测试数据
	universities := []TestUniversity{
		{
			Code:         "10003",
			Name:         "清华大学",
			Province:     "北京",
			Type:         "理工类",
			Level:        "985",
			NationalRank: 1,
			IsActive:     true,
		},
		{
			Code:         "10001",
			Name:         "北京大学",
			Province:     "北京",
			Type:         "综合类",
			Level:        "985",
			NationalRank: 2,
			IsActive:     true,
		},
		{
			Code:         "10246",
			Name:         "复旦大学",
			Province:     "上海",
			Type:         "综合类",
			Level:        "985",
			NationalRank: 3,
			IsActive:     true,
		},
		{
			Code:         "10248",
			Name:         "上海交通大学",
			Province:     "上海",
			Type:         "理工类",
			Level:        "985",
			NationalRank: 4,
			IsActive:     true,
		},
		{
			Code:         "10335",
			Name:         "浙江大学",
			Province:     "浙江",
			Type:         "综合类",
			Level:        "985",
			NationalRank: 5,
			IsActive:     true,
		},
	}

	// 清空现有数据并插入新数据
	db.Where("1 = 1").Delete(&TestUniversity{})
	result := db.Create(&universities)
	if result.Error != nil {
		log.Fatal("插入数据失败:", result.Error)
	}

	fmt.Printf("✅ 成功插入 %d 条高校数据\n", len(universities))

	// 4. 设置Gin路由
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 添加CORS中间件
	r.Use(func(c *gin.Context) {
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
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"timestamp": time.Now(),
			"service":   "data-service",
		})
	})

	// 高校列表API
	r.GET("/v1/universities", func(c *gin.Context) {
		var universities []TestUniversity
		var total int64

		// 查询总数
		db.Model(&TestUniversity{}).Count(&total)

		// 查询数据
		result := db.Find(&universities)
		if result.Error != nil {
			c.JSON(500, APIResponse{
				Success: false,
				Message: "查询失败: " + result.Error.Error(),
			})
			return
		}

		c.JSON(200, APIResponse{
			Success: true,
			Data:    universities,
			Message: "查询成功",
			Total:   total,
		})
	})

	// 高校类型统计API
	r.GET("/v1/universities/types", func(c *gin.Context) {
		var types []map[string]interface{}

		db.Model(&TestUniversity{}).
			Select("type, COUNT(*) as count").
			Group("type").
			Find(&types)

		c.JSON(200, APIResponse{
			Success: true,
			Data:    types,
			Message: "统计成功",
		})
	})

	// 高校统计API
	r.GET("/v1/universities/statistics", func(c *gin.Context) {
		var total int64
		var by985 int64
		var by211 int64

		db.Model(&TestUniversity{}).Count(&total)
		db.Model(&TestUniversity{}).Where("level = ?", "985").Count(&by985)
		db.Model(&TestUniversity{}).Where("level = ?", "211").Count(&by211)

		stats := map[string]interface{}{
			"total":     total,
			"985_count": by985,
			"211_count": by211,
			"provinces": []string{"北京", "上海", "浙江"},
			"types":     []string{"综合类", "理工类", "师范类"},
		}

		c.JSON(200, APIResponse{
			Success: true,
			Data:    stats,
			Message: "统计成功",
		})
	})

	// 根据ID获取高校
	r.GET("/v1/universities/:id", func(c *gin.Context) {
		id := c.Param("id")
		var university TestUniversity

		result := db.First(&university, "id = ?", id)
		if result.Error != nil {
			c.JSON(404, APIResponse{
				Success: false,
				Message: "高校不存在",
			})
			return
		}

		c.JSON(200, APIResponse{
			Success: true,
			Data:    university,
			Message: "查询成功",
		})
	})

	// 5. 启动服务器
	port := "8082"
	fmt.Printf("🌐 服务器启动在端口 %s\n", port)
	fmt.Println("📋 可用的API端点:")
	fmt.Println("  GET /health                    - 健康检查")
	fmt.Println("  GET /v1/universities           - 高校列表")
	fmt.Println("  GET /v1/universities/types     - 高校类型统计")
	fmt.Println("  GET /v1/universities/statistics - 高校统计")
	fmt.Println("  GET /v1/universities/:id       - 获取单个高校")
	fmt.Println()
	fmt.Println("🧪 测试命令:")
	fmt.Printf("  curl http://localhost:%s/health\n", port)
	fmt.Printf("  curl http://localhost:%s/v1/universities\n", port)
	fmt.Println()

	log.Fatal(http.ListenAndServe(":"+port, r))
}
