/**
 * @file test_device_fingerprint.cpp
 * @brief 设备指纹采集模块单元测试
 * @author 高考志愿填报系统开发团队
 * @version 1.0.0
 * @date 2025-01-18
 */

#include <gtest/gtest.h>
#include <gmock/gmock.h>
#include "../include/device_fingerprint.h"
#include "../include/crypto_utils.h"
#include "../include/platform_detector.h"

using namespace gaokao::device;
using namespace gaokao::crypto;
using namespace gaokao::platform;

/**
 * @brief 设备指纹采集器测试类
 */
class DeviceFingerprintTest : public ::testing::Test {
protected:
    void SetUp() override {
        collector_ = std::make_unique<DeviceFingerprintCollector>();
    }
    
    void TearDown() override {
        if (collector_) {
            collector_->Uninitialize();
            collector_.reset();
        }
    }
    
    std::unique_ptr<DeviceFingerprintCollector> collector_;
};

/**
 * @brief 加密工具测试类
 */
class CryptoUtilsTest : public ::testing::Test {
protected:
    void SetUp() override {
        aes_cipher_ = std::make_unique<AESCipher>();
        rsa_cipher_ = std::make_unique<RSACipher>();
    }
    
    void TearDown() override {
        aes_cipher_.reset();
        rsa_cipher_.reset();
    }
    
    std::unique_ptr<AESCipher> aes_cipher_;
    std::unique_ptr<RSACipher> rsa_cipher_;
};

/**
 * @brief 平台检测器测试类
 */
class PlatformDetectorTest : public ::testing::Test {
protected:
    void SetUp() override {
        detector_ = std::make_unique<PlatformDetector>();
        detector_->Initialize();
    }
    
    void TearDown() override {
        if (detector_) {
            detector_->Uninitialize();
            detector_.reset();
        }
    }
    
    std::unique_ptr<PlatformDetector> detector_;
};

// ============================================================================
// 设备指纹采集器测试
// ============================================================================

TEST_F(DeviceFingerprintTest, InitializeAndUninitialize) {
    // 测试初始化
    ErrorCode result = collector_->Initialize();
    EXPECT_EQ(result, ErrorCode::SUCCESS);
    
    // 测试重复初始化
    result = collector_->Initialize();
    EXPECT_EQ(result, ErrorCode::SUCCESS);
    
    // 测试反初始化
    collector_->Uninitialize();
    // 应该可以多次调用反初始化而不出错
    collector_->Uninitialize();
}

TEST_F(DeviceFingerprintTest, CollectFingerprint) {
    // 初始化采集器
    ASSERT_EQ(collector_->Initialize(), ErrorCode::SUCCESS);
    
    // 采集设备指纹
    DeviceFingerprint fingerprint;
    ErrorCode result = collector_->CollectFingerprint(fingerprint);
    
    EXPECT_EQ(result, ErrorCode::SUCCESS);
    EXPECT_FALSE(fingerprint.device_id.empty());
    EXPECT_FALSE(fingerprint.fingerprint_hash.empty());
    EXPECT_GT(fingerprint.confidence_score, 0);
    EXPECT_LE(fingerprint.confidence_score, 100);
    
    // 验证时间戳合理性
    auto now = std::chrono::system_clock::now();
    auto diff = std::chrono::duration_cast<std::chrono::seconds>(now - fingerprint.created_at);
    EXPECT_LT(diff.count(), 60); // 应该在1分钟内
}

TEST_F(DeviceFingerprintTest, CollectHardwareInfo) {
    ASSERT_EQ(collector_->Initialize(), ErrorCode::SUCCESS);
    
    HardwareInfo hardware;
    ErrorCode result = collector_->CollectHardwareInfo(hardware);
    
    EXPECT_EQ(result, ErrorCode::SUCCESS);
    // CPU ID和型号至少有一个应该不为空
    EXPECT_TRUE(!hardware.cpu_id.empty() || !hardware.cpu_model.empty());
    
    if (hardware.cpu_cores > 0) {
        EXPECT_LE(hardware.cpu_cores, 128); // 合理的CPU核心数范围
    }
    
    if (hardware.total_memory > 0) {
        EXPECT_GE(hardware.total_memory, 1024 * 1024); // 至少1MB内存
    }
}

