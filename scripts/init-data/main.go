package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// University 高校模型（对齐 data-service 模型的核心字段）
type University struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	Name        string         `json:"name" gorm:"size:200;not null;index"`
	Code        string         `json:"code" gorm:"size:20;index"`
	EnglishName string         `json:"english_name" gorm:"size:300"`
	Alias       string         `json:"alias" gorm:"size:500"`
	Type        string         `json:"type" gorm:"size:50;index"`
	Level       string         `json:"level" gorm:"size:50;index"`
	Nature      string         `json:"nature" gorm:"size:50;index"`
	Category    string         `json:"category" gorm:"size:100;index"`
	Province    string         `json:"province" gorm:"size:50;index"`
	City        string         `json:"city" gorm:"size:50;index"`
	District    string         `json:"district" gorm:"size:50"`
	Address     string         `json:"address" gorm:"size:500"`
	PostalCode  string         `json:"postal_code" gorm:"size:20"`
	Website     string         `json:"website" gorm:"size:255"`
	Phone       string         `json:"phone" gorm:"size:50"`
	Email       string         `json:"email" gorm:"size:100"`
	Established *time.Time    `json:"established"`
	Description string         `json:"description" gorm:"type:text"`
	Motto       string         `json:"motto" gorm:"size:500"`
	Logo        string         `json:"logo" gorm:"size:255"`
	NationalRank int           `json:"national_rank" gorm:"index"`
	Status      string         `json:"status" gorm:"size:20;default:active;index"`
	IsActive    bool           `json:"is_active" gorm:"default:true;index"`
	IsRecruiting bool          `json:"is_recruiting" gorm:"default:true;index"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (u *University) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// Major 专业模型
type Major struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	UniversityID  uuid.UUID      `json:"university_id" gorm:"type:uuid;not null;index"`
	Name          string         `json:"name" gorm:"size:200;not null;index"`
	Code          string         `json:"code" gorm:"size:20;index"`
	Category      string         `json:"category" gorm:"size:100;index"`
	Discipline    string         `json:"discipline" gorm:"size:100;index"`
	DegreeType    string         `json:"degree_type" gorm:"size:50;index"`
	Duration      int            `json:"duration"`
	Description   string         `json:"description" gorm:"type:text"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

