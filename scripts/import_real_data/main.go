package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/tvvshow/gokao/pkg/shared"
	_ "github.com/lib/pq"
)

// 数据库配置
const (
	DBHost     = "localhost"
	DBPort     = 5432
	DBUser     = "postgres"
	DBPassword = "password"
	DBName     = "gaokao_data"
)

// 大学数据结构（对应API返回）
type UniversityData struct {
	Name        string `json:"name"`
	ID          string `json:"id"`
	Province    string `json:"province"`
	City        string `json:"city"`
	Level       string `json:"level"`       // 985/211/普通本科等
	Type        string `json:"type"`        // 综合/理工/师范等
	FoundedYear int    `json:"founded_year"`
	Website     string `json:"website"`
}

// 专业数据结构（对应API返回）
type MajorData struct {
	EducationLevel          string `json:"EducationLevel"`
	DisciplinaryCategory    string `json:"DisciplinaryCategory"`
	DisciplinarySubCategory string `json:"DisciplinarySubCategory"`
	MajorCode               string `json:"MajorCode"`
	MajorName               string `json:"MajorName"`
	MajorIntroduction       string `json:"MajorIntroduction"`
	GraduateScale           string `json:"GraduateScale"`
	MaleFemaleRatio         string `json:"MaleFemaleRatio"`
}

// 录取数据结构
type AdmissionRecord struct {
	Year         int    `json:"year"`
	Province     string `json:"province"`
	University   string `json:"university"`
	UniversityID string `json:"university_id"`
	Major        string `json:"major"`
	MajorCode    string `json:"major_code"`
	Batch        string `json:"batch"`
	MinScore     int    `json:"min_score"`
	MaxScore     int    `json:"max_score"`
	AvgScore     int    `json:"avg_score"`
	AdmitNum     int    `json:"admit_num"`
	SubjectType  string `json:"subject_type"`
}

// 数据库连接
func connectDB() (*sql.DB, error) {
	return shared.ConnectDB()
}

// 创建数据库表
func createTables(db *sql.DB) error {
	// 创建大学表
	universityTable := `
	CREATE TABLE IF NOT EXISTS universities (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL UNIQUE,
		code VARCHAR(50) UNIQUE,
		province VARCHAR(100),
		city VARCHAR(100),
		level VARCHAR(50),
		type VARCHAR(100),
		founded_year INTEGER,
		website VARCHAR(500),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	// 创建专业表
	majorTable := `
	CREATE TABLE IF NOT EXISTS majors (
		id SERIAL PRIMARY KEY,
		code VARCHAR(50) NOT NULL UNIQUE,
		name VARCHAR(255) NOT NULL,
		category VARCHAR(100),
		sub_category VARCHAR(100),
		education_level VARCHAR(50),
		description TEXT,
		graduate_scale VARCHAR(50),
		gender_ratio VARCHAR(50),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	// 创建录取数据表
	admissionTable := `
	CREATE TABLE IF NOT EXISTS admissions (
		id SERIAL PRIMARY KEY,
		year INTEGER NOT NULL,
		province VARCHAR(100) NOT NULL,
		university_id INTEGER REFERENCES universities(id),
		major_id INTEGER REFERENCES majors(id),
		batch VARCHAR(100),
		min_score INTEGER,
		max_score INTEGER,
		avg_score INTEGER,
		admit_count INTEGER,
		subject_type VARCHAR(50),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(year, province, university_id, major_id, batch, subject_type)
	);
	`

	// 创建索引
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_universities_name ON universities(name);",
		"CREATE INDEX IF NOT EXISTS idx_universities_province ON universities(province);",
		"CREATE INDEX IF NOT EXISTS idx_majors_code ON majors(code);",
		"CREATE INDEX IF NOT EXISTS idx_majors_name ON majors(name);",
		"CREATE INDEX IF NOT EXISTS idx_admissions_year ON admissions(year);",
		"CREATE INDEX IF NOT EXISTS idx_admissions_province ON admissions(province);",
		"CREATE INDEX IF NOT EXISTS idx_admissions_university ON admissions(university_id);",
		"CREATE INDEX IF NOT EXISTS idx_admissions_major ON admissions(major_id);",
	}

	// 执行建表语句
	tables := []string{universityTable, majorTable, admissionTable}
	for i, table := range tables {
		fmt.Printf("创建表 %d/3...\n", i+1)
		_, err := db.Exec(table)
		if err != nil {
			return fmt.Errorf("创建表失败: %v", err)
		}
	}

	// 创建索引
	fmt.Println("创建索引...")
	for _, index := range indexes {
		_, err := db.Exec(index)
		if err != nil {
			return fmt.Errorf("创建索引失败: %v", err)
		}
	}

	return nil
}

