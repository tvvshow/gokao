package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// DeviceAuthClient 设备认证服务客户端
type DeviceAuthClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *logrus.Logger
}

// DeviceFingerprintResponse 设备指纹响应
type DeviceFingerprintResponse struct {
	DeviceID         string    `json:"device_id"`
	DeviceType       string    `json:"device_type"`
	CPUID            string    `json:"cpu_id"`
	CPUModel         string    `json:"cpu_model"`
	CPUCores         uint32    `json:"cpu_cores"`
	TotalMemory      uint64    `json:"total_memory"`
	OSType           string    `json:"os_type"`
	OSVersion        string    `json:"os_version"`
	Hostname         string    `json:"hostname"`
	Username         string    `json:"username"`
	ScreenResolution string    `json:"screen_resolution"`
	FingerprintHash  string    `json:"fingerprint_hash"`
	ConfidenceScore  uint32    `json:"confidence_score"`
	CollectedAt      time.Time `json:"collected_at"`
}

// LicenseValidationRequest 许可证验证请求
type LicenseValidationRequest struct {
	LicenseData string `json:"license_data"`
	DeviceID    string `json:"device_id"`
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

// NewDeviceAuthClient 创建设备认证客户端
func NewDeviceAuthClient(baseURL string, logger *logrus.Logger) *DeviceAuthClient {
	return &DeviceAuthClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// CollectDeviceFingerprint 采集设备指纹
func (c *DeviceAuthClient) CollectDeviceFingerprint(ctx context.Context) (*DeviceFingerprintResponse, error) {
	url := fmt.Sprintf("%s/api/v1/device/fingerprint", c.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device auth service returned status %d", resp.StatusCode)
	}
	
	var fingerprint DeviceFingerprintResponse
	if err := json.NewDecoder(resp.Body).Decode(&fingerprint); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &fingerprint, nil
}

// ValidateLicense 验证许可证
func (c *DeviceAuthClient) ValidateLicense(ctx context.Context, licenseData, deviceID string) (*LicenseValidationResponse, error) {
	url := fmt.Sprintf("%s/api/v1/license/validate", c.baseURL)
	
	request := LicenseValidationRequest{
		LicenseData: licenseData,
		DeviceID:    deviceID,
	}
	
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device auth service returned status %d", resp.StatusCode)
	}
	
	var response LicenseValidationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &response, nil
}