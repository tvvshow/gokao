package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"user-service/internal/cpp"
	"user-service/internal/models"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// DeviceService 设备服务结构体
type DeviceService struct {
	db                *gorm.DB
	logger            *logrus.Logger
	fingerprintCache  sync.Map
	licenseCache      sync.Map
	collector         *cpp.DeviceFingerprintCollector
	cryptoService     *cpp.CryptoService
	licenseService    *cpp.LicenseService
	config            *DeviceServiceConfig
	mutex             sync.RWMutex
	performanceStats  *PerformanceStatistics
}

// DeviceServiceConfig 设备服务配置
type DeviceServiceConfig struct {
	EnableCache           bool          `json:"enable_cache"`
	CacheTTL             time.Duration `json:"cache_ttl"`
	EnableEncryption     bool          `json:"enable_encryption"`
	EncryptionKey        string        `json:"encryption_key"`
	EnableSignature      bool          `json:"enable_signature"`
	PrivateKey           string        `json:"private_key"`
	PublicKey            string        `json:"public_key"`
	MaxConcurrentTasks   int           `json:"max_concurrent_tasks"`
	EnablePerformanceLog bool          `json:"enable_performance_log"`
	SecurityLevel        int           `json:"security_level"`
}

// PerformanceStatistics 性能统计
type PerformanceStatistics struct {
	TotalRequests        int64         `json:"total_requests"`
	SuccessfulRequests   int64         `json:"successful_requests"`
	FailedRequests       int64         `json:"failed_requests"`
	AverageResponseTime  time.Duration `json:"average_response_time"`
	LastRequestTime      time.Time     `json:"last_request_time"`
	CacheHitRate         float64       `json:"cache_hit_rate"`
	TotalCacheHits       int64         `json:"total_cache_hits"`
	TotalCacheMisses     int64         `json:"total_cache_misses"`
	mutex                sync.RWMutex
}

// DeviceInfo 设备信息
type DeviceInfo struct {
	ID                string                    `json:"id"`
	UserID            uint                      `json:"user_id"`
	Fingerprint       *cpp.DeviceFingerprint    `json:"fingerprint"`
	LicenseInfo       *cpp.LicenseInfo          `json:"license_info,omitempty"`
	SecurityStatus    *SecurityStatus           `json:"security_status"`
	PerformanceInfo   *cpp.PerformanceStats     `json:"performance_info,omitempty"`
	LastSeen          time.Time                 `json:"last_seen"`
	Status            DeviceStatus              `json:"status"`
	Metadata          map[string]interface{}    `json:"metadata,omitempty"`
}

// SecurityStatus 安全状态
type SecurityStatus struct {
	SecurityLevel     int      `json:"security_level"`
	RiskFactors       []string `json:"risk_factors"`
	IsDebuggerPresent bool     `json:"is_debugger_present"`
	IsVirtualMachine  bool     `json:"is_virtual_machine"`
	ThreatLevel       string   `json:"threat_level"`
	Recommendations   []string `json:"recommendations"`
}

// DeviceStatus 设备状态枚举
type DeviceStatus string

const (
	DeviceStatusActive      DeviceStatus = "active"
	DeviceStatusInactive    DeviceStatus = "inactive"
	DeviceStatusSuspended   DeviceStatus = "suspended"
	DeviceStatusBlacklisted DeviceStatus = "blacklisted"
)