// 导入大学数据（使用COPY命令批量导入）
func importUniversities(db *sql.DB, filename string) error {
	fmt.Printf("正在导入大学数据从 %s (使用COPY命令)...\n", filename)

	// 检查文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Printf("⚠️  文件 %s 不存在，跳过大学数据导入\n", filename)
		return nil
	}

	// 读取JSON文件
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	var universities []UniversityData
	err = json.Unmarshal(data, &universities)
	if err != nil {
		return err
	}

	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 创建临时表
	_, err = tx.Exec(`
		CREATE TEMP TABLE temp_universities (
			name TEXT,
			code TEXT,
			province TEXT,
			city TEXT,
			level TEXT,
			type TEXT,
			founded_year INTEGER,
			website TEXT
		)
	`)
	if err != nil {
		return fmt.Errorf("创建临时表失败: %v", err)
	}

	// 准备COPY数据
	var copyData strings.Builder
	for _, uni := range universities {
		code := generateUniversityCode(uni.Name, uni.ID)
		copyData.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%d\t%s\n",
			uni.Name,
			code,
			uni.Province,
			uni.City,
			uni.Level,
			uni.Type,
			uni.FoundedYear,
			uni.Website))
	}

	// 使用COPY命令批量插入到临时表
	copyStmt := `COPY temp_universities FROM STDIN WITH (FORMAT text, DELIMITER E'\t')`
	_, err = tx.Exec(copyStmt, copyData.String())
	if err != nil {
		return fmt.Errorf("COPY命令执行失败: %v", err)
	}

	// 从临时表插入到主表（使用UPSERT）
	_, err = tx.Exec(`
		INSERT INTO universities (name, code, province, city, level, type, founded_year, website)
		SELECT name, code, province, city, level, type, founded_year, website
		FROM temp_universities
		ON CONFLICT (name) DO UPDATE SET
			code = EXCLUDED.code,
			province = EXCLUDED.province,
			city = EXCLUDED.city,
			level = EXCLUDED.level,
			type = EXCLUDED.type,
			founded_year = EXCLUDED.founded_year,
			website = EXCLUDED.website,
			updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return fmt.Errorf("从临时表插入失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return err
	}

	fmt.Printf("✅ 成功使用COPY命令导入 %d 所大学数据\n", len(universities))
	return nil
}

// 导入专业数据
func importMajors(db *sql.DB, filename string) error {
	fmt.Printf("正在导入专业数据从 %s...\n", filename)

	// 检查文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Printf("⚠️  文件 %s 不存在，跳过专业数据导入\n", filename)
		return nil
	}

	// 读取JSON文件
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	var majors []MajorData
	err = json.Unmarshal(data, &majors)
	if err != nil {
		return err
	}

	// 批量插入
	stmt, err := db.Prepare(`
		INSERT INTO majors (code, name, category, sub_category, education_level, description, graduate_scale, gender_ratio) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
		ON CONFLICT (code) DO UPDATE SET 
			name = EXCLUDED.name,
			category = EXCLUDED.category,
			sub_category = EXCLUDED.sub_category,
			education_level = EXCLUDED.education_level,
			description = EXCLUDED.description,
			graduate_scale = EXCLUDED.graduate_scale,
			gender_ratio = EXCLUDED.gender_ratio,
			updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	successCount := 0
	for i, major := range majors {
		_, err = stmt.Exec(
			major.MajorCode,
			major.MajorName,
			major.DisciplinaryCategory,
			major.DisciplinarySubCategory,
			major.EducationLevel,
			major.MajorIntroduction,
			major.GraduateScale,
			major.MaleFemaleRatio,
		)
		if err != nil {
			fmt.Printf("插入专业数据失败 [%d]: %v\n", i, err)
			continue
		}
		successCount++
		
		if (i+1)%100 == 0 {
			fmt.Printf("已处理 %d/%d 个专业\n", i+1, len(majors))
		}
	}

	fmt.Printf("✅ 成功导入 %d/%d 个专业\n", successCount, len(majors))
	return nil
}

