package models

import (
	"time"

	"github.com/google/uuid"
)

// UniversityStatistics 存储院校统计信息
type UniversityStatistics struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	UniversityID    uuid.UUID `gorm:"type:uuid;not null;index" json:"university_id"`
	NationalRank    int       `json:"national_rank"`
	ProvinceRank    int       `json:"province_rank"`
	QSRank          int       `json:"qs_rank"`
	USNewsRank      int       `json:"us_news_rank"`
	EmploymentRate  float64   `json:"employment_rate"`
	AverageSalary   float64   `json:"average_salary"`
	TopEmployers    string    `gorm:"type:text" json:"top_employers,omitempty"`
	TeachingQuality float64   `json:"teaching_quality"`
	ResearchQuality float64   `json:"research_quality"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// MajorStatistics 存储专业统计信息
type MajorStatistics struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	MajorID        uuid.UUID `gorm:"type:uuid;not null;index" json:"major_id"`
	AverageScore   float64   `json:"average_score"`
	MinScore       float64   `json:"min_score"`
	MaxScore       float64   `json:"max_score"`
	EmploymentRate float64   `json:"employment_rate"`
	AverageSalary  float64   `json:"average_salary"`
	Popularity     int       `json:"popularity"`
	UpdatedAt      time.Time `json:"updated_at"`
}
