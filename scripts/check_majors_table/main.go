package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/gaokao/shared"
	_ "github.com/lib/pq"
)

// 数据库连接配置
const (
	DBHost     = "localhost"
	DBPort     = 5432
	DBUser     = "postgres"
	DBPassword = "password"
	DBName     = "gaokao_data"
)

// 连接数据库
func connectDB() (*sql.DB, error) {
	db, err := shared.ConnectDB()
	if err != nil {
		return nil, err
	}

	fmt.Println("成功连接到数据库!")
	return db, nil
}

func main() {
	// 连接数据库
	db, err := connectDB()
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	defer db.Close()

	// 查询majors表结构
	rows, err := db.Query(`
		SELECT column_name, data_type, is_nullable, column_default
		FROM information_schema.columns 
		WHERE table_name = 'majors' AND table_schema = 'public'
		ORDER BY ordinal_position
	`)
	if err != nil {
		log.Fatal("查询表结构失败:", err)
	}
	defer rows.Close()

	fmt.Println("majors表结构:")
	fmt.Println("列名\t\t\t数据类型\t\t可空\t\t默认值")
	fmt.Println("------------------------------------------------------------")

	for rows.Next() {
		var columnName, dataType, isNullable string
		var columnDefault *string
		err := rows.Scan(&columnName, &dataType, &isNullable, &columnDefault)
		if err != nil {
			log.Fatal("扫描结果失败:", err)
		}

		defaultValue := "NULL"
		if columnDefault != nil {
			defaultValue = *columnDefault
		}

		fmt.Printf("%-20s\t%-15s\t%-8s\t%s\n", columnName, dataType, isNullable, defaultValue)
	}

	// 检查是否存在majors表
	var tableExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' AND table_name = 'majors'
		)
	`).Scan(&tableExists)
	if err != nil {
		log.Fatal("检查表存在性失败:", err)
	}

	if !tableExists {
		fmt.Println("\n警告: majors表不存在!")
	} else {
		fmt.Println("\n✓ majors表存在")
	}
}