package cpp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCryptoService_EncryptDecrypt 测试加密解密功能
func TestCryptoService_EncryptDecrypt(t *testing.T) {
	cryptoService := NewCryptoService()
	
	testData := []byte("Hello, World! This is a test message for encryption.")
	testKey := "test_encryption_key_123456"
	
	// 测试加密
	encryptedData, err := cryptoService.EncryptData(testData, testKey)
	require.NoError(t, err)
	assert.NotEmpty(t, encryptedData)
	assert.NotEqual(t, testData, encryptedData)
	
	// 测试解密
	decryptedData, err := cryptoService.DecryptData(encryptedData, testKey)
	require.NoError(t, err)
	assert.Equal(t, testData, decryptedData)
}

// TestCryptoService_SignVerify 测试签名验证功能
func TestCryptoService_SignVerify(t *testing.T) {
	cryptoService := NewCryptoService()
	
	testData := []byte("Test data for digital signature")
	privateKey := "test_private_key_for_signing"
	publicKey := "test_public_key_for_verification"
	
	// 测试签名
	signature, err := cryptoService.SignData(testData, privateKey)
	require.NoError(t, err)
	assert.NotEmpty(t, signature)
	
	// 测试验证
	isValid, err := cryptoService.VerifySignature(testData, signature, publicKey)
	require.NoError(t, err)
	assert.True(t, isValid)
	
	// 测试错误的签名
	wrongSignature := []byte("wrong_signature")
	isValid, err = cryptoService.VerifySignature(testData, wrongSignature, publicKey)
	require.NoError(t, err)
	assert.False(t, isValid)
}

// TestCryptoService_EncryptDecryptFingerprint 测试指纹加密解密
func TestCryptoService_EncryptDecryptFingerprint(t *testing.T) {
	cryptoService := NewCryptoService()
	
	// 创建测试指纹
	fingerprint := &DeviceFingerprint{
		DeviceID:          "test_device_123",
		DeviceType:        DeviceTypeDesktop,
		CPUID:             "test_cpu_id",
		CPUModel:          "Test CPU Model",
		CPUCores:          8,
		TotalMemory:       16777216000,
		MotherboardSerial: "test_mb_serial",
		OSType:            "windows",
		OSVersion:         "Windows 10",
		Hostname:          "test_host",
		Username:          "test_user",
		ScreenResolution:  "1920x1080",
		FingerprintHash:   "test_hash_123",
		ConfidenceScore:   95,
	}
	
	testKey := "fingerprint_encryption_key"
	
	// 测试指纹加密
	encryptedFingerprint, err := cryptoService.EncryptFingerprint(fingerprint, testKey)
	require.NoError(t, err)
	assert.NotNil(t, encryptedFingerprint)
	assert.NotEqual(t, fingerprint.DeviceID, encryptedFingerprint.DeviceID)
	
	// 测试指纹解密
	decryptedFingerprint, err := cryptoService.DecryptFingerprint(encryptedFingerprint, testKey)
	require.NoError(t, err)
	assert.Equal(t, fingerprint.DeviceID, decryptedFingerprint.DeviceID)
	assert.Equal(t, fingerprint.CPUModel, decryptedFingerprint.CPUModel)
	assert.Equal(t, fingerprint.ConfidenceScore, decryptedFingerprint.ConfidenceScore)
}

// TestCryptoService_HashData 测试数据哈希
func TestCryptoService_HashData(t *testing.T) {
	cryptoService := NewCryptoService()
	
	testData := []byte("Test data for hashing")
	
	// 测试哈希计算
	hash1, err := cryptoService.HashData(testData)
	require.NoError(t, err)
	assert.NotEmpty(t, hash1)
	
	// 相同数据应该产生相同哈希
	hash2, err := cryptoService.HashData(testData)
	require.NoError(t, err)
	assert.Equal(t, hash1, hash2)
	
	// 不同数据应该产生不同哈希
	differentData := []byte("Different test data")
	hash3, err := cryptoService.HashData(differentData)
	require.NoError(t, err)
	assert.NotEqual(t, hash1, hash3)
}