// NewDeviceService 创建新的设备服务
func NewDeviceService(db *gorm.DB, logger *logrus.Logger, config *DeviceServiceConfig) (*DeviceService, error) {
	if db == nil {
		return nil, errors.New("database connection is required")
	}
	
	if logger == nil {
		return nil, errors.New("logger is required")
	}
	
	if config == nil {
		config = &DeviceServiceConfig{
			EnableCache:          true,
			CacheTTL:            10 * time.Minute,
			EnableEncryption:    true,
			EnableSignature:     true,
			MaxConcurrentTasks:  10,
			EnablePerformanceLog: true,
			SecurityLevel:       80,
		}
	}
	
	// 初始化设备指纹采集器
	collector := cpp.NewDeviceFingerprintCollector()
	if err := collector.Initialize(""); err != nil {
		return nil, fmt.Errorf("failed to initialize fingerprint collector: %w", err)
	}
	
	// 配置采集器
	collectorConfig := &cpp.Configuration{
		CollectSensitiveInfo: config.SecurityLevel >= 70,
		EnableEncryption:     config.EnableEncryption,
		EnableSignature:      config.EnableSignature,
		EncryptionKey:        config.EncryptionKey,
		TimeoutSeconds:       30,
	}
	
	if err := collector.SetConfiguration(collectorConfig); err != nil {
		logger.Warnf("Failed to set collector configuration: %v", err)
	}
	
	// 启用性能监控
	if config.EnablePerformanceLog {
		if err := collector.SetPerformanceMonitoring(true); err != nil {
			logger.Warnf("Failed to enable performance monitoring: %v", err)
		}
	}
	
	service := &DeviceService{
		db:               db,
		logger:           logger,
		collector:        collector,
		cryptoService:    cpp.NewCryptoService(),
		licenseService:   cpp.NewLicenseService(),
		config:           config,
		performanceStats: &PerformanceStatistics{},
	}
	
	// 启动后台任务
	go service.startBackgroundTasks()
	
	return service, nil
}

// CollectDeviceFingerprint 采集设备指纹
func (s *DeviceService) CollectDeviceFingerprint(ctx context.Context, userID uint) (*DeviceInfo, error) {
	startTime := time.Now()
	defer s.updatePerformanceStats(startTime, true)
	
	s.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"action":  "collect_fingerprint",
	}).Info("开始采集设备指纹")
	
	// 检查缓存
	if s.config.EnableCache {
		if cached, found := s.fingerprintCache.Load(userID); found {
			s.performanceStats.incrementCacheHit()
			s.logger.Debug("从缓存返回设备指纹")
			return cached.(*DeviceInfo), nil
		}
		s.performanceStats.incrementCacheMiss()
	}
	
	// 采集设备指纹
	fingerprint, err := s.collector.CollectFingerprint()
	if err != nil {
		s.updatePerformanceStats(startTime, false)
		return nil, fmt.Errorf("failed to collect device fingerprint: %w", err)
	}
	
	// 检查安全状态
	securityStatus, err := s.checkSecurityStatus()
	if err != nil {
		s.logger.Warnf("Failed to check security status: %v", err)
		securityStatus = &SecurityStatus{
			SecurityLevel: 50,
			ThreatLevel:   "unknown",
		}
	}
	
	// 获取性能信息
	perfInfo, err := s.collector.GetPerformanceStats()
	if err != nil {
		s.logger.Warnf("Failed to get performance stats: %v", err)
	}
	
	// 创建设备信息
	deviceInfo := &DeviceInfo{
		ID:              fingerprint.DeviceID,
		UserID:          userID,
		Fingerprint:     fingerprint,
		SecurityStatus:  securityStatus,
		PerformanceInfo: perfInfo,
		LastSeen:        time.Now(),
		Status:          DeviceStatusActive,
		Metadata:        make(map[string]interface{}),
	}
	
	// 保存到数据库
	if err := s.saveDeviceInfo(ctx, deviceInfo); err != nil {
		s.logger.Errorf("Failed to save device info: %v", err)
	}
	
	// 更新缓存
	if s.config.EnableCache {
		s.fingerprintCache.Store(userID, deviceInfo)
		time.AfterFunc(s.config.CacheTTL, func() {
			s.fingerprintCache.Delete(userID)
		})
	}
	
	s.logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"device_id":  deviceInfo.ID,
		"confidence": fingerprint.ConfidenceScore,
		"duration":   time.Since(startTime),
	}).Info("设备指纹采集完成")
	
	return deviceInfo, nil
}