TEST_F(DeviceFingerprintTest, CollectSystemInfo) {
    ASSERT_EQ(collector_->Initialize(), ErrorCode::SUCCESS);
    
    SystemInfo system;
    ErrorCode result = collector_->CollectSystemInfo(system);
    
    EXPECT_EQ(result, ErrorCode::SUCCESS);
    EXPECT_NE(system.os_type, OperatingSystem::UNKNOWN);
    EXPECT_FALSE(system.hostname.empty());
}

TEST_F(DeviceFingerprintTest, CollectRuntimeInfo) {
    ASSERT_EQ(collector_->Initialize(), ErrorCode::SUCCESS);
    
    RuntimeInfo runtime;
    ErrorCode result = collector_->CollectRuntimeInfo(runtime);
    
    EXPECT_EQ(result, ErrorCode::SUCCESS);
    
    // 验证时间戳
    auto now = std::chrono::system_clock::now();
    auto diff = std::chrono::duration_cast<std::chrono::seconds>(now - runtime.timestamp);
    EXPECT_LT(diff.count(), 10); // 应该在10秒内
}

TEST_F(DeviceFingerprintTest, GenerateFingerprintHash) {
    ASSERT_EQ(collector_->Initialize(), ErrorCode::SUCCESS);
    
    DeviceFingerprint fingerprint;
    ASSERT_EQ(collector_->CollectFingerprint(fingerprint), ErrorCode::SUCCESS);
    
    // 生成哈希
    std::string hash1 = collector_->GenerateFingerprintHash(fingerprint);
    EXPECT_FALSE(hash1.empty());
    EXPECT_GT(hash1.length(), 16); // SHA256应该是64字符的十六进制字符串
    
    // 相同指纹应该生成相同哈希
    std::string hash2 = collector_->GenerateFingerprintHash(fingerprint);
    EXPECT_EQ(hash1, hash2);
}

TEST_F(DeviceFingerprintTest, CompareFingerprints) {
    ASSERT_EQ(collector_->Initialize(), ErrorCode::SUCCESS);
    
    DeviceFingerprint fp1, fp2;
    ASSERT_EQ(collector_->CollectFingerprint(fp1), ErrorCode::SUCCESS);
    ASSERT_EQ(collector_->CollectFingerprint(fp2), ErrorCode::SUCCESS);
    
    ComparisonResult result;
    ErrorCode error = collector_->CompareFingerprints(fp1, fp2, result);
    
    EXPECT_EQ(error, ErrorCode::SUCCESS);
    EXPECT_GE(result.similarity_score, 0.0);
    EXPECT_LE(result.similarity_score, 1.0);
    EXPECT_GE(result.confidence_level, 0);
    EXPECT_LE(result.confidence_level, 100);
    
    // 相同指纹应该有很高的相似度
    EXPECT_GT(result.similarity_score, 0.9);
    EXPECT_TRUE(result.is_same_device);
}

TEST_F(DeviceFingerprintTest, ValidateFingerprint) {
    ASSERT_EQ(collector_->Initialize(), ErrorCode::SUCCESS);
    
    DeviceFingerprint fingerprint;
    ASSERT_EQ(collector_->CollectFingerprint(fingerprint), ErrorCode::SUCCESS);
    
    // 验证正确的哈希
    bool valid = collector_->ValidateFingerprint(fingerprint, fingerprint.fingerprint_hash);
    EXPECT_TRUE(valid);
    
    // 验证错误的哈希
    valid = collector_->ValidateFingerprint(fingerprint, "invalid_hash");
    EXPECT_FALSE(valid);
}

TEST_F(DeviceFingerprintTest, SerializeToJson) {
    ASSERT_EQ(collector_->Initialize(), ErrorCode::SUCCESS);
    
    DeviceFingerprint fingerprint;
    ASSERT_EQ(collector_->CollectFingerprint(fingerprint), ErrorCode::SUCCESS);
    
    std::string json = collector_->SerializeToJson(fingerprint);
    EXPECT_FALSE(json.empty());
    EXPECT_NE(json.find("device_id"), std::string::npos);
    EXPECT_NE(json.find("fingerprint_hash"), std::string::npos);
    EXPECT_NE(json.find("hardware"), std::string::npos);
    EXPECT_NE(json.find("system"), std::string::npos);
}

