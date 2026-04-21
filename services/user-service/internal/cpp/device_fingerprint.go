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

// ErrorCode Go版本的错误码枚举
type ErrorCode int

const (
	Success                  ErrorCode = 0
	ErrorInitFailed          ErrorCode = 1001
	ErrorInvalidParam        ErrorCode = 1002
	ErrorMemoryAlloc         ErrorCode = 1003
	ErrorHardwareAccess      ErrorCode = 1004
	ErrorSystemInfo          ErrorCode = 1005
	ErrorEncryption          ErrorCode = 1006
	ErrorPermissionDenied    ErrorCode = 1007
	ErrorPlatformUnsupported ErrorCode = 1008
	ErrorUnknown             ErrorCode = 9999
)

// DeviceType 设备类型枚举
type DeviceType string

const (
	DeviceTypeDesktop DeviceType = "desktop"
	DeviceTypeLaptop  DeviceType = "laptop"
	DeviceTypeTablet  DeviceType = "tablet"
	DeviceTypeMobile  DeviceType = "mobile"
	DeviceTypeServer  DeviceType = "server"
	DeviceTypeUnknown DeviceType = "unknown"
)

// DeviceFingerprint Go版本的设备指纹结构体
type DeviceFingerprint struct {
	DeviceID          string     `json:"device_id"`
	DeviceType        DeviceType `json:"device_type"`
	CPUID             string     `json:"cpu_id"`
	CPUModel          string     `json:"cpu_model"`
	CPUCores          uint32     `json:"cpu_cores"`
	TotalMemory       uint64     `json:"total_memory"`
	MotherboardSerial string     `json:"motherboard_serial"`
	OSType            string     `json:"os_type"`
	OSVersion         string     `json:"os_version"`
	Hostname          string     `json:"hostname"`
	Username          string     `json:"username"`
	ScreenResolution  string     `json:"screen_resolution"`
	FingerprintHash   string     `json:"fingerprint_hash"`
	ConfidenceScore   uint32     `json:"confidence_score"`
	ErrorMessage      string     `json:"error_message,omitempty"`
	CollectedAt       time.Time  `json:"collected_at"`
}

// Configuration Go版本的配置结构体
type Configuration struct {
	CollectSensitiveInfo bool   `json:"collect_sensitive_info"`
	EnableEncryption     bool   `json:"enable_encryption"`
	EnableSignature      bool   `json:"enable_signature"`
	EncryptionKey        string `json:"encryption_key"`
	TimeoutSeconds       int    `json:"timeout_seconds"`
}

// ComparisonResult 指纹比较结果
type ComparisonResult struct {
	SimilarityScore float64 `json:"similarity_score"`
	IsSameDevice    bool    `json:"is_same_device"`
	ConfidenceLevel uint32  `json:"confidence_level"`
}

// PerformanceStats 性能统计信息
type PerformanceStats struct {
	CollectTimeUs    uint64 `json:"collect_time_us"`
	HashTimeUs       uint64 `json:"hash_time_us"`
	EncryptionTimeUs uint64 `json:"encryption_time_us"`
	TotalCalls       uint32 `json:"total_calls"`
	SuccessCalls     uint32 `json:"success_calls"`
	ErrorCalls       uint32 `json:"error_calls"`
}

// DeviceFingerprintCollector Go版本的设备指纹采集器
type DeviceFingerprintCollector struct {
	initialized bool
}

// NewDeviceFingerprintCollector 创建新的设备指纹采集器
func NewDeviceFingerprintCollector() *DeviceFingerprintCollector {
	return &DeviceFingerprintCollector{
		initialized: false,
	}
}

// Initialize 初始化采集器
func (d *DeviceFingerprintCollector) Initialize(configPath string) error {
	var cConfigPath *C.char
	if configPath != "" {
		cConfigPath = C.CString(configPath)
		defer C.free(unsafe.Pointer(cConfigPath))
	}

	result := C.DeviceFingerprint_Initialize(cConfigPath)
	if result != C.C_SUCCESS {
		return d.convertError(result)
	}

	d.initialized = true

	// 设置终结器确保资源释放
	runtime.SetFinalizer(d, (*DeviceFingerprintCollector).finalize)

	return nil
}

