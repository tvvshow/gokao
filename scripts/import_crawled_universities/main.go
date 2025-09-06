package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/gaokaohub/gaokao/pkg/scripts"
)

// UniversityInfo 高校信息结构（与爬虫脚本保持一致）
type UniversityInfo struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Code         string    `json:"code"`
	Province     string    `json:"province"`
	City         string    `json:"city"`
	Type         string    `json:"type"`
	Level        string    `json:"level"`
	Supervisor   string    `json:"supervisor"`
	Is985        bool      `json:"is_985"`
	Is211        bool      `json:"is_211"`
	IsDoubleFirst bool     `json:"is_double_first"`
	Website      string    `json:"website"`
	Source       string    `json:"source"`
	CrawledAt    time.Time `json:"crawled_at"`
}

// ImportStats 导入统计
type ImportStats struct {
	TotalRecords     int           `json:"total_records"`
	SuccessfulImports int          `json:"successful_imports"`
	FailedImports    int           `json:"failed_imports"`
	DuplicateSkipped int           `json:"duplicate_skipped"`
	StartTime        time.Time     `json:"start_time"`
	EndTime          time.Time     `json:"end_time"`
	Duration         time.Duration `json:"duration"`
	Errors           []string      `json:"errors"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	SSLMode  string `json:"sslmode"`
}

func main() {
	fmt.Println("🗄️  高校数据导入工具")
	fmt.Println("将爬取的高校数据导入PostgreSQL数据库")
	fmt.Println("==========================================\n")

	// 检查数据文件
	dataFile := "all_universities.json"
	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		log.Fatalf("❌ 数据文件不存在: %s\n请先运行 advanced_university_crawler.go", dataFile)
	}

	// 读取高校数据
	fmt.Printf("📖 读取数据文件: %s\n", dataFile)
	universities, err := loadUniversities(dataFile)
	if err != nil {
		log.Fatalf("❌ 读取数据失败: %v", err)
	}
	fmt.Printf("✅ 成功读取 %d 条高校记录\n\n", len(universities))

	// 连接数据库
	fmt.Println("🔌 连接数据库...")
	db, err := connectDatabase()
	if err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}
	defer db.Close()
	fmt.Println("✅ 数据库连接成功\n")

	// 检查并创建表
	err = ensureTableExists(db)
	if err != nil {
		log.Fatalf("❌ 表创建失败: %v", err)
	}

	// 执行数据导入
	stats := importUniversities(db, universities)

	// 打印导入统计
	printImportStats(stats)

	// 保存导入报告
	saveImportReport(stats)

	// 验证导入结果
	verifyImportResults(db)

	fmt.Println("\n🎉 数据导入完成！")
}

// loadUniversities 读取高校数据
func loadUniversities(filename string) ([]UniversityInfo, error) {
	fileUtil := scripts.NewFileWriter()
	data, err := fileUtil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}

	var universities []UniversityInfo
	err = json.Unmarshal(data, &universities)
	if err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	return universities, nil
}

// connectDatabase 连接数据库
func connectDatabase() (*sql.DB, error) {
	// 读取数据库配置
	config := getDefaultDatabaseConfig()
	
	// 尝试从环境变量读取配置
	if host := os.Getenv("DB_HOST"); host != "" {
		config.Host = host
	}
	if user := os.Getenv("DB_USER"); user != "" {
		config.User = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		config.Password = password
	}
	if dbname := os.Getenv("DB_NAME"); dbname != "" {
		config.DBName = dbname
	}

	// 构建连接字符串
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)

	fmt.Printf("   📍 连接到: %s@%s:%d/%s\n", config.User, config.Host, config.Port, config.DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("打开数据库连接失败: %v", err)
	}

	// 测试连接
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("数据库连接测试失败: %v", err)
	}

	return db, nil
}

// getDefaultDatabaseConfig 获取默认数据库配置
func getDefaultDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "123456",
		DBName:   "gaokao_user_db",
		SSLMode:  "disable",
	}
}

// ensureTableExists 确保表存在
func ensureTableExists(db *sql.DB) error {
	fmt.Println("🏗️  检查数据表...")

	// 检查表是否存在
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'universities'
		)
	`).Scan(&exists)

	if err != nil {
		return fmt.Errorf("检查表存在性失败: %v", err)
	}

	if exists {
		fmt.Println("   ✅ universities 表已存在")
		return nil
	}

	// 创建表
	fmt.Println("   🔨 创建 universities 表...")
	createTableSQL := `
		CREATE TABLE universities (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			code VARCHAR(50),
			province VARCHAR(50) NOT NULL,
			city VARCHAR(100),
			type VARCHAR(50),
			level VARCHAR(50),
			supervisor VARCHAR(100),
			is_985 BOOLEAN DEFAULT FALSE,
			is_211 BOOLEAN DEFAULT FALSE,
			is_double_first BOOLEAN DEFAULT FALSE,
			website VARCHAR(255),
			source VARCHAR(255),
			crawled_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		-- 创建索引
		CREATE INDEX idx_universities_province ON universities(province);
		CREATE INDEX idx_universities_type ON universities(type);
		CREATE INDEX idx_universities_985 ON universities(is_985);
		CREATE INDEX idx_universities_211 ON universities(is_211);
		CREATE INDEX idx_universities_double_first ON universities(is_double_first);
	`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("创建表失败: %v", err)
	}

	fmt.Println("   ✅ universities 表创建成功")
	return nil
}