// ValidateDeviceAccess 验证设备访问权限
func (s *DeviceService) ValidateDeviceAccess(ctx context.Context, userID uint, deviceID string) (bool, error) {
	startTime := time.Now()
	defer s.updatePerformanceStats(startTime, true)
	
	s.logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"device_id": deviceID,
		"action":    "validate_access",
	}).Info("验证设备访问权限")
	
	// 获取当前设备指纹
	currentDevice, err := s.CollectDeviceFingerprint(ctx, userID)
	if err != nil {
		return false, err
	}
	
	// 从数据库获取存储的设备信息
	storedDevice, err := s.getStoredDeviceInfo(ctx, userID, deviceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 设备未注册
			return false, nil
		}
		return false, err
	}
	
	// 比较设备指纹
	comparison, err := s.collector.CompareFingerprints(currentDevice.Fingerprint, storedDevice.Fingerprint)
	if err != nil {
		return false, fmt.Errorf("failed to compare fingerprints: %w", err)
	}
	
	// 判断是否为同一设备
	isValid := comparison.IsSameDevice && comparison.SimilarityScore >= 0.8
	
	s.logger.WithFields(logrus.Fields{
		"user_id":         userID,
		"device_id":       deviceID,
		"similarity":      comparison.SimilarityScore,
		"is_same_device":  comparison.IsSameDevice,
		"access_granted":  isValid,
	}).Info("设备访问验证完成")
	
	return isValid, nil
}

// GenerateDeviceLicense 生成设备许可证
func (s *DeviceService) GenerateDeviceLicense(ctx context.Context, userID uint, deviceID string, licenseType string, validDays int) (*cpp.LicenseInfo, error) {
	startTime := time.Now()
	defer s.updatePerformanceStats(startTime, true)
	
	s.logger.WithFields(logrus.Fields{
		"user_id":      userID,
		"device_id":    deviceID,
		"license_type": licenseType,
		"valid_days":   validDays,
		"action":       "generate_license",
	}).Info("生成设备许可证")
	
	if s.config.PrivateKey == "" {
		return nil, errors.New("private key not configured")
	}
	
	// 验证设备是否存在
	deviceInfo, err := s.getStoredDeviceInfo(ctx, userID, deviceID)
	if err != nil {
		return nil, fmt.Errorf("device not found: %w", err)
	}
	
	expiresAt := time.Now().Add(time.Duration(validDays) * 24 * time.Hour)
	features := s.getLicenseFeatures(licenseType)
	
	licenseData, err := s.licenseService.GenerateLicense(deviceID, expiresAt, s.config.PrivateKey, licenseType, features)
	if err != nil {
		s.updatePerformanceStats(startTime, false)
		return nil, fmt.Errorf("failed to generate license: %w", err)
	}
	
	// 解析许可证信息
	licenseInfo, err := s.licenseService.GetLicenseInfo(licenseData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse license info: %w", err)
	}
	
	// 保存许可证到数据库
	if err := s.saveLicenseInfo(ctx, userID, deviceID, licenseData); err != nil {
		s.logger.Errorf("Failed to save license info: %v", err)
	}
	
	// 更新缓存
	if s.config.EnableCache {
		cacheKey := fmt.Sprintf("license_%d_%s", userID, deviceID)
		s.licenseCache.Store(cacheKey, licenseInfo)
	}
	
	s.logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"device_id":  deviceID,
		"expires_at": licenseInfo.ExpiresAt,
		"duration":   time.Since(startTime),
	}).Info("设备许可证生成完成")
	
	return licenseInfo, nil
}