// TestCryptoService_ValidateDataIntegrity 测试数据完整性验证
func TestCryptoService_ValidateDataIntegrity(t *testing.T) {
	cryptoService := NewCryptoService()
	
	testData := []byte("Test data for integrity validation")
	
	// 计算预期哈希
	expectedHash, err := cryptoService.HashData(testData)
	require.NoError(t, err)
	
	// 测试完整性验证
	isValid, err := cryptoService.ValidateDataIntegrity(testData, expectedHash)
	require.NoError(t, err)
	assert.True(t, isValid)
	
	// 测试错误的哈希
	wrongHash := "wrong_hash_value"
	isValid, err = cryptoService.ValidateDataIntegrity(testData, wrongHash)
	require.NoError(t, err)
	assert.False(t, isValid)
}

// TestCryptoService_GenerateSecureKey 测试安全密钥生成
func TestCryptoService_GenerateSecureKey(t *testing.T) {
	cryptoService := NewCryptoService()
	
	// 测试不同长度的密钥生成
	testLengths := []int{16, 32, 64, 128}
	
	for _, length := range testLengths {
		key, err := cryptoService.GenerateSecureKey(length)
		require.NoError(t, err)
		assert.NotEmpty(t, key)
		assert.LessOrEqual(t, len(key), length)
	}
}

// TestCryptoService_ValidateKey 测试密钥验证
func TestCryptoService_ValidateKey(t *testing.T) {
	cryptoService := NewCryptoService()
	
	// 测试有效密钥
	validKey := "valid_key_12345678"
	err := cryptoService.ValidateKey(validKey)
	assert.NoError(t, err)
	
	// 测试空密钥
	err = cryptoService.ValidateKey("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
	
	// 测试过短密钥
	shortKey := "short"
	err = cryptoService.ValidateKey(shortKey)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too short")
	
	// 测试过长密钥
	longKey := make([]byte, 300)
	for i := range longKey {
		longKey[i] = 'a'
	}
	err = cryptoService.ValidateKey(string(longKey))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too long")
}

// TestCryptoService_HexConversion 测试十六进制转换
func TestCryptoService_HexConversion(t *testing.T) {
	cryptoService := NewCryptoService()
	
	testData := []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF}
	expectedHex := "0123456789abcdef"
	
	// 测试字节到十六进制
	hexStr := cryptoService.bytesToHex(testData)
	assert.Equal(t, expectedHex, hexStr)
	
	// 测试十六进制到字节
	convertedData, err := cryptoService.hexToBytes(hexStr)
	require.NoError(t, err)
	assert.Equal(t, testData, convertedData)
	
	// 测试无效十六进制
	invalidHex := "invalid_hex_string"
	_, err = cryptoService.hexToBytes(invalidHex)
	assert.Error(t, err)
	
	// 测试奇数长度十六进制
	oddLengthHex := "123"
	_, err = cryptoService.hexToBytes(oddLengthHex)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "length")
}

// TestCryptoService_SecureClearMemory 测试安全内存清除
func TestCryptoService_SecureClearMemory(t *testing.T) {
	cryptoService := NewCryptoService()
	
	sensitiveData := []byte("sensitive_password_123")
	originalData := make([]byte, len(sensitiveData))
	copy(originalData, sensitiveData)
	
	// 清除内存
	cryptoService.SecureClearMemory(sensitiveData)
	
	// 验证数据已被清除
	for _, b := range sensitiveData {
		assert.Equal(t, byte(0), b)
	}
	
	// 确保原始数据不同
	assert.NotEqual(t, originalData, sensitiveData)
}

