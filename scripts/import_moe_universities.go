package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// 数据库配置
const (
	DBHost     = "localhost"
	DBPort     = 5432
	DBUser     = "gaokao_user"
	DBPassword = "gaokao_password"
	DBName     = "gaokao_user_db"
)

// MOE大学数据结构（对应moe_universities.json）
type MOEUniversity struct {
	ID             int    `json:"id"`
	Sequence       int    `json:"sequence"`
	Name           string `json:"name"`
	Code           string `json:"code"`
	Supervisor     string `json:"supervisor"`
	Location       string `json:"location"`
	Level          string `json:"level"`
	Remark         string `json:"remark"`
	Province       string `json:"province"`
	City           string `json:"city"`
	Type           string `json:"type"`
	Is985          bool   `json:"is_985"`
	Is211          bool   `json:"is_211"`
	IsDoubleFirst  bool   `json:"is_double_first"`
}

// 连接数据库
func connectDB() (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		DBHost, DBPort, DBUser, DBPassword, DBName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

// 创建或更新数据库表结构
func createOrUpdateTables(db *sql.DB) error {
	// 创建增强版大学表（包含985/211/双一流标识）
	universityTable := `
	CREATE TABLE IF NOT EXISTS universities (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL UNIQUE,
		code VARCHAR(50) UNIQUE,
		supervisor VARCHAR(200),
		location VARCHAR(200),
		province VARCHAR(100),
		city VARCHAR(100),
		level VARCHAR(50),
		type VARCHAR(100),
		remark TEXT,
		is_985 BOOLEAN DEFAULT FALSE,
		is_211 BOOLEAN DEFAULT FALSE,
		is_double_first BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	// 添加缺失的列（如果表已存在）
	alterStatements := []string{
		"ALTER TABLE universities ADD COLUMN IF NOT EXISTS supervisor VARCHAR(200);",
		"ALTER TABLE universities ADD COLUMN IF NOT EXISTS location VARCHAR(200);",
		"ALTER TABLE universities ADD COLUMN IF NOT EXISTS remark TEXT;",
		"ALTER TABLE universities ADD COLUMN IF NOT EXISTS is_985 BOOLEAN DEFAULT FALSE;",
		"ALTER TABLE universities ADD COLUMN IF NOT EXISTS is_211 BOOLEAN DEFAULT FALSE;",
		"ALTER TABLE universities ADD COLUMN IF NOT EXISTS is_double_first BOOLEAN DEFAULT FALSE;",
	}

	// 创建表
	_, err := db.Exec(universityTable)
	if err != nil {
		return fmt.Errorf("创建universities表失败: %w", err)
	}

	// 添加缺失的列
	for _, stmt := range alterStatements {
		_, err := db.Exec(stmt)
		if err != nil {
			log.Printf("执行ALTER语句失败 (%s): %v", stmt, err)
			// 继续执行其他语句，不中断
		}
	}

	// 创建索引
	indexStatements := []string{
		"CREATE INDEX IF NOT EXISTS idx_universities_province ON universities(province);",
		"CREATE INDEX IF NOT EXISTS idx_universities_city ON universities(city);",
		"CREATE INDEX IF NOT EXISTS idx_universities_level ON universities(level);",
		"CREATE INDEX IF NOT EXISTS idx_universities_type ON universities(type);",
		"CREATE INDEX IF NOT EXISTS idx_universities_985 ON universities(is_985);",
		"CREATE INDEX IF NOT EXISTS idx_universities_211 ON universities(is_211);",
		"CREATE INDEX IF NOT EXISTS idx_universities_double_first ON universities(is_double_first);",
	}

	for _, stmt := range indexStatements {
		_, err := db.Exec(stmt)
		if err != nil {
			log.Printf("创建索引失败 (%s): %v", stmt, err)
		}
	}

	return nil
}

// 导入MOE大学数据
func importMOEUniversities(db *sql.DB, filename string) error {
	fmt.Printf("正在导入MOE大学数据: %s\n", filename)

	// 检查文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: %s", filename)
	}

	// 读取JSON文件
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	// 解析JSON数据
	var universities []MOEUniversity
	err = json.Unmarshal(data, &universities)
	if err != nil {
		return fmt.Errorf("解析JSON失败: %w", err)
	}

	fmt.Printf("解析到 %d 所大学数据\n", len(universities))

	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	// 准备插入语句（使用UPSERT）
	stmt, err := tx.Prepare(`
		INSERT INTO universities (
			name, code, supervisor, location, province, city, 
			level, type, remark, is_985, is_211, is_double_first,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (code) DO UPDATE SET
			name = EXCLUDED.name,
			supervisor = EXCLUDED.supervisor,
			location = EXCLUDED.location,
			province = EXCLUDED.province,
			city = EXCLUDED.city,
			level = EXCLUDED.level,
			type = EXCLUDED.type,
			remark = EXCLUDED.remark,
			is_985 = EXCLUDED.is_985,
			is_211 = EXCLUDED.is_211,
			is_double_first = EXCLUDED.is_double_first,
			updated_at = EXCLUDED.updated_at
	`)
	if err != nil {
		return fmt.Errorf("准备插入语句失败: %w", err)
	}
	defer stmt.Close()

	// 统计计数器
	var (
		insertedCount = 0
		updatedCount  = 0
		errorCount    = 0
		count985      = 0
		count211      = 0
		countDouble   = 0
	)

	// 批量插入数据
	for i, uni := range universities {
		now := time.Now()
		
		// 执行插入/更新
		result, err := stmt.Exec(
			uni.Name,
			uni.Code,
			uni.Supervisor,
			uni.Location,
			uni.Province,
			uni.City,
			uni.Level,
			uni.Type,
			uni.Remark,
			uni.Is985,
			uni.Is211,
			uni.IsDoubleFirst,
			now,
			now,
		)
		
		if err != nil {
			log.Printf("插入第 %d 条记录失败 (%s): %v", i+1, uni.Name, err)
			errorCount++
			continue
		}

		// 检查是否为插入还是更新
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			insertedCount++
		} else {
			updatedCount++
		}

		// 统计985/211/双一流
		if uni.Is985 {
			count985++
		}
		if uni.Is211 {
			count211++
		}
		if uni.IsDoubleFirst {
			countDouble++
		}

		// 进度显示
		if (i+1)%10 == 0 || i == len(universities)-1 {
			fmt.Printf("\r进度: %d/%d (%.1f%%) - 插入: %d, 更新: %d, 错误: %d",
				i+1, len(universities), float64(i+1)/float64(len(universities))*100,
				insertedCount, updatedCount, errorCount)
		}
	}
	fmt.Println() // 换行

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	// 输出统计信息
	fmt.Println("\n=== 导入统计 ===")
	fmt.Printf("总计处理: %d 所大学\n", len(universities))
	fmt.Printf("成功插入: %d 条记录\n", insertedCount)
	fmt.Printf("成功更新: %d 条记录\n", updatedCount)
	fmt.Printf("失败记录: %d 条\n", errorCount)
	fmt.Printf("985高校: %d 所\n", count985)
	fmt.Printf("211高校: %d 所\n", count211)
	fmt.Printf("双一流高校: %d 所\n", countDouble)

	return nil
}

// 验证导入的数据
func verifyImportedData(db *sql.DB) error {
	fmt.Println("\n=== 数据验证 ===")

	// 查询总数
	var totalCount int
	err := db.QueryRow("SELECT COUNT(*) FROM universities").Scan(&totalCount)
	if err != nil {
		return fmt.Errorf("查询总数失败: %w", err)
	}
	fmt.Printf("数据库中总计: %d 所大学\n", totalCount)

	// 查询985高校数量
	var count985 int
	err = db.QueryRow("SELECT COUNT(*) FROM universities WHERE is_985 = true").Scan(&count985)
	if err != nil {
		return fmt.Errorf("查询985高校数量失败: %w", err)
	}
	fmt.Printf("985高校: %d 所\n", count985)

	// 查询211高校数量
	var count211 int
	err = db.QueryRow("SELECT COUNT(*) FROM universities WHERE is_211 = true").Scan(&count211)
	if err != nil {
		return fmt.Errorf("查询211高校数量失败: %w", err)
	}
	fmt.Printf("211高校: %d 所\n", count211)

	// 查询双一流高校数量
	var countDouble int
	err = db.QueryRow("SELECT COUNT(*) FROM universities WHERE is_double_first = true").Scan(&countDouble)
	if err != nil {
		return fmt.Errorf("查询双一流高校数量失败: %w", err)
	}
	fmt.Printf("双一流高校: %d 所\n", countDouble)

	// 按省份统计
	fmt.Println("\n按省份分布:")
	rows, err := db.Query(`
		SELECT province, COUNT(*) as count 
		FROM universities 
		WHERE province IS NOT NULL AND province != '' 
		GROUP BY province 
		ORDER BY count DESC 
		LIMIT 10
	`)
	if err != nil {
		return fmt.Errorf("查询省份分布失败: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var province string
		var count int
		err := rows.Scan(&province, &count)
		if err != nil {
			log.Printf("扫描省份数据失败: %v", err)
			continue
		}
		fmt.Printf("  %s: %d 所\n", province, count)
	}

	// 检查数据质量
	fmt.Println("\n数据质量检查:")
	
	// 检查缺失代码的记录
	var missingCode int
	err = db.QueryRow("SELECT COUNT(*) FROM universities WHERE code IS NULL OR code = ''").Scan(&missingCode)
	if err == nil {
		fmt.Printf("  缺失院校代码: %d 条\n", missingCode)
	}

	// 检查缺失名称的记录
	var missingName int
	err = db.QueryRow("SELECT COUNT(*) FROM universities WHERE name IS NULL OR name = ''").Scan(&missingName)
	if err == nil {
		fmt.Printf("  缺失院校名称: %d 条\n", missingName)
	}

	// 检查重复代码
	var duplicateCode int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM (
			SELECT code FROM universities 
			WHERE code IS NOT NULL AND code != '' 
			GROUP BY code HAVING COUNT(*) > 1
		) as duplicates
	`).Scan(&duplicateCode)
	if err == nil {
		fmt.Printf("  重复院校代码: %d 个\n", duplicateCode)
	}

	return nil
}

func main() {
	fmt.Println("🎓 MOE大学数据导入工具")
	fmt.Println("====================")
	fmt.Println()

	// 连接数据库
	fmt.Println("📡 连接数据库...")
	db, err := connectDB()
	if err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}
	defer db.Close()
	fmt.Println("✅ 数据库连接成功")

	// 创建或更新表结构
	fmt.Println("\n🏗️  创建/更新数据库表结构...")
	err = createOrUpdateTables(db)
	if err != nil {
		log.Fatalf("❌ 创建表失败: %v", err)
	}
	fmt.Println("✅ 表结构准备完成")

	// 导入MOE大学数据
	fmt.Println("\n📥 开始导入MOE大学数据...")
	err = importMOEUniversities(db, "moe_universities.json")
	if err != nil {
		log.Fatalf("❌ 导入MOE大学数据失败: %v", err)
	}
	fmt.Println("✅ MOE大学数据导入完成")

	// 验证导入的数据
	err = verifyImportedData(db)
	if err != nil {
		log.Printf("⚠️  数据验证失败: %v", err)
	} else {
		fmt.Println("✅ 数据验证通过")
	}

	fmt.Println("\n🎉 MOE大学数据导入完成！")
	fmt.Println("\n📝 使用说明：")
	fmt.Println("1. 数据已成功导入到 PostgreSQL 数据库")
	fmt.Println("2. 包含985/211/双一流标识字段")
	fmt.Println("3. 支持按省份、城市、类型等维度查询")
	fmt.Println("4. 可通过 data-service API 进行查询")
	fmt.Println("5. 下一步可以获取专业数据和录取数据")
}