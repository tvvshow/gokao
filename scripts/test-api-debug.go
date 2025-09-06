package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// University 简化的院校模型
type University struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Code     string    `gorm:"uniqueIndex;not null;size:20" json:"code"`
	Name     string    `gorm:"index;not null;size:200" json:"name"`
	Province string    `gorm:"size:50;index" json:"province"`
	Type     string    `gorm:"size:50;index" json:"type"`
	Level    string    `gorm:"size:50;index" json:"level"`
}

func (University) TableName() string {
	return "universities"
}

func main() {
	fmt.Println("🔍 API调试测试")

	// 1. 测试数据库连接
	fmt.Println("\n1. 测试数据库连接...")
	dsn := "host=localhost user=gaokao_user password=gaokao_pass dbname=gaokao_data port=5432 sslmode=disable"
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		dsn = dbURL
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	fmt.Println("✅ 数据库连接成功")

	// 2. 测试表是否存在
	fmt.Println("\n2. 测试表结构...")
	var tableExists bool
	result := db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'universities')").Scan(&tableExists)
	if result.Error != nil {
		log.Fatal("检查表存在性失败:", result.Error)
	}
	if !tableExists {
		log.Fatal("universities表不存在")
	}
	fmt.Println("✅ universities表存在")

	// 3. 测试数据查询
	fmt.Println("\n3. 测试数据查询...")
	var count int64
	result = db.Model(&University{}).Count(&count)
	if result.Error != nil {
		log.Fatal("查询数据数量失败:", result.Error)
	}
	fmt.Printf("✅ 数据库中有 %d 条高校记录\n", count)

	// 4. 测试具体数据
	fmt.Println("\n4. 测试具体数据...")
	var universities []University
	result = db.Limit(3).Find(&universities)
	if result.Error != nil {
		log.Fatal("查询高校数据失败:", result.Error)
	}

	fmt.Printf("✅ 成功查询到 %d 条记录:\n", len(universities))
	for _, u := range universities {
		fmt.Printf("  - %s (%s) - %s %s\n", u.Name, u.Code, u.Province, u.Level)
	}

	// 5. 测试API端点
	fmt.Println("\n5. 测试API端点...")

	// 测试健康检查
	resp, err := http.Get("http://localhost:8082/health")
	if err != nil {
		fmt.Printf("❌ 健康检查失败: %v\n", err)
	} else {
		fmt.Printf("✅ 健康检查成功: %d\n", resp.StatusCode)
		resp.Body.Close()
	}

	// 测试高校列表API
	resp, err = http.Get("http://localhost:8082/v1/universities")
	if err != nil {
		fmt.Printf("❌ 高校列表API请求失败: %v\n", err)
	} else {
		fmt.Printf("📊 高校列表API响应: %d\n", resp.StatusCode)
		if resp.StatusCode == 200 {
			var result map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&result)
			fmt.Printf("✅ API响应成功: %+v\n", result)
		} else {
			fmt.Printf("❌ API返回错误状态码: %d\n", resp.StatusCode)
		}
		resp.Body.Close()
	}

	// 6. 测试不同的API路径
	fmt.Println("\n6. 测试不同的API路径...")
	testPaths := []string{
		"http://localhost:8082/",
		"http://localhost:8082/v1",
		"http://localhost:8082/v1/universities/types",
		"http://localhost:8082/api/v1/universities",
	}

	for _, path := range testPaths {
		resp, err := http.Get(path)
		if err != nil {
			fmt.Printf("❌ %s - 请求失败: %v\n", path, err)
		} else {
			fmt.Printf("📊 %s - 状态码: %d\n", path, resp.StatusCode)
			resp.Body.Close()
		}
	}

	fmt.Println("\n🎉 调试测试完成")
}
