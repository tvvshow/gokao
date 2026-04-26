//go:build cgo && devicefingerprint && !windows
// +build cgo,devicefingerprint,!windows

package cpp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDeviceFingerprintCollector_Initialize 测试初始化功能
func TestDeviceFingerprintCollector_Initialize(t *testing.T) {
	collector := NewDeviceFingerprintCollector()

	// 测试初始化
	err := collector.Initialize("")
	require.NoError(t, err)
	assert.True(t, collector.IsInitialized())

	// 清理
	collector.Uninitialize()
	assert.False(t, collector.IsInitialized())
}

// TestDeviceFingerprintCollector_CollectFingerprint 测试指纹采集
func TestDeviceFingerprintCollector_CollectFingerprint(t *testing.T) {
	collector := NewDeviceFingerprintCollector()
	err := collector.Initialize("")
	require.NoError(t, err)
	defer collector.Uninitialize()

	// 测试采集指纹
	fingerprint, err := collector.CollectFingerprint()
	require.NoError(t, err)
	require.NotNil(t, fingerprint)

	// 验证指纹字段
	assert.NotEmpty(t, fingerprint.DeviceID)
	assert.NotEmpty(t, fingerprint.FingerprintHash)
	assert.Greater(t, fingerprint.ConfidenceScore, uint32(0))
	assert.NotEqual(t, DeviceTypeUnknown, fingerprint.DeviceType)
}

// TestQuickCollectFingerprint 测试快速采集
func TestQuickCollectFingerprint(t *testing.T) {
	fingerprint, err := QuickCollectFingerprint()
	require.NoError(t, err)
	require.NotNil(t, fingerprint)

	assert.NotEmpty(t, fingerprint.DeviceID)
	assert.NotEmpty(t, fingerprint.FingerprintHash)
}

// TestDeviceFingerprintCollector_Configuration 测试配置设置
func TestDeviceFingerprintCollector_Configuration(t *testing.T) {
	collector := NewDeviceFingerprintCollector()
	err := collector.Initialize("")
	require.NoError(t, err)
	defer collector.Uninitialize()

	// 设置配置
	config := &Configuration{
		CollectSensitiveInfo: true,
		EnableEncryption:     true,
		EnableSignature:      true,
		EncryptionKey:        "test_key_123",
		TimeoutSeconds:       30,
	}

	err = collector.SetConfiguration(config)
	require.NoError(t, err)

	// 获取配置验证
	retrievedConfig, err := collector.GetConfiguration()
	require.NoError(t, err)
	assert.Equal(t, config.CollectSensitiveInfo, retrievedConfig.CollectSensitiveInfo)
	assert.Equal(t, config.EnableEncryption, retrievedConfig.EnableEncryption)
	assert.Equal(t, config.EnableSignature, retrievedConfig.EnableSignature)
}

// TestDeviceFingerprintCollector_HashGeneration 测试哈希生成
func TestDeviceFingerprintCollector_HashGeneration(t *testing.T) {
	collector := NewDeviceFingerprintCollector()
	err := collector.Initialize("")
	require.NoError(t, err)
	defer collector.Uninitialize()

	// 采集指纹
	fingerprint, err := collector.CollectFingerprint()
	require.NoError(t, err)

	// 生成哈希
	hash, err := collector.GenerateHash(fingerprint)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.Equal(t, fingerprint.FingerprintHash, hash)
}

// TestDeviceFingerprintCollector_Comparison 测试指纹比较
func TestDeviceFingerprintCollector_Comparison(t *testing.T) {
	collector := NewDeviceFingerprintCollector()
	err := collector.Initialize("")
	require.NoError(t, err)
	defer collector.Uninitialize()

	// 采集两次指纹
	fp1, err := collector.CollectFingerprint()
	require.NoError(t, err)

	fp2, err := collector.CollectFingerprint()
	require.NoError(t, err)

	// 比较指纹
	comparison, err := collector.CompareFingerprints(fp1, fp2)
	require.NoError(t, err)

	// 同一设备的指纹应该相似
	assert.True(t, comparison.IsSameDevice)
	assert.Greater(t, comparison.SimilarityScore, 0.9)
	assert.Greater(t, comparison.ConfidenceLevel, uint32(90))
}