// 导入录取数据
func importAdmissions(db *sql.DB, filename string) error {
	fmt.Printf("正在导入录取数据从 %s...\n", filename)

	// 检查文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Printf("⚠️  文件 %s 不存在，跳过录取数据导入\n", filename)
		return nil
	}

	// 读取JSON文件
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	var admissions []AdmissionRecord
	err = json.Unmarshal(data, &admissions)
	if err != nil {
		return err
	}

	// 创建大学名称到ID的映射
	universityMap := make(map[string]int)
	rows, err := db.Query("SELECT id, name FROM universities")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			continue
		}
		universityMap[name] = id
	}

	// 创建专业代码到ID的映射
	majorMap := make(map[string]int)
	rows, err = db.Query("SELECT id, code FROM majors")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var code string
		err = rows.Scan(&id, &code)
		if err != nil {
			continue
		}
		majorMap[code] = id
	}

	// 批量插入录取数据
	stmt, err := db.Prepare(`
		INSERT INTO admissions (year, province, university_id, major_id, batch, min_score, max_score, avg_score, admit_count, subject_type) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) 
		ON CONFLICT (year, province, university_id, major_id, batch, subject_type) DO UPDATE SET 
			min_score = EXCLUDED.min_score,
			max_score = EXCLUDED.max_score,
			avg_score = EXCLUDED.avg_score,
			admit_count = EXCLUDED.admit_count,
			updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	successCount := 0
	skippedCount := 0

	for i, admission := range admissions {
		// 查找大学ID
		universityID, exists := universityMap[admission.University]
		if !exists {
			skippedCount++
			continue
		}

		// 查找专业ID
		majorID, exists := majorMap[admission.MajorCode]
		if !exists {
			skippedCount++
			continue
		}

		_, err = stmt.Exec(
			admission.Year,
			admission.Province,
			universityID,
			majorID,
			admission.Batch,
			admission.MinScore,
			admission.MaxScore,
			admission.AvgScore,
			admission.AdmitNum,
			admission.SubjectType,
		)
		if err != nil {
			fmt.Printf("插入录取数据失败 [%d]: %v\n", i, err)
			skippedCount++
			continue
		}
		successCount++
		
		if (i+1)%1000 == 0 {
			fmt.Printf("已处理 %d/%d 条录取记录\n", i+1, len(admissions))
		}
	}

	fmt.Printf("✅ 成功导入 %d/%d 条录取记录（跳过 %d 条）\n", successCount, len(admissions), skippedCount)
	return nil
}

// 验证数据完整性
func verifyData(db *sql.DB) error {
	fmt.Println("\n🔍 验证数据完整性...")

	// 统计各表数据量
	tables := []string{"universities", "majors", "admissions"}
	for _, table := range tables {
		var count int
		// 使用参数化查询防止SQL注入
		err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count)
		if err != nil {
			return err
		}
		fmt.Printf("- %s 表: %d 条记录\n", table, count)
	}

	// 检查数据关联性
	var orphanAdmissions int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM admissions a 
		WHERE NOT EXISTS (SELECT 1 FROM universities u WHERE u.id = a.university_id)
		   OR NOT EXISTS (SELECT 1 FROM majors m WHERE m.id = a.major_id)
	`).Scan(&orphanAdmissions)
	if err != nil {
		return err
	}

	if orphanAdmissions > 0 {
		fmt.Printf("⚠️  发现 %d 条孤立的录取记录（缺少对应的大学或专业）\n", orphanAdmissions)
	} else {
		fmt.Println("✅ 数据关联性检查通过")
	}

	// 检查数据范围
	var minYear, maxYear int
	err = db.QueryRow("SELECT MIN(year), MAX(year) FROM admissions").Scan(&minYear, &maxYear)
	if err == nil {
		fmt.Printf("- 录取数据年份范围: %d - %d\n", minYear, maxYear)
	}

	return nil
}

// 生成大学代码（简化处理，实际应根据教育部编码规则）
func generateUniversityCode(name, id string) string {
	if id != "" {
		return id
	}
	// 如果没有提供ID，使用名称的哈希值生成一个唯一代码
	// 这里简化处理，实际应使用更复杂的算法
	return fmt.Sprintf("U%06d", len(name))
}

