package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/tvvshow/gokao/services/data-service/internal/database"
	"github.com/tvvshow/gokao/services/data-service/internal/models"

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

// ProcessUniversityData 处理高校数据（兼容入口，内部委托给流式版本）。
// 保留这个签名让 data_handler.ProcessData 的 JSON-in-body 路径继续工作。
func (s *DataProcessingService) ProcessUniversityData(data []byte) error {
	return s.ProcessUniversityDataStream(bytes.NewReader(data))
}

// ProcessUniversityDataStream 以流式解析顶层 JSON 数组，逐条 upsert。
// 内存峰值仅为单条 University + Decoder 内部 buffer（约 KB 级），
// 不再受文件大小约束 —— P-24 修复：原 io.ReadAll 在 100MB 上会顶到 100MB 峰值。
func (s *DataProcessingService) ProcessUniversityDataStream(r io.Reader) error {
	dec := json.NewDecoder(r)
	if err := expectArrayStart(dec); err != nil {
		return fmt.Errorf("解析高校数据失败: %w", err)
	}

	tx := s.db.PostgreSQL.Begin()
	defer func() {
		if rec := recover(); rec != nil {
			tx.Rollback()
			panic(rec)
		}
	}()
	if tx.Error != nil {
		return tx.Error
	}

	var idx int
	for dec.More() {
		var uni models.University
		if err := dec.Decode(&uni); err != nil {
			tx.Rollback()
			return fmt.Errorf("解析第 %d 条高校记录失败: %w", idx+1, err)
		}
		if err := upsertUniversity(tx, &uni); err != nil {
			tx.Rollback()
			return err
		}
		idx++
	}

	return tx.Commit().Error
}

// ProcessMajorData 处理专业数据（兼容入口）。
func (s *DataProcessingService) ProcessMajorData(data []byte) error {
	return s.ProcessMajorDataStream(bytes.NewReader(data))
}

// ProcessMajorDataStream 流式解析专业数据。
func (s *DataProcessingService) ProcessMajorDataStream(r io.Reader) error {
	dec := json.NewDecoder(r)
	if err := expectArrayStart(dec); err != nil {
		return fmt.Errorf("解析专业数据失败: %w", err)
	}

	tx := s.db.PostgreSQL.Begin()
	defer func() {
		if rec := recover(); rec != nil {
			tx.Rollback()
			panic(rec)
		}
	}()
	if tx.Error != nil {
		return tx.Error
	}

	var idx int
	for dec.More() {
		var major models.Major
		if err := dec.Decode(&major); err != nil {
			tx.Rollback()
			return fmt.Errorf("解析第 %d 条专业记录失败: %w", idx+1, err)
		}
		if err := upsertMajor(tx, &major); err != nil {
			tx.Rollback()
			return err
		}
		idx++
	}

	return tx.Commit().Error
}

// ProcessAdmissionData 处理录取数据（兼容入口）。
func (s *DataProcessingService) ProcessAdmissionData(data []byte) error {
	return s.ProcessAdmissionDataStream(bytes.NewReader(data))
}

// ProcessAdmissionDataStream 流式解析录取数据。
func (s *DataProcessingService) ProcessAdmissionDataStream(r io.Reader) error {
	dec := json.NewDecoder(r)
	if err := expectArrayStart(dec); err != nil {
		return fmt.Errorf("解析录取数据失败: %w", err)
	}

	tx := s.db.PostgreSQL.Begin()
	defer func() {
		if rec := recover(); rec != nil {
			tx.Rollback()
			panic(rec)
		}
	}()
	if tx.Error != nil {
		return tx.Error
	}

	var idx int
	for dec.More() {
		var admission models.AdmissionData
		if err := dec.Decode(&admission); err != nil {
			tx.Rollback()
			return fmt.Errorf("解析第 %d 条录取记录失败: %w", idx+1, err)
		}
		if err := upsertAdmission(tx, &admission); err != nil {
			tx.Rollback()
			return err
		}
		idx++
	}

	return tx.Commit().Error
}

// expectArrayStart 读取首个 token 并断言为 '['。
// 显式校验避免后续 dec.More() 在非数组结构（对象、单值）下行为不定。
func expectArrayStart(dec *json.Decoder) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '[' {
		return fmt.Errorf("期望 JSON 数组起始 '['，但读到 %v", tok)
	}
	return nil
}

// upsertUniversity 按 code 唯一键 upsert 单条高校记录。
// 提取出来避免流式与兼容入口重复逻辑。
func upsertUniversity(tx *gorm.DB, uni *models.University) error {
	var existing models.University
	result := tx.Where("code = ?", uni.Code).First(&existing)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			if err := tx.Create(uni).Error; err != nil {
				return fmt.Errorf("创建高校记录失败: %w", err)
			}
			return nil
		}
		return fmt.Errorf("查询高校记录失败: %w", result.Error)
	}
	uni.ID = existing.ID
	if err := tx.Save(uni).Error; err != nil {
		return fmt.Errorf("更新高校记录失败: %w", err)
	}
	return nil
}

// upsertMajor 按 (code, university_id) 复合键 upsert 单条专业记录。
func upsertMajor(tx *gorm.DB, major *models.Major) error {
	var existing models.Major
	result := tx.Where("code = ? AND university_id = ?", major.Code, major.UniversityID).First(&existing)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			if err := tx.Create(major).Error; err != nil {
				return fmt.Errorf("创建专业记录失败: %w", err)
			}
			return nil
		}
		return fmt.Errorf("查询专业记录失败: %w", result.Error)
	}
	major.ID = existing.ID
	if err := tx.Save(major).Error; err != nil {
		return fmt.Errorf("更新专业记录失败: %w", err)
	}
	return nil
}

// upsertAdmission 按 (university_id, major_id, year, province) upsert 单条录取记录。
func upsertAdmission(tx *gorm.DB, admission *models.AdmissionData) error {
	var existing models.AdmissionData
	result := tx.Where("university_id = ? AND major_id = ? AND year = ? AND province = ?",
		admission.UniversityID, admission.MajorID, admission.Year, admission.Province).First(&existing)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			if err := tx.Create(admission).Error; err != nil {
				return fmt.Errorf("创建录取记录失败: %w", err)
			}
			return nil
		}
		return fmt.Errorf("查询录取记录失败: %w", result.Error)
	}
	admission.ID = existing.ID
	if err := tx.Save(admission).Error; err != nil {
		return fmt.Errorf("更新录取记录失败: %w", err)
	}
	return nil
}
