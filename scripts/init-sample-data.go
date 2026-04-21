package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("🚀 开始初始化高考志愿填报系统示例数据...")

	// 连接数据库
	dsn := "host=localhost user=gaokao_user password=gaokao_password dbname=gaokao_user_db port=5432 sslmode=disable"
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		dsn = dbURL
	}

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

	// 创建基础表结构
	fmt.Println("📋 创建基础表结构...")
	if err := createTables(db); err != nil {
		log.Fatal("创建表结构失败:", err)
	}

	// 插入示例数据
	fmt.Println("📊 插入示例数据...")
	if err := insertSampleData(db); err != nil {
		log.Fatal("插入示例数据失败:", err)
	}

	fmt.Println("✅ 数据初始化完成!")
}

func createTables(db *sql.DB) error {
	// 删除现有表（如果存在）
	_, err := db.Exec("DROP TABLE IF EXISTS admission_data CASCADE")
	if err != nil {
		return fmt.Errorf("删除admission_data表失败: %w", err)
	}

	_, err = db.Exec("DROP TABLE IF EXISTS majors CASCADE")
	if err != nil {
		return fmt.Errorf("删除majors表失败: %w", err)
	}

	_, err = db.Exec("DROP TABLE IF EXISTS universities CASCADE")
	if err != nil {
		return fmt.Errorf("删除universities表失败: %w", err)
	}

	// 删除可能存在的约束和索引
	_, _ = db.Exec("DROP INDEX IF EXISTS idx_universities_code")
	_, _ = db.Exec("DROP INDEX IF EXISTS uni_universities_code")

	// 创建高校表
	universityTable := `
	CREATE TABLE universities (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		code VARCHAR(20) UNIQUE NOT NULL,
		name VARCHAR(200) NOT NULL,
		english_name VARCHAR(300),
		alias VARCHAR(500),
		type VARCHAR(50),
		level VARCHAR(50),
		nature VARCHAR(50),
		category VARCHAR(100),
		province VARCHAR(50),
		city VARCHAR(50),
		district VARCHAR(50),
		address VARCHAR(500),
		postal_code VARCHAR(20),
		website VARCHAR(255),
		phone VARCHAR(50),
		email VARCHAR(100),
		established TIMESTAMP,
		description TEXT,
		motto VARCHAR(500),
		logo VARCHAR(255),
		campus_area FLOAT,
		student_count INTEGER,
		teacher_count INTEGER,
		academician_count INTEGER,
		national_rank INTEGER,
		province_rank INTEGER,
		qs_rank INTEGER,
		us_news_rank INTEGER,
		overall_score FLOAT,
		teaching_score FLOAT,
		research_score FLOAT,
		employment_score FLOAT,
		status VARCHAR(20) DEFAULT 'active',
		is_active BOOLEAN DEFAULT true,
		is_recruiting BOOLEAN DEFAULT true,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_universities_code ON universities(code);
	CREATE INDEX IF NOT EXISTS idx_universities_name ON universities(name);
	CREATE INDEX IF NOT EXISTS idx_universities_type ON universities(type);
	CREATE INDEX IF NOT EXISTS idx_universities_level ON universities(level);
	CREATE INDEX IF NOT EXISTS idx_universities_nature ON universities(nature);
	CREATE INDEX IF NOT EXISTS idx_universities_category ON universities(category);
	CREATE INDEX IF NOT EXISTS idx_universities_province ON universities(province);
	CREATE INDEX IF NOT EXISTS idx_universities_city ON universities(city);
	CREATE INDEX IF NOT EXISTS idx_universities_national_rank ON universities(national_rank);
	CREATE INDEX IF NOT EXISTS idx_universities_status ON universities(status);
	CREATE INDEX IF NOT EXISTS idx_universities_is_active ON universities(is_active);
	CREATE INDEX IF NOT EXISTS idx_universities_is_recruiting ON universities(is_recruiting);
	CREATE INDEX IF NOT EXISTS idx_universities_deleted_at ON universities(deleted_at);
	`

	// 创建专业表
	majorTable := `
	CREATE TABLE majors (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		university_id UUID NOT NULL REFERENCES universities(id),
		code VARCHAR(20),
		name VARCHAR(200) NOT NULL,
		english_name VARCHAR(300),
		category VARCHAR(100),
		discipline VARCHAR(100),
		sub_discipline VARCHAR(100),
		degree_type VARCHAR(50),
		duration INTEGER,
		description TEXT,
		core_courses TEXT,
		requirements TEXT,
		career_prospects TEXT,
		employment_rate FLOAT,
		average_salary FLOAT,
		top_employers TEXT,
		is_recruiting BOOLEAN DEFAULT true,
		status VARCHAR(20) DEFAULT 'active',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_majors_university_id ON majors(university_id);
	CREATE INDEX IF NOT EXISTS idx_majors_code ON majors(code);
	CREATE INDEX IF NOT EXISTS idx_majors_name ON majors(name);
	CREATE INDEX IF NOT EXISTS idx_majors_category ON majors(category);
	CREATE INDEX IF NOT EXISTS idx_majors_discipline ON majors(discipline);
	CREATE INDEX IF NOT EXISTS idx_majors_degree_type ON majors(degree_type);
	CREATE INDEX IF NOT EXISTS idx_majors_is_recruiting ON majors(is_recruiting);
	CREATE INDEX IF NOT EXISTS idx_majors_status ON majors(status);
	CREATE INDEX IF NOT EXISTS idx_majors_deleted_at ON majors(deleted_at);
	`

	// 创建录取数据表
	admissionTable := `
	CREATE TABLE admission_data (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		university_id UUID NOT NULL REFERENCES universities(id),
		major_id UUID REFERENCES majors(id),
		year INTEGER NOT NULL,
		province VARCHAR(50),
		batch VARCHAR(50),
		category VARCHAR(50),
		min_score FLOAT,
		max_score FLOAT,
		avg_score FLOAT,
		median_score FLOAT,
		min_rank INTEGER,
		max_rank INTEGER,
		avg_rank INTEGER,
		planned_count INTEGER,
		actual_count INTEGER,
		difficulty VARCHAR(20),
		competition FLOAT,
		admission_rate FLOAT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_admission_data_university_id ON admission_data(university_id);
	CREATE INDEX IF NOT EXISTS idx_admission_data_major_id ON admission_data(major_id);
	CREATE INDEX IF NOT EXISTS idx_admission_data_year ON admission_data(year);
	CREATE INDEX IF NOT EXISTS idx_admission_data_province ON admission_data(province);
	CREATE INDEX IF NOT EXISTS idx_admission_data_batch ON admission_data(batch);
	CREATE INDEX IF NOT EXISTS idx_admission_data_category ON admission_data(category);
	CREATE INDEX IF NOT EXISTS idx_admission_data_min_score ON admission_data(min_score);
	CREATE INDEX IF NOT EXISTS idx_admission_data_avg_score ON admission_data(avg_score);
	CREATE INDEX IF NOT EXISTS idx_admission_data_min_rank ON admission_data(min_rank);
	CREATE INDEX IF NOT EXISTS idx_admission_data_difficulty ON admission_data(difficulty);
	CREATE INDEX IF NOT EXISTS idx_admission_data_deleted_at ON admission_data(deleted_at);
	`

	tables := []string{universityTable, majorTable, admissionTable}

	for _, table := range tables {
		if _, err := db.Exec(table); err != nil {
			return fmt.Errorf("创建表失败: %w", err)
		}
	}

	return nil
}