// 使用COPY命令批量导入大学数据
func importUniversitiesWithCopy(db *sql.DB, filename string) error {
	fmt.Printf("正在导入大学数据从 %s (使用COPY命令)...\n", filename)

	// 检查文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Printf("⚠️  文件 %s 不存在，跳过大学数据导入\n", filename)
		return nil
	}

	// 读取JSON文件
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	var universities []UniversityData
	err = json.Unmarshal(data, &universities)
	if err != nil {
		return err
	}

	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 创建临时表
	_, err = tx.Exec(`
		CREATE TEMP TABLE temp_universities (
			name TEXT,
			code TEXT,
			province TEXT,
			city TEXT,
			level TEXT,
			type TEXT,
			founded_year INTEGER,
			website TEXT
		)
	`)
	if err != nil {
		return fmt.Errorf("创建临时表失败: %v", err)
	}

	// 准备COPY数据
	var copyData strings.Builder
	for _, uni := range universities {
		code := generateUniversityCode(uni.Name, uni.ID)
		copyData.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%d\t%s\n",
			uni.Name,
			code,
			uni.Province,
			uni.City,
			uni.Level,
			uni.Type,
			uni.FoundedYear,
			uni.Website))
	}

	// 使用COPY命令批量插入到临时表
	copyStmt := `COPY temp_universities FROM STDIN WITH (FORMAT text, DELIMITER E'\t')`
	_, err = tx.Exec(copyStmt, copyData.String())
	if err != nil {
		return fmt.Errorf("COPY命令执行失败: %v", err)
	}

	// 从临时表插入到主表（使用UPSERT）
	_, err = tx.Exec(`
		INSERT INTO universities (name, code, province, city, level, type, founded_year, website)
		SELECT name, code, province, city, level, type, founded_year, website
		FROM temp_universities
		ON CONFLICT (name) DO UPDATE SET
			code = EXCLUDED.code,
			province = EXCLUDED.province,
			city = EXCLUDED.city,
			level = EXCLUDED.level,
			type = EXCLUDED.type,
			founded_year = EXCLUDED.founded_year,
			website = EXCLUDED.website,
			updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return fmt.Errorf("从临时表插入失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return err
	}

	fmt.Printf("✅ 成功使用COPY命令导入 %d 所大学数据\n", len(universities))
	return nil
}

// 使用COPY命令批量导入专业数据
func importMajorsWithCopy(db *sql.DB, filename string) error {
	fmt.Printf("正在导入专业数据从 %s (使用COPY命令)...\n", filename)

	// 检查文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Printf("⚠️  文件 %s 不存在，跳过专业数据导入\n", filename)
		return nil
	}

	// 读取JSON文件
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	var majors []MajorData
	err = json.Unmarshal(data, &majors)
	if err != nil {
		return err
	}

	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 创建临时表
	_, err = tx.Exec(`
		CREATE TEMP TABLE temp_majors (
			code TEXT,
			name TEXT,
			category TEXT,
			sub_category TEXT,
			education_level TEXT,
			description TEXT,
			graduate_scale TEXT,
			gender_ratio TEXT
		)
	`)
	if err != nil {
		return fmt.Errorf("创建临时表失败: %v", err)
	}

	// 准备COPY数据
	var copyData strings.Builder
	for _, major := range majors {
		copyData.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			major.MajorCode,
			major.MajorName,
			major.DisciplinaryCategory,
			major.DisciplinarySubCategory,
			major.EducationLevel,
			major.MajorIntroduction,
			major.GraduateScale,
			major.MaleFemaleRatio))
	}

	// 使用COPY命令批量插入到临时表
	copyStmt := `COPY temp_majors FROM STDIN WITH (FORMAT text, DELIMITER E'\t')`
	_, err = tx.Exec(copyStmt, copyData.String())
	if err != nil {
		return fmt.Errorf("COPY命令执行失败: %v", err)
	}

	// 从临时表插入到主表（使用UPSERT）
	_, err = tx.Exec(`
		INSERT INTO majors (code, name, category, sub_category, education_level, description, graduate_scale, gender_ratio)
		SELECT code, name, category, sub_category, education_level, description, graduate_scale, gender_ratio
		FROM temp_majors
		ON CONFLICT (code) DO UPDATE SET
			name = EXCLUDED.name,
			category = EXCLUDED.category,
			sub_category = EXCLUDED.sub_category,
			education_level = EXCLUDED.education_level,
			description = EXCLUDED.description,
			graduate_scale = EXCLUDED.graduate_scale,
			gender_ratio = EXCLUDED.gender_ratio,
			updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return fmt.Errorf("从临时表插入失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return err
	}

	fmt.Printf("✅ 成功使用COPY命令导入 %d 个专业\n", len(majors))
	return nil
}

