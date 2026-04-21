package cpp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLicenseService_GenerateAndValidate 测试许可证生成和验证
func TestLicenseService_GenerateAndValidate(t *testing.T) {
	licenseService := NewLicenseService()
	
	deviceID := "test_device_12345"
	expiresAt := time.Now().Add(30 * 24 * time.Hour) // 30天后过期
	privateKey := "test_private_key_for_license"
	licenseType := "commercial"
	features := []string{"feature1", "feature2", "feature3"}
	
	// 生成许可证
	licenseData, err := licenseService.GenerateLicense(deviceID, expiresAt, privateKey, licenseType, features)
	require.NoError(t, err)
	assert.NotEmpty(t, licenseData)
	
	// 验证许可证
	licenseInfo, err := licenseService.ValidateLicense(licenseData, deviceID)
	require.NoError(t, err)
	require.NotNil(t, licenseInfo)
	
	assert.Equal(t, deviceID, licenseInfo.DeviceID)
	assert.Equal(t, licenseType, licenseInfo.LicenseType)
	assert.True(t, licenseInfo.IsValid)
	assert.Equal(t, features, licenseInfo.Features)
	assert.True(t, licenseInfo.ExpiresAt.After(time.Now()))
}

// TestLicenseService_VerifySignature 测试许可证签名验证
func TestLicenseService_VerifySignature(t *testing.T) {
	licenseService := NewLicenseService()
	
	deviceID := "test_device_signature"
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	privateKey := "test_private_key_signature"
	publicKey := "test_public_key_signature"
	licenseType := "trial"
	features := []string{"basic"}
	
	// 生成许可证
	licenseData, err := licenseService.GenerateLicense(deviceID, expiresAt, privateKey, licenseType, features)
	require.NoError(t, err)
	
	// 验证签名
	isValid, err := licenseService.VerifyLicenseSignature(licenseData, publicKey)
	require.NoError(t, err)
	assert.True(t, isValid)
	
	// 测试错误的公钥
	wrongPublicKey := "wrong_public_key"
	isValid, err = licenseService.VerifyLicenseSignature(licenseData, wrongPublicKey)
	require.NoError(t, err)
	assert.False(t, isValid)
}

// TestLicenseService_CheckExpiry 测试许可证过期检查
func TestLicenseService_CheckExpiry(t *testing.T) {
	licenseService := NewLicenseService()
	
	deviceID := "test_device_expiry"
	privateKey := "test_private_key_expiry"
	licenseType := "trial"
	features := []string{"basic"}
	
	// 测试未过期许可证
	futureExpiry := time.Now().Add(24 * time.Hour)
	licenseData, err := licenseService.GenerateLicense(deviceID, futureExpiry, privateKey, licenseType, features)
	require.NoError(t, err)
	
	isExpired, remaining, err := licenseService.CheckLicenseExpiry(licenseData)
	require.NoError(t, err)
	assert.False(t, isExpired)
	assert.Greater(t, remaining, time.Duration(0))
	
	// 测试已过期许可证
	pastExpiry := time.Now().Add(-24 * time.Hour)
	expiredLicenseData, err := licenseService.GenerateLicense(deviceID, pastExpiry, privateKey, licenseType, features)
	require.NoError(t, err)
	
	isExpired, remaining, err = licenseService.CheckLicenseExpiry(expiredLicenseData)
	require.NoError(t, err)
	assert.True(t, isExpired)
	assert.Equal(t, time.Duration(0), remaining)
}

// TestLicenseService_DeviceBinding 测试设备绑定验证
func TestLicenseService_DeviceBinding(t *testing.T) {
	licenseService := NewLicenseService()
	
	deviceID := "test_device_binding"
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	privateKey := "test_private_key_binding"
	licenseType := "commercial"
	features := []string{"full_access"}
	
	// 生成许可证
	licenseData, err := licenseService.GenerateLicense(deviceID, expiresAt, privateKey, licenseType, features)
	require.NoError(t, err)
	
	// 测试正确的设备ID
	isValid, err := licenseService.ValidateDeviceBinding(licenseData, deviceID)
	require.NoError(t, err)
	assert.True(t, isValid)
	
	// 测试错误的设备ID
	wrongDeviceID := "wrong_device_id"
	isValid, err = licenseService.ValidateDeviceBinding(licenseData, wrongDeviceID)
	require.NoError(t, err)
	assert.False(t, isValid)
}