// ValidateDeviceLicense 验证设备许可证
func (s *DeviceService) ValidateDeviceLicense(ctx context.Context, userID uint, deviceID string, licenseData string) (bool, error) {
	startTime := time.Now()
	defer s.updatePerformanceStats(startTime, true)
	
	s.logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"device_id": deviceID,
		"action":    "validate_license",
	}).Info("验证设备许可证")
	
	// 检查缓存
	if s.config.EnableCache {
		cacheKey := fmt.Sprintf("license_%d_%s", userID, deviceID)
		if cached, found := s.licenseCache.Load(cacheKey); found {
			licenseInfo := cached.(*cpp.LicenseInfo)
			if licenseInfo.ExpiresAt.After(time.Now()) {
				s.performanceStats.incrementCacheHit()
				return licenseInfo.IsValid, nil
			}
			s.licenseCache.Delete(cacheKey)
		}
		s.performanceStats.incrementCacheMiss()
	}
	
	// 验证许可证
	licenseInfo, err := s.licenseService.ValidateLicense(licenseData, deviceID)
	if err != nil {
		s.updatePerformanceStats(startTime, false)
		return false, fmt.Errorf("failed to validate license: %w", err)
	}
	
	// 验证签名（如果配置了公钥）
	if s.config.PublicKey != "" {
		signatureValid, err := s.licenseService.VerifyLicenseSignature(licenseData, s.config.PublicKey)
		if err != nil {
			s.logger.Errorf("Failed to verify license signature: %v", err)
		} else if !signatureValid {
			s.logger.Warn("License signature verification failed")
			return false, nil
		}
	}
	
	// 检查设备绑定
	deviceBound, err := s.licenseService.ValidateDeviceBinding(licenseData, deviceID)
	if err != nil {
		s.logger.Errorf("Failed to validate device binding: %v", err)
	} else if !deviceBound {
		s.logger.Warn("Device binding validation failed")
		return false, nil
	}
	
	isValid := licenseInfo.IsValid && !licenseInfo.ExpiresAt.Before(time.Now())
	
	s.logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"device_id":  deviceID,
		"is_valid":   isValid,
		"expires_at": licenseInfo.ExpiresAt,
		"duration":   time.Since(startTime),
	}).Info("设备许可证验证完成")
	
	return isValid, nil
}

// GetDeviceSecurityStatus 获取设备安全状态
func (s *DeviceService) GetDeviceSecurityStatus(ctx context.Context, userID uint) (*SecurityStatus, error) {
	startTime := time.Now()
	defer s.updatePerformanceStats(startTime, true)
	
	return s.checkSecurityStatus()
}

// EncryptDeviceData 加密设备数据
func (s *DeviceService) EncryptDeviceData(ctx context.Context, data []byte) ([]byte, error) {
	if s.config.EncryptionKey == "" {
		return nil, errors.New("encryption key not configured")
	}
	
	return s.cryptoService.EncryptData(data, s.config.EncryptionKey)
}

// DecryptDeviceData 解密设备数据
func (s *DeviceService) DecryptDeviceData(ctx context.Context, encryptedData []byte) ([]byte, error) {
	if s.config.EncryptionKey == "" {
		return nil, errors.New("encryption key not configured")
	}
	
	return s.cryptoService.DecryptData(encryptedData, s.config.EncryptionKey)
}

// GetPerformanceStatistics 获取性能统计信息
func (s *DeviceService) GetPerformanceStatistics() *PerformanceStatistics {
	s.performanceStats.mutex.RLock()
	defer s.performanceStats.mutex.RUnlock()
	
	// 返回副本以避免并发修改
	stats := *s.performanceStats
	return &stats
}

