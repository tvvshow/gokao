package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

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

// ExtendedAdmissionData 结构体，对应extended-data-service.go中的数据
type ExtendedAdmissionData struct {
	ID           int     `json:"id"`
	UniversityID int     `json:"university_id"`
	MajorID      int     `json:"major_id"`
	Year         int     `json:"year"`
	Province     string  `json:"province"`
	MinScore     float64 `json:"min_score"`
	AvgScore     float64 `json:"avg_score"`
	MaxScore     float64 `json:"max_score"`
	MinRank      int     `json:"min_rank"`
	AvgRank      int     `json:"avg_rank"`
	MaxRank      int     `json:"max_rank"`
}

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

// 检查admission_data表是否存在
func checkAdmissionTable(db *sql.DB) error {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'admission_data'
		)
	`).Scan(&exists)

	if err != nil {
		return fmt.Errorf("检查表存在性失败: %v", err)
	}

	if !exists {
		return fmt.Errorf("admission_data表不存在")
	}

	fmt.Println("✓ admission_data表存在")
	return nil
}

// 生成扩展录取数据（模拟extended-data-service.go中的generateExtendedAdmissionData函数）
func generateExtendedAdmissionData() []ExtendedAdmissionData {
	rand.Seed(time.Now().UnixNano())

	provinces := []string{"北京", "上海", "广东", "江苏", "浙江", "山东", "河南", "四川", "湖北", "湖南"}
	years := []int{2020, 2021, 2022, 2023}

	var data []ExtendedAdmissionData

	// 生成200条录取数据
	for i := 1; i <= 200; i++ {
		year := years[rand.Intn(len(years))]
		province := provinces[rand.Intn(len(provinces))]

		// 生成分数数据
		minScore := 500 + rand.Float64()*200 // 500-700分
		maxScore := minScore + 50 + rand.Float64()*100 // 比最低分高50-150分
		avgScore := (minScore + maxScore) / 2

		// 生成排名数据
		minRank := 1000 + rand.Intn(50000) // 1000-51000名
		maxRank := minRank + 5000 + rand.Intn(15000) // 比最高排名低5000-20000名
		avgRank := (minRank + maxRank) / 2

		admission := ExtendedAdmissionData{
			ID:           i,
			UniversityID: 1 + rand.Intn(100), // 假设有100所大学
			MajorID:      1 + rand.Intn(67),  // 假设有67个专业
			Year:         year,
			Province:     province,
			MinScore:     minScore,
			AvgScore:     avgScore,
			MaxScore:     maxScore,
			MinRank:      minRank,
			AvgRank:      avgRank,
			MaxRank:      maxRank,
		}

		data = append(data, admission)
	}

	return data
}

// 获取大学UUID映射
func getUniversityUUIDs(db *sql.DB) (map[int]string, error) {
	uuidMap := make(map[int]string)

	rows, err := db.Query("SELECT id, name FROM universities ORDER BY id LIMIT 100")
	if err != nil {
		return nil, fmt.Errorf("查询大学失败: %v", err)
	}
	defer rows.Close()

	i := 1
	for rows.Next() {
		var uuid, name string
		err := rows.Scan(&uuid, &name)
		if err != nil {
			return nil, fmt.Errorf("扫描大学数据失败: %v", err)
		}
		uuidMap[i] = uuid
		i++
	}

	return uuidMap, nil
}

// 获取专业UUID映射
func getMajorUUIDs(db *sql.DB) (map[int]string, error) {
	uuidMap := make(map[int]string)

	rows, err := db.Query("SELECT id, name FROM majors ORDER BY id LIMIT 67")
	if err != nil {
		return nil, fmt.Errorf("查询专业失败: %v", err)
	}
	defer rows.Close()

	i := 1
	for rows.Next() {
		var uuid, name string
		err := rows.Scan(&uuid, &name)
		if err != nil {
			return nil, fmt.Errorf("扫描专业数据失败: %v", err)
		}
		uuidMap[i] = uuid
		i++
	}

	return uuidMap, nil
}

// 导入录取数据
func importAdmissions(db *sql.DB) error {
	fmt.Println("开始导入录取数据...")

	// 获取UUID映射
	universityUUIDs, err := getUniversityUUIDs(db)
	if err != nil {
		return err
	}

	majorUUIDs, err := getMajorUUIDs(db)
	if err != nil {
		return err
	}

	// 生成录取数据
	admissions := generateExtendedAdmissionData()

	successCount := 0
	failCount := 0

	for _, admission := range admissions {
		// 获取对应的UUID
		universityUUID, exists := universityUUIDs[admission.UniversityID]
		if !exists {
			fmt.Printf("警告: 找不到大学ID %d 对应的UUID\n", admission.UniversityID)
			failCount++
			continue
		}

		majorUUID, exists := majorUUIDs[admission.MajorID]
		if !exists {
			fmt.Printf("警告: 找不到专业ID %d 对应的UUID\n", admission.MajorID)
			failCount++
			continue
		}

		// 插入数据
		_, err := db.Exec(`
			INSERT INTO admission_data (
				university_id, major_id, year, province, batch, category,
				min_score, max_score, avg_score, min_rank, avg_rank, max_rank,
				difficulty, admission_rate, competition
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		`, universityUUID, majorUUID, admission.Year, admission.Province, "本科一批", "理科",
			admission.MinScore, admission.MaxScore, admission.AvgScore,
			admission.MinRank, admission.AvgRank, admission.MaxRank,
			"中等", 0.8, 0.15)

		if err != nil {
			fmt.Printf("插入录取数据失败: %v\n", err)
			failCount++
		} else {
			successCount++
		}
	}

	fmt.Printf("录取数据导入完成: 成功 %d 条，失败 %d 条\n", successCount, failCount)
	return nil
}

// 验证录取数据
func verifyAdmissionData(db *sql.DB) error {
	// 统计总数
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM admission_data").Scan(&count)
	if err != nil {
		return fmt.Errorf("查询录取数据总数失败: %v", err)
	}
	fmt.Printf("数据库中共有 %d 条录取数据\n\n", count)

	// 按年份统计
	fmt.Println("按年份统计:")
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

	// 按省份统计
	fmt.Println("\n按省份统计:")
	rows2, err := db.Query(`
		SELECT province, COUNT(*) as count
		FROM admission_data
		GROUP BY province
		ORDER BY count DESC
	`)
	if err != nil {
		return fmt.Errorf("按省份统计失败: %v", err)
	}
	defer rows2.Close()

	for rows2.Next() {
		var province string
		var count int
		err := rows2.Scan(&province, &count)
		if err != nil {
			return fmt.Errorf("扫描省份统计失败: %v", err)
		}
		fmt.Printf("  %s: %d 条记录\n", province, count)
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

	fmt.Println("成功连接到数据库!")

	// 检查表是否存在
	if err := checkAdmissionTable(db); err != nil {
		log.Fatal(err)
	}

	// 导入录取数据
	if err := importAdmissions(db); err != nil {
		log.Fatal("导入录取数据失败:", err)
	}

	// 验证数据
	if err := verifyAdmissionData(db); err != nil {
		log.Fatal("验证录取数据失败:", err)
	}

	fmt.Println("\n录取数据导入完成!")
}