// TestLicenseService_FeatureAccess 测试功能访问验证
func TestLicenseService_FeatureAccess(t *testing.T) {
	licenseService := NewLicenseService()
	
	deviceID := "test_device_features"
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	privateKey := "test_private_key_features"
	licenseType := "commercial"
	features := []string{"feature_a", "feature_b", "feature_c"}
	
	// 生成许可证
	licenseData, err := licenseService.GenerateLicense(deviceID, expiresAt, privateKey, licenseType, features)
	require.NoError(t, err)
	
	// 测试已授权功能
	hasAccess, err := licenseService.ValidateFeatureAccess(licenseData, "feature_a")
	require.NoError(t, err)
	assert.True(t, hasAccess)
	
	hasAccess, err = licenseService.ValidateFeatureAccess(licenseData, "feature_b")
	require.NoError(t, err)
	assert.True(t, hasAccess)
	
	// 测试未授权功能
	hasAccess, err = licenseService.ValidateFeatureAccess(licenseData, "feature_x")
	require.NoError(t, err)
	assert.False(t, hasAccess)
	
	// 测试通配符权限
	allFeaturesLicenseData, err := licenseService.GenerateLicense(deviceID, expiresAt, privateKey, "enterprise", []string{"*"})
	require.NoError(t, err)
	
	hasAccess, err = licenseService.ValidateFeatureAccess(allFeaturesLicenseData, "any_feature")
	require.NoError(t, err)
	assert.True(t, hasAccess)
}

// TestLicenseService_RenewLicense 测试许可证续期
func TestLicenseService_RenewLicense(t *testing.T) {
	licenseService := NewLicenseService()
	
	deviceID := "test_device_renew"
	originalExpiry := time.Now().Add(7 * 24 * time.Hour)
	privateKey := "test_private_key_renew"
	licenseType := "commercial"
	features := []string{"feature1", "feature2"}
	
	// 生成原始许可证
	originalLicenseData, err := licenseService.GenerateLicense(deviceID, originalExpiry, privateKey, licenseType, features)
	require.NoError(t, err)
	
	// 续期许可证
	newExpiry := time.Now().Add(60 * 24 * time.Hour) // 60天
	renewedLicenseData, err := licenseService.RenewLicense(originalLicenseData, newExpiry, privateKey)
	require.NoError(t, err)
	assert.NotEmpty(t, renewedLicenseData)
	
	// 验证续期后的许可证
	renewedLicenseInfo, err := licenseService.ValidateLicense(renewedLicenseData, deviceID)
	require.NoError(t, err)
	
	assert.Equal(t, deviceID, renewedLicenseInfo.DeviceID)
	assert.Equal(t, licenseType, renewedLicenseInfo.LicenseType)
	assert.Equal(t, features, renewedLicenseInfo.Features)
	assert.True(t, renewedLicenseInfo.ExpiresAt.After(originalExpiry))
}

// TestLicenseService_RevokeLicense 测试许可证吊销
func TestLicenseService_RevokeLicense(t *testing.T) {
	licenseService := NewLicenseService()
	
	deviceID := "test_device_revoke"
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	privateKey := "test_private_key_revoke"
	licenseType := "commercial"
	features := []string{"feature1"}
	
	// 生成许可证
	licenseData, err := licenseService.GenerateLicense(deviceID, expiresAt, privateKey, licenseType, features)
	require.NoError(t, err)
	
	// 验证许可证有效
	licenseInfo, err := licenseService.ValidateLicense(licenseData, deviceID)
	require.NoError(t, err)
	assert.True(t, licenseInfo.IsValid)
	
	// 吊销许可证
	err = licenseService.RevokeLicense(licenseData, privateKey)
	require.NoError(t, err)
	
	// 注意：在实际实现中，吊销后需要检查黑名单
	// 这里仅测试吊销操作不出错
}