// 使用COPY命令批量导入录取数据
func importAdmissionsWithCopy(db *sql.DB, filename string) error {
	fmt.Printf("正在导入录取数据从 %s (使用COPY命令)...\n", filename)

	// 检查文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Printf("⚠️  文件 %s 不存在，跳过录取数据导入\n", filename)
		return nil
	}

	// 读取JSON文件
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	var admissions []AdmissionRecord
	err = json.Unmarshal(data, &admissions)
	if err != nil {
		return err
	}

	// 创建大学名称到ID的映射
	universityMap := make(map[string]int)
	rows, err := db.Query("SELECT id, name FROM universities")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			continue
		}
		universityMap[name] = id
	}

	// 创建专业代码到ID的映射
	majorMap := make(map[string]int)
	rows, err = db.Query("SELECT id, code FROM majors")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var code string
		err = rows.Scan(&id, &code)
		if err != nil {
			continue
		}
		majorMap[code] = id
	}

	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 创建临时表
	_, err = tx.Exec(`
		CREATE TEMP TABLE temp_admissions (
			year INTEGER,
			province TEXT,
			university_name TEXT,
			major_code TEXT,
			batch TEXT,
			min_score INTEGER,
			max_score INTEGER,
			avg_score INTEGER,
			admit_count INTEGER,
			subject_type TEXT
		)
	`)
	if err != nil {
		return fmt.Errorf("创建临时表失败: %v", err)
	}

	// 准备COPY数据
	var copyData strings.Builder
	for _, admission := range admissions {
		copyData.WriteString(fmt.Sprintf("%d\t%s\t%s\t%s\t%s\t%d\t%d\t%d\t%d\t%s\n",
			admission.Year,
			admission.Province,
			admission.University,
			admission.MajorCode,
			admission.Batch,
			admission.MinScore,
			admission.MaxScore,
			admission.AvgScore,
			admission.AdmitNum,
			admission.SubjectType))
	}

	// 使用COPY命令批量插入到临时表
	copyStmt := `COPY temp_admissions FROM STDIN WITH (FORMAT text, DELIMITER E'\t')`
	_, err = tx.Exec(copyStmt, copyData.String())
	if err != nil {
		return fmt.Errorf("COPY命令执行失败: %v", err)
	}

	// 从临时表插入到主表（使用UPSERT）
	_, err = tx.Exec(`
		INSERT INTO admissions (year, province, university_id, major_id, batch, min_score, max_score, avg_score, admit_count, subject_type)
		SELECT 
			year, 
			province, 
			(SELECT id FROM universities WHERE name = temp_admissions.university_name),
			(SELECT id FROM majors WHERE code = temp_admissions.major_code),
			batch, 
			min_score, 
			max_score, 
			avg_score, 
			admit_count, 
			subject_type
		FROM temp_admissions
		WHERE EXISTS (SELECT 1 FROM universities WHERE name = temp_admissions.university_name)
		  AND EXISTS (SELECT 1 FROM majors WHERE code = temp_admissions.major_code)
		ON CONFLICT (year, province, university_id, major_id, batch, subject_type) DO UPDATE SET
			min_score = EXCLUDED.min_score,
			max_score = EXCLUDED.max_score,
			avg_score = EXCLUDED.avg_score,
			admit_count = EXCLUDED.admit_count,
			updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return fmt.Errorf("从临时表插入失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return err
	}

	fmt.Printf("✅ 成功使用COPY命令导入 %d 条录取记录\n", len(admissions))
	return nil
}

func main() {
	fmt.Println("开始导入真实高考数据到PostgreSQL数据库...")
	fmt.Println("")

	// 连接数据库
	fmt.Println("连接数据库...")
	db, err := connectDB()
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()
	fmt.Println("✅ 数据库连接成功")

	// 创建表结构
	fmt.Println("\n创建数据库表结构...")
	err = createTables(db)
	if err != nil {
		log.Fatalf("创建表失败: %v", err)
	}
	fmt.Println("✅ 表结构创建完成")

	// 导入数据
	fmt.Println("\n开始导入数据...")
	
	// 导入大学数据（使用COPY命令批量导入）
	err = importUniversitiesWithCopy(db, "real_universities_data.json")
	if err != nil {
		log.Fatalf("Failed to import universities: %v", err)
	}

	// 导入专业数据（使用COPY命令批量导入）
	err = importMajorsWithCopy(db, "real_majors_data.json")
	if err != nil {
		log.Fatalf("Failed to import majors: %v", err)
	}

	// 导入录取数据（使用COPY命令批量导入）
	err = importAdmissionsWithCopy(db, "real_admission_data.json")
	if err != nil {
		log.Fatalf("Failed to import admissions: %v", err)
	}

	// 验证数据完整性
	err = verifyData(db)
	if err != nil {
		log.Printf("数据验证失败: %v", err)
	}

	fmt.Println("\n🎉 真实数据导入完成！")
	fmt.Println("\n📝 使用说明：")
	fmt.Println("1. 确保已运行 fetch_real_university_data.go 生成大学数据")
	fmt.Println("2. 确保已运行 fetch_real_major_data.go 生成专业数据")
	fmt.Println("3. 确保已运行 fetch_real_admission_data.go 生成录取数据")
	fmt.Println("4. 数据已导入到 PostgreSQL 数据库中")
	fmt.Println("5. 可以通过 data-service API 查询使用")
}