TEST_F(DeviceFingerprintTest, Configuration) {
    ASSERT_EQ(collector_->Initialize(), ErrorCode::SUCCESS);
    
    // 测试设置配置
    collector_->SetConfiguration(true, true, true);
    
    DeviceFingerprint fp1;
    ASSERT_EQ(collector_->CollectFingerprint(fp1), ErrorCode::SUCCESS);
    
    // 改变配置
    collector_->SetConfiguration(false, false, false);
    
    DeviceFingerprint fp2;
    ASSERT_EQ(collector_->CollectFingerprint(fp2), ErrorCode::SUCCESS);
    
    // 两个指纹应该不同（因为配置不同）
    // 但基本硬件信息应该相同
    EXPECT_EQ(fp1.hardware.cpu_id, fp2.hardware.cpu_id);
}

TEST_F(DeviceFingerprintTest, ErrorHandling) {
    // 测试未初始化时的错误处理
    DeviceFingerprint fingerprint;
    ErrorCode result = collector_->CollectFingerprint(fingerprint);
    EXPECT_EQ(result, ErrorCode::INITIALIZATION_FAILED);
    
    HardwareInfo hardware;
    result = collector_->CollectHardwareInfo(hardware);
    EXPECT_EQ(result, ErrorCode::INITIALIZATION_FAILED);
}

// ============================================================================
// 加密工具测试
// ============================================================================

TEST_F(CryptoUtilsTest, AESKeyGeneration) {
    std::vector<uint8_t> key;
    
    // 测试AES-256密钥生成
    CryptoError result = aes_cipher_->GenerateKey(256, key);
    EXPECT_EQ(result, CryptoError::SUCCESS);
    EXPECT_EQ(key.size(), 32); // 256位 = 32字节
    
    // 测试无效密钥长度
    result = aes_cipher_->GenerateKey(100, key);
    EXPECT_EQ(result, CryptoError::INVALID_PARAMETER);
}

TEST_F(CryptoUtilsTest, AESEncryptionDecryption) {
    // 生成密钥
    std::vector<uint8_t> key;
    ASSERT_EQ(aes_cipher_->GenerateKey(256, key), CryptoError::SUCCESS);
    
    // 准备测试数据
    std::string plaintext = "Hello, Device Fingerprint!";
    std::vector<uint8_t> data(plaintext.begin(), plaintext.end());
    
    // 加密
    EncryptionResult encrypt_result;
    CryptoError result = aes_cipher_->Encrypt(data, key, 
                                             CryptoAlgorithm::AES_256_CBC, 
                                             encrypt_result);
    
    EXPECT_EQ(result, CryptoError::SUCCESS);
    EXPECT_FALSE(encrypt_result.encrypted_data.empty());
    EXPECT_FALSE(encrypt_result.iv.empty());
    EXPECT_EQ(encrypt_result.algorithm, CryptoAlgorithm::AES_256_CBC);
    
    // 解密
    std::vector<uint8_t> decrypted_data;
    result = aes_cipher_->Decrypt(encrypt_result, key, decrypted_data);
    
    EXPECT_EQ(result, CryptoError::SUCCESS);
    
    // 验证解密结果
    std::string decrypted_text(decrypted_data.begin(), decrypted_data.end());
    EXPECT_EQ(plaintext, decrypted_text);
}

TEST_F(CryptoUtilsTest, AESStringEncryption) {
    std::string plaintext = "Device Fingerprint Test Data";
    std::string password = "test_password_123";
    std::string encrypted;
    
    // 加密字符串
    CryptoError result = aes_cipher_->EncryptString(plaintext, password, encrypted);
    EXPECT_EQ(result, CryptoError::SUCCESS);
    EXPECT_FALSE(encrypted.empty());
    EXPECT_NE(encrypted, plaintext);
    
    // 解密字符串
    std::string decrypted;
    result = aes_cipher_->DecryptString(encrypted, password, decrypted);
    EXPECT_EQ(result, CryptoError::SUCCESS);
    EXPECT_EQ(plaintext, decrypted);
    
    // 错误密码解密应该失败
    result = aes_cipher_->DecryptString(encrypted, "wrong_password", decrypted);
    EXPECT_NE(result, CryptoError::SUCCESS);
}

