package models

import (
	"time"

	"gorm.io/gorm"
	"github.com/google/uuid"
)

// ===== 增强后的University模型 =====
type University struct {
	ID                      uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()" form:"id"`
	Name                    string         `json:"name" gorm:"type:varchar(255);not null;index" form:"name"`
	EnglishName            string         `json:"english_name" gorm:"type:varchar(255)" form:"english_name"`
	Logo                    string         `json:"logo" gorm:"type:varchar(255)" form:"logo"`
	CoverImage             string         `json:"cover_image" gorm:"type:varchar(255)" form:"cover_image"`
	Province               string         `json:"province" gorm:"type:varchar(50);not null;index" form:"province"`
	City                   string         `json:"city" gorm:"type:varchar(100)" form:"city"`
	Type                   string         `json:"type" gorm:"type:varchar(50);not null;index" form:"type"` // 985/211/普通/专科
	Is985                  bool           `json:"is_985" gorm:"default:false;index" form:"is_985"`
	Is211                  bool           `json:"is_211" gorm:"default:false;index" form:"is_211"`
	IsDoubleFirstClass     bool           `json:"is_double_first_class" gorm:"default:false;index" form:"is_double_first_class"`
	WorldClassCount        int            `json:"world_class_count" gorm:"default:0" form:"world_class_count"`
	DisciplineEvaluation   string         `json:"discipline_evaluation" gorm:"type:text" form:"discipline_evaluation"`
	Website                string         `json:"website" gorm:"type:varchar(255)" form:"website"`
	Address                string         `json:"address" gorm:"type:text" form:"address"`
	ContactPhone           string         `json:"contact_phone" gorm:"type:varchar(50)" form:"contact_phone"`
	ContactEmail           string         `json:"contact_email" gorm:"type:varchar(100)" form:"contact_email"`

	// 原有字段
	Established            *time.Time     `json:"established" gorm:"type:date" form:"established"`
	NationalRank          int            `json:"national_rank" gorm:"default:0" form:"national_rank"`
	ProvinceRank          int            `json:"province_rank" gorm:"default:0" form:"province_rank"`
	QSRank                int            `json:"qs_rank" gorm:"default:0" form:"qs_rank"`
	USNewsRank            int            `json:"us_news_rank" gorm:"default:0" form:"us_news_rank"`
	OverallScore          float64        `json:"overall_score" gorm:"default:0" form:"overall_score"`
	TeachingScore         float64        `json:"teaching_score" gorm:"default:0" form:"teaching_score"`
	ResearchScore          float64        `json:"research_score" gorm:"default:0" form:"research_score"`
	EmploymentScore        float64        `json:"employment_score" gorm:"default:0" form:"employment_score"`

	// 关联
	Majors                []Major        `json:"majors" gorm:"foreignKey:UniversityID"`
	Statistics           UniversityStatistics `json:"statistics" gorm:"foreignKey:UniversityID"`
	Admissions           []AdmissionData `json:"admissions" gorm:"foreignKey:UniversityID"`

	// Gorm字段
	CreatedAt             time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt             time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt             gorm.DeletedAt  `json:"deleted_at" gorm:"index" form:"deleted_at"`
}

// ===== 新增UniversityStatistics模型 =====
type UniversityStatistics struct {
	ID              uuid.UUID    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UniversityID    uuid.UUID    `json:"university_id" gorm:"type:uuid;not null;index" form:"university_id"`

	// 排名数据
	NationalRank    int         `json:"national_rank" gorm:"default:0"`
	ProvinceRank    int         `json:"province_rank" gorm:"default:0"`
	QSRank          int         `json:"qs_rank" gorm:"default:0"`
	USNewsRank      int         `json:"us_news_rank" gorm:"default:0"`

	// 就业数据
	EmploymentRate  float64     `json:"employment_rate" gorm:"default:0"`
	AverageSalary   float64     `json:"average_salary" gorm:"default:0"`
	TopEmployers    string      `json:"top_employers" gorm:"type:text"`

	// 质量评估
	TeachingQuality float64     `json:"teaching_quality" gorm:"default:0"`
	ResearchQuality float64     `json:"research_quality" gorm:"default:0"`

	// 更新时间
	UpdatedAt       time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
}