// TestLicenseService_BatchValidation 测试批量验证
func TestLicenseService_BatchValidation(t *testing.T) {
	licenseService := NewLicenseService()
	
	deviceID := "test_device_batch"
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	privateKey := "test_private_key_batch"
	
	// 生成多个许可证
	var licenses []string
	for i := 0; i < 5; i++ {
		licenseData, err := licenseService.GenerateLicense(deviceID, expiresAt, privateKey, "trial", []string{"basic"})
		require.NoError(t, err)
		licenses = append(licenses, licenseData)
	}
	
	// 添加一个无效许可证
	licenses = append(licenses, "invalid_license_data")
	
	// 批量验证
	results, err := licenseService.BatchValidateLicenses(licenses, deviceID)
	require.NoError(t, err)
	require.Len(t, results, 6)
	
	// 前5个应该有效，最后一个应该无效
	for i := 0; i < 5; i++ {
		assert.True(t, results[i], "License %d should be valid", i)
	}
	assert.False(t, results[5], "Invalid license should be false")
}

// TestLicenseService_TrialLicense 测试试用许可证
func TestLicenseService_TrialLicense(t *testing.T) {
	licenseService := NewLicenseService()
	
	deviceID := "test_device_trial"
	trialDays := 14
	privateKey := "test_private_key_trial"
	
	// 生成试用许可证
	trialLicenseData, err := licenseService.GenerateTrialLicense(deviceID, trialDays, privateKey)
	require.NoError(t, err)
	assert.NotEmpty(t, trialLicenseData)
	
	// 验证试用许可证
	licenseInfo, err := licenseService.ValidateLicense(trialLicenseData, deviceID)
	require.NoError(t, err)
	
	assert.Equal(t, deviceID, licenseInfo.DeviceID)
	assert.Equal(t, "trial", licenseInfo.LicenseType)
	assert.True(t, licenseInfo.IsValid)
	assert.Contains(t, licenseInfo.Features, "trial_mode")
	
	// 检查过期时间约为14天
	expectedExpiry := time.Now().Add(time.Duration(trialDays) * 24 * time.Hour)
	timeDiff := licenseInfo.ExpiresAt.Sub(expectedExpiry)
	assert.Less(t, timeDiff.Abs(), time.Hour) // 允许1小时误差
}

// TestLicenseService_CommercialLicense 测试商业许可证
func TestLicenseService_CommercialLicense(t *testing.T) {
	licenseService := NewLicenseService()
	
	deviceID := "test_device_commercial"
	validYears := 2
	privateKey := "test_private_key_commercial"
	features := []string{"premium_feature1", "premium_feature2"}
	
	// 生成商业许可证
	commercialLicenseData, err := licenseService.GenerateCommercialLicense(deviceID, validYears, privateKey, features)
	require.NoError(t, err)
	assert.NotEmpty(t, commercialLicenseData)
	
	// 验证商业许可证
	licenseInfo, err := licenseService.ValidateLicense(commercialLicenseData, deviceID)
	require.NoError(t, err)
	
	assert.Equal(t, deviceID, licenseInfo.DeviceID)
	assert.Equal(t, "commercial", licenseInfo.LicenseType)
	assert.True(t, licenseInfo.IsValid)
	assert.Equal(t, features, licenseInfo.Features)
	
	// 检查过期时间约为2年
	expectedExpiry := time.Now().Add(time.Duration(validYears) * 365 * 24 * time.Hour)
	timeDiff := licenseInfo.ExpiresAt.Sub(expectedExpiry)
	assert.Less(t, timeDiff.Abs(), 24*time.Hour) // 允许1天误差
}

// TestLicenseService_GetLicenseInfo 测试许可证信息获取
func TestLicenseService_GetLicenseInfo(t *testing.T) {
	licenseService := NewLicenseService()
	
	deviceID := "test_device_info"
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	privateKey := "test_private_key_info"
	licenseType := "enterprise"
	features := []string{"advanced_feature1", "advanced_feature2"}
	
	// 生成许可证
	licenseData, err := licenseService.GenerateLicense(deviceID, expiresAt, privateKey, licenseType, features)
	require.NoError(t, err)
	
	// 获取许可证信息
	licenseInfo, err := licenseService.GetLicenseInfo(licenseData)
	require.NoError(t, err)
	require.NotNil(t, licenseInfo)
	
	assert.Equal(t, deviceID, licenseInfo.DeviceID)
	assert.Equal(t, licenseType, licenseInfo.LicenseType)
	assert.Equal(t, features, licenseInfo.Features)
	assert.NotEmpty(t, licenseInfo.Signature)
}

