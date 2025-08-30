package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// University 院校模型
type University struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	Code        string         `gorm:"uniqueIndex;not null;size:20" json:"code"`
	Name        string         `gorm:"index;not null;size:200" json:"name"`
	EnglishName string         `gorm:"size:300" json:"english_name,omitempty"`
	Alias       string         `gorm:"size:500" json:"alias,omitempty"`
	
	// 基本信息
	Type          string `gorm:"size:50;index" json:"type"`
	Level         string `gorm:"size:50;index" json:"level"`
	Nature        string `gorm:"size:50;index" json:"nature"`
	Category      string `gorm:"size:100;index" json:"category,omitempty"`
	
	// 地理位置
	Province     string `gorm:"size:50;index" json:"province"`
	City         string `gorm:"size:50;index" json:"city"`
	District     string `gorm:"size:50" json:"district,omitempty"`
	Address      string `gorm:"size:500" json:"address,omitempty"`
	PostalCode   string `gorm:"size:20" json:"postal_code,omitempty"`
	
	// 联系信息
	Website      string `gorm:"size:255" json:"website,omitempty"`
	Phone        string `gorm:"size:50" json:"phone,omitempty"`
	Email        string `gorm:"size:100" json:"email,omitempty"`
	
	// 院校详情
	Established      *time.Time `json:"established,omitempty"`
	Description      string     `gorm:"type:text" json:"description,omitempty"`
	Motto            string     `gorm:"size:500" json:"motto,omitempty"`
	Logo             string     `gorm:"size:255" json:"logo,omitempty"`
	CampusArea       float64    `json:"campus_area,omitempty"`    // 占地面积（亩）
	StudentCount     int        `json:"student_count,omitempty"`  // 在校学生数
	TeacherCount     int        `json:"teacher_count,omitempty"`  // 教师数
	AcademicianCount int        `json:"academician_count,omitempty"` // 院士数
	
	// 排名信息
	NationalRank     int     `gorm:"index" json:"national_rank,omitempty"`
	ProvinceRank     int     `json:"province_rank,omitempty"`
	QSRank           int     `json:"qs_rank,omitempty"`
	USNewsRank       int     `json:"us_news_rank,omitempty"`
	OverallScore     float64 `json:"overall_score,omitempty"`
	TeachingScore    float64 `json:"teaching_score,omitempty"`
	ResearchScore    float64 `json:"research_score,omitempty"`
	EmploymentScore  float64 `json:"employment_score,omitempty"`
	
	// 状态
	Status         string `gorm:"default:active;size:20;index" json:"status"`
	IsActive       bool   `gorm:"default:true;index" json:"is_active"`
	IsRecruiting   bool   `gorm:"default:true;index" json:"is_recruiting"`
	
	// 审计字段
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	Majors         []Major         `gorm:"foreignKey:UniversityID" json:"majors,omitempty"`
	AdmissionData  []AdmissionData `gorm:"foreignKey:UniversityID" json:"admission_data,omitempty"`
}

// Major 专业模型
type Major struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	UniversityID  uuid.UUID      `gorm:"type:uuid;not null;index" json:"university_id"`
	Code          string         `gorm:"size:20;index" json:"code"`
	Name          string         `gorm:"index;not null;size:200" json:"name"`
	EnglishName   string         `gorm:"size:300" json:"english_name,omitempty"`
	
	// 专业分类
	Category      string `gorm:"size:100;index" json:"category"`
	Discipline    string `gorm:"size:100;index" json:"discipline"`
	SubDiscipline string `gorm:"size:100;index" json:"sub_discipline,omitempty"`
	
	// 学位信息
	DegreeType    string `gorm:"size:50;index" json:"degree_type"`
	Duration      int    `json:"duration,omitempty"` // 学制年数
	
	// 专业详情
	Description   string `gorm:"type:text" json:"description,omitempty"`
	CoreCourses   string `gorm:"type:text" json:"core_courses,omitempty"` // JSON格式存储
	Requirements  string `gorm:"type:text" json:"requirements,omitempty"`
	
	// 就业信息
	CareerProspects    string  `gorm:"type:text" json:"career_prospects,omitempty"`
	EmploymentRate     float64 `json:"employment_rate,omitempty"`
	AverageSalary      float64 `json:"average_salary,omitempty"`
	TopEmployers       string  `gorm:"type:text" json:"top_employers,omitempty"` // JSON格式
	
	// 招生信息
	IsRecruiting      bool `gorm:"default:true;index" json:"is_recruiting"`
	RecruitmentQuota  int  `json:"recruitment_quota,omitempty"`
	
	// 热度统计
	ViewCount         uint64  `gorm:"default:0" json:"view_count"`
	SearchCount       uint64  `gorm:"default:0" json:"search_count"`
	PopularityScore   float64 `json:"popularity_score,omitempty"`
	
	// 状态
	Status    string `gorm:"default:active;size:20;index" json:"status"`
	IsActive  bool   `gorm:"default:true;index" json:"is_active"`
	
	// 审计字段
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	University    University      `gorm:"foreignKey:UniversityID" json:"university,omitempty"`
	AdmissionData []AdmissionData `gorm:"foreignKey:MajorID" json:"admission_data,omitempty"`
}