// Uninitialize 反初始化采集器
func (d *DeviceFingerprintCollector) Uninitialize() {
	if d.initialized {
		C.DeviceFingerprint_Uninitialize()
		d.initialized = false
		runtime.SetFinalizer(d, nil)
	}
}

// finalize 终结器函数
func (d *DeviceFingerprintCollector) finalize() {
	d.Uninitialize()
}

// IsInitialized 检查是否已初始化
func (d *DeviceFingerprintCollector) IsInitialized() bool {
	return d.initialized
}

// CollectFingerprint 采集设备指纹
func (d *DeviceFingerprintCollector) CollectFingerprint() (*DeviceFingerprint, error) {
	if !d.initialized {
		return nil, errors.New("collector not initialized")
	}

	var cFingerprint C.CDeviceFingerprint
	result := C.DeviceFingerprint_Collect(&cFingerprint)

	if result != C.C_SUCCESS {
		errorMsg := C.GoString(&cFingerprint.error_message[0])
		if errorMsg == "" {
			errorMsg = d.getErrorDescription(result)
		}
		return nil, errors.New(errorMsg)
	}

	return d.convertCFingerprint(&cFingerprint), nil
}

// QuickCollectFingerprint 快速采集设备指纹(无需初始化)
func QuickCollectFingerprint() (*DeviceFingerprint, error) {
	var cFingerprint C.CDeviceFingerprint
	result := C.DeviceFingerprint_QuickCollect(&cFingerprint)

	if result != C.C_SUCCESS {
		errorMsg := C.GoString(&cFingerprint.error_message[0])
		if errorMsg == "" {
			errorMsg = getErrorDescription(result)
		}
		return nil, errors.New(errorMsg)
	}

	collector := &DeviceFingerprintCollector{}
	return collector.convertCFingerprint(&cFingerprint), nil
}

// SetConfiguration 设置采集配置
func (d *DeviceFingerprintCollector) SetConfiguration(config *Configuration) error {
	if !d.initialized {
		return errors.New("collector not initialized")
	}

	cConfig := d.convertGoConfig(config)
	result := C.DeviceFingerprint_SetConfiguration(&cConfig)

	if result != C.C_SUCCESS {
		return d.convertError(result)
	}

	return nil
}

// GetConfiguration 获取当前配置
func (d *DeviceFingerprintCollector) GetConfiguration() (*Configuration, error) {
	if !d.initialized {
		return nil, errors.New("collector not initialized")
	}

	var cConfig C.CConfiguration
	result := C.DeviceFingerprint_GetConfiguration(&cConfig)

	if result != C.C_SUCCESS {
		return nil, d.convertError(result)
	}

	return d.convertCConfig(&cConfig), nil
}

// GenerateHash 生成设备指纹哈希
func (d *DeviceFingerprintCollector) GenerateHash(fingerprint *DeviceFingerprint) (string, error) {
	if !d.initialized {
		return "", errors.New("collector not initialized")
	}

	cFingerprint := d.convertGoFingerprint(fingerprint)
	hashBuffer := make([]byte, 128)

	result := C.DeviceFingerprint_GenerateHash(&cFingerprint,
		(*C.char)(unsafe.Pointer(&hashBuffer[0])),
		C.size_t(len(hashBuffer)))

	if result != C.C_SUCCESS {
		return "", d.convertError(result)
	}

	return C.GoString((*C.char)(unsafe.Pointer(&hashBuffer[0]))), nil
}

// CompareFingerprints 比较两个设备指纹
func (d *DeviceFingerprintCollector) CompareFingerprints(fp1, fp2 *DeviceFingerprint) (*ComparisonResult, error) {
	if !d.initialized {
		return nil, errors.New("collector not initialized")
	}

	cFp1 := d.convertGoFingerprint(fp1)
	cFp2 := d.convertGoFingerprint(fp2)

	var similarityScore C.double
	var isSameDevice C.int

	result := C.DeviceFingerprint_Compare(&cFp1, &cFp2, &similarityScore, &isSameDevice)

	if result != C.C_SUCCESS {
		return nil, d.convertError(result)
	}

	return &ComparisonResult{
		SimilarityScore: float64(similarityScore),
		IsSameDevice:    isSameDevice != 0,
		ConfidenceLevel: uint32(float64(similarityScore) * 100),
	}, nil
}