// importUniversities 导入高校数据（使用COPY命令优化性能）
func importUniversities(db *sql.DB, universities []UniversityInfo) *ImportStats {
	fmt.Printf("\n📥 开始导入 %d 条高校数据（使用COPY命令）...\n", len(universities))

	stats := &ImportStats{
		TotalRecords: len(universities),
		StartTime:    time.Now(),
		Errors:       make([]string, 0),
	}

	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		stats.Errors = append(stats.Errors, fmt.Sprintf("开始事务失败: %v", err))
		return stats
	}
	defer tx.Rollback()

	// 创建临时表
	_, err = tx.Exec(`
		CREATE TEMP TABLE temp_universities (
			name TEXT,
			code TEXT,
			province TEXT,
			city TEXT,
			type TEXT,
			level TEXT,
			supervisor TEXT,
			is_985 BOOLEAN,
			is_211 BOOLEAN,
			is_double_first BOOLEAN,
			website TEXT,
			source TEXT,
			crawled_at TIMESTAMP
		)
	`)
	if err != nil {
		stats.Errors = append(stats.Errors, fmt.Sprintf("创建临时表失败: %v", err))
		return stats
	}

	// 准备COPY数据
	var copyData strings.Builder
	for i, uni := range universities {
		copyData.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%t\t%t\t%t\t%s\t%s\t%s\n",
			uni.Name,
			uni.Code,
			uni.Province,
			uni.City,
			uni.Type,
			uni.Level,
			uni.Supervisor,
			uni.Is985,
			uni.Is211,
			uni.IsDoubleFirst,
			uni.Website,
			uni.Source,
			uni.CrawledAt.Format("2006-01-02 15:04:05"),
		))

		// 每1000条记录输出进度
		if (i+1)%1000 == 0 {
			fmt.Printf("   📊 已准备 %d/%d 条记录\n", i+1, len(universities))
		}
	}

	// 使用COPY命令批量插入到临时表
	copyStmt := `COPY temp_universities FROM STDIN WITH (FORMAT text, DELIMITER E'\t')`
	_, err = tx.Exec(copyStmt, copyData.String())
	if err != nil {
		stats.Errors = append(stats.Errors, fmt.Sprintf("COPY命令执行失败: %v", err))
		return stats
	}

	// 从临时表插入到主表（使用UPSERT）
	insertSQL := `
		INSERT INTO universities (
			name, code, province, city, type, level, supervisor,
			is_985, is_211, is_double_first, website, source, crawled_at
		)
		SELECT 
			name, code, province, city, type, level, supervisor,
			is_985, is_211, is_double_first, website, source, crawled_at
		FROM temp_universities
		ON CONFLICT (name) DO UPDATE SET
			code = EXCLUDED.code,
			province = EXCLUDED.province,
			city = EXCLUDED.city,
			type = EXCLUDED.type,
			level = EXCLUDED.level,
			supervisor = EXCLUDED.supervisor,
			is_985 = EXCLUDED.is_985,
			is_211 = EXCLUDED.is_211,
			is_double_first = EXCLUDED.is_double_first,
			website = EXCLUDED.website,
			source = EXCLUDED.source,
			crawled_at = EXCLUDED.crawled_at,
			updated_at = CURRENT_TIMESTAMP
	`

	result, err := tx.Exec(insertSQL)
	if err != nil {
		stats.Errors = append(stats.Errors, fmt.Sprintf("从临时表插入失败: %v", err))
		return stats
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		stats.Errors = append(stats.Errors, fmt.Sprintf("获取影响行数失败: %v", err))
	} else {
		stats.SuccessfulImports = int(rowsAffected)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		stats.Errors = append(stats.Errors, fmt.Sprintf("提交事务失败: %v", err))
		return stats
	}

	stats.EndTime = time.Now()
	stats.Duration = stats.EndTime.Sub(stats.StartTime)

	fmt.Printf("✅ COPY导入完成! 成功导入: %d 条记录, 耗时: %v\n",
		stats.SuccessfulImports, stats.Duration)

	return stats
}

