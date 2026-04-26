package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/oktetopython/gaokao/services/data-service/internal/models"
	"gorm.io/gorm"
)

// DataSource 定义统一数据源接口
type DataSource interface {
	FetchUniversities() ([]models.University, error)
	FetchMajors() ([]models.Major, error)
	FetchAdmissionData() ([]models.AdmissionData, error)
	Validate() error
	GetSourceName() string
}

// ===== 统一数据导入器 =====
type UnifiedDataImporter struct {
	db            *gorm.DB
	validator     *DataValidationService
	dataSources   map[string]DataSource
	importHistory []ImportRecord
}

// 导入记录
type ImportRecord struct {
	ID           uuid.UUID  `json:"id"`
	Source       string     `json:"source"`
	DataType     string     `json:"data_type"` // university, major, admission
	RecordCount  int        `json:"record_count"`
	Status       string     `json:"status"` // success, failed, partial
	StartTime    time.Time  `json:"start_time"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

func NewUnifiedDataImporter(db *gorm.DB, validator *DataValidationService) *UnifiedDataImporter {
	return &UnifiedDataImporter{
		db:          db,
		validator:   validator,
		dataSources: make(map[string]DataSource),
	}
}

// 添加数据源
func (i *UnifiedDataImporter) AddDataSource(name string, source DataSource) {
	i.dataSources[name] = source
}

// 导入所有数据
func (i *UnifiedDataImporter) ImportAllData() (map[string]ImportRecord, error) {
	results := make(map[string]ImportRecord)

	// 1. 导入大学数据
	if universitySource, exists := i.dataSources["教育部"]; exists {
		record, err := i.ImportUniversities(universitySource)
		if err != nil {
			results["universities"] = record
			return results, err
		}
		results["universities"] = record
	}

	// 2. 导入专业数据
	if majorSource, exists := i.dataSources["教育部"]; exists {
		record, err := i.ImportMajors(majorSource)
		if err != nil {
			results["majors"] = record
			return results, err
		}
		results["majors"] = record
	}

	// 3. 导入录取数据
	if admissionSource, exists := i.dataSources["考试院"]; exists {
		record, err := i.ImportAdmissionData(admissionSource)
		if err != nil {
			results["admissions"] = record
			return results, err
		}
		results["admissions"] = record
	}

	return results, nil
}

// 导入大学数据
func (i *UnifiedDataImporter) ImportUniversities(source DataSource) (ImportRecord, error) {
	startTime := time.Now()
	record := ImportRecord{
		ID:        uuid.New(),
		Source:    source.GetSourceName(),
		DataType:  "university",
		Status:    "success",
		StartTime: startTime,
	}

	// 1. 获取数据
	universities, err := source.FetchUniversities()
	if err != nil {
		record.Status = "failed"
		record.ErrorMessage = err.Error()
		return record, err
	}

	if len(universities) == 0 {
		record.Status = "success"
		record.RecordCount = 0
		endTime := time.Now()
		record.EndTime = &endTime
		i.importHistory = append(i.importHistory, record)
		return record, nil
	}

	// 2. 验证数据
	errors := i.validator.BatchValidateUniversities(universities)
	if len(errors) > 0 {
		record.Status = "partial"
		record.ErrorMessage = fmt.Sprintf("发现%d个错误", len(errors))
	}

	// 3. 批量导入
	successCount := 0
	for _, university := range universities {
		// 数据清洗
		i.validator.CleanUniversityData(&university)

		// 检查是否已存在
		var existingUniversity models.University
		err = i.db.Where("name = ? AND province = ?", university.Name, university.Province).First(&existingUniversity).Error

		if err == nil {
			// 更新现有记录
			university.ID = existingUniversity.ID
			if updateErr := i.db.Save(&university).Error; updateErr == nil {
				successCount++
			}
		} else {
			// 创建新记录
			if createErr := i.db.Create(&university).Error; createErr == nil {
				successCount++
			}
		}
	}

	// 4. 创建关联的统计数据
	for _, university := range universities {
		stats := models.UniversityStatistics{
			ID:           uuid.New(),
			UniversityID: university.ID,
			UpdatedAt:    time.Now(),
		}
		i.db.Create(&stats)
	}

	// 5. 记录导入结果
	record.RecordCount = successCount
	endTime := time.Now()
	record.EndTime = &endTime
	i.importHistory = append(i.importHistory, record)

	if record.Status == "partial" {
		return record, fmt.Errorf("部分数据导入失败，成功导入%d条", successCount)
	}

	return record, nil
}

// 导入专业数据
func (i *UnifiedDataImporter) ImportMajors(source DataSource) (ImportRecord, error) {
	startTime := time.Now()
	record := ImportRecord{
		ID:        uuid.New(),
		Source:    source.GetSourceName(),
		DataType:  "major",
		Status:    "success",
		StartTime: startTime,
	}

	// 1. 获取数据
	majors, err := source.FetchMajors()
	if err != nil {
		record.Status = "failed"
		record.ErrorMessage = err.Error()
		return record, err
	}

	if len(majors) == 0 {
		record.Status = "success"
		record.RecordCount = 0
		endTime := time.Now()
		record.EndTime = &endTime
		i.importHistory = append(i.importHistory, record)
		return record, nil
	}

	// 2. 验证数据
	errors := i.validator.BatchValidateMajors(majors)
	if len(errors) > 0 {
		record.Status = "partial"
		record.ErrorMessage = fmt.Sprintf("发现%d个错误", len(errors))
	}

	// 3. 批量导入
	successCount := 0
	for _, major := range majors {
		// 数据清洗
		i.validator.CleanMajorData(&major)

		// 检查是否已存在
		var existingMajor models.Major
		err = i.db.Where("code = ?", major.Code).First(&existingMajor).Error

		if err == nil {
			// 更新现有记录
			major.ID = existingMajor.ID
			if updateErr := i.db.Save(&major).Error; updateErr == nil {
				successCount++
			}
		} else {
			// 创建新记录
			if createErr := i.db.Create(&major).Error; createErr == nil {
				successCount++
			}
		}
	}

	// 4. 创建关联的统计数据
	for _, major := range majors {
		stats := models.MajorStatistics{
			ID:        uuid.New(),
			MajorID:   major.ID,
			UpdatedAt: time.Now(),
		}
		i.db.Create(&stats)
	}

	// 5. 记录导入结果
	record.RecordCount = successCount
	endTime := time.Now()
	record.EndTime = &endTime
	i.importHistory = append(i.importHistory, record)

	if record.Status == "partial" {
		return record, fmt.Errorf("部分数据导入失败，成功导入%d条", successCount)
	}

	return record, nil
}

// 导入录取数据
func (i *UnifiedDataImporter) ImportAdmissionData(source DataSource) (ImportRecord, error) {
	startTime := time.Now()
	record := ImportRecord{
		ID:        uuid.New(),
		Source:    source.GetSourceName(),
		DataType:  "admission",
		Status:    "success",
		StartTime: startTime,
	}

	// 1. 获取数据
	admissions, err := source.FetchAdmissionData()
	if err != nil {
		record.Status = "failed"
		record.ErrorMessage = err.Error()
		return record, err
	}

	if len(admissions) == 0 {
		record.Status = "success"
		record.RecordCount = 0
		endTime := time.Now()
		record.EndTime = &endTime
		i.importHistory = append(i.importHistory, record)
		return record, nil
	}

	// 2. 验证数据
	errors := i.validator.BatchValidateAdmissionData(admissions)
	if len(errors) > 0 {
		record.Status = "partial"
		record.ErrorMessage = fmt.Sprintf("发现%d个错误", len(errors))
	}

	// 3. 批量导入
	successCount := 0
	for _, admission := range admissions {
		// 数据清洗
		i.validator.CleanAdmissionData(&admission)

		// 检查大学是否存在
		if admission.UniversityID != uuid.Nil {
			var university models.University
			if err := i.db.First(&university, admission.UniversityID).Error; err != nil {
				continue // 跳过无大学的录取数据
			}
		}

		// 检查专业是否存在（如果非空）
		if admission.MajorID != nil && *admission.MajorID != uuid.Nil {
			var major models.Major
			if err := i.db.First(&major, *admission.MajorID).Error; err != nil {
				admission.MajorID = nil // 清除无效的专业ID
			}
		}

		// 创建或更新记录
		if createErr := i.db.Create(&admission).Error; createErr == nil {
			successCount++
		}
	}

	// 5. 记录导入结果
	record.RecordCount = successCount
	endTime := time.Now()
	record.EndTime = &endTime
	i.importHistory = append(i.importHistory, record)

	if record.Status == "partial" {
		return record, fmt.Errorf("部分数据导入失败，成功导入%d条", successCount)
	}

	return record, nil
}

// 获取导入历史
func (i *UnifiedDataImporter) GetImportHistory() []ImportRecord {
	return i.importHistory
}

// 清除导入历史
func (i *UnifiedDataImporter) ClearImportHistory() {
	i.importHistory = []ImportRecord{}
}

// ===== HTTP数据源实现 =====
type MOEDataSource struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
}

func NewMOEDataSource(baseURL, apiKey string) *MOEDataSource {
	return &MOEDataSource{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey: apiKey,
	}
}

func (m *MOEDataSource) GetSourceName() string {
	return "教育部"
}

func (m *MOEDataSource) Validate() error {
	if m.baseURL == "" {
		return errors.New("MOE数据源URL不能为空")
	}
	return nil
}

func (m *MOEDataSource) FetchUniversities() ([]models.University, error) {
	url := fmt.Sprintf("%s/api/universities", m.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MOE数据源返回状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var universities []models.University
	if err := json.Unmarshal(body, &universities); err != nil {
		return nil, err
	}

	return universities, nil
}

func (m *MOEDataSource) FetchMajors() ([]models.Major, error) {
	url := fmt.Sprintf("%s/api/majors", m.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MOE数据源返回状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var majors []models.Major
	if err := json.Unmarshal(body, &majors); err != nil {
		return nil, err
	}

	return majors, nil
}

func (m *MOEDataSource) FetchAdmissionData() ([]models.AdmissionData, error) {
	url := fmt.Sprintf("%s/api/admissions", m.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MOE数据源返回状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var admissions []models.AdmissionData
	if err := json.Unmarshal(body, &admissions); err != nil {
		return nil, err
	}

	return admissions, nil
}

// ===== 本地文件数据源 =====
type LocalFileDataSource struct {
	filePath string
}

func NewLocalFileDataSource(filePath string) *LocalFileDataSource {
	return &LocalFileDataSource{
		filePath: filePath,
	}
}

func (l *LocalFileDataSource) GetSourceName() string {
	return "本地文件"
}

func (l *LocalFileDataSource) Validate() error {
	// 这里可以实现文件存在性检查
	return nil
}

func (l *LocalFileDataSource) FetchUniversities() ([]models.University, error) {
	// 实现从本地文件读取大学数据
	return []models.University{}, nil
}

func (l *LocalFileDataSource) FetchMajors() ([]models.Major, error) {
	// 实现从本地文件读取专业数据
	return []models.Major{}, nil
}

func (l *LocalFileDataSource) FetchAdmissionData() ([]models.AdmissionData, error) {
	// 实现从本地文件读取录取数据
	return []models.AdmissionData{}, nil
}
