package services

import (
	"fmt"
	"mime/multipart"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/tvvshow/gokao/services/data-service/internal/database"
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

// ImportFromFile 从上传文件流式导入数据。
// multipart.File 实现 io.Reader，直接传给 ProcessXxxDataStream，
// 不再走 io.ReadAll —— 100MB 文件不再顶到 100MB 内存峰值（P-24）。
func (s *DataImportService) ImportFromFile(file multipart.File, fileType string) error {
	processingService := NewDataProcessingService(s.db, s.logger)
	switch fileType {
	case "universities":
		return processingService.ProcessUniversityDataStream(file)
	case "majors":
		return processingService.ProcessMajorDataStream(file)
	case "admissions":
		return processingService.ProcessAdmissionDataStream(file)
	default:
		return fmt.Errorf("不支持的文件类型: %s", fileType)
	}
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
