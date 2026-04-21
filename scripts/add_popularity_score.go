package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// 数据库连接配置
	dsn := "host=localhost user=postgres password=password dbname=gaokao_data port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	
	// 连接数据库
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatal("数据库连接测试失败:", err)
	}

	fmt.Println("✅ 数据库连接成功")

	// 添加 popularity_score 字段
	query := `ALTER TABLE majors ADD COLUMN IF NOT EXISTS popularity_score INTEGER DEFAULT 0;`
	
	_, err = db.Exec(query)
	if err != nil {
		log.Fatal("添加 popularity_score 字段失败:", err)
	}

	fmt.Println("✅ 成功添加 popularity_score 字段到 majors 表")

	// 更新现有记录的 popularity_score 值（基于专业名称设置不同的热度分数）
	updateQueries := []string{
		`UPDATE majors SET popularity_score = 95 WHERE name LIKE '%计算机%' OR name LIKE '%软件%' OR name LIKE '%人工智能%';`,
		`UPDATE majors SET popularity_score = 90 WHERE name LIKE '%电子%' OR name LIKE '%通信%' OR name LIKE '%自动化%';`,
		`UPDATE majors SET popularity_score = 85 WHERE name LIKE '%金融%' OR name LIKE '%经济%' OR name LIKE '%管理%';`,
		`UPDATE majors SET popularity_score = 80 WHERE name LIKE '%医学%' OR name LIKE '%临床%' OR name LIKE '%护理%';`,
		`UPDATE majors SET popularity_score = 75 WHERE name LIKE '%机械%' OR name LIKE '%土木%' OR name LIKE '%建筑%';`,
		`UPDATE majors SET popularity_score = 70 WHERE popularity_score = 0;`, // 其他专业默认70分
	}

	for i, updateQuery := range updateQueries {
		_, err = db.Exec(updateQuery)
		if err != nil {
			log.Printf("更新专业热度分数失败 (查询 %d): %v", i+1, err)
		} else {
			fmt.Printf("✅ 完成专业热度分数更新 (查询 %d)\n", i+1)
		}
	}

	// 验证字段是否添加成功
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM majors WHERE popularity_score IS NOT NULL").Scan(&count)
	if err != nil {
		log.Fatal("验证字段失败:", err)
	}

	fmt.Printf("✅ 验证成功: %d 条专业记录包含 popularity_score 字段\n", count)
	fmt.Println("🎉 数据库字段添加和数据更新完成!")
}