// printImportStats 打印导入统计
func printImportStats(stats *ImportStats) {
	fmt.Printf("\n📊 导入完成统计:\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("⏱️  总用时: %v\n", stats.Duration)
	fmt.Printf("📝 总记录数: %d\n", stats.TotalRecords)
	fmt.Printf("✅ 成功导入: %d (%.1f%%)\n", stats.SuccessfulImports, float64(stats.SuccessfulImports)/float64(stats.TotalRecords)*100)
	fmt.Printf("❌ 导入失败: %d (%.1f%%)\n", stats.FailedImports, float64(stats.FailedImports)/float64(stats.TotalRecords)*100)
	fmt.Printf("⚡ 导入速度: %.1f 记录/秒\n", float64(stats.TotalRecords)/stats.Duration.Seconds())

	if len(stats.Errors) > 0 {
		fmt.Printf("\n⚠️  错误详情 (前10条):\n")
		for i, err := range stats.Errors {
			if i >= 10 {
				fmt.Printf("   ... 还有 %d 个错误\n", len(stats.Errors)-10)
				break
			}
			fmt.Printf("   %d. %s\n", i+1, err)
		}
	}
}

// saveImportReport 保存导入报告
func saveImportReport(stats *ImportStats) {
	reportFile := "import_report.json"
	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		log.Printf("❌ 生成导入报告失败: %v", err)
		return
	}

	fileUtil := scripts.NewFileWriter()
	err = fileUtil.SaveToFile(data, reportFile)
	if err != nil {
		log.Printf("❌ 保存导入报告失败: %v", err)
		return
	}

	fmt.Printf("✅ 导入报告已保存到: %s\n", reportFile)
}

// verifyImportResults 验证导入结果
func verifyImportResults(db *sql.DB) {
	fmt.Printf("\n🔍 验证导入结果...\n")

	// 统计总数
	var totalCount int
	err := db.QueryRow("SELECT COUNT(*) FROM universities").Scan(&totalCount)
	if err != nil {
		log.Printf("❌ 查询总数失败: %v", err)
		return
	}
	fmt.Printf("   📊 数据库中总高校数: %d\n", totalCount)

	// 按省份统计
	rows, err := db.Query(`
		SELECT province, COUNT(*) as count 
		FROM universities 
		GROUP BY province 
		ORDER BY count DESC 
		LIMIT 10
	`)
	if err != nil {
		log.Printf("❌ 查询省份统计失败: %v", err)
		return
	}
	defer rows.Close()

	fmt.Printf("\n🗺️  各省高校数量 (前10):\n")
	for rows.Next() {
		var province string
		var count int
		err := rows.Scan(&province, &count)
		if err != nil {
			log.Printf("❌ 读取省份统计失败: %v", err)
			continue
		}
		fmt.Printf("   %s: %d 所\n", province, count)
	}

	// 985/211统计
	var count985, count211, countDoubleFirst int
	db.QueryRow("SELECT COUNT(*) FROM universities WHERE is_985 = true").Scan(&count985)
	db.QueryRow("SELECT COUNT(*) FROM universities WHERE is_211 = true").Scan(&count211)
	db.QueryRow("SELECT COUNT(*) FROM universities WHERE is_double_first = true").Scan(&countDoubleFirst)

	fmt.Printf("\n🏆 重点高校统计:\n")
	fmt.Printf("   985高校: %d 所\n", count985)
	fmt.Printf("   211高校: %d 所\n", count211)
	fmt.Printf("   双一流: %d 所\n", countDoubleFirst)

	// 按类型统计
	rows2, err := db.Query(`
		SELECT type, COUNT(*) as count 
		FROM universities 
		WHERE type IS NOT NULL AND type != ''
		GROUP BY type 
		ORDER BY count DESC
	`)
	if err != nil {
		log.Printf("❌ 查询类型统计失败: %v", err)
		return
	}
	defer rows2.Close()

	fmt.Printf("\n📚 高校类型分布:\n")
	for rows2.Next() {
		var uType string
		var count int
		err := rows2.Scan(&uType, &count)
		if err != nil {
			log.Printf("❌ 读取类型统计失败: %v", err)
			continue
		}
		fmt.Printf("   %s: %d 所\n", uType, count)
	}

	fmt.Println("\n✅ 数据验证完成")
}