// TestDeviceFingerprintCollector_Validation 测试指纹验证
func TestDeviceFingerprintCollector_Validation(t *testing.T) {
	collector := NewDeviceFingerprintCollector()
	err := collector.Initialize("")
	require.NoError(t, err)
	defer collector.Uninitialize()

	// 采集指纹
	fingerprint, err := collector.CollectFingerprint()
	require.NoError(t, err)

	// 验证指纹
	isValid, err := collector.ValidateFingerprint(fingerprint, fingerprint.FingerprintHash)
	require.NoError(t, err)
	assert.True(t, isValid)

	// 使用错误的哈希验证
	isValid, err = collector.ValidateFingerprint(fingerprint, "wrong_hash")
	require.NoError(t, err)
	assert.False(t, isValid)
}

// TestDeviceFingerprintCollector_Serialization 测试序列化
func TestDeviceFingerprintCollector_Serialization(t *testing.T) {
	collector := NewDeviceFingerprintCollector()
	err := collector.Initialize("")
	require.NoError(t, err)
	defer collector.Uninitialize()

	// 采集指纹
	originalFp, err := collector.CollectFingerprint()
	require.NoError(t, err)

	// 序列化为JSON
	jsonData, err := collector.SerializeToJSON(originalFp)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// 从JSON反序列化
	deserializedFp, err := collector.DeserializeFromJSON(jsonData)
	require.NoError(t, err)

	// 验证反序列化的数据
	assert.Equal(t, originalFp.DeviceID, deserializedFp.DeviceID)
	assert.Equal(t, originalFp.FingerprintHash, deserializedFp.FingerprintHash)
	assert.Equal(t, originalFp.ConfidenceScore, deserializedFp.ConfidenceScore)
}

// TestDeviceFingerprintCollector_SecurityChecks 测试安全检查
func TestDeviceFingerprintCollector_SecurityChecks(t *testing.T) {
	collector := NewDeviceFingerprintCollector()
	err := collector.Initialize("")
	require.NoError(t, err)
	defer collector.Uninitialize()

	// 检查调试器
	isDebugger, err := collector.IsDebuggerPresent()
	require.NoError(t, err)
	t.Logf("Debugger present: %v", isDebugger)

	// 检查虚拟机
	isVM, err := collector.IsVirtualMachine()
	require.NoError(t, err)
	t.Logf("Virtual machine: %v", isVM)

	// 检查安全性
	securityLevel, riskFactors, err := collector.CheckSecurity()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, securityLevel, 0)
	assert.LessOrEqual(t, securityLevel, 100)
	t.Logf("Security level: %d, Risk factors: %s", securityLevel, riskFactors)
}

// TestDeviceFingerprintCollector_PerformanceStats 测试性能统计
func TestDeviceFingerprintCollector_PerformanceStats(t *testing.T) {
	collector := NewDeviceFingerprintCollector()
	err := collector.Initialize("")
	require.NoError(t, err)
	defer collector.Uninitialize()

	// 启用性能监控
	err = collector.SetPerformanceMonitoring(true)
	require.NoError(t, err)

	// 重置统计
	err = collector.ResetPerformanceStats()
	require.NoError(t, err)

	// 执行一些操作
	_, err = collector.CollectFingerprint()
	require.NoError(t, err)

	// 获取性能统计
	stats, err := collector.GetPerformanceStats()
	require.NoError(t, err)
	require.NotNil(t, stats)

	assert.Greater(t, stats.TotalCalls, uint32(0))
	assert.Greater(t, stats.SuccessCalls, uint32(0))
	assert.GreaterOrEqual(t, stats.CollectTimeUs, uint64(0))
}

// TestGetVersion 测试版本获取
func TestGetVersion(t *testing.T) {
	version, err := GetVersion()
	require.NoError(t, err)
	assert.NotEmpty(t, version)
	t.Logf("Library version: %s", version)
}

// TestGetSupportedPlatforms 测试支持的平台
func TestGetSupportedPlatforms(t *testing.T) {
	platforms, err := GetSupportedPlatforms()
	require.NoError(t, err)
	assert.NotEmpty(t, platforms)
	t.Logf("Supported platforms: %v", platforms)
}

// TestDeviceFingerprintCollector_ErrorHandling 测试错误处理
func TestDeviceFingerprintCollector_ErrorHandling(t *testing.T) {
	collector := NewDeviceFingerprintCollector()

	// 未初始化时调用方法应该失败
	_, err := collector.CollectFingerprint()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	// 无效配置测试
	err = collector.SetConfiguration(nil)
	assert.Error(t, err)
}