// AdmissionData 录取数据模型
type AdmissionData struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	UniversityID uuid.UUID      `gorm:"type:uuid;not null;index" json:"university_id"`
	MajorID      *uuid.UUID     `gorm:"type:uuid;index" json:"major_id,omitempty"`
	
	// 年份和地区
	Year          int    `gorm:"not null;index" json:"year"`
	Province      string `gorm:"size:50;index" json:"province"`
	
	// 录取批次
	Batch         string `gorm:"size:50;index" json:"batch"`
	Category      string `gorm:"size:50;index" json:"category"`
	
	// 分数线信息
	MinScore      float64 `gorm:"index" json:"min_score,omitempty"`
	MaxScore      float64 `json:"max_score,omitempty"`
	AvgScore      float64 `gorm:"index" json:"avg_score,omitempty"`
	MedianScore   float64 `json:"median_score,omitempty"`
	
	// 排名信息
	MinRank       int `gorm:"index" json:"min_rank,omitempty"`
	MaxRank       int `json:"max_rank,omitempty"`
	AvgRank       int `json:"avg_rank,omitempty"`
	
	// 招生计划
	PlannedCount  int `json:"planned_count,omitempty"`
	ActualCount   int `json:"actual_count,omitempty"`
	
	// 录取概率相关
	Difficulty    string  `gorm:"size:20;index" json:"difficulty"`
	Competition   float64 `json:"competition,omitempty"`    // 竞争激烈程度
	AdmissionRate float64 `json:"admission_rate,omitempty"` // 录取率
	
	// 审计字段
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	University University `gorm:"foreignKey:UniversityID" json:"university,omitempty"`
	Major      *Major     `gorm:"foreignKey:MajorID" json:"major,omitempty"`
}

// SearchIndex 搜索索引模型
type SearchIndex struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	Type        string         `gorm:"size:50;index" json:"type"`
	EntityID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"entity_id"`
	Title       string         `gorm:"index;not null;size:500" json:"title"`
	Content     string         `gorm:"type:text" json:"content"`
	Keywords    string         `gorm:"type:text" json:"keywords"`
	Tags        string         `gorm:"type:text" json:"tags"`
	Province    string         `gorm:"size:50;index" json:"province,omitempty"`
	Category    string         `gorm:"size:100;index" json:"category,omitempty"`
	
	// 搜索权重
	SearchWeight float64 `gorm:"default:1.0" json:"search_weight"`
	
	// 统计信息
	ViewCount   uint64    `gorm:"default:0" json:"view_count"`
	SearchCount uint64    `gorm:"default:0" json:"search_count"`
	LastViewed  time.Time `json:"last_viewed"`
	
	// 审计字段
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// AnalysisResult 分析结果模型
type AnalysisResult struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	UserID      *uuid.UUID     `gorm:"type:uuid;index" json:"user_id,omitempty"`
	RequestID   string         `gorm:"size:100;index" json:"request_id"`
	
	// 分析参数
	Province    string  `gorm:"size:50;index" json:"province"`
	Score       float64 `json:"score"`
	Rank        int     `json:"rank,omitempty"`
	Category    string  `gorm:"size:50" json:"category"`
	Preferences string  `gorm:"type:text" json:"preferences,omitempty"` // JSON格式
	
	// 分析结果
	Results     string `gorm:"type:text" json:"results"` // JSON格式存储推荐结果
	Confidence  float64 `json:"confidence,omitempty"`
	
	// 处理信息
	ProcessTime float64   `json:"process_time,omitempty"` // 处理时间（毫秒）
	Algorithm   string    `gorm:"size:50" json:"algorithm,omitempty"`
	Version     string    `gorm:"size:20" json:"version,omitempty"`
	
	// 审计字段
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// HotSearch 热门搜索模型
type HotSearch struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Keyword     string    `gorm:"size:200;index" json:"keyword"`
	SearchCount uint64    `gorm:"default:0" json:"search_count"`
	Category    string    `gorm:"size:50;index" json:"category,omitempty"`
	Date        time.Time `gorm:"index" json:"date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DataStatistics 数据统计模型
type DataStatistics struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	StatType        string    `gorm:"size:50;index" json:"stat_type"`
	StatKey         string    `gorm:"size:200;index" json:"stat_key"`
	StatValue       float64   `json:"stat_value"`
	StringValue     string    `gorm:"size:500" json:"string_value,omitempty"`
	JsonValue       string    `gorm:"type:text" json:"json_value,omitempty"`
	Date            time.Time `gorm:"index" json:"date"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName 设置表名
func (University) TableName() string {
	return "universities"
}

func (Major) TableName() string {
	return "majors"
}

func (AdmissionData) TableName() string {
	return "admission_data"
}

func (SearchIndex) TableName() string {
	return "search_indices"
}

func (AnalysisResult) TableName() string {
	return "analysis_results"
}

func (HotSearch) TableName() string {
	return "hot_searches"
}

func (DataStatistics) TableName() string {
	return "data_statistics"
}

// BeforeCreate GORM钩子：创建前
func (u *University) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (m *Major) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

func (ad *AdmissionData) BeforeCreate(tx *gorm.DB) error {
	if ad.ID == uuid.Nil {
		ad.ID = uuid.New()
	}
	return nil
}

func (si *SearchIndex) BeforeCreate(tx *gorm.DB) error {
	if si.ID == uuid.Nil {
		si.ID = uuid.New()
	}
	return nil
}

func (ar *AnalysisResult) BeforeCreate(tx *gorm.DB) error {
	if ar.ID == uuid.Nil {
		ar.ID = uuid.New()
	}
	return nil
}