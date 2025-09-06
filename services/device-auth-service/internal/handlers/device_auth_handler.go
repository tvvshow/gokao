package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"device-auth-service/internal/models"
	"device-auth-service/internal/services"
)

// HealthCheckResponse 健康检查响应
type HealthCheckResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// DeviceFingerprintResponse 设备指纹响应
type DeviceFingerprintResponse struct {
	DeviceID        string    `json:"device_id"`
	DeviceType      string    `json:"device_type"`
	CPUID           string    `json:"cpu_id"`
	CPUModel        string    `json:"cpu_model"`
	CPUCores        uint32    `json:"cpu_cores"`
	TotalMemory     uint64    `json:"total_memory"`
	OSType          string    `json:"os_type"`
	OSVersion       string    `json:"os_version"`
	Hostname        string    `json:"hostname"`
	Username        string    `json:"username"`
	ScreenResolution string   `json:"screen_resolution"`
	FingerprintHash  string   `json:"fingerprint_hash"`
	ConfidenceScore  uint32   `json:"confidence_score"`
	CollectedAt      time.Time `json:"collected_at"`
}

// LicenseValidationResponse 许可证验证响应
type LicenseValidationResponse struct {
	DeviceID     string    `json:"device_id"`
	LicenseType  string    `json:"license_type"`
	IssuedAt     time.Time `json:"issued_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	MaxDevices   int       `json:"max_devices"`
	Features     []string  `json:"features"`
	IsValid      bool      `json:"is_valid"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

// DeviceAuthHandler 设备认证处理器
type DeviceAuthHandler struct {
	deviceAuthService *services.DeviceAuthService
}

// NewDeviceAuthHandler 创建设备认证处理器实例
func NewDeviceAuthHandler(deviceAuthService *services.DeviceAuthService) *DeviceAuthHandler {
	return &DeviceAuthHandler{
		deviceAuthService: deviceAuthService,
	}
}

// HealthCheckHandler 健康检查处理器
func (h *DeviceAuthHandler) HealthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, HealthCheckResponse{
		Status:  "ok",
		Message: "Device Auth Service is running",
	})
}

// CollectDeviceFingerprintHandler 采集设备指纹处理器
func (h *DeviceAuthHandler) CollectDeviceFingerprintHandler(c *gin.Context) {
	// 采集设备指纹
	fingerprint, err := h.deviceAuthService.CollectDeviceFingerprint()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 转换为响应格式
	response := DeviceFingerprintResponse{
		DeviceID:         fingerprint.DeviceID,
		DeviceType:       string(fingerprint.DeviceType),
		CPUID:            fingerprint.CPUID,
		CPUModel:         fingerprint.CPUModel,
		CPUCores:         fingerprint.CPUCores,
		TotalMemory:      fingerprint.TotalMemory,
		OSType:           fingerprint.OSType,
		OSVersion:        fingerprint.OSVersion,
		Hostname:         fingerprint.Hostname,
		Username:         fingerprint.Username,
		ScreenResolution: fingerprint.ScreenResolution,
		FingerprintHash:  fingerprint.FingerprintHash,
		ConfidenceScore:  fingerprint.ConfidenceScore,
		CollectedAt:      fingerprint.CollectedAt,
	}

	c.JSON(http.StatusOK, response)
}

// RegisterDeviceHandler 注册设备处理器
func (h *DeviceAuthHandler) RegisterDeviceHandler(c *gin.Context) {
	// 生成设备ID并存储设备信息
	deviceID, err := h.deviceAuthService.GenerateDeviceID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	device := &models.Device{
		ID:         uuid.New(),
		DeviceID:   deviceID,
		DeviceType: "unknown", // 实际应用中应从指纹中获取
	}

	if err := h.deviceAuthService.StoreDevice(device); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"device_id": deviceID})
}

// ValidateLicenseHandler 验证许可证处理器
func (h *DeviceAuthHandler) ValidateLicenseHandler(c *gin.Context) {
	// 获取请求参数
	var req struct {
		LicenseData string `json:"license_data"`
		DeviceID    string `json:"device_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证许可证
	licenseInfo, err := h.deviceAuthService.ValidateLicense(req.LicenseData, req.DeviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 转换为响应格式
	response := LicenseValidationResponse{
		DeviceID:     licenseInfo.DeviceID,
		LicenseType:  licenseInfo.LicenseType,
		IssuedAt:     licenseInfo.IssuedAt,
		ExpiresAt:    licenseInfo.ExpiresAt,
		MaxDevices:   licenseInfo.MaxDevices,
		Features:     licenseInfo.Features,
		IsValid:      licenseInfo.IsValid,
		ErrorMessage: licenseInfo.ErrorMessage,
	}

	c.JSON(http.StatusOK, response)
}

// GenerateLicenseHandler 生成许可证处理器
func (h *DeviceAuthHandler) GenerateLicenseHandler(c *gin.Context) {
	// 获取请求参数
	var req struct {
		DeviceID    string    `json:"device_id"`
		ExpiresAt   time.Time `json:"expires_at"`
		PrivateKey  string    `json:"private_key"`
		LicenseType string    `json:"license_type"`
		Features    []string  `json:"features"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 生成许可证
	licenseData, err := h.deviceAuthService.GenerateLicense(
		req.DeviceID,
		req.ExpiresAt,
		req.PrivateKey,
		req.LicenseType,
		req.Features,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"license_data": licenseData})
}