package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// 数据库连接配置
const (
	DBHost     = "localhost"
	DBPort     = 5432
	DBUser     = "gaokao_user"
	DBPassword = "gaokao_password"
	DBName     = "gaokao_user_db"
)

// 连接数据库
func connectDB() (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		DBHost, DBPort, DBUser, DBPassword, DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// 验证大学数据
func verifyUniversityData(db *sql.DB) error {
	fmt.Println("=== 验证大学数据 ===")

	// 统计总数
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM universities").Scan(&count)
	if err != nil {
		return fmt.Errorf("查询大学总数失败: %v", err)
	}
	fmt.Printf("✓ 大学总数: %d\n", count)

	// 检查必填字段
	var nullCount int
	err = db.QueryRow("SELECT COUNT(*) FROM universities WHERE name IS NULL OR name = ''").Scan(&nullCount)
	if err != nil {
		return fmt.Errorf("检查大学名称失败: %v", err)
	}
	if nullCount > 0 {
		fmt.Printf("⚠️  发现 %d 所大学名称为空\n", nullCount)
	} else {
		fmt.Println("✓ 所有大学都有名称")
	}

	// 检查重复名称
	var duplicateCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM (
			SELECT name, COUNT(*) as cnt
			FROM universities
			GROUP BY name
			HAVING COUNT(*) > 1
		) duplicates
	`).Scan(&duplicateCount)
	if err != nil {
		return fmt.Errorf("检查重复大学名称失败: %v", err)
	}
	if duplicateCount > 0 {
		fmt.Printf("⚠️  发现 %d 个重复的大学名称\n", duplicateCount)
	} else {
		fmt.Println("✓ 没有重复的大学名称")
	}

	// 按类型统计
	fmt.Println("\n按类型统计:")
	rows, err := db.Query(`
		SELECT type, COUNT(*) as count
		FROM universities
		WHERE type IS NOT NULL
		GROUP BY type
		ORDER BY count DESC
	`)
	if err != nil {
		return fmt.Errorf("按类型统计失败: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var uType string
		var count int
		err := rows.Scan(&uType, &count)
		if err != nil {
			return fmt.Errorf("扫描类型统计失败: %v", err)
		}
		fmt.Printf("  %s: %d 所\n", uType, count)
	}

	return nil
}

// 验证专业数据
func verifyMajorData(db *sql.DB) error {
	fmt.Println("\n=== 验证专业数据 ===")

	// 统计总数
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM majors").Scan(&count)
	if err != nil {
		return fmt.Errorf("查询专业总数失败: %v", err)
	}
	fmt.Printf("✓ 专业总数: %d\n", count)

	// 检查必填字段
	var nullCount int
	err = db.QueryRow("SELECT COUNT(*) FROM majors WHERE name IS NULL OR name = ''").Scan(&nullCount)
	if err != nil {
		return fmt.Errorf("检查专业名称失败: %v", err)
	}
	if nullCount > 0 {
		fmt.Printf("⚠️  发现 %d 个专业名称为空\n", nullCount)
	} else {
		fmt.Println("✓ 所有专业都有名称")
	}

	// 检查重复代码
	var duplicateCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM (
			SELECT code, COUNT(*) as cnt
			FROM majors
			WHERE code IS NOT NULL AND code != ''
			GROUP BY code
			HAVING COUNT(*) > 1
		) duplicates
	`).Scan(&duplicateCount)
	if err != nil {
		return fmt.Errorf("检查重复专业代码失败: %v", err)
	}
	if duplicateCount > 0 {
		fmt.Printf("⚠️  发现 %d 个重复的专业代码\n", duplicateCount)
	} else {
		fmt.Println("✓ 没有重复的专业代码")
	}

	// 按学科门类统计
	fmt.Println("\n按学科门类统计:")
	rows, err := db.Query(`
		SELECT category, COUNT(*) as count
		FROM majors
		WHERE category IS NOT NULL
		GROUP BY category
		ORDER BY count DESC
	`)
	if err != nil {
		return fmt.Errorf("按学科门类统计失败: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var category string
		var count int
		err := rows.Scan(&category, &count)
		if err != nil {
			return fmt.Errorf("扫描学科门类统计失败: %v", err)
		}
		fmt.Printf("  %s: %d 个专业\n", category, count)
	}

	// 检查外键关系
	var orphanCount int
	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM majors m
		LEFT JOIN universities u ON m.university_id = u.id
		WHERE m.university_id IS NOT NULL AND u.id IS NULL
	`).Scan(&orphanCount)
	if err != nil {
		return fmt.Errorf("检查专业外键关系失败: %v", err)
	}
	if orphanCount > 0 {
		fmt.Printf("⚠️  发现 %d 个专业的大学ID无效\n", orphanCount)
	} else {
		fmt.Println("✓ 所有专业的大学ID都有效")
	}

	return nil
}

// 验证录取数据
func verifyAdmissionData(db *sql.DB) error {
	fmt.Println("\n=== 验证录取数据 ===")

	// 统计总数
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM admission_data").Scan(&count)
	if err != nil {
		return fmt.Errorf("查询录取数据总数失败: %v", err)
	}
	fmt.Printf("✓ 录取数据总数: %d\n", count)

	// 检查必填字段
	var nullUniversityCount int
	err = db.QueryRow("SELECT COUNT(*) FROM admission_data WHERE university_id IS NULL").Scan(&nullUniversityCount)
	if err != nil {
		return fmt.Errorf("检查大学ID失败: %v", err)
	}
	if nullUniversityCount > 0 {
		fmt.Printf("⚠️  发现 %d 条录取数据缺少大学ID\n", nullUniversityCount)
	} else {
		fmt.Println("✓ 所有录取数据都有大学ID")
	}

	var nullYearCount int
	err = db.QueryRow("SELECT COUNT(*) FROM admission_data WHERE year IS NULL").Scan(&nullYearCount)
	if err != nil {
		return fmt.Errorf("检查年份失败: %v", err)
	}
	if nullYearCount > 0 {
		fmt.Printf("⚠️  发现 %d 条录取数据缺少年份\n", nullYearCount)
	} else {
		fmt.Println("✓ 所有录取数据都有年份")
	}

	// 检查分数数据合理性
	var invalidScoreCount int
	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM admission_data
		WHERE min_score IS NOT NULL AND max_score IS NOT NULL
		AND min_score > max_score
	`).Scan(&invalidScoreCount)
	if err != nil {
		return fmt.Errorf("检查分数合理性失败: %v", err)
	}
	if invalidScoreCount > 0 {
		fmt.Printf("⚠️  发现 %d 条录取数据最低分大于最高分\n", invalidScoreCount)
	} else {
		fmt.Println("✓ 所有录取数据分数范围合理")
	}

	// 检查排名数据合理性
	var invalidRankCount int
	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM admission_data
		WHERE min_rank IS NOT NULL AND max_rank IS NOT NULL
		AND min_rank > max_rank
	`).Scan(&invalidRankCount)
	if err != nil {
		return fmt.Errorf("检查排名合理性失败: %v", err)
	}
	if invalidRankCount > 0 {
		fmt.Printf("⚠️  发现 %d 条录取数据最高排名大于最低排名\n", invalidRankCount)
	} else {
		fmt.Println("✓ 所有录取数据排名范围合理")
	}

	// 按年份统计
	fmt.Println("\n按年份统计:")
	rows, err := db.Query(`
		SELECT year, COUNT(*) as count
		FROM admission_data
		GROUP BY year
		ORDER BY year
	`)
	if err != nil {
		return fmt.Errorf("按年份统计失败: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var year, count int
		err := rows.Scan(&year, &count)
		if err != nil {
			return fmt.Errorf("扫描年份统计失败: %v", err)
		}
		fmt.Printf("  %d年: %d 条记录\n", year, count)
	}

	// 检查外键关系
	var orphanUniversityCount int
	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM admission_data a
		LEFT JOIN universities u ON a.university_id = u.id
		WHERE a.university_id IS NOT NULL AND u.id IS NULL
	`).Scan(&orphanUniversityCount)
	if err != nil {
		return fmt.Errorf("检查录取数据大学外键关系失败: %v", err)
	}
	if orphanUniversityCount > 0 {
		fmt.Printf("⚠️  发现 %d 条录取数据的大学ID无效\n", orphanUniversityCount)
	} else {
		fmt.Println("✓ 所有录取数据的大学ID都有效")
	}

	var orphanMajorCount int
	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM admission_data a
		LEFT JOIN majors m ON a.major_id = m.id
		WHERE a.major_id IS NOT NULL AND m.id IS NULL
	`).Scan(&orphanMajorCount)
	if err != nil {
		return fmt.Errorf("检查录取数据专业外键关系失败: %v", err)
	}
	if orphanMajorCount > 0 {
		fmt.Printf("⚠️  发现 %d 条录取数据的专业ID无效\n", orphanMajorCount)
	} else {
		fmt.Println("✓ 所有录取数据的专业ID都有效")
	}

	return nil
}

// 数据关联性验证
func verifyDataRelationships(db *sql.DB) error {
	fmt.Println("\n=== 验证数据关联性 ===")

	// 检查有录取数据的大学数量
	var universitiesWithAdmissions int
	err := db.QueryRow(`
		SELECT COUNT(DISTINCT university_id)
		FROM admission_data
		WHERE university_id IS NOT NULL
	`).Scan(&universitiesWithAdmissions)
	if err != nil {
		return fmt.Errorf("查询有录取数据的大学数量失败: %v", err)
	}
	fmt.Printf("✓ 有录取数据的大学数量: %d\n", universitiesWithAdmissions)

	// 检查有录取数据的专业数量
	var majorsWithAdmissions int
	err = db.QueryRow(`
		SELECT COUNT(DISTINCT major_id)
		FROM admission_data
		WHERE major_id IS NOT NULL
	`).Scan(&majorsWithAdmissions)
	if err != nil {
		return fmt.Errorf("查询有录取数据的专业数量失败: %v", err)
	}
	fmt.Printf("✓ 有录取数据的专业数量: %d\n", majorsWithAdmissions)

	// 检查数据覆盖率
	var totalUniversities int
	err = db.QueryRow("SELECT COUNT(*) FROM universities").Scan(&totalUniversities)
	if err != nil {
		return fmt.Errorf("查询大学总数失败: %v", err)
	}

	var totalMajors int
	err = db.QueryRow("SELECT COUNT(*) FROM majors").Scan(&totalMajors)
	if err != nil {
		return fmt.Errorf("查询专业总数失败: %v", err)
	}

	universityCoverage := float64(universitiesWithAdmissions) / float64(totalUniversities) * 100
	majorCoverage := float64(majorsWithAdmissions) / float64(totalMajors) * 100

	fmt.Printf("✓ 大学数据覆盖率: %.1f%% (%d/%d)\n", universityCoverage, universitiesWithAdmissions, totalUniversities)
	fmt.Printf("✓ 专业数据覆盖率: %.1f%% (%d/%d)\n", majorCoverage, majorsWithAdmissions, totalMajors)

	return nil
}

// 生成数据质量报告
func generateQualityReport(db *sql.DB) error {
	fmt.Println("\n=== 数据质量报告 ===")

	// 计算数据完整性得分
	var completenessScore float64 = 100.0

	// 检查各表的空值比例
	tables := []string{"universities", "majors", "admission_data"}
	for _, table := range tables {
		var totalRows, nullRows int

		// 获取总行数
		err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&totalRows)
		if err != nil {
			return fmt.Errorf("查询%s总行数失败: %v", table, err)
		}

		// 根据表检查关键字段的空值
		switch table {
		case "universities":
			err = db.QueryRow("SELECT COUNT(*) FROM universities WHERE name IS NULL OR name = ''").Scan(&nullRows)
		case "majors":
			err = db.QueryRow("SELECT COUNT(*) FROM majors WHERE name IS NULL OR name = ''").Scan(&nullRows)
		case "admission_data":
			err = db.QueryRow("SELECT COUNT(*) FROM admission_data WHERE university_id IS NULL OR year IS NULL").Scan(&nullRows)
		}

		if err != nil {
			return fmt.Errorf("查询%s空值失败: %v", table, err)
		}

		if totalRows > 0 {
			nullPercentage := float64(nullRows) / float64(totalRows) * 100
			fmt.Printf("  %s表关键字段完整性: %.1f%% (空值: %d/%d)\n", table, 100-nullPercentage, nullRows, totalRows)
			completenessScore -= nullPercentage / 3 // 平均分配到三个表
		}
	}

	fmt.Printf("\n📊 总体数据质量评分: %.1f/100\n", completenessScore)

	if completenessScore >= 90 {
		fmt.Println("🎉 数据质量优秀!")
	} else if completenessScore >= 80 {
		fmt.Println("✅ 数据质量良好")
	} else if completenessScore >= 70 {
		fmt.Println("⚠️  数据质量一般，建议改进")
	} else {
		fmt.Println("❌ 数据质量较差，需要修复")
	}

	return nil
}

func main() {
	// 连接数据库
	db, err := connectDB()
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	fmt.Println("🔍 开始数据完整性验证...")
	fmt.Println("✅ 数据库连接成功")

	// 验证各表数据
	if err := verifyUniversityData(db); err != nil {
		log.Printf("验证大学数据失败: %v", err)
	}

	if err := verifyMajorData(db); err != nil {
		log.Printf("验证专业数据失败: %v", err)
	}

	if err := verifyAdmissionData(db); err != nil {
		log.Printf("验证录取数据失败: %v", err)
	}

	if err := verifyDataRelationships(db); err != nil {
		log.Printf("验证数据关联性失败: %v", err)
	}

	if err := generateQualityReport(db); err != nil {
		log.Printf("生成质量报告失败: %v", err)
	}

	fmt.Println("\n✅ 数据完整性验证完成!")
}