// ValidateFingerprint 验证设备指纹
func (d *DeviceFingerprintCollector) ValidateFingerprint(fingerprint *DeviceFingerprint, referenceHash string) (bool, error) {
	if !d.initialized {
		return false, errors.New("collector not initialized")
	}

	cFingerprint := d.convertGoFingerprint(fingerprint)
	cReferenceHash := C.CString(referenceHash)
	defer C.free(unsafe.Pointer(cReferenceHash))

	var isValid C.int
	result := C.DeviceFingerprint_Validate(&cFingerprint, cReferenceHash, &isValid)

	if result != C.C_SUCCESS {
		return false, d.convertError(result)
	}

	return isValid != 0, nil
}

// SerializeToJSON 序列化设备指纹为JSON
func (d *DeviceFingerprintCollector) SerializeToJSON(fingerprint *DeviceFingerprint) (string, error) {
	if !d.initialized {
		return "", errors.New("collector not initialized")
	}

	cFingerprint := d.convertGoFingerprint(fingerprint)
	jsonBuffer := make([]byte, 4096)

	result := C.DeviceFingerprint_SerializeToJson(&cFingerprint,
		(*C.char)(unsafe.Pointer(&jsonBuffer[0])),
		C.size_t(len(jsonBuffer)))

	if result != C.C_SUCCESS {
		return "", d.convertError(result)
	}

	return C.GoString((*C.char)(unsafe.Pointer(&jsonBuffer[0]))), nil
}

// DeserializeFromJSON 从JSON反序列化设备指纹
func (d *DeviceFingerprintCollector) DeserializeFromJSON(jsonData string) (*DeviceFingerprint, error) {
	if !d.initialized {
		return nil, errors.New("collector not initialized")
	}

	cJsonData := C.CString(jsonData)
	defer C.free(unsafe.Pointer(cJsonData))

	var cFingerprint C.CDeviceFingerprint
	result := C.DeviceFingerprint_DeserializeFromJson(cJsonData, &cFingerprint)

	if result != C.C_SUCCESS {
		return nil, d.convertError(result)
	}

	return d.convertCFingerprint(&cFingerprint), nil
}

// IsDebuggerPresent 检查是否存在调试器
func (d *DeviceFingerprintCollector) IsDebuggerPresent() (bool, error) {
	if !d.initialized {
		return false, errors.New("collector not initialized")
	}

	var isPresent C.int
	result := C.DeviceFingerprint_IsDebuggerPresent(&isPresent)

	if result != C.C_SUCCESS {
		return false, d.convertError(result)
	}

	return isPresent != 0, nil
}

// IsVirtualMachine 检查是否在虚拟机中运行
func (d *DeviceFingerprintCollector) IsVirtualMachine() (bool, error) {
	if !d.initialized {
		return false, errors.New("collector not initialized")
	}

	var isVM C.int
	result := C.DeviceFingerprint_IsVirtualMachine(&isVM)

	if result != C.C_SUCCESS {
		return false, d.convertError(result)
	}

	return isVM != 0, nil
}

// CheckSecurity 检查运行环境安全性
func (d *DeviceFingerprintCollector) CheckSecurity() (int, string, error) {
	if !d.initialized {
		return 0, "", errors.New("collector not initialized")
	}

	var securityLevel C.int
	riskFactorsBuffer := make([]byte, 1024)

	result := C.DeviceFingerprint_CheckSecurity(&securityLevel,
		(*C.char)(unsafe.Pointer(&riskFactorsBuffer[0])),
		C.size_t(len(riskFactorsBuffer)))

	if result != C.C_SUCCESS {
		return 0, "", d.convertError(result)
	}

	riskFactors := C.GoString((*C.char)(unsafe.Pointer(&riskFactorsBuffer[0])))
	return int(securityLevel), riskFactors, nil
}

