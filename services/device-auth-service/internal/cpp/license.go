//go:build cgo && !windows
// +build cgo,!windows

package cpp

/*
#cgo CPPFLAGS: -I../../../cpp-modules/device-fingerprint/include
#cgo LDFLAGS: -L../../../cpp-modules/device-fingerprint/lib -ldevice_fingerprint -lstdc++

#include "c_interface.h"
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"
	"errors"
	"runtime"
	"time"
	"unsafe"
)

// LicenseInfo 许可证信息结构体
type LicenseInfo struct {
	DeviceID     string    `json:"device_id"`
	LicenseType  string    `json:"license_type"`
	IssuedAt     time.Time `json:"issued_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	MaxDevices   int       `json:"max_devices"`
	Features     []string  `json:"features"`
	Signature    string    `json:"signature"`
	IsValid      bool      `json:"is_valid"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

// LicenseService 许可证服务
type LicenseService struct {
	cryptoService *CryptoService
	initialized   bool
}

// NewLicenseService 创建新的许可证服务
func NewLicenseService() *LicenseService {
	return &LicenseService{
		cryptoService: NewCryptoService(),
		initialized:   true,
	}
}

// ValidateLicense 验证许可证
func (l *LicenseService) ValidateLicense(licenseData string, deviceID string) (*LicenseInfo, error) {
	if !l.initialized {
		return nil, errors.New("license service not initialized")
	}

	if licenseData == "" {
		return nil, errors.New("license data is empty")
	}

	if deviceID == "" {
		return nil, errors.New("device ID is empty")
	}

	cLicenseData := C.CString(licenseData)
	defer C.free(unsafe.Pointer(cLicenseData))

	cDeviceID := C.CString(deviceID)
	defer C.free(unsafe.Pointer(cDeviceID))

	var isValid C.int
	var expiresAt C.longlong

	result := C.DeviceFingerprint_ValidateLicense(cLicenseData, cDeviceID, &isValid, &expiresAt)

	if result != C.C_SUCCESS {
		return &LicenseInfo{
			DeviceID:     deviceID,
			IsValid:      false,
			ErrorMessage: getErrorDescription(result),
		}, nil
	}

	// 解析许可证数据获取详细信息
	licenseInfo, err := l.parseLicenseData(licenseData)
	if err != nil {
		return &LicenseInfo{
			DeviceID:  deviceID,
			IsValid:   isValid != 0,
			ExpiresAt: time.Unix(int64(expiresAt), 0),
		}, nil
	}

	licenseInfo.DeviceID = deviceID
	licenseInfo.IsValid = isValid != 0
	licenseInfo.ExpiresAt = time.Unix(int64(expiresAt), 0)

	return licenseInfo, nil
}

// GenerateLicense 生成设备许可证
func (l *LicenseService) GenerateLicense(deviceID string, expiresAt time.Time, privateKey string, licenseType string, features []string) (string, error) {
	if !l.initialized {
		return "", errors.New("license service not initialized")
	}

	if deviceID == "" {
		return "", errors.New("device ID is empty")
	}

	if privateKey == "" {
		return "", errors.New("private key is empty")
	}

	cDeviceID := C.CString(deviceID)
	defer C.free(unsafe.Pointer(cDeviceID))

	cPrivateKey := C.CString(privateKey)
	defer C.free(unsafe.Pointer(cPrivateKey))

	licenseBuffer := make([]byte, 4096)

	result := C.DeviceFingerprint_GenerateLicense(cDeviceID,
		C.longlong(expiresAt.Unix()),
		cPrivateKey,
		(*C.char)(unsafe.Pointer(&licenseBuffer[0])),
		C.size_t(len(licenseBuffer)))

	if result != C.C_SUCCESS {
		return "", l.convertError(result)
	}

	// 扩展许可证信息
	licenseInfo := &LicenseInfo{
		DeviceID:    deviceID,
		LicenseType: licenseType,
		IssuedAt:    time.Now(),
		ExpiresAt:   expiresAt,
		MaxDevices:  1,
		Features:    features,
		IsValid:     true,
	}

	// 序列化完整许可证信息
	fullLicenseData, err := json.Marshal(licenseInfo)
	if err != nil {
		return "", err
	}

	// 对完整许可证数据进行签名
	signature, err := l.cryptoService.SignData(fullLicenseData, privateKey)
	if err != nil {
		return "", err
	}

	licenseInfo.Signature = l.cryptoService.bytesToHex(signature)

	// 重新序列化包含签名的许可证
	finalLicenseData, err := json.Marshal(licenseInfo)
	if err != nil {
		return "", err
	}

	return string(finalLicenseData), nil
}

// VerifyLicenseSignature 验证许可证签名
func (l *LicenseService) VerifyLicenseSignature(licenseData string, publicKey string) (bool, error) {
	if !l.initialized {
		return false, errors.New("license service not initialized")
	}

	if licenseData == "" {
		return false, errors.New("license data is empty")
	}

	if publicKey == "" {
		return false, errors.New("public key is empty")
	}

	// 解析许可证信息
	var licenseInfo LicenseInfo
	if err := json.Unmarshal([]byte(licenseData), &licenseInfo); err != nil {
		return false, err
	}

	if licenseInfo.Signature == "" {
		return false, errors.New("license signature is empty")
	}

	// 提取签名
	signature, err := l.cryptoService.hexToBytes(licenseInfo.Signature)
	if err != nil {
		return false, err
	}

	// 清除签名字段后重新序列化用于验证
	tempLicenseInfo := licenseInfo
	tempLicenseInfo.Signature = ""

	dataToVerify, err := json.Marshal(tempLicenseInfo)
	if err != nil {
		return false, err
	}

	// 验证签名
	return l.cryptoService.VerifySignature(dataToVerify, signature, publicKey)
}

// CheckLicenseExpiry 检查许可证是否过期
func (l *LicenseService) CheckLicenseExpiry(licenseData string) (bool, time.Duration, error) {
	if licenseData == "" {
		return false, 0, errors.New("license data is empty")
	}

	licenseInfo, err := l.parseLicenseData(licenseData)
	if err != nil {
		return false, 0, err
	}

	now := time.Now()

	if licenseInfo.ExpiresAt.Before(now) {
		return true, 0, nil // 已过期
	}

	remaining := licenseInfo.ExpiresAt.Sub(now)
	return false, remaining, nil
}

// ValidateDeviceBinding 验证设备绑定
func (l *LicenseService) ValidateDeviceBinding(licenseData string, currentDeviceID string) (bool, error) {
	if licenseData == "" {
		return false, errors.New("license data is empty")
	}

	if currentDeviceID == "" {
		return false, errors.New("current device ID is empty")
	}

	licenseInfo, err := l.parseLicenseData(licenseData)
	if err != nil {
		return false, err
	}

	return licenseInfo.DeviceID == currentDeviceID, nil
}

// ValidateFeatureAccess 验证功能访问权限
func (l *LicenseService) ValidateFeatureAccess(licenseData string, feature string) (bool, error) {
	if licenseData == "" {
		return false, errors.New("license data is empty")
	}

	if feature == "" {
		return false, errors.New("feature is empty")
	}

	licenseInfo, err := l.parseLicenseData(licenseData)
	if err != nil {
		return false, err
	}

	// 检查是否有该功能的访问权限
	for _, f := range licenseInfo.Features {
		if f == feature || f == "*" { // "*" 表示所有功能
			return true, nil
		}
	}

	return false, nil
}

// GetLicenseInfo 获取许可证详细信息
func (l *LicenseService) GetLicenseInfo(licenseData string) (*LicenseInfo, error) {
	if licenseData == "" {
		return nil, errors.New("license data is empty")
	}

	return l.parseLicenseData(licenseData)
}

// RenewLicense 续期许可证
func (l *LicenseService) RenewLicense(oldLicenseData string, newExpiresAt time.Time, privateKey string) (string, error) {
	if !l.initialized {
		return "", errors.New("license service not initialized")
	}

	if oldLicenseData == "" {
		return "", errors.New("old license data is empty")
	}

	if privateKey == "" {
		return "", errors.New("private key is empty")
	}

	// 解析旧许可证信息
	oldLicenseInfo, err := l.parseLicenseData(oldLicenseData)
	if err != nil {
		return "", err
	}

	// 生成新许可证
	return l.GenerateLicense(oldLicenseInfo.DeviceID, newExpiresAt, privateKey, oldLicenseInfo.LicenseType, oldLicenseInfo.Features)
}

// RevokeLicense 吊销许可证
func (l *LicenseService) RevokeLicense(licenseData string, privateKey string) error {
	if !l.initialized {
		return errors.New("license service not initialized")
	}

	if licenseData == "" {
		return errors.New("license data is empty")
	}

	if privateKey == "" {
		return errors.New("private key is empty")
	}

	// 解析许可证信息
	licenseInfo, err := l.parseLicenseData(licenseData)
	if err != nil {
		return err
	}

	// 标记为已吊销（设置过期时间为过去）
	licenseInfo.ExpiresAt = time.Now().Add(-24 * time.Hour)
	licenseInfo.IsValid = false

	// 重新签名
	licenseInfo.Signature = ""
	dataToSign, err := json.Marshal(licenseInfo)
	if err != nil {
		return err
	}

	signature, err := l.cryptoService.SignData(dataToSign, privateKey)
	if err != nil {
		return err
	}

	licenseInfo.Signature = l.cryptoService.bytesToHex(signature)

	// 这里应该将吊销的许可证存储到黑名单或数据库
	// 实际实现中需要持久化存储

	return nil
}

// BatchValidateLicenses 批量验证许可证
func (l *LicenseService) BatchValidateLicenses(licenses []string, deviceID string) ([]bool, error) {
	if !l.initialized {
		return nil, errors.New("license service not initialized")
	}

	if len(licenses) == 0 {
		return []bool{}, nil
	}

	if deviceID == "" {
		return nil, errors.New("device ID is empty")
	}

	results := make([]bool, len(licenses))

	for i, license := range licenses {
		licenseInfo, err := l.ValidateLicense(license, deviceID)
		if err != nil {
			results[i] = false
		} else {
			results[i] = licenseInfo.IsValid
		}
	}

	return results, nil
}

// GenerateTrialLicense 生成试用许可证
func (l *LicenseService) GenerateTrialLicense(deviceID string, trialDays int, privateKey string) (string, error) {
	if trialDays <= 0 {
		return "", errors.New("trial days must be positive")
	}

	if trialDays > 365 {
		return "", errors.New("trial days cannot exceed 365")
	}

	expiresAt := time.Now().Add(time.Duration(trialDays) * 24 * time.Hour)
	features := []string{"basic_features", "trial_mode"}

	return l.GenerateLicense(deviceID, expiresAt, privateKey, "trial", features)
}

// GenerateCommercialLicense 生成商业许可证
func (l *LicenseService) GenerateCommercialLicense(deviceID string, validYears int, privateKey string, features []string) (string, error) {
	if validYears <= 0 {
		return "", errors.New("valid years must be positive")
	}

	if validYears > 10 {
		return "", errors.New("valid years cannot exceed 10")
	}

	expiresAt := time.Now().Add(time.Duration(validYears) * 365 * 24 * time.Hour)

	if len(features) == 0 {
		features = []string{"*"} // 默认所有功能
	}

	return l.GenerateLicense(deviceID, expiresAt, privateKey, "commercial", features)
}

// 内部辅助函数

// parseLicenseData 解析许可证数据
func (l *LicenseService) parseLicenseData(licenseData string) (*LicenseInfo, error) {
	var licenseInfo LicenseInfo
	if err := json.Unmarshal([]byte(licenseData), &licenseInfo); err != nil {
		// 如果不是JSON格式，尝试作为简单许可证处理
		return &LicenseInfo{
			LicenseType: "simple",
			IssuedAt:    time.Now(),
			ExpiresAt:   time.Now().Add(30 * 24 * time.Hour), // 默认30天
			MaxDevices:  1,
			Features:    []string{"basic"},
			IsValid:     true,
		}, nil
	}

	return &licenseInfo, nil
}

// convertError 转换C错误码为Go错误
func (l *LicenseService) convertError(cError C.CErrorCode) error {
	return errors.New(getErrorDescription(cError))
}

// validateLicenseFields 验证许可证字段
func (l *LicenseService) validateLicenseFields(licenseInfo *LicenseInfo) error {
	if licenseInfo.DeviceID == "" {
		return errors.New("device ID is required")
	}

	if licenseInfo.LicenseType == "" {
		return errors.New("license type is required")
	}

	if licenseInfo.ExpiresAt.IsZero() {
		return errors.New("expiration date is required")
	}

	if licenseInfo.ExpiresAt.Before(time.Now()) {
		return errors.New("license has expired")
	}

	if licenseInfo.MaxDevices <= 0 {
		return errors.New("max devices must be positive")
	}

	return nil
}

// isLicenseTypeValid 检查许可证类型是否有效
func (l *LicenseService) isLicenseTypeValid(licenseType string) bool {
	validTypes := []string{"trial", "commercial", "enterprise", "developer", "simple"}

	for _, validType := range validTypes {
		if licenseType == validType {
			return true
		}
	}

	return false
}

// calculateLicenseStrength 计算许可证强度（安全级别）
func (l *LicenseService) calculateLicenseStrength(licenseInfo *LicenseInfo) int {
	strength := 0

	// 签名存在加分
	if licenseInfo.Signature != "" {
		strength += 30
	}

	// 设备绑定加分
	if licenseInfo.DeviceID != "" {
		strength += 20
	}

	// 许可证类型加分
	switch licenseInfo.LicenseType {
	case "commercial", "enterprise":
		strength += 30
	case "developer":
		strength += 20
	case "trial":
		strength += 10
	}

	// 功能限制加分
	if len(licenseInfo.Features) > 0 && licenseInfo.Features[0] != "*" {
		strength += 20
	}

	return strength
}