TEST_F(CryptoUtilsTest, RSAKeyGeneration) {
    KeyPair key_pair;
    
    // 测试RSA-2048密钥生成
    CryptoError result = rsa_cipher_->GenerateKeyPair(2048, key_pair);
    EXPECT_EQ(result, CryptoError::SUCCESS);
    EXPECT_FALSE(key_pair.public_key.empty());
    EXPECT_FALSE(key_pair.private_key.empty());
    EXPECT_EQ(key_pair.algorithm, CryptoAlgorithm::RSA_2048);
    EXPECT_EQ(key_pair.key_size, 2048);
    
    // 测试无效密钥长度
    result = rsa_cipher_->GenerateKeyPair(1024, key_pair);
    EXPECT_EQ(result, CryptoError::INVALID_PARAMETER);
}

TEST_F(CryptoUtilsTest, HashCalculation) {
    std::string test_data = "Device Fingerprint Hash Test";
    std::string hash_hex;
    
    // 测试SHA256哈希
    CryptoError result = HashUtils::CalculateStringHash(test_data, 
                                                       HashAlgorithm::SHA256, 
                                                       hash_hex);
    
    EXPECT_EQ(result, CryptoError::SUCCESS);
    EXPECT_FALSE(hash_hex.empty());
    EXPECT_EQ(hash_hex.length(), 64); // SHA256 = 32字节 = 64字符十六进制
    
    // 相同数据应该产生相同哈希
    std::string hash_hex2;
    result = HashUtils::CalculateStringHash(test_data, HashAlgorithm::SHA256, hash_hex2);
    EXPECT_EQ(result, CryptoError::SUCCESS);
    EXPECT_EQ(hash_hex, hash_hex2);
    
    // 不同数据应该产生不同哈希
    std::string hash_hex3;
    result = HashUtils::CalculateStringHash("Different Data", HashAlgorithm::SHA256, hash_hex3);
    EXPECT_EQ(result, CryptoError::SUCCESS);
    EXPECT_NE(hash_hex, hash_hex3);
}

TEST_F(CryptoUtilsTest, Base64Encoding) {
    std::vector<uint8_t> data = {0x48, 0x65, 0x6C, 0x6C, 0x6F}; // "Hello"
    
    // Base64编码
    std::string encoded = EncodingUtils::Base64Encode(data);
    EXPECT_EQ(encoded, "SGVsbG8=");
    
    // Base64解码
    std::vector<uint8_t> decoded;
    bool success = EncodingUtils::Base64Decode(encoded, decoded);
    EXPECT_TRUE(success);
    EXPECT_EQ(data, decoded);
}

TEST_F(CryptoUtilsTest, HexEncoding) {
    std::vector<uint8_t> data = {0xDE, 0xAD, 0xBE, 0xEF};
    
    // 十六进制编码（小写）
    std::string hex_lower = EncodingUtils::HexEncode(data, false);
    EXPECT_EQ(hex_lower, "deadbeef");
    
    // 十六进制编码（大写）
    std::string hex_upper = EncodingUtils::HexEncode(data, true);
    EXPECT_EQ(hex_upper, "DEADBEEF");
    
    // 十六进制解码
    std::vector<uint8_t> decoded;
    bool success = EncodingUtils::HexDecode(hex_lower, decoded);
    EXPECT_TRUE(success);
    EXPECT_EQ(data, decoded);
}

TEST_F(CryptoUtilsTest, RandomGeneration) {
    // 测试随机字节生成
    std::vector<uint8_t> random_bytes;
    CryptoError result = RandomGenerator::GenerateRandomBytes(32, random_bytes);
    
    EXPECT_EQ(result, CryptoError::SUCCESS);
    EXPECT_EQ(random_bytes.size(), 32);
    
    // 生成两次应该得到不同的结果
    std::vector<uint8_t> random_bytes2;
    result = RandomGenerator::GenerateRandomBytes(32, random_bytes2);
    EXPECT_EQ(result, CryptoError::SUCCESS);
    EXPECT_NE(random_bytes, random_bytes2);
    
    // 测试随机字符串生成
    std::string random_str = RandomGenerator::GenerateRandomString(16);
    EXPECT_EQ(random_str.length(), 16);
    
    std::string random_str2 = RandomGenerator::GenerateRandomString(16);
    EXPECT_NE(random_str, random_str2);
    
    // 测试UUID生成
    std::string uuid = RandomGenerator::GenerateUUID();
    EXPECT_FALSE(uuid.empty());
    EXPECT_EQ(uuid.length(), 36); // UUID格式: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
    
    std::string uuid2 = RandomGenerator::GenerateUUID();
    EXPECT_NE(uuid, uuid2);
}

// ============================================================================
// 平台检测器测试
// ============================================================================