// GetPerformanceStats 获取性能统计信息
func (d *DeviceFingerprintCollector) GetPerformanceStats() (*PerformanceStats, error) {
	if !d.initialized {
		return nil, errors.New("collector not initialized")
	}

	var cStats C.CPerformanceStats
	result := C.DeviceFingerprint_GetPerformanceStats(&cStats)

	if result != C.C_SUCCESS {
		return nil, d.convertError(result)
	}

	return &PerformanceStats{
		CollectTimeUs:    uint64(cStats.collect_time_us),
		HashTimeUs:       uint64(cStats.hash_time_us),
		EncryptionTimeUs: uint64(cStats.encryption_time_us),
		TotalCalls:       uint32(cStats.total_calls),
		SuccessCalls:     uint32(cStats.success_calls),
		ErrorCalls:       uint32(cStats.error_calls),
	}, nil
}

// ResetPerformanceStats 重置性能统计信息
func (d *DeviceFingerprintCollector) ResetPerformanceStats() error {
	if !d.initialized {
		return errors.New("collector not initialized")
	}

	result := C.DeviceFingerprint_ResetPerformanceStats()
	if result != C.C_SUCCESS {
		return d.convertError(result)
	}

	return nil
}

// SetPerformanceMonitoring 启用/禁用性能监控
func (d *DeviceFingerprintCollector) SetPerformanceMonitoring(enable bool) error {
	if !d.initialized {
		return errors.New("collector not initialized")
	}

	var cEnable C.int
	if enable {
		cEnable = 1
	} else {
		cEnable = 0
	}

	result := C.DeviceFingerprint_SetPerformanceMonitoring(cEnable)
	if result != C.C_SUCCESS {
		return d.convertError(result)
	}

	return nil
}

// GetVersion 获取库版本
func GetVersion() (string, error) {
	versionBuffer := make([]byte, 64)
	result := C.DeviceFingerprint_GetVersion(
		(*C.char)(unsafe.Pointer(&versionBuffer[0])),
		C.size_t(len(versionBuffer)))

	if result != C.C_SUCCESS {
		return "", errors.New("failed to get version")
	}

	return C.GoString((*C.char)(unsafe.Pointer(&versionBuffer[0]))), nil
}

// GetSupportedPlatforms 获取支持的平台列表
func GetSupportedPlatforms() ([]string, error) {
	platformsBuffer := make([]byte, 256)
	result := C.DeviceFingerprint_GetSupportedPlatforms(
		(*C.char)(unsafe.Pointer(&platformsBuffer[0])),
		C.size_t(len(platformsBuffer)))

	if result != C.C_SUCCESS {
		return nil, errors.New("failed to get supported platforms")
	}

	platformsStr := C.GoString((*C.char)(unsafe.Pointer(&platformsBuffer[0])))
	if platformsStr == "" {
		return []string{}, nil
	}

	var platforms []string
	if err := json.Unmarshal([]byte(platformsStr), &platforms); err != nil {
		// 如果不是JSON格式，按逗号分割
		return []string{platformsStr}, nil
	}

	return platforms, nil
}

// 内部辅助函数

// convertError 转换C错误码为Go错误
func (d *DeviceFingerprintCollector) convertError(cError C.CErrorCode) error {
	return errors.New(d.getErrorDescription(cError))
}

// getErrorDescription 获取错误描述
func (d *DeviceFingerprintCollector) getErrorDescription(cError C.CErrorCode) string {
	return getErrorDescription(cError)
}

// getErrorDescription 全局函数获取错误描述
func getErrorDescription(cError C.CErrorCode) string {
	errorBuffer := make([]byte, 256)
	result := C.DeviceFingerprint_GetErrorDescription(cError,
		(*C.char)(unsafe.Pointer(&errorBuffer[0])),
		C.size_t(len(errorBuffer)))

	if result != C.C_SUCCESS {
		return "Unknown error"
	}

	return C.GoString((*C.char)(unsafe.Pointer(&errorBuffer[0])))
}

