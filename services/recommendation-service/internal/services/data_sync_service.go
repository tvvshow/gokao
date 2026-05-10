package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tvvshow/gokao/services/recommendation-service/pkg/cppbridge"
)

// DataSyncService 数据同步服务
// 负责从数据服务获取真实录取数据并同步到推荐引擎
type DataSyncService struct {
	dataServiceURL string
	apiKey         string
	cache          map[string][]cppbridge.AdmissionRecord
	mu             sync.RWMutex
	logger         *logrus.Logger
	lastSyncTime   time.Time
	syncInterval   time.Duration
}

// NewDataSyncService 创建数据同步服务
func NewDataSyncService(dataServiceURL, apiKey string, syncInterval time.Duration, logger *logrus.Logger) *DataSyncService {
	return &DataSyncService{
		dataServiceURL: dataServiceURL,
		apiKey:         apiKey,
		cache:          make(map[string][]cppbridge.AdmissionRecord),
		logger:         logger,
		syncInterval:   syncInterval,
	}
}

// Start 启动数据同步服务
func (s *DataSyncService) Start(ctx context.Context) {
	s.logger.Info("启动数据同步服务...")
	
	// 立即执行一次同步
	if err := s.SyncData(); err != nil {
		s.logger.Warnf("初始数据同步失败: %v", err)
	}
	
	// 启动定时同步
	ticker := time.NewTicker(s.syncInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("数据同步服务已停止")
			return
		case <-ticker.C:
			if err := s.SyncData(); err != nil {
				s.logger.Errorf("定时数据同步失败: %v", err)
			}
		}
	}
}

// SyncData 同步数据
func (s *DataSyncService) SyncData() error {
	startTime := time.Now()
	s.logger.Info("开始同步录取数据...")
	
	// 获取大学数据
	universities, err := s.fetchUniversities()
	if err != nil {
		return fmt.Errorf("获取大学数据失败: %v", err)
	}
	
	// 获取专业数据
	majors, err := s.fetchMajors()
	if err != nil {
		return fmt.Errorf("获取专业数据失败: %v", err)
	}
	
	// 获取录取数据
	admissionData, err := s.fetchAdmissionData()
	if err != nil {
		return fmt.Errorf("获取录取数据失败: %v", err)
	}
	
	// 更新缓存
	s.mu.Lock()
	s.cache = admissionData
	s.lastSyncTime = time.Now()
	s.mu.Unlock()
	
	duration := time.Since(startTime)
	s.logger.Infof("数据同步完成，耗时: %v，大学: %d，专业: %d，录取记录: %d", 
		duration, len(universities), len(majors), len(admissionData))
	
	return nil
}

// fetchUniversities 获取大学数据
func (s *DataSyncService) fetchUniversities() ([]cppbridge.UniversityData, error) {
	url := fmt.Sprintf("%s/api/v1/universities?limit=1000", s.dataServiceURL)
	
	resp, err := s.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var response struct {
		Success bool                         `json:"success"`
		Data    []cppbridge.UniversityData `json:"data"`
		Message string                       `json:"message"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("解析大学数据失败: %v", err)
	}
	
	if !response.Success {
		return nil, fmt.Errorf("获取大学数据失败: %s", response.Message)
	}
	
	return response.Data, nil
}

// fetchMajors 获取专业数据
func (s *DataSyncService) fetchMajors() ([]cppbridge.MajorData, error) {
	url := fmt.Sprintf("%s/api/v1/majors?limit=1000", s.dataServiceURL)
	
	resp, err := s.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var response struct {
		Success bool                    `json:"success"`
		Data    []cppbridge.MajorData `json:"data"`
		Message string                  `json:"message"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("解析专业数据失败: %v", err)
	}
	
	if !response.Success {
		return nil, fmt.Errorf("获取专业数据失败: %s", response.Message)
	}
	
	return response.Data, nil
}

// fetchAdmissionData 获取录取数据
func (s *DataSyncService) fetchAdmissionData() (map[string][]cppbridge.AdmissionRecord, error) {
	url := fmt.Sprintf("%s/api/v1/admissions?years=2021,2022,2023&limit=5000", s.dataServiceURL)
	
	resp, err := s.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var response struct {
		Success bool                                `json:"success"`
		Data    []cppbridge.AdmissionRecord       `json:"data"`
		Message string                              `json:"message"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("解析录取数据失败: %v", err)
	}
	
	if !response.Success {
		return nil, fmt.Errorf("获取录取数据失败: %s", response.Message)
	}
	
	// 按大学-专业组合组织数据
	result := make(map[string][]cppbridge.AdmissionRecord)
	for _, record := range response.Data {
		key := fmt.Sprintf("%s_%s", record.UniversityID, record.MajorID)
		result[key] = append(result[key], record)
	}
	
	return result, nil
}

// makeRequest 发送HTTP请求
func (s *DataSyncService) makeRequest(method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	
	// 添加认证头
	if s.apiKey != "" {
		req.Header.Add("Authorization", "Bearer "+s.apiKey)
	}
	req.Header.Add("Content-Type", "application/json")
	
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP请求失败: %s", resp.Status)
	}
	
	return resp, nil
}

// GetAdmissionData 获取录取数据
func (s *DataSyncService) GetAdmissionData() map[string][]cppbridge.AdmissionRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cache
}

// GetLastSyncTime 获取最后同步时间
func (s *DataSyncService) GetLastSyncTime() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastSyncTime
}

// GetCacheStats 获取缓存统计
func (s *DataSyncService) GetCacheStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return map[string]interface{}{
		"cache_size":       len(s.cache),
		"last_sync_time":   s.lastSyncTime.Format(time.RFC3339),
		"data_service_url": s.dataServiceURL,
		"sync_interval":    s.syncInterval.String(),
	}
}

// ClearCache 清除缓存
func (s *DataSyncService) ClearCache() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache = make(map[string][]cppbridge.AdmissionRecord)
	s.logger.Info("数据缓存已清除")
}