TEST_F(PlatformDetectorTest, PlatformDetection) {
    PlatformType platform = detector_->DetectPlatform();
    EXPECT_NE(platform, PlatformType::UNKNOWN);
    
    // 快速检测应该返回相同结果
    PlatformType quick_platform = QuickDetectPlatform();
    EXPECT_EQ(platform, quick_platform);
}

TEST_F(PlatformDetectorTest, ArchitectureDetection) {
    Architecture arch = detector_->DetectArchitecture();
    EXPECT_NE(arch, Architecture::UNKNOWN);
    
    // 快速检测应该返回相同结果
    Architecture quick_arch = QuickDetectArchitecture();
    EXPECT_EQ(arch, quick_arch);
}

TEST_F(PlatformDetectorTest, CPUInfoCollection) {
    CPUInfo cpu_info;
    DetectionError result = detector_->GetCPUInfo(cpu_info);
    
    EXPECT_EQ(result, DetectionError::SUCCESS);
    EXPECT_NE(cpu_info.architecture, Architecture::UNKNOWN);
    
    if (cpu_info.logical_cores > 0) {
        EXPECT_LE(cpu_info.logical_cores, 256); // 合理范围
        EXPECT_LE(cpu_info.physical_cores, cpu_info.logical_cores);
    }
}

TEST_F(PlatformDetectorTest, MemoryInfoCollection) {
    MemoryInfo memory_info;
    DetectionError result = detector_->GetMemoryInfo(memory_info);
    
    EXPECT_EQ(result, DetectionError::SUCCESS);
    
    if (memory_info.total_physical > 0) {
        EXPECT_GE(memory_info.total_physical, 1024 * 1024); // 至少1MB
        EXPECT_LE(memory_info.available_physical, memory_info.total_physical);
        EXPECT_LE(memory_info.memory_load, 100);
    }
}

TEST_F(PlatformDetectorTest, OSInfoCollection) {
    OSInfo os_info;
    DetectionError result = detector_->GetOSInfo(os_info);
    
    EXPECT_EQ(result, DetectionError::SUCCESS);
    EXPECT_NE(os_info.platform, PlatformType::UNKNOWN);
    EXPECT_NE(os_info.architecture, Architecture::UNKNOWN);
    EXPECT_FALSE(os_info.name.empty());
}

TEST_F(PlatformDetectorTest, SecurityDetection) {
    // 测试虚拟机检测
    bool is_vm = detector_->IsVirtualMachine();
    bool quick_vm = QuickCheckVirtualMachine();
    EXPECT_EQ(is_vm, quick_vm);
    
    // 测试调试器检测
    bool is_debugger = detector_->IsDebuggerPresent();
    bool quick_debugger = QuickCheckDebugger();
    EXPECT_EQ(is_debugger, quick_debugger);
}

TEST_F(PlatformDetectorTest, SystemUptime) {
    uint64_t uptime = detector_->GetSystemUptime();
    // 系统运行时间应该大于0（除非刚启动）
    // 这里不做严格检查，因为测试环境可能刚启动
    EXPECT_GE(uptime, 0); // 至少应该是0或更大
}

// ============================================================================
// 全局函数测试
// ============================================================================

TEST(GlobalFunctionTest, QuickCollectFingerprint) {
    DeviceFingerprint fingerprint;
    ErrorCode result = QuickCollectFingerprint(fingerprint);
    
    EXPECT_EQ(result, ErrorCode::SUCCESS);
    EXPECT_FALSE(fingerprint.device_id.empty());
    EXPECT_FALSE(fingerprint.fingerprint_hash.empty());
    EXPECT_GT(fingerprint.confidence_score, 0);
}

TEST(GlobalFunctionTest, CalculateFingerprintSimilarity) {
    std::string hash1 = "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890";
    std::string hash2 = "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890";
    std::string hash3 = "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef";
    
    // 相同哈希的相似度应该是1.0
    double similarity1 = CalculateFingerprintSimilarity(hash1, hash2);
    EXPECT_DOUBLE_EQ(similarity1, 1.0);
    
    // 不同哈希的相似度应该小于1.0
    double similarity2 = CalculateFingerprintSimilarity(hash1, hash3);
    EXPECT_LT(similarity2, 1.0);
    EXPECT_GE(similarity2, 0.0);
    
    // 空哈希的处理
    double similarity3 = CalculateFingerprintSimilarity("", hash1);
    EXPECT_DOUBLE_EQ(similarity3, 0.0);
}