// ResetPerformanceStatistics 重置性能统计信息
func (s *DeviceService) ResetPerformanceStatistics() {
	s.performanceStats.mutex.Lock()
	defer s.performanceStats.mutex.Unlock()
	
	s.performanceStats.TotalRequests = 0
	s.performanceStats.SuccessfulRequests = 0
	s.performanceStats.FailedRequests = 0
	s.performanceStats.AverageResponseTime = 0
	s.performanceStats.TotalCacheHits = 0
	s.performanceStats.TotalCacheMisses = 0
	s.performanceStats.CacheHitRate = 0
	
	// 重置C++层面的性能统计
	if err := s.collector.ResetPerformanceStats(); err != nil {
		s.logger.Errorf("Failed to reset C++ performance stats: %v", err)
	}
}

// Close 关闭设备服务
func (s *DeviceService) Close() {
	s.logger.Info("关闭设备服务")
	
	if s.collector != nil {
		s.collector.Uninitialize()
	}
	
	// 清理缓存
	s.fingerprintCache.Range(func(key, value interface{}) bool {
		s.fingerprintCache.Delete(key)
		return true
	})
	
	s.licenseCache.Range(func(key, value interface{}) bool {
		s.licenseCache.Delete(key)
		return true
	})
}

// 内部辅助方法

// checkSecurityStatus 检查安全状态
func (s *DeviceService) checkSecurityStatus() (*SecurityStatus, error) {
	securityLevel, riskFactors, err := s.collector.CheckSecurity()
	if err != nil {
		return nil, err
	}
	
	isDebugger, _ := s.collector.IsDebuggerPresent()
	isVM, _ := s.collector.IsVirtualMachine()
	
	status := &SecurityStatus{
		SecurityLevel:     securityLevel,
		RiskFactors:       []string{riskFactors},
		IsDebuggerPresent: isDebugger,
		IsVirtualMachine:  isVM,
		ThreatLevel:       s.calculateThreatLevel(securityLevel, isDebugger, isVM),
		Recommendations:   s.generateSecurityRecommendations(securityLevel, isDebugger, isVM),
	}
	
	return status, nil
}

// calculateThreatLevel 计算威胁级别
func (s *DeviceService) calculateThreatLevel(securityLevel int, isDebugger, isVM bool) string {
	if securityLevel >= 80 && !isDebugger && !isVM {
		return "low"
	} else if securityLevel >= 60 && !isDebugger {
		return "medium"
	} else if securityLevel >= 40 {
		return "high"
	} else {
		return "critical"
	}
}

// generateSecurityRecommendations 生成安全建议
func (s *DeviceService) generateSecurityRecommendations(securityLevel int, isDebugger, isVM bool) []string {
	var recommendations []string
	
	if securityLevel < 60 {
		recommendations = append(recommendations, "提升系统安全级别")
	}
	
	if isDebugger {
		recommendations = append(recommendations, "检测到调试器，请确保在安全环境中运行")
	}
	
	if isVM {
		recommendations = append(recommendations, "检测到虚拟机环境，建议在物理机上运行")
	}
	
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "当前安全状态良好")
	}
	
	return recommendations
}

// getLicenseFeatures 根据许可证类型获取功能列表
func (s *DeviceService) getLicenseFeatures(licenseType string) []string {
	switch licenseType {
	case "trial":
		return []string{"basic_features", "trial_mode"}
	case "commercial":
		return []string{"*"}
	case "enterprise":
		return []string{"*", "advanced_analytics", "premium_support"}
	default:
		return []string{"basic_features"}
	}
}

// saveDeviceInfo 保存设备信息到数据库
func (s *DeviceService) saveDeviceInfo(ctx context.Context, deviceInfo *DeviceInfo) error {
	// 序列化设备信息
	data, err := json.Marshal(deviceInfo)
	if err != nil {
		return err
	}
	
	// 加密数据（如果启用）
	if s.config.EnableEncryption && s.config.EncryptionKey != "" {
		encryptedData, err := s.cryptoService.EncryptData(data, s.config.EncryptionKey)
		if err != nil {
			s.logger.Errorf("Failed to encrypt device data: %v", err)
		} else {
			data = encryptedData
		}
	}
	
	// 这里需要根据实际的数据库模型来保存数据
	// 暂时省略具体实现
	
	return nil
}

