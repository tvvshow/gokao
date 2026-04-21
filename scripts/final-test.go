package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// TestUniversity 测试用高校模型
type TestUniversity struct {
	ID           uint   `json:"id"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	Province     string `json:"province"`
	Type         string `json:"type"`
	Level        string `json:"level"`
	NationalRank int    `json:"national_rank"`
	IsActive     bool   `json:"is_active"`
}

// APIResponse 统一API响应格式
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Total   int64       `json:"total,omitempty"`
}

// 内存数据存储
var universities = []TestUniversity{
	{ID: 1, Code: "10003", Name: "清华大学", Province: "北京", Type: "理工类", Level: "985", NationalRank: 1, IsActive: true},
	{ID: 2, Code: "10001", Name: "北京大学", Province: "北京", Type: "综合类", Level: "985", NationalRank: 2, IsActive: true},
	{ID: 3, Code: "10246", Name: "复旦大学", Province: "上海", Type: "综合类", Level: "985", NationalRank: 3, IsActive: true},
	{ID: 4, Code: "10248", Name: "上海交通大学", Province: "上海", Type: "理工类", Level: "985", NationalRank: 4, IsActive: true},
	{ID: 5, Code: "10335", Name: "浙江大学", Province: "浙江", Type: "综合类", Level: "985", NationalRank: 5, IsActive: true},
	{ID: 6, Code: "10358", Name: "中国科学技术大学", Province: "安徽", Type: "理工类", Level: "985", NationalRank: 6, IsActive: true},
	{ID: 7, Code: "10284", Name: "南京大学", Province: "江苏", Type: "综合类", Level: "985", NationalRank: 7, IsActive: true},
	{ID: 8, Code: "10487", Name: "华中科技大学", Province: "湖北", Type: "理工类", Level: "985", NationalRank: 8, IsActive: true},
	{ID: 9, Code: "10698", Name: "西安交通大学", Province: "陕西", Type: "理工类", Level: "985", NationalRank: 9, IsActive: true},
	{ID: 10, Code: "10213", Name: "哈尔滨工业大学", Province: "黑龙江", Type: "理工类", Level: "985", NationalRank: 10, IsActive: true},
}

func main() {
	fmt.Println("🚀 启动快速前后端数据交互测试")
	fmt.Printf("✅ 成功加载 %d 条高校数据\n", len(universities))

	// 设置Gin路由
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
			"version":   "1.0.0",
		})
	})

	// 高校列表API
	r.GET("/v1/universities", func(c *gin.Context) {
		c.JSON(200, APIResponse{
			Success: true,
			Data:    universities,
			Message: "查询成功",
			Total:   int64(len(universities)),
		})
	})

	// 高校类型统计API
	r.GET("/v1/universities/types", func(c *gin.Context) {
		typeCount := make(map[string]int)
		for _, u := range universities {
			typeCount[u.Type]++
		}

		var types []map[string]interface{}
		for t, count := range typeCount {
			types = append(types, map[string]interface{}{
				"type":  t,
				"count": count,
			})
		}

		c.JSON(200, APIResponse{
			Success: true,
			Data:    types,
			Message: "统计成功",
		})
	})

	// 高校统计API
	r.GET("/v1/universities/statistics", func(c *gin.Context) {
		total := len(universities)
		by985 := 0
		provinces := make(map[string]bool)
		types := make(map[string]bool)

		for _, u := range universities {
			if u.Level == "985" {
				by985++
			}
			provinces[u.Province] = true
			types[u.Type] = true
		}

		var provinceList []string
		for p := range provinces {
			provinceList = append(provinceList, p)
		}

		var typeList []string
		for t := range types {
			typeList = append(typeList, t)
		}

		stats := map[string]interface{}{
			"total":       total,
			"985_count":   by985,
			"211_count":   total - by985, // 简化处理
			"provinces":   provinceList,
			"types":       typeList,
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
		
		for _, u := range universities {
			if fmt.Sprintf("%d", u.ID) == id {
				c.JSON(200, APIResponse{
					Success: true,
					Data:    u,
					Message: "查询成功",
				})
				return
			}
		}

		c.JSON(404, APIResponse{
			Success: false,
			Message: "高校不存在",
		})
	})

	// 搜索高校API
	r.GET("/v1/universities/search", func(c *gin.Context) {
		keyword := c.Query("keyword")
		province := c.Query("province")
		utype := c.Query("type")

		var results []TestUniversity
		for _, u := range universities {
			match := true
			
			if keyword != "" && u.Name != keyword && u.Code != keyword {
				match = false
			}
			if province != "" && u.Province != province {
				match = false
			}
			if utype != "" && u.Type != utype {
				match = false
			}
			
			if match {
				results = append(results, u)
			}
		}

		c.JSON(200, APIResponse{
			Success: true,
			Data:    results,
			Message: "搜索成功",
			Total:   int64(len(results)),
		})
	})

	// 启动服务器
	port := "8082"
	fmt.Printf("🌐 服务器启动在端口 %s\n", port)
	fmt.Println("📋 可用的API端点:")
	fmt.Println("  GET /health                      - 健康检查")
	fmt.Println("  GET /v1/universities             - 高校列表")
	fmt.Println("  GET /v1/universities/types       - 高校类型统计")
	fmt.Println("  GET /v1/universities/statistics  - 高校统计")
	fmt.Println("  GET /v1/universities/:id         - 获取单个高校")
	fmt.Println("  GET /v1/universities/search      - 搜索高校")
	fmt.Println()
	fmt.Println("🧪 测试命令:")
	fmt.Printf("  curl http://localhost:%s/health\n", port)
	fmt.Printf("  curl http://localhost:%s/v1/universities\n", port)
	fmt.Println()

	log.Fatal(http.ListenAndServe(":"+port, r))
}
