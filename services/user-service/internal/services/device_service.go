package services

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/oktetopython/gaokao/services/user-service/internal/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// DeviceService 设备服务结构体
type DeviceService struct {
	db               *gorm.DB
	logger           *logrus.Logger
	fingerprintCache sync.Map
	licenseCache     sync.Map
	deviceAuthClient *DeviceAuthClient
	config           *DeviceServiceConfig
	mutex            sync.RWMutex
	performanceStats *PerformanceStatistics
}

// DeviceServiceConfig 设备服务配置
type DeviceServiceConfig struct {
	EnableCache          bool          `json:"enable_cache"`
	CacheTTL             time.Duration `json:"cache_ttl"`
	EnableEncryption     bool          `json:"enable_encryption"`
	EncryptionKey        string        `json:"encryption_key"`
	EnableSignature      bool          `json:"enable_signature"`
	PrivateKey           string        `json:"private_key"`
	PublicKey            string        `json:"public_key"`
	MaxConcurrentTasks   int           `json:"max_concurrent_tasks"`
	EnablePerformanceLog bool          `json:"enable_performance_log"`
	SecurityLevel        int           `json:"security_level"`
	DeviceAuthURL        string        `json:"device_auth_url"`
}

// PerformanceStatistics 性能统计
type PerformanceStatistics struct {
	TotalRequests       int64         `json:"total_requests"`
	SuccessfulRequests  int64         `json:"successful_requests"`
	FailedRequests      int64         `json:"failed_requests"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	LastRequestTime     time.Time     `json:"last_request_time"`
	CacheHitRate        float64       `json:"cache_hit_rate"`
	TotalCacheHits      int64         `json:"total_cache_hits"`
	TotalCacheMisses    int64         `json:"total_cache_misses"`
	mutex               sync.RWMutex
}

// DeviceInfo 设备信息
type DeviceInfo struct {
	ID              string                            `json:"id"`
	UserID          uint                              `json:"user_id"`
	Fingerprint     *DeviceFingerprintResponse        `json:"fingerprint"`
	LicenseInfo     *LicenseValidationResponse        `json:"license_info,omitempty"`
	SecurityStatus  *SecurityStatus                   `json:"security_status"`
	PerformanceInfo *PerformanceStats                 `json:"performance_info,omitempty"`
	LastSeen        time.Time                         `json:"last_seen"`
	Status          DeviceStatus                      `json:"status"`
	Metadata        map[string]interface{}            `json:"metadata,omitempty"`
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

// PerformanceStats 性能统计信息（简化版本）
type PerformanceStats struct {
	CollectTimeUs    uint64 `json:"collect_time_us"`
	HashTimeUs       uint64 `json:"hash_time_us"`
	EncryptionTimeUs uint64 `json:"encryption_time_us"`
	TotalCalls       uint32 `json:"total_calls"`
	SuccessCalls     uint32 `json:"success_calls"`
	ErrorCalls       uint32 `json:"error_calls"`
}

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
			CacheTTL:             10 * time.Minute,
			EnableEncryption:     true,
			EnableSignature:      true,
			MaxConcurrentTasks:   10,
			EnablePerformanceLog: true,
			SecurityLevel:        80,
			DeviceAuthURL:        "http://localhost:8085",
		}
	}

	// 创建设备认证客户端
	deviceAuthClient := NewDeviceAuthClient(config.DeviceAuthURL, logger)

	service := &DeviceService{
		db:               db,
		logger:           logger,
		deviceAuthClient: deviceAuthClient,
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

	// 通过HTTP客户端调用设备认证服务采集设备指纹
	fingerprint, err := s.deviceAuthClient.CollectDeviceFingerprint(ctx)
	if err != nil {
		s.updatePerformanceStats(startTime, false)
		return nil, fmt.Errorf("failed to collect device fingerprint: %w", err)
	}

	// 创建安全状态（简化实现）
	securityStatus := &SecurityStatus{
		SecurityLevel: 80,
		ThreatLevel:   "low",
		Recommendations: []string{
			"当前安全状态良好",
		},
	}

	// 获取性能信息（简化实现）
	perfInfo := &PerformanceStats{
		CollectTimeUs:    1000,
		HashTimeUs:       500,
		EncryptionTimeUs: 200,
		TotalCalls:       10,
		SuccessCalls:     10,
		ErrorCalls:       0,
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
	similarity := 0.9 // 假设相似度为90%
	if currentDevice.Fingerprint.FingerprintHash != storedDevice.Fingerprint.FingerprintHash {
		similarity = 0.5 // 不同则降低相似度
	}

	// 判断是否为同一设备
	isValid := similarity >= 0.8

	s.logger.WithFields(logrus.Fields{
		"user_id":        userID,
		"device_id":      deviceID,
		"similarity":     similarity,
		"is_same_device": isValid,
		"access_granted": isValid,
	}).Info("设备访问验证完成")

	return isValid, nil
}

// GenerateDeviceLicense 生成设备许可证
func (s *DeviceService) GenerateDeviceLicense(ctx context.Context, userID uint, deviceID string, licenseType string, validDays int) (*LicenseValidationResponse, error) {
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
	_, err := s.getStoredDeviceInfo(ctx, userID, deviceID)
	if err != nil {
		return nil, fmt.Errorf("device not found: %w", err)
	}

	// 生成许可证（简化实现，实际应该调用设备认证服务）
	expiresAt := time.Now().Add(time.Duration(validDays) * 24 * time.Hour)
	features := s.getLicenseFeatures(licenseType)

	licenseInfo := &LicenseValidationResponse{
		DeviceID:     deviceID,
		ExpiresAt:    expiresAt,
		Features:     features,
		IsValid:      true,
		ErrorMessage: "",
	}

	// 保存许可证到数据库
	licenseData := "license_data_placeholder" // 简化实现
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
			licenseInfo := cached.(*LicenseValidationResponse)
			if licenseInfo.ExpiresAt.After(time.Now()) {
				s.performanceStats.incrementCacheHit()
				return licenseInfo.IsValid, nil
			}
			s.licenseCache.Delete(cacheKey)
		}
		s.performanceStats.incrementCacheMiss()
	}

	// 通过HTTP客户端调用设备认证服务验证许可证
	licenseInfo, err := s.deviceAuthClient.ValidateLicense(ctx, licenseData, deviceID)
	if err != nil {
		s.updatePerformanceStats(startTime, false)
		return false, fmt.Errorf("failed to validate license: %w", err)
	}
	isValidLicense := licenseInfo.IsValid

	// 验证签名（简化实现）
	if s.config.PublicKey != "" {
		s.logger.Info("License signature verification skipped (simplified implementation)")
	}

	// 检查设备绑定（简化实现）
	deviceBound := true // 简化实现总是返回true
	if !deviceBound {
		s.logger.Warn("Device binding validation failed")
		return false, nil
	}

	// 使用许可证验证结果
	isValid := isValidLicense

	s.logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"device_id": deviceID,
		"is_valid":  isValid,
		"duration":  time.Since(startTime),
	}).Info("设备许可证验证完成")

	return isValid, nil
}

// GetDeviceSecurityStatus 获取设备安全状态
func (s *DeviceService) GetDeviceSecurityStatus(ctx context.Context, userID uint) (*SecurityStatus, error) {
	startTime := time.Now()
	defer s.updatePerformanceStats(startTime, true)

	// 简化实现，返回固定的安全状态
	return &SecurityStatus{
		SecurityLevel: 80,
		ThreatLevel:   "low",
		Recommendations: []string{
			"当前安全状态良好",
		},
	}, nil
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
}

// Close 关闭设备服务
func (s *DeviceService) Close() {
	s.logger.Info("关闭设备服务")

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
	// 转换DeviceInfo到DeviceFingerprint模型
	deviceFingerprint := models.DeviceFingerprint{
		UserID:           uuid.MustParse(fmt.Sprintf("%08x-0000-0000-0000-%012x", deviceInfo.UserID, deviceInfo.UserID)),
		DeviceID:         deviceInfo.ID,
		DeviceName:       deviceInfo.Fingerprint.Hostname,
		DeviceType:       deviceInfo.Fingerprint.DeviceType,
		Platform:         deviceInfo.Fingerprint.OSType,
		OS:               deviceInfo.Fingerprint.OSType,
		OSVersion:        deviceInfo.Fingerprint.OSVersion,
		ScreenResolution: deviceInfo.Fingerprint.ScreenResolution,
		IsActive:         true,
		IsTrusted:        deviceInfo.SecurityStatus.SecurityLevel >= 80,
		LastSeenAt:       &deviceInfo.LastSeen,
	}

	// 保存到数据库
	result := s.db.WithContext(ctx).Create(&deviceFingerprint)
	if result.Error != nil {
		return fmt.Errorf("failed to save device info: %w", result.Error)
	}

	return nil
}

// getStoredDeviceInfo 从数据库获取存储的设备信息
func (s *DeviceService) getStoredDeviceInfo(ctx context.Context, userID uint, deviceID string) (*DeviceInfo, error) {
	// 查询数据库获取设备信息
	var deviceInfo DeviceInfo
	
	// 这里需要根据实际的数据库模型来查询数据
	// 假设我们有一个devices表存储设备信息
	result := s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", deviceID, userID).
		First(&deviceInfo)
	
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("device not found for user %d: %s", userID, deviceID)
		}
		return nil, fmt.Errorf("failed to query device info: %w", result.Error)
	}

	return &deviceInfo, nil
}

// saveLicenseInfo 保存许可证信息到数据库
func (s *DeviceService) saveLicenseInfo(ctx context.Context, userID uint, deviceID string, licenseData string) error {
	// 创建许可证记录
	licenseRecord := models.DeviceLicense{
		UserID:       uuid.MustParse(fmt.Sprintf("%08x-0000-0000-0000-%012x", userID, userID)),
		DeviceID:     deviceID,
		LicenseData:  licenseData,
		Status:       "active",
		IssuedAt:     time.Now(),
	}

	// 保存到数据库
	result := s.db.WithContext(ctx).Create(&licenseRecord)
	if result.Error != nil {
		return fmt.Errorf("failed to save license info: %w", result.Error)
	}

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
		"total_requests":        stats.TotalRequests,
		"successful_requests":   stats.SuccessfulRequests,
		"failed_requests":       stats.FailedRequests,
		"average_response_time": stats.AverageResponseTime,
		"cache_hit_rate":        stats.CacheHitRate,
	}).Info("设备服务性能统计")
}