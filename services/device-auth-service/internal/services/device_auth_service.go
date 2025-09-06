package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"device-auth-service/internal/cpp"
	"device-auth-service/internal/models"
)

// DeviceAuthService 设备认证服务
type DeviceAuthService struct {
	db                    *gorm.DB
	redis                 *redis.Client
	deviceFingerprintCollector *cpp.DeviceFingerprintCollector
	licenseService        *cpp.LicenseService
}

// NewDeviceAuthService 创建设备认证服务实例
func NewDeviceAuthService(db *gorm.DB, redis *redis.Client) *DeviceAuthService {
	collector := cpp.NewDeviceFingerprintCollector()
	// 初始化采集器
	if err := collector.Initialize(""); err != nil {
		// 处理初始化错误，这里仅记录日志
		// 实际应用中可能需要更复杂的错误处理
	}

	licenseService := cpp.NewLicenseService()

	return &DeviceAuthService{
		db:                    db,
		redis:                 redis,
		deviceFingerprintCollector: collector,
		licenseService:        licenseService,
	}
}

// CollectDeviceFingerprint 采集设备指纹
func (s *DeviceAuthService) CollectDeviceFingerprint() (*cpp.DeviceFingerprint, error) {
	if s.deviceFingerprintCollector == nil {
		return nil, errors.New("device fingerprint collector not initialized")
	}

	return s.deviceFingerprintCollector.CollectFingerprint()
}

// ValidateDeviceFingerprint 验证设备指纹
func (s *DeviceAuthService) ValidateDeviceFingerprint(fingerprint *cpp.DeviceFingerprint, referenceHash string) (bool, error) {
	if s.deviceFingerprintCollector == nil {
		return false, errors.New("device fingerprint collector not initialized")
	}

	return s.deviceFingerprintCollector.ValidateFingerprint(fingerprint, referenceHash)
}

// GenerateDeviceID 生成设备ID
func (s *DeviceAuthService) GenerateDeviceID() (string, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return uuid.String(), nil
}

// StoreDevice 存储设备信息
func (s *DeviceAuthService) StoreDevice(device *models.Device) error {
	// 序列化指纹信息
	fingerprint, err := s.deviceFingerprintCollector.CollectFingerprint()
	if err != nil {
		return err
	}

	serializedFingerprint, err := s.deviceFingerprintCollector.SerializeToJSON(fingerprint)
	if err != nil {
		return err
	}

	device.Fingerprint = serializedFingerprint
	device.LastSeen = time.Now()

	return s.db.Create(device).Error
}

// GetDeviceByDeviceID 根据设备ID获取设备信息
func (s *DeviceAuthService) GetDeviceByDeviceID(deviceID string) (*models.Device, error) {
	var device models.Device
	if err := s.db.Where("device_id = ?", deviceID).First(&device).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("device not found")
		}
		return nil, err
	}
	return &device, nil
}

// ValidateLicense 验证许可证
func (s *DeviceAuthService) ValidateLicense(licenseData string, deviceID string) (*cpp.LicenseInfo, error) {
	if s.licenseService == nil {
		return nil, errors.New("license service not initialized")
	}

	return s.licenseService.ValidateLicense(licenseData, deviceID)
}

// GenerateLicense 生成许可证
func (s *DeviceAuthService) GenerateLicense(deviceID string, expiresAt time.Time, privateKey string, licenseType string, features []string) (string, error) {
	if s.licenseService == nil {
		return "", errors.New("license service not initialized")
	}

	return s.licenseService.GenerateLicense(deviceID, expiresAt, privateKey, licenseType, features)
}

// StoreLicense 存储许可证
func (s *DeviceAuthService) StoreLicense(license *models.License) error {
	return s.db.Create(license).Error
}

// GetLicenseByDeviceID 根据设备ID获取许可证
func (s *DeviceAuthService) GetLicenseByDeviceID(deviceID string) (*models.License, error) {
	var license models.License
	if err := s.db.Where("device_id = ?", deviceID).First(&license).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("license not found")
		}
		return nil, err
	}
	return &license, nil
}