func insertSampleData(db *sql.DB) error {
	// 检查是否已有数据
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM universities").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		fmt.Printf("  高校数据已存在 (%d 条)，跳过初始化\n", count)
		return nil
	}

	// 插入示例高校数据
	universities := []struct {
		name, code, province, city, utype, level, nature, website, description string
		national_rank int
	}{
		{"清华大学", "10003", "北京", "北京", "理工类", "985", "public", "https://www.tsinghua.edu.cn", "清华大学是中国著名高等学府，坐落于北京西北郊风景秀丽的清华园。", 1},
		{"北京大学", "10001", "北京", "北京", "综合类", "985", "public", "https://www.pku.edu.cn", "北京大学创办于1898年，初名京师大学堂，是中国第一所国立综合性大学。", 2},
		{"复旦大学", "10246", "上海", "上海", "综合类", "985", "public", "https://www.fudan.edu.cn", "复旦大学校名取自《尚书大传》之日月光华，旦复旦兮。", 3},
		{"上海交通大学", "10248", "上海", "上海", "理工类", "985", "public", "https://www.sjtu.edu.cn", "上海交通大学是我国历史最悠久、享誉海内外的高等学府之一。", 4},
		{"浙江大学", "10335", "浙江", "杭州", "综合类", "985", "public", "https://www.zju.edu.cn", "浙江大学是一所历史悠久、声誉卓著的高等学府。", 5},
		{"中国科学技术大学", "10358", "安徽", "合肥", "理工类", "985", "public", "https://www.ustc.edu.cn", "中国科学技术大学是中国科学院所属的一所以前沿科学和高新技术为主的大学。", 6},
		{"南京大学", "10284", "江苏", "南京", "综合类", "985", "public", "https://www.nju.edu.cn", "南京大学是一所历史悠久、声誉卓著的百年名校。", 7},
		{"华中科技大学", "10487", "湖北", "武汉", "理工类", "985", "public", "https://www.hust.edu.cn", "华中科技大学是国家教育部直属重点综合性大学。", 8},
		{"西安交通大学", "10698", "陕西", "西安", "理工类", "985", "public", "https://www.xjtu.edu.cn", "西安交通大学是国家教育部直属重点大学。", 9},
		{"哈尔滨工业大学", "10213", "黑龙江", "哈尔滨", "理工类", "985", "public", "https://www.hit.edu.cn", "哈尔滨工业大学是一所以理工为主的全国重点大学。", 10},
	}

	for _, univ := range universities {
		_, err := db.Exec(`
			INSERT INTO universities (name, code, province, city, type, level, nature, website, description, national_rank)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, univ.name, univ.code, univ.province, univ.city, univ.utype, univ.level, univ.nature, univ.website, univ.description, univ.national_rank)
		if err != nil {
			return fmt.Errorf("插入高校数据失败: %w", err)
		}
	}

	fmt.Printf("  成功插入 %d 条高校数据\n", len(universities))

	// 获取插入的大学ID用于专业关联
	var universityIDs []string
	rows, err := db.Query("SELECT id FROM universities ORDER BY name")
	if err != nil {
		return fmt.Errorf("查询大学ID失败: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("扫描大学ID失败: %w", err)
		}
		universityIDs = append(universityIDs, id)
	}

	// 插入示例专业数据
	majors := []struct {
		name, code, category, discipline, degree string
		duration                                 int
		description                              string
	}{
		{"计算机科学与技术", "080901", "工学", "计算机类", "工学学士", 4, "培养具有良好的科学素养，系统地、较好地掌握计算机科学与技术的基本理论、基本知识和基本技能。"},
		{"软件工程", "080902", "工学", "计算机类", "工学学士", 4, "培养适应计算机应用学科的发展，特别是软件产业的发展的专门人才。"},
		{"人工智能", "080717T", "工学", "计算机类", "工学学士", 4, "培养具备人工智能基础理论、基本知识、基本技能的专门人才。"},
		{"金融学", "020301K", "经济学", "金融学类", "经济学学士", 4, "培养具有金融学理论知识及专业技能的专门人才。"},
		{"临床医学", "100201K", "医学", "临床医学类", "医学学士", 5, "培养具备基础医学、临床医学的基本理论和医疗预防的基本技能。"},
		{"电子信息工程", "080701", "工学", "电子信息类", "工学学士", 4, "培养具备电子技术和信息系统的基础知识。"},
		{"机械工程", "080201", "工学", "机械类", "工学学士", 4, "培养具备机械设计、制造、自动化基础知识与应用能力。"},
		{"土木工程", "081001", "工学", "土木类", "工学学士", 4, "培养掌握土木工程学科的基本理论和基本知识。"},
		{"建筑学", "082801", "工学", "建筑类", "建筑学学士", 5, "培养具备建筑设计、城市设计、室内设计等方面的知识。"},
		{"法学", "030101K", "法学", "法学类", "法学学士", 4, "培养系统掌握法学知识，熟悉我国法律和党的相关政策。"},
	}

	// 为每个大学插入专业（每个大学插入一个专业）
	for i, major := range majors {
		if i >= len(universityIDs) {
			break
		}
		_, err := db.Exec(`
			INSERT INTO majors (university_id, name, code, category, discipline, degree_type, duration, description)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, universityIDs[i], major.name, major.code, major.category, major.discipline, major.degree, major.duration, major.description)
		if err != nil {
			return fmt.Errorf("插入专业数据失败: %w", err)
		}
	}

	fmt.Printf("  成功插入 %d 条专业数据\n", len(majors))

	// 获取专业ID用于录取数据关联
	var majorData []struct {
		id           string
		universityID string
	}
	majorRows, err := db.Query("SELECT id, university_id FROM majors")
	if err != nil {
		return fmt.Errorf("查询专业ID失败: %w", err)
	}
	defer majorRows.Close()

	for majorRows.Next() {
		var major struct {
			id           string
			universityID string
		}
		if err := majorRows.Scan(&major.id, &major.universityID); err != nil {
			return fmt.Errorf("扫描专业ID失败: %w", err)
		}
		majorData = append(majorData, major)
	}

	// 插入示例录取数据
	provinces := []string{"北京", "上海", "广东", "江苏", "浙江"}
	batches := []string{"本科一批", "本科二批"}
	categories := []string{"理科", "文科"}
	difficulties := []string{"容易", "中等", "困难"}

	count = 0
	for year := 2020; year <= 2023; year++ {
		for _, province := range provinces {
			for _, batch := range batches {
				for _, category := range categories {
					for i, major := range majorData {
						if i >= 15 { // 限制录取数据数量
							break
						}
						baseScore := 500
						if batch == "本科一批" {
							baseScore = 580
						}
						if category == "理科" {
							baseScore += 20
						}

						minScore := float64(baseScore + i*5)
						maxScore := minScore + 50
						avgScore := minScore + 25
						minRank := 5000 + i*1000
						maxRank := minRank + 10000
						avgRank := minRank + 5000
						plannedCount := 50 + i*5
						actualCount := plannedCount - 2
						difficulty := difficulties[i%len(difficulties)]
						competition := 0.7 + float64(i%3)*0.1
						admissionRate := 0.2 - float64(i%3)*0.05

						_, err := db.Exec(`
							INSERT INTO admission_data (university_id, major_id, year, province, batch, category, min_score, max_score, avg_score, min_rank, max_rank, avg_rank, planned_count, actual_count, difficulty, competition, admission_rate)
							VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
						`, major.universityID, major.id, year, province, batch, category, minScore, maxScore, avgScore, minRank, maxRank, avgRank, plannedCount, actualCount, difficulty, competition, admissionRate)
						if err != nil {
							return fmt.Errorf("插入录取数据失败: %w", err)
						}
						count++
					}
				}
			}
		}
	}

	fmt.Printf("  成功插入 %d 条录取数据\n", count)
	return nil
}