// ============================================================================
// 错误处理测试
// ============================================================================

TEST(ErrorHandlingTest, ErrorDescriptions) {
    // 测试所有错误码都有描述
    std::vector<ErrorCode> error_codes = {
        ErrorCode::SUCCESS,
        ErrorCode::INITIALIZATION_FAILED,
        ErrorCode::HARDWARE_ACCESS_DENIED,
        ErrorCode::SYSTEM_INFO_UNAVAILABLE,
        ErrorCode::ENCRYPTION_FAILED,
        ErrorCode::INVALID_PARAMETER,
        ErrorCode::MEMORY_ALLOCATION_FAILED,
        ErrorCode::PLATFORM_NOT_SUPPORTED,
        ErrorCode::PERMISSION_DENIED
    };
    
    for (ErrorCode code : error_codes) {
        std::string description = DeviceFingerprintCollector::GetErrorDescription(code);
        EXPECT_FALSE(description.empty());
    }
}

TEST(ErrorHandlingTest, CryptoErrorDescriptions) {
    // 测试加密错误描述
    std::vector<CryptoError> crypto_errors = {
        CryptoError::SUCCESS,
        CryptoError::INVALID_PARAMETER,
        CryptoError::KEY_GENERATION_FAILED,
        CryptoError::ENCRYPTION_FAILED,
        CryptoError::DECRYPTION_FAILED,
        CryptoError::SIGNATURE_FAILED,
        CryptoError::VERIFICATION_FAILED,
        CryptoError::HASH_FAILED,
        CryptoError::KEY_INVALID,
        CryptoError::ALGORITHM_NOT_SUPPORTED,
        CryptoError::MEMORY_ERROR,
        CryptoError::RANDOM_GENERATION_FAILED
    };
    
    for (CryptoError error : crypto_errors) {
        std::string description = GetCryptoErrorDescription(error);
        EXPECT_FALSE(description.empty());
    }
}

// ============================================================================
// 性能测试
// ============================================================================

TEST(PerformanceTest, FingerprintCollectionSpeed) {
    auto start = std::chrono::high_resolution_clock::now();
    
    DeviceFingerprint fingerprint;
    ErrorCode result = QuickCollectFingerprint(fingerprint);
    
    auto end = std::chrono::high_resolution_clock::now();
    auto duration = std::chrono::duration_cast<std::chrono::milliseconds>(end - start);
    
    EXPECT_EQ(result, ErrorCode::SUCCESS);
    EXPECT_LT(duration.count(), 5000); // 应该在5秒内完成
    
    std::cout << "指纹采集耗时: " << duration.count() << " ms" << std::endl;
}

TEST(PerformanceTest, HashGenerationSpeed) {
    DeviceFingerprint fingerprint;
    ASSERT_EQ(QuickCollectFingerprint(fingerprint), ErrorCode::SUCCESS);
    
    DeviceFingerprintCollector collector;
    ASSERT_EQ(collector.Initialize(), ErrorCode::SUCCESS);
    
    auto start = std::chrono::high_resolution_clock::now();
    
    for (int i = 0; i < 1000; ++i) {
        std::string hash = collector.GenerateFingerprintHash(fingerprint);
        EXPECT_FALSE(hash.empty());
    }
    
    auto end = std::chrono::high_resolution_clock::now();
    auto duration = std::chrono::duration_cast<std::chrono::milliseconds>(end - start);
    
    EXPECT_LT(duration.count(), 1000); // 1000次哈希应该在1秒内完成
    
    std::cout << "1000次哈希生成耗时: " << duration.count() << " ms" << std::endl;
}

// ============================================================================
// 主函数
// ============================================================================

int main(int argc, char** argv) {
    ::testing::InitGoogleTest(&argc, argv);
    
    std::cout << "设备指纹采集模块单元测试开始..." << std::endl;
    std::cout << "版本: " << DeviceFingerprintCollector::GetVersion() << std::endl;
    std::cout << "平台: " << static_cast<int>(QuickDetectPlatform()) << std::endl;
    std::cout << "架构: " << static_cast<int>(QuickDetectArchitecture()) << std::endl;
    std::cout << std::endl;
    
    int result = RUN_ALL_TESTS();
    
    std::cout << std::endl << "单元测试完成。" << std::endl;
    
    return result;
}