// TestLicenseService_ErrorHandling 测试错误处理
func TestLicenseService_ErrorHandling(t *testing.T) {
	licenseService := NewLicenseService()
	
	// 测试空设备ID
	_, err := licenseService.GenerateLicense("", time.Now().Add(time.Hour), "key", "trial", []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "device ID")
	
	// 测试空私钥
	_, err = licenseService.GenerateLicense("device", time.Now().Add(time.Hour), "", "trial", []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "private key")
	
	// 测试空许可证数据验证
	_, err = licenseService.ValidateLicense("", "device")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "license data")
	
	// 测试无效试用天数
	_, err = licenseService.GenerateTrialLicense("device", 0, "key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "positive")
	
	_, err = licenseService.GenerateTrialLicense("device", 400, "key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "365")
	
	// 测试无效商业许可证年数
	_, err = licenseService.GenerateCommercialLicense("device", 0, "key", []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "positive")
	
	_, err = licenseService.GenerateCommercialLicense("device", 15, "key", []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "10")
}

// BenchmarkLicenseService_GenerateLicense 基准测试：许可证生成
func BenchmarkLicenseService_GenerateLicense(b *testing.B) {
	licenseService := NewLicenseService()
	
	deviceID := "benchmark_device"
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	privateKey := "benchmark_private_key"
	licenseType := "commercial"
	features := []string{"feature1", "feature2"}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_, err := licenseService.GenerateLicense(deviceID, expiresAt, privateKey, licenseType, features)
		if err != nil {
			b.Fatalf("GenerateLicense failed: %v", err)
		}
	}
}

// BenchmarkLicenseService_ValidateLicense 基准测试：许可证验证
func BenchmarkLicenseService_ValidateLicense(b *testing.B) {
	licenseService := NewLicenseService()
	
	deviceID := "benchmark_device"
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	privateKey := "benchmark_private_key"
	licenseType := "commercial"
	features := []string{"feature1", "feature2"}
	
	// 预先生成许可证
	licenseData, err := licenseService.GenerateLicense(deviceID, expiresAt, privateKey, licenseType, features)
	require.NoError(b, err)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_, err := licenseService.ValidateLicense(licenseData, deviceID)
		if err != nil {
			b.Fatalf("ValidateLicense failed: %v", err)
		}
	}
}

// BenchmarkLicenseService_VerifySignature 基准测试：签名验证
func BenchmarkLicenseService_VerifySignature(b *testing.B) {
	licenseService := NewLicenseService()
	
	deviceID := "benchmark_device"
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	privateKey := "benchmark_private_key"
	publicKey := "benchmark_public_key"
	licenseType := "commercial"
	features := []string{"feature1", "feature2"}
	
	// 预先生成许可证
	licenseData, err := licenseService.GenerateLicense(deviceID, expiresAt, privateKey, licenseType, features)
	require.NoError(b, err)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_, err := licenseService.VerifyLicenseSignature(licenseData, publicKey)
		if err != nil {
			b.Fatalf("VerifyLicenseSignature failed: %v", err)
		}
	}
}

// BenchmarkLicenseService_BatchValidation 基准测试：批量验证
func BenchmarkLicenseService_BatchValidation(b *testing.B) {
	licenseService := NewLicenseService()
	
	deviceID := "benchmark_device"
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	privateKey := "benchmark_private_key"
	
	// 预先生成100个许可证
	var licenses []string
	for i := 0; i < 100; i++ {
		licenseData, err := licenseService.GenerateLicense(deviceID, expiresAt, privateKey, "trial", []string{"basic"})
		require.NoError(b, err)
		licenses = append(licenses, licenseData)
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_, err := licenseService.BatchValidateLicenses(licenses, deviceID)
		if err != nil {
			b.Fatalf("BatchValidateLicenses failed: %v", err)
		}
	}
}