// ===== 增强后的Major模型 =====
type Major struct {
	ID              uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()" form:"id"`
	Name            string         `json:"name" gorm:"type:varchar(255);not null;index" form:"name"`
	EnglishName     string         `json:"english_name" gorm:"type:varchar(255)" form:"english_name"`
	Code            string         `json:"code" gorm:"type:varchar(50);not null;uniqueIndex" form:"code"`
	Category        string         `json:"category" gorm:"type:varchar(100);not null;index" form:"category"` // 工学、理学、医学等
	Degree          string         `json:"degree" gorm:"type:varchar(50);not null;index" form:"degree"`    // 本科、硕士、博士
	Duration        int            `json:"duration" gorm:"default:4" form:"duration"`                   // 学制年限
	Description     string         `json:"description" gorm:"type:text" form:"description"`
	IsDoubleFirstClass bool        `json:"is_double_first_class" gorm:"default:false" form:"is_double_first_class"`

	// 原有字段
	EmploymentRate  float64        `json:"employment_rate" gorm:"default:0" form:"employment_rate"`
	AverageSalary   float64        `json:"average_salary" gorm:"default:0" form:"average_salary"`
	TopEmployers    string         `json:"top_employers" gorm:"type:text" form:"top_employers"`

	// 关联
	UniversityID    uuid.UUID      `json:"university_id" gorm:"type:uuid;not null;index" form:"university_id"`
	University      University    `json:"university" gorm:"foreignKey:UniversityID"`

	// 统计数据
	Statistics      MajorStatistics `json:"statistics" gorm:"foreignKey:MajorID"`
	Admissions      []AdmissionData `json:"admissions" gorm:"foreignKey:MajorID"`

	// Gorm字段
	CreatedAt       time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index" form:"deleted_at"`
}

// ===== 新增MajorStatistics模型 =====
type MajorStatistics struct {
	ID              uuid.UUID    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	MajorID         uuid.UUID    `json:"major_id" gorm:"type:uuid;not null;index" form:"major_id"`

	// 录取统计
	AverageScore    float64     `json:"average_score" gorm:"default:0"`
	MinScore        float64     `json:"min_score" gorm:"default:0"`
	MaxScore        float64     `json:"max_score" gorm:"default:0"`

	// 就业统计
	EmploymentRate  float64     `json:"employment_rate" gorm:"default:0"`
	AverageSalary   float64     `json:"average_salary" gorm:"default:0"`

	// 热门程度
	Popularity      int         `json:"popularity" gorm:"default:0"`

	// 更新时间
	UpdatedAt       time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
}

// ===== 增强后的AdmissionData模型 =====
type AdmissionData struct {
	ID              uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()" form:"id"`
	UniversityID    uuid.UUID      `json:"university_id" gorm:"type:uuid;not null;index" form:"university_id"`
	MajorID         uuid.UUID      `json:"major_id" gorm:"type:uuid;index" form:"major_id"` // 可为空

	// 科类信息
	ScienceType     string         `json:"science_type" gorm:"type:varchar(20);not null;index" form:"science_type"` // 理科/文科/新高考/综合
	SubjectRequirements []string     `json:"subject_requirements" gorm:"type:text" form:"subject_requirements"`

	// 批次信息
	BatchType       string         `json:"batch_type" gorm:"type:varchar(20);not null;index" form:"batch_type"` // 本科一批/二批/专科
	Year            int            `json:"year" gorm:"type:int;not null;index" form:"year"`

	// 录取数据
	EnrollmentQuota int            `json:"enrollment_quota" gorm:"default:0" form:"enrollment_quota"`
	ActualEnrollment int           `json:"actual_enrollment" gorm:"default:0" form:"actual_enrollment"`

	// 分数线
	CutoffScore     float64        `json:"cutoff_score" gorm:"default:0" form:"cutoff_score"`
	AverageScore    float64        `json:"average_score" gorm:"default:0" form:"average_score"`
	HighestScore    float64        `json:"highest_score" gorm:"default:0" form:"highest_score"`
	LowestScore     float64        `json:"lowest_score" gorm:"default:0" form:"lowest_score"`

	// 排名
	CutoffRank      int            `json:"cutoff_rank" gorm:"default:0" form:"cutoff_rank"`
	AverageRank     int            `json:"average_rank" gorm:"default:0" form:"average_rank"`

	// 关联
	University      University    `json:"university" gorm:"foreignKey:UniversityID"`
	Major           *Major         `json:"major,omitempty" gorm:"foreignKey:MajorID"`

	// Gorm字段
	CreatedAt       time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
}

// ===== 数据验证服务接口 =====
type DataValidationService interface {
	ValidateUniversity(data *University) error
	ValidateMajor(data *Major) error
	ValidateAdmissionData(data *AdmissionData) error
	CheckDataConsistency() error
	EnforceForeignKeyConstraints() error
}

// ===== 数据导入服务接口 =====
type DataSource interface {
	FetchUniversities() ([]University, error)
	FetchMajors() ([]Major, error)
	FetchAdmissionData() ([]AdmissionData, error)
	Validate() error
	GetSourceName() string
}

// ===== 冲稳保策略数据结构 =====
type StrategyData struct {
	ReachUniversities []University `json:"reach_universities"`
	MatchUniversities []University `json:"match_universities"`
	SafetyUniversities []University `json:"safety_universities"`
}

type RecommendationConfig struct {
	UserScore       float64 `json:"user_score"`
	UserRank        int     `json:"user_rank"`
	Province        string  `json:"province"`
	ScienceType     string  `json:"science_type"`
	MajorPreference []string `json:"major_preference"`
	Budget          float64 `json:"budget"`
}