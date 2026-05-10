package main

import (
	"fmt"
	"github.com/tvvshow/gokao/services/data-service/internal/config"
	"github.com/tvvshow/gokao/services/data-service/internal/database"
	"log"

	"github.com/sirupsen/logrus"
)

func main() {
	// 加载配置
	cfg := config.Load()
	logger := logrus.New()

	// 创建数据库连接
	db, err := database.NewDB(cfg, logger)
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	fmt.Println("✅ 数据库连接成功")

	// 添加 popularity_score 字段
	sql := `ALTER TABLE majors ADD COLUMN IF NOT EXISTS popularity_score INTEGER DEFAULT 0;`

	result := db.PostgreSQL.Exec(sql)
	if result.Error != nil {
		log.Fatal("添加 popularity_score 字段失败:", result.Error)
	}

	fmt.Println("✅ 成功添加 popularity_score 字段到 majors 表")

	// 更新现有记录的 popularity_score 值
	updateQueries := []string{
		`UPDATE majors SET popularity_score = 95 WHERE name LIKE '%计算机%' OR name LIKE '%软件%' OR name LIKE '%人工智能%';`,
		`UPDATE majors SET popularity_score = 90 WHERE name LIKE '%电子%' OR name LIKE '%通信%' OR name LIKE '%自动化%';`,
		`UPDATE majors SET popularity_score = 85 WHERE name LIKE '%金融%' OR name LIKE '%经济%' OR name LIKE '%管理%';`,
		`UPDATE majors SET popularity_score = 80 WHERE name LIKE '%医学%' OR name LIKE '%临床%' OR name LIKE '%护理%';`,
		`UPDATE majors SET popularity_score = 75 WHERE name LIKE '%机械%' OR name LIKE '%土木%' OR name LIKE '%建筑%';`,
		`UPDATE majors SET popularity_score = 70 WHERE popularity_score = 0;`, // 其他专业默认70分
	}

	for i, updateQuery := range updateQueries {
		result := db.PostgreSQL.Exec(updateQuery)
		if result.Error != nil {
			log.Printf("更新专业热度分数失败 (查询 %d): %v", i+1, result.Error)
		} else {
			fmt.Printf("✅ 完成专业热度分数更新 (查询 %d)\n", i+1)
		}
	}

	// 验证字段是否添加成功
	var count int64
	result = db.PostgreSQL.Raw("SELECT COUNT(*) FROM majors WHERE popularity_score IS NOT NULL").Scan(&count)
	if result.Error != nil {
		log.Fatal("验证字段失败:", result.Error)
	}

	fmt.Printf("✅ 验证成功: %d 条专业记录包含 popularity_score 字段\n", count)
	fmt.Println("🎉 数据库字段添加和数据更新完成!")
}