// TestCryptoService_ErrorHandling 测试错误处理
func TestCryptoService_ErrorHandling(t *testing.T) {
	cryptoService := NewCryptoService()
	
	// 测试空数据加密
	_, err := cryptoService.EncryptData(nil, "test_key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
	
	// 测试空密钥加密
	testData := []byte("test data")
	_, err = cryptoService.EncryptData(testData, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
	
	// 测试空数据哈希
	_, err = cryptoService.HashData(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
	
	// 测试无效长度密钥生成
	_, err = cryptoService.GenerateSecureKey(0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "positive")
	
	_, err = cryptoService.GenerateSecureKey(300)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too large")
}

// BenchmarkCryptoService_EncryptData 基准测试：数据加密
func BenchmarkCryptoService_EncryptData(b *testing.B) {
	cryptoService := NewCryptoService()
	testData := make([]byte, 1024) // 1KB数据
	testKey := "benchmark_encryption_key"
	
	// 填充测试数据
	for i := range testData {
		testData[i] = byte(i % 256)
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_, err := cryptoService.EncryptData(testData, testKey)
		if err != nil {
			b.Fatalf("EncryptData failed: %v", err)
		}
	}
}

// BenchmarkCryptoService_DecryptData 基准测试：数据解密
func BenchmarkCryptoService_DecryptData(b *testing.B) {
	cryptoService := NewCryptoService()
	testData := make([]byte, 1024) // 1KB数据
	testKey := "benchmark_encryption_key"
	
	// 填充测试数据
	for i := range testData {
		testData[i] = byte(i % 256)
	}
	
	// 预先加密数据
	encryptedData, err := cryptoService.EncryptData(testData, testKey)
	require.NoError(b, err)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_, err := cryptoService.DecryptData(encryptedData, testKey)
		if err != nil {
			b.Fatalf("DecryptData failed: %v", err)
		}
	}
}

// BenchmarkCryptoService_HashData 基准测试：数据哈希
func BenchmarkCryptoService_HashData(b *testing.B) {
	cryptoService := NewCryptoService()
	testData := make([]byte, 1024) // 1KB数据
	
	// 填充测试数据
	for i := range testData {
		testData[i] = byte(i % 256)
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_, err := cryptoService.HashData(testData)
		if err != nil {
			b.Fatalf("HashData failed: %v", err)
		}
	}
}

// BenchmarkCryptoService_SignData 基准测试：数据签名
func BenchmarkCryptoService_SignData(b *testing.B) {
	cryptoService := NewCryptoService()
	testData := make([]byte, 256) // 256字节数据
	privateKey := "benchmark_private_key"
	
	// 填充测试数据
	for i := range testData {
		testData[i] = byte(i % 256)
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_, err := cryptoService.SignData(testData, privateKey)
		if err != nil {
			b.Fatalf("SignData failed: %v", err)
		}
	}
}

// BenchmarkCryptoService_VerifySignature 基准测试：签名验证
func BenchmarkCryptoService_VerifySignature(b *testing.B) {
	cryptoService := NewCryptoService()
	testData := make([]byte, 256) // 256字节数据
	privateKey := "benchmark_private_key"
	publicKey := "benchmark_public_key"
	
	// 填充测试数据
	for i := range testData {
		testData[i] = byte(i % 256)
	}
	
	// 预先生成签名
	signature, err := cryptoService.SignData(testData, privateKey)
	require.NoError(b, err)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_, err := cryptoService.VerifySignature(testData, signature, publicKey)
		if err != nil {
			b.Fatalf("VerifySignature failed: %v", err)
		}
	}
}

// BenchmarkCryptoService_GenerateSecureKey 基准测试：安全密钥生成
func BenchmarkCryptoService_GenerateSecureKey(b *testing.B) {
	cryptoService := NewCryptoService()
	keyLength := 32
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_, err := cryptoService.GenerateSecureKey(keyLength)
		if err != nil {
			b.Fatalf("GenerateSecureKey failed: %v", err)
		}
	}
}

// BenchmarkCryptoService_BytesToHex 基准测试：字节转十六进制
func BenchmarkCryptoService_BytesToHex(b *testing.B) {
	cryptoService := NewCryptoService()
	testData := make([]byte, 1024) // 1KB数据
	
	// 填充测试数据
	for i := range testData {
		testData[i] = byte(i % 256)
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_ = cryptoService.bytesToHex(testData)
	}
}

// BenchmarkCryptoService_HexToBytes 基准测试：十六进制转字节
func BenchmarkCryptoService_HexToBytes(b *testing.B) {
	cryptoService := NewCryptoService()
	testData := make([]byte, 1024) // 1KB数据
	
	// 填充测试数据
	for i := range testData {
		testData[i] = byte(i % 256)
	}
	
	hexStr := cryptoService.bytesToHex(testData)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_, err := cryptoService.hexToBytes(hexStr)
		if err != nil {
			b.Fatalf("HexToBytes failed: %v", err)
		}
	}
}