// TestDeviceFingerprintCollector_ConcurrentAccess 测试并发访问
func TestDeviceFingerprintCollector_ConcurrentAccess(t *testing.T) {
	collector := NewDeviceFingerprintCollector()
	err := collector.Initialize("")
	require.NoError(t, err)
	defer collector.Uninitialize()

	// 并发采集指纹
	const goroutines = 10
	results := make(chan *DeviceFingerprint, goroutines)
	errors := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			fp, err := collector.CollectFingerprint()
			if err != nil {
				errors <- err
			} else {
				results <- fp
			}
		}()
	}

	// 收集结果
	for i := 0; i < goroutines; i++ {
		select {
		case fp := <-results:
			assert.NotNil(t, fp)
			assert.NotEmpty(t, fp.DeviceID)
		case err := <-errors:
			t.Errorf("Concurrent access failed: %v", err)
		case <-time.After(10 * time.Second):
			t.Fatal("Timeout waiting for concurrent operations")
		}
	}
}

// BenchmarkDeviceFingerprintCollector_CollectFingerprint 基准测试：指纹采集
func BenchmarkDeviceFingerprintCollector_CollectFingerprint(b *testing.B) {
	collector := NewDeviceFingerprintCollector()
	err := collector.Initialize("")
	require.NoError(b, err)
	defer collector.Uninitialize()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := collector.CollectFingerprint()
		if err != nil {
			b.Fatalf("CollectFingerprint failed: %v", err)
		}
	}
}

// BenchmarkQuickCollectFingerprint 基准测试：快速采集
func BenchmarkQuickCollectFingerprint(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := QuickCollectFingerprint()
		if err != nil {
			b.Fatalf("QuickCollectFingerprint failed: %v", err)
		}
	}
}

// BenchmarkDeviceFingerprintCollector_GenerateHash 基准测试：哈希生成
func BenchmarkDeviceFingerprintCollector_GenerateHash(b *testing.B) {
	collector := NewDeviceFingerprintCollector()
	err := collector.Initialize("")
	require.NoError(b, err)
	defer collector.Uninitialize()

	// 准备测试数据
	fingerprint, err := collector.CollectFingerprint()
	require.NoError(b, err)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := collector.GenerateHash(fingerprint)
		if err != nil {
			b.Fatalf("GenerateHash failed: %v", err)
		}
	}
}

// BenchmarkDeviceFingerprintCollector_CompareFingerprints 基准测试：指纹比较
func BenchmarkDeviceFingerprintCollector_CompareFingerprints(b *testing.B) {
	collector := NewDeviceFingerprintCollector()
	err := collector.Initialize("")
	require.NoError(b, err)
	defer collector.Uninitialize()

	// 准备测试数据
	fp1, err := collector.CollectFingerprint()
	require.NoError(b, err)

	fp2, err := collector.CollectFingerprint()
	require.NoError(b, err)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := collector.CompareFingerprints(fp1, fp2)
		if err != nil {
			b.Fatalf("CompareFingerprints failed: %v", err)
		}
	}
}

// BenchmarkDeviceFingerprintCollector_Serialization 基准测试：序列化
func BenchmarkDeviceFingerprintCollector_Serialization(b *testing.B) {
	collector := NewDeviceFingerprintCollector()
	err := collector.Initialize("")
	require.NoError(b, err)
	defer collector.Uninitialize()

	// 准备测试数据
	fingerprint, err := collector.CollectFingerprint()
	require.NoError(b, err)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		jsonData, err := collector.SerializeToJSON(fingerprint)
		if err != nil {
			b.Fatalf("SerializeToJSON failed: %v", err)
		}

		_, err = collector.DeserializeFromJSON(jsonData)
		if err != nil {
			b.Fatalf("DeserializeFromJSON failed: %v", err)
		}
	}
}

// BenchmarkDeviceFingerprintCollector_ConcurrentCollect 基准测试：并发采集
func BenchmarkDeviceFingerprintCollector_ConcurrentCollect(b *testing.B) {
	collector := NewDeviceFingerprintCollector()
	err := collector.Initialize("")
	require.NoError(b, err)
	defer collector.Uninitialize()

	b.ResetTimer()
	b.ReportAllocs()
	b.SetParallelism(10)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := collector.CollectFingerprint()
			if err != nil {
				b.Fatalf("Concurrent CollectFingerprint failed: %v", err)
			}
		}
	})
}

// TestMain 测试主函数
func TestMain(m *testing.M) {
	// 在所有测试之前执行的设置
	code := m.Run()
	// 在所有测试之后执行的清理
	cleanupTestResources()
	exit(code)
}

// cleanupTestResources 清理测试资源
func cleanupTestResources() {
	// 清理测试过程中创建的资源
}

// exit 退出函数（用于测试）
var exit = func(code int) {
	// 在实际测试中，这里会调用 os.Exit(code)
	// 为了便于测试，这里使用变量
}
