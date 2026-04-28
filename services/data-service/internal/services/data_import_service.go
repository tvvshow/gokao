package services

import (
	"fmt"
	"github.com/oktetopython/gaokao/services/data-service/internal/database"
	"io"
	"mime/multipart"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// DataImportService 数据导入服务
type DataImportService struct {
	db     *database.DB
	logger *logrus.Logger
}

// NewDataImportService 创建新的数据导入服务
func NewDataImportService(db *database.DB, logger *logrus.Logger) *DataImportService {
	return &DataImportService{
		db:     db,
		logger: logger,
	}
}

// ImportFromFile 从文件导入数据
func (s *DataImportService) ImportFromFile(file multipart.File, fileType string) error {
	// 读取文件内容
	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	// 根据文件类型处理数据
	switch fileType {
	case "universities":
		return s.importUniversities(data)
	case "majors":
		return s.importMajors(data)
	case "admissions":
		return s.importAdmissions(data)
	default:
		return fmt.Errorf("不支持的文件类型: %s", fileType)
	}
}

// importUniversities 导入高校数据
func (s *DataImportService) importUniversities(data []byte) error {
	// 使用数据处理服务处理高校数据
	processingService := NewDataProcessingService(s.db, s.logger)
	return processingService.ProcessUniversityData(data)
}

// importMajors 导入专业数据
func (s *DataImportService) importMajors(data []byte) error {
	// 使用数据处理服务处理专业数据
	processingService := NewDataProcessingService(s.db, s.logger)
	return processingService.ProcessMajorData(data)
}

// importAdmissions 导入录取数据
func (s *DataImportService) importAdmissions(data []byte) error {
	// 使用数据处理服务处理录取数据
	processingService := NewDataProcessingService(s.db, s.logger)
	return processingService.ProcessAdmissionData(data)
}

// ValidateFile 验证上传的文件
func (s *DataImportService) ValidateFile(fileHeader *multipart.FileHeader) error {
	// 检查文件大小（限制为100MB）
	if fileHeader.Size > 100*1024*1024 {
		return fmt.Errorf("文件大小不能超过100MB")
	}

	// 检查文件扩展名
	ext := filepath.Ext(fileHeader.Filename)
	if ext != ".json" {
		return fmt.Errorf("只支持JSON文件")
	}

	return nil
}
