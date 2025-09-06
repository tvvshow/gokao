/**
 * @file c_interface.cpp
 * @brief C接口实现，供Go语言通过CGO调用
 * @author 高考志愿填报系统开发团队
 * @version 1.0.0
 * @date 2025-01-18
 */

#include "device_fingerprint.h"
#include "crypto_utils.h"
#include "platform_detector.h"

#include <cstring>
#include <cstdlib>
#include <memory>

extern "C" {

// 错误码定义(C风格)
typedef enum {
    C_SUCCESS = 0,
    C_ERROR_INIT_FAILED = 1001,
    C_ERROR_INVALID_PARAM = 1002,
    C_ERROR_MEMORY_ALLOC = 1003,
    C_ERROR_HARDWARE_ACCESS = 1004,
    C_ERROR_SYSTEM_INFO = 1005,
    C_ERROR_ENCRYPTION = 1006,
    C_ERROR_UNKNOWN = 9999
} CErrorCode;

// C风格设备指纹结构体
typedef struct {
    char device_id[64];             ///< 设备唯一标识
    char device_type[32];           ///< 设备类型
    char cpu_id[64];                ///< CPU标识符
    char cpu_model[128];            ///< CPU型号
    unsigned int cpu_cores;         ///< CPU核心数
    unsigned long long total_memory; ///< 总内存大小(字节)
    char motherboard_serial[64];    ///< 主板序列号
    char os_type[32];               ///< 操作系统类型
    char os_version[64];            ///< 操作系统版本
    char hostname[64];              ///< 主机名
    char username[64];              ///< 用户名
    char screen_resolution[32];     ///< 屏幕分辨率
    char fingerprint_hash[128];     ///< 指纹哈希值
    unsigned int confidence_score;  ///< 置信度分数(0-100)
    char error_message[256];        ///< 错误信息
} CDeviceFingerprint;

// C风格配置结构体
typedef struct {
    int collect_sensitive_info;     ///< 是否采集敏感信息(1是/0否)
    int enable_encryption;          ///< 是否启用加密(1是/0否)
    int enable_signature;           ///< 是否启用签名(1是/0否)
    char encryption_key[64];        ///< 加密密钥
    int timeout_seconds;            ///< 超时时间(秒)
} CConfiguration;

// C风格性能统计结构体
typedef struct {
    unsigned long long collect_time_us;    ///< 采集耗时(微秒)
    unsigned long long hash_time_us;       ///< 哈希计算耗时(微秒)
    unsigned long long encryption_time_us; ///< 加密耗时(微秒)
    unsigned int total_calls;              ///< 总调用次数
    unsigned int success_calls;            ///< 成功调用次数
    unsigned int error_calls;              ///< 错误调用次数
} CPerformanceStats;

// 全局采集器实例
static std::unique_ptr<gaokao::device::DeviceFingerprintCollector> g_collector = nullptr;

// 内部辅助函数
static CErrorCode ConvertErrorCode(gaokao::device::ErrorCode error) {
    switch (error) {
        case gaokao::device::ErrorCode::SUCCESS:
            return C_SUCCESS;
        case gaokao::device::ErrorCode::INITIALIZATION_FAILED:
            return C_ERROR_INIT_FAILED;
        case gaokao::device::ErrorCode::INVALID_PARAMETER:
            return C_ERROR_INVALID_PARAM;
        case gaokao::device::ErrorCode::MEMORY_ALLOCATION_FAILED:
            return C_ERROR_MEMORY_ALLOC;
        case gaokao::device::ErrorCode::HARDWARE_ACCESS_DENIED:
            return C_ERROR_HARDWARE_ACCESS;
        case gaokao::device::ErrorCode::SYSTEM_INFO_UNAVAILABLE:
            return C_ERROR_SYSTEM_INFO;
        case gaokao::device::ErrorCode::ENCRYPTION_FAILED:
            return C_ERROR_ENCRYPTION;
        default:
            return C_ERROR_UNKNOWN;
    }
}

static void ConvertDeviceTypeToC(gaokao::device::DeviceType type, char* buffer, size_t size) {
    const char* type_str = "unknown";
    switch (type) {
        case gaokao::device::DeviceType::DESKTOP:
            type_str = "desktop";
            break;
        case gaokao::device::DeviceType::LAPTOP:
            type_str = "laptop";
            break;
        case gaokao::device::DeviceType::TABLET:
            type_str = "tablet";
            break;
        case gaokao::device::DeviceType::MOBILE:
            type_str = "mobile";
            break;
        case gaokao::device::DeviceType::SERVER:
            type_str = "server";
            break;
        default:
            type_str = "unknown";
            break;
    }
    
    strncpy(buffer, type_str, size - 1);
    buffer[size - 1] = '\0';
}

static gaokao::device::DeviceType ConvertDeviceTypeFromC(const char* type_str) {
    if (strcmp(type_str, "desktop") == 0) {
        return gaokao::device::DeviceType::DESKTOP;
    } else if (strcmp(type_str, "laptop") == 0) {
        return gaokao::device::DeviceType::LAPTOP;
    } else if (strcmp(type_str, "tablet") == 0) {
        return gaokao::device::DeviceType::TABLET;
    } else if (strcmp(type_str, "mobile") == 0) {
        return gaokao::device::DeviceType::MOBILE;
    } else if (strcmp(type_str, "server") == 0) {
        return gaokao::device::DeviceType::SERVER;
    }
    return gaokao::device::DeviceType::UNKNOWN;
}

static void ConvertOperatingSystemToC(gaokao::device::OperatingSystem os, char* buffer, size_t size) {
    const char* os_str = "unknown";
    switch (os) {
        case gaokao::device::OperatingSystem::WINDOWS:
            os_str = "windows";
            break;
        case gaokao::device::OperatingSystem::LINUX:
            os_str = "linux";
            break;
        case gaokao::device::OperatingSystem::MACOS:
            os_str = "macos";
            break;
        case gaokao::device::OperatingSystem::ANDROID:
            os_str = "android";
            break;
        case gaokao::device::OperatingSystem::IOS:
            os_str = "ios";
            break;
        default:
            os_str = "unknown";
            break;
    }
    
    strncpy(buffer, os_str, size - 1);
    buffer[size - 1] = '\0';
}

static gaokao::device::OperatingSystem ConvertOperatingSystemFromC(const char* os_str) {
    if (strcmp(os_str, "windows") == 0) {
        return gaokao::device::OperatingSystem::WINDOWS;
    } else if (strcmp(os_str, "linux") == 0) {
        return gaokao::device::OperatingSystem::LINUX;
    } else if (strcmp(os_str, "macos") == 0) {
        return gaokao::device::OperatingSystem::MACOS;
    } else if (strcmp(os_str, "android") == 0) {
        return gaokao::device::OperatingSystem::ANDROID;
    } else if (strcmp(os_str, "ios") == 0) {
        return gaokao::device::OperatingSystem::IOS;
    }
    return gaokao::device::OperatingSystem::UNKNOWN;
}

static void ConvertToFingerprint(const gaokao::device::DeviceFingerprint& cpp_fingerprint, 
                                CDeviceFingerprint* c_fingerprint) {
    // 设备ID
    strncpy(c_fingerprint->device_id, cpp_fingerprint.device_id.c_str(), sizeof(c_fingerprint->device_id) - 1);
    c_fingerprint->device_id[sizeof(c_fingerprint->device_id) - 1] = '\0';
    
    // 设备类型
    ConvertDeviceTypeToC(cpp_fingerprint.device_type, c_fingerprint->device_type, sizeof(c_fingerprint->device_type));
    
    // 硬件信息
    strncpy(c_fingerprint->cpu_id, cpp_fingerprint.hardware.cpu_id.c_str(), sizeof(c_fingerprint->cpu_id) - 1);
    c_fingerprint->cpu_id[sizeof(c_fingerprint->cpu_id) - 1] = '\0';
    
    strncpy(c_fingerprint->cpu_model, cpp_fingerprint.hardware.cpu_model.c_str(), sizeof(c_fingerprint->cpu_model) - 1);
    c_fingerprint->cpu_model[sizeof(c_fingerprint->cpu_model) - 1] = '\0';
    
    c_fingerprint->cpu_cores = cpp_fingerprint.hardware.cpu_cores;
    c_fingerprint->total_memory = cpp_fingerprint.hardware.total_memory;
    
    strncpy(c_fingerprint->motherboard_serial, cpp_fingerprint.hardware.motherboard_serial.c_str(), sizeof(c_fingerprint->motherboard_serial) - 1);
    c_fingerprint->motherboard_serial[sizeof(c_fingerprint->motherboard_serial) - 1] = '\0';
    
    // 系统信息
    ConvertOperatingSystemToC(cpp_fingerprint.system.os_type, c_fingerprint->os_type, sizeof(c_fingerprint->os_type));
    
    strncpy(c_fingerprint->os_version, cpp_fingerprint.system.os_version.c_str(), sizeof(c_fingerprint->os_version) - 1);
    c_fingerprint->os_version[sizeof(c_fingerprint->os_version) - 1] = '\0';
    
    strncpy(c_fingerprint->hostname, cpp_fingerprint.system.hostname.c_str(), sizeof(c_fingerprint->hostname) - 1);
    c_fingerprint->hostname[sizeof(c_fingerprint->hostname) - 1] = '\0';
    
    strncpy(c_fingerprint->username, cpp_fingerprint.system.username.c_str(), sizeof(c_fingerprint->username) - 1);
    c_fingerprint->username[sizeof(c_fingerprint->username) - 1] = '\0';
    
    // 运行时信息
    strncpy(c_fingerprint->screen_resolution, cpp_fingerprint.runtime.screen_resolution.c_str(), sizeof(c_fingerprint->screen_resolution) - 1);
    c_fingerprint->screen_resolution[sizeof(c_fingerprint->screen_resolution) - 1] = '\0';
    
    // 指纹哈希
    strncpy(c_fingerprint->fingerprint_hash, cpp_fingerprint.fingerprint_hash.c_str(), sizeof(c_fingerprint->fingerprint_hash) - 1);
    c_fingerprint->fingerprint_hash[sizeof(c_fingerprint->fingerprint_hash) - 1] = '\0';
    
    // 置信度分数
    c_fingerprint->confidence_score = cpp_fingerprint.confidence_score;
    
    // 错误信息（这里清空，因为DeviceFingerprint结构体中没有错误信息字段）
    c_fingerprint->error_message[0] = '\0';
}

static void ConvertFromFingerprint(const CDeviceFingerprint& c_fingerprint, 
                                  gaokao::device::DeviceFingerprint& cpp_fingerprint) {
    // 设备ID
    cpp_fingerprint.device_id = std::string(c_fingerprint.device_id);
    
    // 设备类型
    cpp_fingerprint.device_type = ConvertDeviceTypeFromC(c_fingerprint.device_type);
    
    // 硬件信息
    cpp_fingerprint.hardware.cpu_id = std::string(c_fingerprint.cpu_id);
    cpp_fingerprint.hardware.cpu_model = std::string(c_fingerprint.cpu_model);
    cpp_fingerprint.hardware.cpu_cores = c_fingerprint.cpu_cores;
    cpp_fingerprint.hardware.total_memory = c_fingerprint.total_memory;
    cpp_fingerprint.hardware.motherboard_serial = std::string(c_fingerprint.motherboard_serial);
    
    // 系统信息
    cpp_fingerprint.system.os_type = ConvertOperatingSystemFromC(c_fingerprint.os_type);
    cpp_fingerprint.system.os_version = std::string(c_fingerprint.os_version);
    cpp_fingerprint.system.hostname = std::string(c_fingerprint.hostname);
    cpp_fingerprint.system.username = std::string(c_fingerprint.username);
    
    // 运行时信息
    cpp_fingerprint.runtime.screen_resolution = std::string(c_fingerprint.screen_resolution);
    
    // 指纹哈希
    cpp_fingerprint.fingerprint_hash = std::string(c_fingerprint.fingerprint_hash);
    
    // 置信度分数
    cpp_fingerprint.confidence_score = c_fingerprint.confidence_score;
}

// =============================================================================
// 核心功能接口实现
// =============================================================================

CErrorCode DeviceFingerprint_Initialize(const char* config_path) {
    try {
        if (g_collector) {
            g_collector.reset();
        }
        
        g_collector = std::make_unique<gaokao::device::DeviceFingerprintCollector>();
        
        std::string config_path_str;
        if (config_path) {
            config_path_str = std::string(config_path);
        }
        
        gaokao::device::ErrorCode result = g_collector->Initialize(config_path_str);
        return ConvertErrorCode(result);
    } catch (const std::exception& e) {
        return C_ERROR_UNKNOWN;
    }
}

void DeviceFingerprint_Uninitialize(void) {
    if (g_collector) {
        g_collector->Uninitialize();
        g_collector.reset();
    }
}

CErrorCode DeviceFingerprint_Collect(CDeviceFingerprint* fingerprint) {
    if (!g_collector || !fingerprint) {
        return C_ERROR_INVALID_PARAM;
    }
    
    try {
        gaokao::device::DeviceFingerprint cpp_fingerprint;
        gaokao::device::ErrorCode result = g_collector->CollectFingerprint(cpp_fingerprint);
        
        if (result == gaokao::device::ErrorCode::SUCCESS) {
            ConvertToFingerprint(cpp_fingerprint, fingerprint);
        }
        
        return ConvertErrorCode(result);
    } catch (const std::exception& e) {
        return C_ERROR_UNKNOWN;
    }
}

CErrorCode DeviceFingerprint_QuickCollect(CDeviceFingerprint* fingerprint) {
    if (!fingerprint) {
        return C_ERROR_INVALID_PARAM;
    }
    
    try {
        gaokao::device::DeviceFingerprint cpp_fingerprint;
        gaokao::device::ErrorCode result = gaokao::device::QuickCollectFingerprint(cpp_fingerprint);
        
        if (result == gaokao::device::ErrorCode::SUCCESS) {
            ConvertToFingerprint(cpp_fingerprint, fingerprint);
        }
        
        return ConvertErrorCode(result);
    } catch (const std::exception& e) {
        return C_ERROR_UNKNOWN;
    }
}

// =============================================================================
// 配置和管理接口实现
// =============================================================================

CErrorCode DeviceFingerprint_SetConfiguration(const CConfiguration* config) {
    if (!g_collector || !config) {
        return C_ERROR_INVALID_PARAM;
    }
    
    try {
        g_collector->SetConfiguration(
            config->collect_sensitive_info != 0,
            config->enable_encryption != 0,
            config->enable_signature != 0
        );
        return C_SUCCESS;
    } catch (const std::exception& e) {
        return C_ERROR_UNKNOWN;
    }
}

CErrorCode DeviceFingerprint_GetConfiguration(CConfiguration* config) {
    // 注意：当前DeviceFingerprintCollector没有提供获取配置的接口
    // 这里返回默认配置
    if (!config) {
        return C_ERROR_INVALID_PARAM;
    }
    
    config->collect_sensitive_info = 1;
    config->enable_encryption = 1;
    config->enable_signature = 1;
    config->encryption_key[0] = '\0';
    config->timeout_seconds = 30;
    
    return C_SUCCESS;
}

// =============================================================================
// 哈希和比较接口实现
// =============================================================================

CErrorCode DeviceFingerprint_GenerateHash(const CDeviceFingerprint* fingerprint,
                                         char* hash_buffer,
                                         size_t buffer_size) {
    if (!g_collector || !fingerprint || !hash_buffer || buffer_size == 0) {
        return C_ERROR_INVALID_PARAM;
    }
    
    try {
        gaokao::device::DeviceFingerprint cpp_fingerprint;
        ConvertFromFingerprint(*fingerprint, cpp_fingerprint);
        
        std::string hash = g_collector->GenerateFingerprintHash(cpp_fingerprint);
        
        if (hash.length() >= buffer_size) {
            return C_ERROR_MEMORY_ALLOC;
        }
        strncpy(hash_buffer, hash.c_str(), buffer_size - 1);
        hash_buffer[buffer_size - 1] = '\0';
        
        return C_SUCCESS;
    } catch (const std::exception& e) {
        return C_ERROR_UNKNOWN;
    }
}

CErrorCode DeviceFingerprint_Compare(const CDeviceFingerprint* fingerprint1,
                                    const CDeviceFingerprint* fingerprint2,
                                    double* similarity_score,
                                    int* is_same_device) {
    if (!g_collector || !fingerprint1 || !fingerprint2 || !similarity_score || !is_same_device) {
        return C_ERROR_INVALID_PARAM;
    }
    
    try {
        gaokao::device::DeviceFingerprint cpp_fingerprint1, cpp_fingerprint2;
        ConvertFromFingerprint(*fingerprint1, cpp_fingerprint1);
        ConvertFromFingerprint(*fingerprint2, cpp_fingerprint2);
        
        gaokao::device::ComparisonResult result;
        gaokao::device::ErrorCode error = g_collector->CompareFingerprints(cpp_fingerprint1, cpp_fingerprint2, result);
        
        if (error == gaokao::device::ErrorCode::SUCCESS) {
            *similarity_score = result.similarity_score;
            *is_same_device = result.is_same_device ? 1 : 0;
        }
        
        return ConvertErrorCode(error);
    } catch (const std::exception& e) {
        return C_ERROR_UNKNOWN;
    }
}

CErrorCode DeviceFingerprint_Validate(const CDeviceFingerprint* fingerprint,
                                     const char* reference_hash,
                                     int* is_valid) {
    if (!g_collector || !fingerprint || !reference_hash || !is_valid) {
        return C_ERROR_INVALID_PARAM;
    }
    
    try {
        gaokao::device::DeviceFingerprint cpp_fingerprint;
        ConvertFromFingerprint(*fingerprint, cpp_fingerprint);
        
        bool valid = g_collector->ValidateFingerprint(cpp_fingerprint, std::string(reference_hash));
        *is_valid = valid ? 1 : 0;
        
        return C_SUCCESS;
    } catch (const std::exception& e) {
        return C_ERROR_UNKNOWN;
    }
}

// =============================================================================
// 加密和签名接口实现
// =============================================================================

CErrorCode DeviceFingerprint_Encrypt(const char* data,
                                    size_t data_size,
                                    const char* key,
                                    char* encrypted_buffer,
                                    size_t buffer_size,
                                    size_t* actual_size) {
    if (!data || data_size == 0 || !key || !encrypted_buffer || buffer_size == 0 || !actual_size) {
        return C_ERROR_INVALID_PARAM;
    }
    
    try {
        // 创建AES加密器实例
        gaokao::crypto::AESCipher aes_cipher;
        
        // 将输入数据转换为字节数组
        std::vector<uint8_t> plaintext(data, data + data_size);
        std::string password(key);
        
        // 执行加密
        std::string encrypted_base64;
        gaokao::crypto::CryptoError result = aes_cipher.EncryptString(
            std::string(data, data_size), password, encrypted_base64);
        
        if (result == gaokao::crypto::CryptoError::SUCCESS) {
            if (encrypted_base64.length() >= buffer_size) {
                return C_ERROR_MEMORY_ALLOC;
            }
            
            memcpy(encrypted_buffer, encrypted_base64.c_str(), encrypted_base64.length());
            *actual_size = encrypted_base64.length();
        }
        
        // 转换加密错误码为设备指纹错误码
        switch (result) {
            case gaokao::crypto::CryptoError::SUCCESS:
                return C_SUCCESS;
            case gaokao::crypto::CryptoError::INVALID_PARAMETER:
                return C_ERROR_INVALID_PARAM;
            case gaokao::crypto::CryptoError::ENCRYPTION_FAILED:
                return C_ERROR_ENCRYPTION;
            default:
                return C_ERROR_UNKNOWN;
        }
    } catch (const std::exception& e) {
        return C_ERROR_UNKNOWN;
    }
}

CErrorCode DeviceFingerprint_Decrypt(const char* encrypted_data,
                                    size_t data_size,
                                    const char* key,
                                    char* decrypted_buffer,
                                    size_t buffer_size,
                                    size_t* actual_size) {
    if (!encrypted_data || data_size == 0 || !key || !decrypted_buffer || buffer_size == 0 || !actual_size) {
        return C_ERROR_INVALID_PARAM;
    }
    
    try {
        // 创建AES解密器实例
        gaokao::crypto::AESCipher aes_cipher;
        
        // 将加密数据转换为字符串
        std::string encrypted_base64(encrypted_data, data_size);
        std::string password(key);
        
        // 执行解密
        std::string plaintext;
        gaokao::crypto::CryptoError result = aes_cipher.DecryptString(
            encrypted_base64, password, plaintext);
        
        if (result == gaokao::crypto::CryptoError::SUCCESS) {
            if (plaintext.length() >= buffer_size) {
                return C_ERROR_MEMORY_ALLOC;
            }
            
            memcpy(decrypted_buffer, plaintext.c_str(), plaintext.length());
            *actual_size = plaintext.length();
        }
        
        // 转换解密错误码为设备指纹错误码
        switch (result) {
            case gaokao::crypto::CryptoError::SUCCESS:
                return C_SUCCESS;
            case gaokao::crypto::CryptoError::INVALID_PARAMETER:
                return C_ERROR_INVALID_PARAM;
            case gaokao::crypto::CryptoError::DECRYPTION_FAILED:
                return C_ERROR_ENCRYPTION;
            default:
                return C_ERROR_UNKNOWN;
        }
    } catch (const std::exception& e) {
        return C_ERROR_UNKNOWN;
    }
}

CErrorCode DeviceFingerprint_Sign(const char* data,
                                 size_t data_size,
                                 const char* private_key,
                                 char* signature_buffer,
                                 size_t buffer_size,
                                 size_t* actual_size) {
    if (!data || data_size == 0 || !private_key || !signature_buffer || buffer_size == 0 || !actual_size) {
        return C_ERROR_INVALID_PARAM;
    }
    
    try {
        // 创建RSA签名器实例
        gaokao::crypto::RSACipher rsa_cipher;
        
        // 将私钥从PEM格式导入
        std::vector<uint8_t> private_key_data;
        gaokao::crypto::CryptoError import_result = rsa_cipher.ImportPEMKey(std::string(private_key), private_key_data);
        if (import_result != gaokao::crypto::CryptoError::SUCCESS) {
            // 转换导入错误码为设备指纹错误码
            switch (import_result) {
                case gaokao::crypto::CryptoError::INVALID_PARAMETER:
                    return C_ERROR_INVALID_PARAM;
                default:
                    return C_ERROR_ENCRYPTION;
            }
        }
        
        // 将输入数据转换为字节数组
        std::vector<uint8_t> data_bytes(data, data + data_size);
        
        // 执行签名
        gaokao::crypto::SignatureResult signature_result;
        gaokao::crypto::CryptoError sign_result = rsa_cipher.Sign(
            data_bytes, private_key_data, gaokao::crypto::HashAlgorithm::SHA256, signature_result);
        
        if (sign_result == gaokao::crypto::CryptoError::SUCCESS) {
            // 将签名数据编码为Base64
            std::string signature_base64 = gaokao::crypto::EncodingUtils::Base64Encode(signature_result.signature);
            
            if (signature_base64.length() >= buffer_size) {
                return C_ERROR_MEMORY_ALLOC;
            }
            
            memcpy(signature_buffer, signature_base64.c_str(), signature_base64.length());
            *actual_size = signature_base64.length();
        }
        
        // 转换签名错误码为设备指纹错误码
        switch (sign_result) {
            case gaokao::crypto::CryptoError::SUCCESS:
                return C_SUCCESS;
            case gaokao::crypto::CryptoError::INVALID_PARAMETER:
                return C_ERROR_INVALID_PARAM;
            case gaokao::crypto::CryptoError::SIGNATURE_FAILED:
                return C_ERROR_ENCRYPTION;
            default:
                return C_ERROR_UNKNOWN;
        }
    } catch (const std::exception& e) {
        return C_ERROR_UNKNOWN;
    }
}

CErrorCode DeviceFingerprint_VerifySignature(const char* data,
                                            size_t data_size,
                                            const char* signature,
                                            size_t signature_size,
                                            const char* public_key,
                                            int* is_valid) {
    if (!data || data_size == 0 || !signature || signature_size == 0 || !public_key || !is_valid) {
        return C_ERROR_INVALID_PARAM;
    }
    
    try {
        // 创建RSA签名验证器实例
        gaokao::crypto::RSACipher rsa_cipher;
        
        // 将公钥从PEM格式导入
        std::vector<uint8_t> public_key_data;
        gaokao::crypto::CryptoError import_result = rsa_cipher.ImportPEMKey(std::string(public_key), public_key_data);
        if (import_result != gaokao::crypto::CryptoError::SUCCESS) {
            // 转换导入错误码为设备指纹错误码
            switch (import_result) {
                case gaokao::crypto::CryptoError::INVALID_PARAMETER:
                    return C_ERROR_INVALID_PARAM;
                default:
                    return C_ERROR_ENCRYPTION;
            }
        }
        
        // 将输入数据转换为字节数组
        std::vector<uint8_t> data_bytes(data, data + data_size);
        
        // 将签名从Base64解码
        std::vector<uint8_t> signature_bytes;
        if (!gaokao::crypto::EncodingUtils::Base64Decode(std::string(signature, signature_size), signature_bytes)) {
            return C_ERROR_INVALID_PARAM;
        }
        
        // 构造签名结果
        gaokao::crypto::SignatureResult signature_result;
        signature_result.signature = signature_bytes;
        signature_result.hash_algorithm = gaokao::crypto::HashAlgorithm::SHA256;
        signature_result.sign_algorithm = gaokao::crypto::CryptoAlgorithm::RSA_2048;
        
        // 执行签名验证
        bool valid = rsa_cipher.VerifySignature(data_bytes, signature_result, public_key_data);
        *is_valid = valid ? 1 : 0;
        
        return C_SUCCESS;
    } catch (const std::exception& e) {
        return C_ERROR_UNKNOWN;
    }
}

// =============================================================================
// 许可证验证接口实现
// =============================================================================

// 注意：许可证功能在当前版本中未实现，这里提供空实现
CErrorCode DeviceFingerprint_ValidateLicense(const char* license_data,
                                            const char* device_id,
                                            int* is_valid,
                                            long long* expires_at) {
    if (!license_data || !device_id || !is_valid || !expires_at) {
        return C_ERROR_INVALID_PARAM;
    }
    
    // 空实现，返回默认值
    *is_valid = 1;
    *expires_at = 0;
    
    return C_SUCCESS;
}

CErrorCode DeviceFingerprint_GenerateLicense(const char* device_id,
                                            long long expires_at,
                                            const char* private_key,
                                            char* license_buffer,
                                            size_t buffer_size) {
    if (!device_id || !private_key || !license_buffer || buffer_size == 0) {
        return C_ERROR_INVALID_PARAM;
    }
    
    // 空实现，返回默认值
    std::string license = "license_placeholder";
    if (license.length() >= buffer_size) {
        return C_ERROR_MEMORY_ALLOC;
    }
    strncpy(license_buffer, license.c_str(), buffer_size - 1);
    license_buffer[buffer_size - 1] = '\0';
    
    return C_SUCCESS;
}

// =============================================================================
// 序列化接口实现
// =============================================================================

CErrorCode DeviceFingerprint_SerializeToJson(const CDeviceFingerprint* fingerprint,
                                            char* json_buffer,
                                            size_t buffer_size) {
    if (!g_collector || !fingerprint || !json_buffer || buffer_size == 0) {
        return C_ERROR_INVALID_PARAM;
    }
    
    try {
        gaokao::device::DeviceFingerprint cpp_fingerprint;
        ConvertFromFingerprint(*fingerprint, cpp_fingerprint);
        
        std::string json_str = g_collector->SerializeToJson(cpp_fingerprint);
        
        if (json_str.length() >= buffer_size) {
            return C_ERROR_MEMORY_ALLOC;
        }
        strncpy(json_buffer, json_str.c_str(), buffer_size - 1);
        json_buffer[buffer_size - 1] = '\0';
        
        return C_SUCCESS;
    } catch (const std::exception& e) {
        return C_ERROR_UNKNOWN;
    }
}

CErrorCode DeviceFingerprint_DeserializeFromJson(const char* json_data,
                                                CDeviceFingerprint* fingerprint) {
    if (!g_collector || !json_data || !fingerprint) {
        return C_ERROR_INVALID_PARAM;
    }
    
    try {
        gaokao::device::DeviceFingerprint cpp_fingerprint;
        gaokao::device::ErrorCode result = g_collector->DeserializeFromJson(std::string(json_data), cpp_fingerprint);
        
        if (result == gaokao::device::ErrorCode::SUCCESS) {
            ConvertToFingerprint(cpp_fingerprint, fingerprint);
        }
        
        return ConvertErrorCode(result);
    } catch (const std::exception& e) {
        return C_ERROR_UNKNOWN;
    }
}

// =============================================================================
// 安全检测接口实现
// =============================================================================

CErrorCode DeviceFingerprint_IsDebuggerPresent(int* is_debugger_present) {
    if (!g_collector || !is_debugger_present) {
        return C_ERROR_INVALID_PARAM;
    }
    
    try {
        bool present = g_collector->IsDebuggerPresent();
        *is_debugger_present = present ? 1 : 0;
        return C_SUCCESS;
    } catch (const std::exception& e) {
        return C_ERROR_UNKNOWN;
    }
}

CErrorCode DeviceFingerprint_IsVirtualMachine(int* is_virtual_machine) {
    if (!g_collector || !is_virtual_machine) {
        return C_ERROR_INVALID_PARAM;
    }
    
    try {
        bool is_vm = g_collector->IsRunningInVirtualMachine();
        *is_virtual_machine = is_vm ? 1 : 0;
        return C_SUCCESS;
    } catch (const std::exception& e) {
        return C_ERROR_UNKNOWN;
    }
}

CErrorCode DeviceFingerprint_CheckSecurity(int* security_level,
                                          char* risk_factors,
                                          size_t buffer_size) {
    // 注意：当前版本未实现完整的安全检查功能
    if (!security_level || !risk_factors || buffer_size == 0) {
        return C_ERROR_INVALID_PARAM;
    }
    
    // 返回默认值
    *security_level = 80;
    strncpy(risk_factors, "normal", buffer_size - 1);
    risk_factors[buffer_size - 1] = '\0';
    
    return C_SUCCESS;
}

// =============================================================================
// 性能监控接口实现
// =============================================================================

// 注意：当前版本未实现性能监控功能
CErrorCode DeviceFingerprint_GetPerformanceStats(CPerformanceStats* stats) {
    if (!stats) {
        return C_ERROR_INVALID_PARAM;
    }
    
    // 返回默认值
    stats->collect_time_us = 0;
    stats->hash_time_us = 0;
    stats->encryption_time_us = 0;
    stats->total_calls = 0;
    stats->success_calls = 0;
    stats->error_calls = 0;
    
    return C_SUCCESS;
}

CErrorCode DeviceFingerprint_ResetPerformanceStats(void) {
    // 空实现
    return C_SUCCESS;
}

CErrorCode DeviceFingerprint_SetPerformanceMonitoring(int enable) {
    // 空实现
    return C_SUCCESS;
}

// =============================================================================
// 工具函数接口实现
// =============================================================================

CErrorCode DeviceFingerprint_GetVersion(char* version_buffer, size_t buffer_size) {
    if (!version_buffer || buffer_size == 0) {
        return C_ERROR_INVALID_PARAM;
    }
    
    try {
        std::string version = gaokao::device::DeviceFingerprintCollector::GetVersion();
        if (version.length() >= buffer_size) {
            return C_ERROR_MEMORY_ALLOC;
        }
        strncpy(version_buffer, version.c_str(), buffer_size - 1);
        version_buffer[buffer_size - 1] = '\0';
        return C_SUCCESS;
    } catch (const std::exception& e) {
        return C_ERROR_UNKNOWN;
    }
}

CErrorCode DeviceFingerprint_GetErrorDescription(CErrorCode error_code,
                                                char* error_buffer,
                                                size_t buffer_size) {
    if (!error_buffer || buffer_size == 0) {
        return C_ERROR_INVALID_PARAM;
    }
    
    try {
        // 错误描述映射
        const char* description = "Unknown error";
        switch (error_code) {
            case C_SUCCESS:
                description = "Success";
                break;
            case C_ERROR_INIT_FAILED:
                description = "Initialization failed";
                break;
            case C_ERROR_INVALID_PARAM:
                description = "Invalid parameter";
                break;
            case C_ERROR_MEMORY_ALLOC:
                description = "Memory allocation failed";
                break;
            case C_ERROR_HARDWARE_ACCESS:
                description = "Hardware access denied";
                break;
            case C_ERROR_SYSTEM_INFO:
                description = "System information unavailable";
                break;
            case C_ERROR_ENCRYPTION:
                description = "Encryption operation failed";
                break;
            case C_ERROR_UNKNOWN:
                description = "Unknown error";
                break;
            default:
                break;
        }
        
        if (strlen(description) >= buffer_size) {
            return C_ERROR_MEMORY_ALLOC;
        }
        strncpy(error_buffer, description, buffer_size - 1);
        error_buffer[buffer_size - 1] = '\0';
        return C_SUCCESS;
    } catch (const std::exception& e) {
        return C_ERROR_UNKNOWN;
    }
}

CErrorCode DeviceFingerprint_IsInitialized(int* is_initialized) {
    if (!is_initialized) {
        return C_ERROR_INVALID_PARAM;
    }
    
    *is_initialized = (g_collector != nullptr) ? 1 : 0;
    return C_SUCCESS;
}

CErrorCode DeviceFingerprint_GetSupportedPlatforms(char* platforms_buffer,
                                                   size_t buffer_size) {
    if (!platforms_buffer || buffer_size == 0) {
        return C_ERROR_INVALID_PARAM;
    }
    
    try {
        // 返回支持的平台列表
        std::string platforms_str = "windows,linux,macos";
        
        if (platforms_str.length() >= buffer_size) {
            return C_ERROR_MEMORY_ALLOC;
        }
        strncpy(platforms_buffer, platforms_str.c_str(), buffer_size - 1);
        platforms_buffer[buffer_size - 1] = '\0';
        return C_SUCCESS;
    } catch (const std::exception& e) {
        return C_ERROR_UNKNOWN;
    }
}

// 简化的C接口函数实现
typedef void DeviceFingerprintHandle;

DeviceFingerprintHandle* device_fingerprint_create(void) {
    try {
        return reinterpret_cast<DeviceFingerprintHandle*>(new gaokao::device::DeviceFingerprintCollector());
    } catch (const std::exception& e) {
        return nullptr;
    }
}

void device_fingerprint_destroy(DeviceFingerprintHandle* handle) {
    if (handle) {
        delete reinterpret_cast<gaokao::device::DeviceFingerprintCollector*>(handle);
    }
}

char* device_fingerprint_collect(DeviceFingerprintHandle* handle) {
    if (!handle) {
        return nullptr;
    }
    
    try {
        gaokao::device::DeviceFingerprintCollector* collector = 
            reinterpret_cast<gaokao::device::DeviceFingerprintCollector*>(handle);
        
        gaokao::device::DeviceFingerprint fingerprint;
        gaokao::device::ErrorCode result = collector->CollectFingerprint(fingerprint);
        
        if (result == gaokao::device::ErrorCode::SUCCESS) {
            std::string fingerprint_str = fingerprint.device_id + ":" + fingerprint.hardware.cpu_id;
            char* cpy = new char[fingerprint_str.length() + 1];
            strcpy(cpy, fingerprint_str.c_str());
            return cpy;
        }
        
        return nullptr;
    } catch (const std::exception& e) {
        return nullptr;
    }
}

char* device_fingerprint_get_hardware_info(DeviceFingerprintHandle* handle) {
    if (!handle) {
        return nullptr;
    }
    
    try {
        gaokao::device::DeviceFingerprintCollector* collector = 
            reinterpret_cast<gaokao::device::DeviceFingerprintCollector*>(handle);
        
        gaokao::device::HardwareInfo hw_info;
        gaokao::device::ErrorCode result = collector->CollectHardwareInfo(hw_info);
        
        if (result == gaokao::device::ErrorCode::SUCCESS) {
            std::string info_str = hw_info.cpu_model + ":" + std::to_string(hw_info.total_memory);
            char* cpy = new char[info_str.length() + 1];
            strcpy(cpy, info_str.c_str());
            return cpy;
        }
        
        return nullptr;
    } catch (const std::exception& e) {
        return nullptr;
    }
}

char* device_fingerprint_get_system_info(DeviceFingerprintHandle* handle) {
    if (!handle) {
        return nullptr;
    }
    
    try {
        gaokao::device::DeviceFingerprintCollector* collector = 
            reinterpret_cast<gaokao::device::DeviceFingerprintCollector*>(handle);
        
        gaokao::device::SystemInfo sys_info;
        gaokao::device::ErrorCode result = collector->CollectSystemInfo(sys_info);
        
        if (result == gaokao::device::ErrorCode::SUCCESS) {
            std::string os_type_str;
            switch (sys_info.os_type) {
                case gaokao::device::OperatingSystem::WINDOWS:
                    os_type_str = "windows";
                    break;
                case gaokao::device::OperatingSystem::LINUX:
                    os_type_str = "linux";
                    break;
                case gaokao::device::OperatingSystem::MACOS:
                    os_type_str = "macos";
                    break;
                case gaokao::device::OperatingSystem::ANDROID:
                    os_type_str = "android";
                    break;
                case gaokao::device::OperatingSystem::IOS:
                    os_type_str = "ios";
                    break;
                default:
                    os_type_str = "unknown";
                    break;
            }
            
            std::string info_str = os_type_str + ":" + sys_info.os_version;
            char* cpy = new char[info_str.length() + 1];
            strcpy(cpy, info_str.c_str());
            return cpy;
        }
        
        return nullptr;
    } catch (const std::exception& e) {
        return nullptr;
    }
}

char* device_fingerprint_get_network_info(DeviceFingerprintHandle* handle) {
    if (!handle) {
        return nullptr;
    }
    
    // 注意：当前版本未实现网络信息采集功能
    try {
        std::string info_str = "network_info_not_implemented";
        char* cpy = new char[info_str.length() + 1];
        strcpy(cpy, info_str.c_str());
        return cpy;
    } catch (const std::exception& e) {
        return nullptr;
    }
}

} // extern "C"