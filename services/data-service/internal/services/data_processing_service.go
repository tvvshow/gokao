package services

import (
	"encoding/json"
	"fmt"
	"github.com/oktetopython/gaokao/services/data-service/internal/database"
	"github.com/oktetopython/gaokao/services/data-service/internal/models"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// DataProcessingService 数据处理服务
type DataProcessingService struct {
	db     *database.DB
	logger *logrus.Logger
}

// NewDataProcessingService 创建新的数据处理服务
func NewDataProcessingService(db *database.DB, logger *logrus.Logger) *DataProcessingService {
	return &DataProcessingService{
		db:     db,
		logger: logger,
	}
}

// ProcessUniversityData 处理高校数据
func (s *DataProcessingService) ProcessUniversityData(data []byte) error {
	var universities []models.University
	if err := json.Unmarshal(data, &universities); err != nil {
		return fmt.Errorf("解析高校数据失败: %w", err)
	}

	// 开始事务
	tx := s.db.PostgreSQL.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return tx.Error
	}

	// 处理每个高校数据
	for _, uni := range universities {
		// 检查是否已存在
		var existing models.University
		result := tx.Where("code = ?", uni.Code).First(&existing)

		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				// 创建新记录
				if err := tx.Create(&uni).Error; err != nil {
					tx.Rollback()
					return fmt.Errorf("创建高校记录失败: %w", err)
				}
			} else {
				tx.Rollback()
				return fmt.Errorf("查询高校记录失败: %w", result.Error)
			}
		} else {
			// 更新现有记录
			uni.ID = existing.ID
			if err := tx.Save(&uni).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("更新高校记录失败: %w", err)
			}
		}
	}

	// 提交事务
	return tx.Commit().Error
}

// ProcessMajorData 处理专业数据
func (s *DataProcessingService) ProcessMajorData(data []byte) error {
	var majors []models.Major
	if err := json.Unmarshal(data, &majors); err != nil {
		return fmt.Errorf("解析专业数据失败: %w", err)
	}

	// 开始事务
	tx := s.db.PostgreSQL.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return tx.Error
	}

	// 处理每个专业数据
	for _, major := range majors {
		// 检查是否已存在
		var existing models.Major
		result := tx.Where("code = ? AND university_id = ?", major.Code, major.UniversityID).First(&existing)

		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				// 创建新记录
				if err := tx.Create(&major).Error; err != nil {
					tx.Rollback()
					return fmt.Errorf("创建专业记录失败: %w", err)
				}
			} else {
				tx.Rollback()
				return fmt.Errorf("查询专业记录失败: %w", result.Error)
			}
		} else {
			// 更新现有记录
			major.ID = existing.ID
			if err := tx.Save(&major).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("更新专业记录失败: %w", err)
			}
		}
	}

	// 提交事务
	return tx.Commit().Error
}

// ProcessAdmissionData 处理录取数据
func (s *DataProcessingService) ProcessAdmissionData(data []byte) error {
	var admissions []models.AdmissionData
	if err := json.Unmarshal(data, &admissions); err != nil {
		return fmt.Errorf("解析录取数据失败: %w", err)
	}

	// 开始事务
	tx := s.db.PostgreSQL.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return tx.Error
	}

	// 处理每个录取数据
	for _, admission := range admissions {
		// 检查是否已存在
		var existing models.AdmissionData
		result := tx.Where("university_id = ? AND major_id = ? AND year = ? AND province = ?",
			admission.UniversityID, admission.MajorID, admission.Year, admission.Province).First(&existing)

		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				// 创建新记录
				if err := tx.Create(&admission).Error; err != nil {
					tx.Rollback()
					return fmt.Errorf("创建录取记录失败: %w", err)
				}
			} else {
				tx.Rollback()
				return fmt.Errorf("查询录取记录失败: %w", result.Error)
			}
		} else {
			// 更新现有记录
			admission.ID = existing.ID
			if err := tx.Save(&admission).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("更新录取记录失败: %w", err)
			}
		}
	}

	// 提交事务
	return tx.Commit().Error
}