func (m *Major) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// AdmissionData 录取数据模型
type AdmissionData struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	UniversityID uuid.UUID      `json:"university_id" gorm:"type:uuid;not null;index"`
	MajorID      *uuid.UUID     `json:"major_id" gorm:"type:uuid;index"`
	Province     string         `json:"province" gorm:"size:50;index"`
	Year         int            `json:"year" gorm:"index"`
	BatchType    string         `json:"batch_type" gorm:"size:50;index"`
	MinScore     int            `json:"min_score"`
	MaxScore     int            `json:"max_score"`
	AvgScore     int            `json:"avg_score"`
	MinRank      int            `json:"min_rank"`
	MaxRank      int            `json:"max_rank"`
	AvgRank      int            `json:"avg_rank"`
	Enrollment   int            `json:"enrollment"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

func (ad *AdmissionData) BeforeCreate(tx *gorm.DB) error {
	if ad.ID == uuid.Nil {
		ad.ID = uuid.New()
	}
	return nil
}

// TableName 设置表名（与 data-service 模型一致）
func (University) TableName() string {
	return "universities"
}

func (Major) TableName() string {
	return "majors"
}

func (AdmissionData) TableName() string {
	return "admission_data"
}

func main() {
	fmt.Println("🚀 开始初始化高考志愿填报系统数据...")

	// 连接数据库（与 data-service 一致，可由 DATABASE_URL 覆盖）
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=gaokao_data port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	// 自动迁移
	fmt.Println("📋 执行数据库迁移...")
	if err := db.AutoMigrate(&University{}, &Major{}, &AdmissionData{}); err != nil {
		log.Fatal("数据库迁移失败:", err)
	}

	// 初始化高校数据
	fmt.Println("🏫 初始化高校数据...")
	if err := initUniversities(db); err != nil {
		log.Fatal("初始化高校数据失败:", err)
	}

	// 初始化专业数据
	fmt.Println("📚 初始化专业数据...")
	if err := initMajors(db); err != nil {
		log.Fatal("初始化专业数据失败:", err)
	}

	// 初始化录取数据
	fmt.Println("📊 初始化录取数据...")
	if err := initAdmissionData(db); err != nil {
		log.Fatal("初始化录取数据失败:", err)
	}

	fmt.Println("✅ 数据初始化完成!")
}

func initUniversities(db *gorm.DB) error {
	var count int64
	db.Model(&University{}).Count(&count)
	if count > 0 {
		fmt.Printf("  高校数据已存在 (%d 条)，跳过初始化\n", count)
		return nil
	}

	universities := []University{
		{
			Name:        "清华大学",
			Code:        "10003",
			Province:    "北京",
			City:        "北京",
			Type:        "理工类",
			Level:       "985",
			Website:     "https://www.tsinghua.edu.cn",
			Description: "清华大学是中国著名高等学府，坐落于北京西北郊风景秀丽的清华园。",
		},
		{
			Name:        "北京大学",
			Code:        "10001",
			Province:    "北京",
			City:        "北京",
			Type:        "综合类",
			Level:       "985",
			Website:     "https://www.pku.edu.cn",
			Description: "北京大学创办于1898年，初名京师大学堂，是中国第一所国立综合性大学。",
		},
		{
			Name:        "复旦大学",
			Code:        "10246",
			Province:    "上海",
			City:        "上海",
			Type:        "综合类",
			Level:       "985",
			Website:     "https://www.fudan.edu.cn",
			Description: "复旦大学校名取自《尚书大传》之日月光华，旦复旦兮。",
		},
		{
			Name:        "上海交通大学",
			Code:        "10248",
			Province:    "上海",
			City:        "上海",
			Type:        "理工类",
			Level:       "985",
			Website:     "https://www.sjtu.edu.cn",
			Description: "上海交通大学是我国历史最悠久、享誉海内外的高等学府之一。",
		},
		{
			Name:        "浙江大学",
			Code:        "10335",
			Province:    "浙江",
			City:        "杭州",
			Type:        "综合类",
			Level:       "985",
			Website:     "https://www.zju.edu.cn",
			Description: "浙江大学是一所历史悠久、声誉卓著的高等学府。",
		},
	}

	if err := db.Create(&universities).Error; err != nil {
		return err
	}

	fmt.Printf("  成功插入 %d 条高校数据\n", len(universities))
	return nil
}

func initMajors(db *gorm.DB) error {
	var count int64
	db.Model(&Major{}).Count(&count)
	if count > 0 {
		fmt.Printf("  专业数据已存在 (%d 条)，跳过初始化\n", count)
		return nil
	}

	// 将所有高校各插入一个示例专业
	var universities []University
	if err := db.Find(&universities).Error; err != nil {
		return err
	}

	if len(universities) == 0 {
		return fmt.Errorf("缺少高校数据")
	}

	majors := []Major{
		{
			UniversityID: universities[0].ID,
			Name:        "计算机科学与技术",
			Code:        "080901",
			Category:    "工学",
			Discipline:  "计算机类",
			DegreeType:  "工学学士",
			Duration:    4,
			Description: "培养具有良好的科学素养，系统地、较好地掌握计算机科学与技术的基本理论、基本知识和基本技能。",
		},
		{
			UniversityID: universities[0].ID,
			Name:        "软件工程",
			Code:        "080902",
			Category:    "工学",
			Discipline:  "计算机类",
			DegreeType:  "工学学士",
			Duration:    4,
			Description: "培养适应计算机应用学科的发展，特别是软件产业的发展的专门人才。",
		},
	}

	if err := db.Create(&majors).Error; err != nil {
		return err
	}

	fmt.Printf("  成功插入 %d 条专业数据\n", len(majors))
	return nil
}

func initAdmissionData(db *gorm.DB) error {
	var count int64
	db.Model(&AdmissionData{}).Count(&count)
	if count > 0 {
		fmt.Printf("  录取数据已存在 (%d 条)，跳过初始化\n", count)
		return nil
	}

	var universities []University
	var majors []Major
	if err := db.Find(&universities).Error; err != nil {
		return err
	}
	if err := db.Find(&majors).Error; err != nil {
		return err
	}

	if len(universities) == 0 || len(majors) == 0 {
		return fmt.Errorf("缺少高校或专业数据")
	}

	admissions := []AdmissionData{
		{
			UniversityID: universities[0].ID,
			MajorID:      &majors[0].ID,
			Province:     "北京",
			Year:         2023,
			BatchType:    "本科一批",
			MinScore:     650,
			MaxScore:     720,
			AvgScore:     680,
			MinRank:      1000,
			MaxRank:      100,
			AvgRank:      500,
			Enrollment:   120,
		},
	}

	if err := db.Create(&admissions).Error; err != nil {
		return err
	}

	fmt.Printf("  成功插入 %d 条录取数据\n", len(admissions))
	return nil
}
