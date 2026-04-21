package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// 数据库连接信息
	connStr := "host=localhost port=5432 user=gaokao_user password=gaokao_password dbname=gaokao_user_db sslmode=disable"

	// 连接数据库
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatal("数据库连接测试失败:", err)
	}

	fmt.Println("✅ 数据库连接成功")

	// 检查admission_data表是否存在
	var exists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'admission_data'
		)
	`).Scan(&exists)

	if err != nil {
		log.Fatal("检查表存在性失败:", err)
	}

	if !exists {
		fmt.Println("❌ admission_data表不存在")
		return
	}

	fmt.Println("✅ admission_data表存在")

	// 查询表结构
	rows, err := db.Query(`
		SELECT column_name, data_type, is_nullable, column_default
		FROM information_schema.columns
		WHERE table_schema = 'public'
		AND table_name = 'admission_data'
		ORDER BY ordinal_position
	`)

	if err != nil {
		log.Fatal("查询表结构失败:", err)
	}
	defer rows.Close()

	fmt.Println("\n📋 admission_data表结构:")
	fmt.Println("列名\t\t\t数据类型\t\t可空\t\t默认值")
	fmt.Println("------------------------------------------------------------")

	for rows.Next() {
		var columnName, dataType, isNullable string
		var columnDefault sql.NullString

		err := rows.Scan(&columnName, &dataType, &isNullable, &columnDefault)
		if err != nil {
			log.Fatal("扫描行失败:", err)
		}

		defaultVal := "NULL"
		if columnDefault.Valid {
			defaultVal = columnDefault.String
		}

		fmt.Printf("%-20s\t%-15s\t%-8s\t%s\n", columnName, dataType, isNullable, defaultVal)
	}

	// 检查表中现有数据数量
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM admission_data").Scan(&count)
	if err != nil {
		log.Fatal("查询数据数量失败:", err)
	}

	fmt.Printf("\n📊 当前admission_data表中有 %d 条记录\n", count)

	// 如果有数据，显示几条示例
	if count > 0 {
		fmt.Println("\n📝 示例数据（前5条）:")
		rows, err := db.Query("SELECT * FROM admission_data LIMIT 5")
		if err != nil {
			log.Fatal("查询示例数据失败:", err)
		}
		defer rows.Close()

		// 获取列名
		columns, err := rows.Columns()
		if err != nil {
			log.Fatal("获取列名失败:", err)
		}

		fmt.Println("列名:", columns)

		// 显示数据
		for rows.Next() {
			// 创建接收数据的切片
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			err := rows.Scan(valuePtrs...)
			if err != nil {
				log.Fatal("扫描数据失败:", err)
			}

			fmt.Println("数据:", values)
		}
	}

	fmt.Println("\n✅ admission_data表检查完成")
}