// getStoredDeviceInfo 从数据库获取存储的设备信息
func (s *DeviceService) getStoredDeviceInfo(ctx context.Context, userID uint, deviceID string) (*DeviceInfo, error) {
	// 这里需要根据实际的数据库模型来查询数据
	// 暂时返回模拟数据
	
	return &DeviceInfo{
		ID:     deviceID,
		UserID: userID,
		Fingerprint: &cpp.DeviceFingerprint{
			DeviceID:        deviceID,
			FingerprintHash: "mock_hash",
		},
		LastSeen: time.Now().Add(-1 * time.Hour),
		Status:   DeviceStatusActive,
	}, nil
}

// saveLicenseInfo 保存许可证信息到数据库
func (s *DeviceService) saveLicenseInfo(ctx context.Context, userID uint, deviceID string, licenseData string) error {
	// 这里需要根据实际的数据库模型来保存许可证数据
	// 暂时省略具体实现
	
	return nil
}

// updatePerformanceStats 更新性能统计
func (s *DeviceService) updatePerformanceStats(startTime time.Time, success bool) {
	s.performanceStats.mutex.Lock()
	defer s.performanceStats.mutex.Unlock()
	
	duration := time.Since(startTime)
	s.performanceStats.TotalRequests++
	s.performanceStats.LastRequestTime = time.Now()
	
	if success {
		s.performanceStats.SuccessfulRequests++
	} else {
		s.performanceStats.FailedRequests++
	}
	
	// 计算平均响应时间
	if s.performanceStats.TotalRequests > 0 {
		totalTime := s.performanceStats.AverageResponseTime * time.Duration(s.performanceStats.TotalRequests-1)
		s.performanceStats.AverageResponseTime = (totalTime + duration) / time.Duration(s.performanceStats.TotalRequests)
	}
	
	// 计算缓存命中率
	totalCacheAccess := s.performanceStats.TotalCacheHits + s.performanceStats.TotalCacheMisses
	if totalCacheAccess > 0 {
		s.performanceStats.CacheHitRate = float64(s.performanceStats.TotalCacheHits) / float64(totalCacheAccess)
	}
}

// incrementCacheHit 增加缓存命中计数
func (stats *PerformanceStatistics) incrementCacheHit() {
	stats.mutex.Lock()
	defer stats.mutex.Unlock()
	stats.TotalCacheHits++
}

// incrementCacheMiss 增加缓存未命中计数
func (stats *PerformanceStatistics) incrementCacheMiss() {
	stats.mutex.Lock()
	defer stats.mutex.Unlock()
	stats.TotalCacheMisses++
}

// startBackgroundTasks 启动后台任务
func (s *DeviceService) startBackgroundTasks() {
	// 定期清理过期缓存
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		
		for range ticker.C {
			s.cleanupExpiredCache()
		}
	}()
	
	// 定期记录性能统计
	if s.config.EnablePerformanceLog {
		go func() {
			ticker := time.NewTicker(10 * time.Minute)
			defer ticker.Stop()
			
			for range ticker.C {
				s.logPerformanceStats()
			}
		}()
	}
}

// cleanupExpiredCache 清理过期缓存
func (s *DeviceService) cleanupExpiredCache() {
	// 实现缓存清理逻辑
	s.logger.Debug("清理过期缓存")
}

// logPerformanceStats 记录性能统计
func (s *DeviceService) logPerformanceStats() {
	stats := s.GetPerformanceStatistics()
	
	s.logger.WithFields(logrus.Fields{
		"total_requests":     stats.TotalRequests,
		"successful_requests": stats.SuccessfulRequests,
		"failed_requests":    stats.FailedRequests,
		"average_response_time": stats.AverageResponseTime,
		"cache_hit_rate":     stats.CacheHitRate,
	}).Info("设备服务性能统计")
}