// convertCFingerprint 转换C结构体为Go结构体
func (d *DeviceFingerprintCollector) convertCFingerprint(cFp *C.CDeviceFingerprint) *DeviceFingerprint {
	return &DeviceFingerprint{
		DeviceID:          C.GoString(&cFp.device_id[0]),
		DeviceType:        DeviceType(C.GoString(&cFp.device_type[0])),
		CPUID:             C.GoString(&cFp.cpu_id[0]),
		CPUModel:          C.GoString(&cFp.cpu_model[0]),
		CPUCores:          uint32(cFp.cpu_cores),
		TotalMemory:       uint64(cFp.total_memory),
		MotherboardSerial: C.GoString(&cFp.motherboard_serial[0]),
		OSType:            C.GoString(&cFp.os_type[0]),
		OSVersion:         C.GoString(&cFp.os_version[0]),
		Hostname:          C.GoString(&cFp.hostname[0]),
		Username:          C.GoString(&cFp.username[0]),
		ScreenResolution:  C.GoString(&cFp.screen_resolution[0]),
		FingerprintHash:   C.GoString(&cFp.fingerprint_hash[0]),
		ConfidenceScore:   uint32(cFp.confidence_score),
		ErrorMessage:      C.GoString(&cFp.error_message[0]),
		CollectedAt:       time.Now(),
	}
}

// convertGoFingerprint 转换Go结构体为C结构体
func (d *DeviceFingerprintCollector) convertGoFingerprint(goFp *DeviceFingerprint) C.CDeviceFingerprint {
	var cFp C.CDeviceFingerprint

	// 安全复制字符串到C结构体
	d.safeStrCopy(cFp.device_id[:], goFp.DeviceID)
	d.safeStrCopy(cFp.device_type[:], string(goFp.DeviceType))
	d.safeStrCopy(cFp.cpu_id[:], goFp.CPUID)
	d.safeStrCopy(cFp.cpu_model[:], goFp.CPUModel)
	d.safeStrCopy(cFp.motherboard_serial[:], goFp.MotherboardSerial)
	d.safeStrCopy(cFp.os_type[:], goFp.OSType)
	d.safeStrCopy(cFp.os_version[:], goFp.OSVersion)
	d.safeStrCopy(cFp.hostname[:], goFp.Hostname)
	d.safeStrCopy(cFp.username[:], goFp.Username)
	d.safeStrCopy(cFp.screen_resolution[:], goFp.ScreenResolution)
	d.safeStrCopy(cFp.fingerprint_hash[:], goFp.FingerprintHash)

	cFp.cpu_cores = C.uint(goFp.CPUCores)
	cFp.total_memory = C.ulonglong(goFp.TotalMemory)
	cFp.confidence_score = C.uint(goFp.ConfidenceScore)

	return cFp
}

// convertGoConfig 转换Go配置为C配置
func (d *DeviceFingerprintCollector) convertGoConfig(goConfig *Configuration) C.CConfiguration {
	var cConfig C.CConfiguration

	if goConfig.CollectSensitiveInfo {
		cConfig.collect_sensitive_info = 1
	}
	if goConfig.EnableEncryption {
		cConfig.enable_encryption = 1
	}
	if goConfig.EnableSignature {
		cConfig.enable_signature = 1
	}

	d.safeStrCopy(cConfig.encryption_key[:], goConfig.EncryptionKey)
	cConfig.timeout_seconds = C.int(goConfig.TimeoutSeconds)

	return cConfig
}

// convertCConfig 转换C配置为Go配置
func (d *DeviceFingerprintCollector) convertCConfig(cConfig *C.CConfiguration) *Configuration {
	return &Configuration{
		CollectSensitiveInfo: cConfig.collect_sensitive_info != 0,
		EnableEncryption:     cConfig.enable_encryption != 0,
		EnableSignature:      cConfig.enable_signature != 0,
		EncryptionKey:        C.GoString(&cConfig.encryption_key[0]),
		TimeoutSeconds:       int(cConfig.timeout_seconds),
	}
}

// safeStrCopy 安全地复制字符串到C数组
func (d *DeviceFingerprintCollector) safeStrCopy(dest []C.char, src string) {
	srcBytes := []byte(src)
	maxLen := len(dest) - 1 // 保留一个字节用于null终止符

	copyLen := len(srcBytes)
	if copyLen > maxLen {
		copyLen = maxLen
	}

	// 清空目标数组
	for i := range dest {
		dest[i] = 0
	}

	// 复制数据
	for i := 0; i < copyLen; i++ {
		dest[i] = C.char(srcBytes[i])
	}
}
