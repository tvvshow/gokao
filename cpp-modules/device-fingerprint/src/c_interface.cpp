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
    char device_id[64];
    char device_type[32];
    char cpu_id[64];
    char cpu_model[128];
    unsigned int cpu_cores;
    unsigned long long total_memory;
    char motherboard_serial[64];
    char os_type[32];
    char os_version[64];
    char hostname[64];
    char username[64];
    char screen_resolution[32];
    char fingerprint_hash[128];
    unsigned int confidence_score;
    char error_message[256];
} CDeviceFingerprint;

// C风格配置结构体
typedef struct {
    int collect_sensitive_info;
    int enable_encryption;
    int enable_signature;
    char encryption_key[64];
} CConfiguration;

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

static void ConvertDeviceType(gaokao::device::DeviceType type, char* buffer, size_t size) {
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

static void ConvertOSType(gaokao::device::OperatingSystem os, char* buffer, size_t size) {
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

static void SafeStrCopy(char* dest, const std::string& src, size_t dest_size) {
    if (dest && dest_size > 0) {
        strncpy(dest, src.c_str(), dest_size - 1);
        dest[dest_size - 1] = '\0';
    }
}

// C接口函数实现

/**
 * @brief 初始化设备指纹采集器
 * @param config_path 配置文件路径(可为NULL)
 * @return 错误码
 */
CErrorCode DeviceFingerprint_Initialize(const char* config_path) {
    try {
        if (!g_collector) {
            g_collector = std::make_unique<gaokao::device::DeviceFingerprintCollector>();
        }
        
        std::string config_str = config_path ? config_path : "";
        auto result = g_collector->Initialize(config_str);
        return ConvertErrorCode(result);
        
    } catch (const std::exception& e) {
        return C_ERROR_INIT_FAILED;
    } catch (...) {
        return C_ERROR_UNKNOWN;
    }
}

/**
 * @brief 反初始化设备指纹采集器
 */
void DeviceFingerprint_Uninitialize() {
    try {
        if (g_collector) {
            g_collector->Uninitialize();
            g_collector.reset();
        }
    } catch (...) {
        // 忽略异常
    }
}

/**
 * @brief 采集设备指纹
 * @param fingerprint 输出的设备指纹结构体指针
 * @return 错误码
 */
CErrorCode DeviceFingerprint_Collect(CDeviceFingerprint* fingerprint) {
    if (!fingerprint) {
        return C_ERROR_INVALID_PARAM;
    }
    
    if (!g_collector) {
        SafeStrCopy(fingerprint->error_message, "采集器未初始化", sizeof(fingerprint->error_message));
        return C_ERROR_INIT_FAILED;
    }
    
    try {
        // 清空结构体
        memset(fingerprint, 0, sizeof(CDeviceFingerprint));
        
        // 采集设备指纹
        gaokao::device::DeviceFingerprint cpp_fingerprint;
        auto result = g_collector->CollectFingerprint(cpp_fingerprint);
        
        if (result != gaokao::device::ErrorCode::SUCCESS) {
            SafeStrCopy(fingerprint->error_message, 
                       gaokao::device::DeviceFingerprintCollector::GetErrorDescription(result),
                       sizeof(fingerprint->error_message));
            return ConvertErrorCode(result);
        }
        
        // 转换数据到C结构体
        SafeStrCopy(fingerprint->device_id, cpp_fingerprint.device_id, sizeof(fingerprint->device_id));
        ConvertDeviceType(cpp_fingerprint.device_type, fingerprint->device_type, sizeof(fingerprint->device_type));
        
        // 硬件信息
        SafeStrCopy(fingerprint->cpu_id, cpp_fingerprint.hardware.cpu_id, sizeof(fingerprint->cpu_id));
        SafeStrCopy(fingerprint->cpu_model, cpp_fingerprint.hardware.cpu_model, sizeof(fingerprint->cpu_model));
        fingerprint->cpu_cores = cpp_fingerprint.hardware.cpu_cores;
        fingerprint->total_memory = cpp_fingerprint.hardware.total_memory;
        SafeStrCopy(fingerprint->motherboard_serial, cpp_fingerprint.hardware.motherboard_serial, 
                   sizeof(fingerprint->motherboard_serial));
        
        // 系统信息
        ConvertOSType(cpp_fingerprint.system.os_type, fingerprint->os_type, sizeof(fingerprint->os_type));
        SafeStrCopy(fingerprint->os_version, cpp_fingerprint.system.os_version, sizeof(fingerprint->os_version));
        SafeStrCopy(fingerprint->hostname, cpp_fingerprint.system.hostname, sizeof(fingerprint->hostname));
        SafeStrCopy(fingerprint->username, cpp_fingerprint.system.username, sizeof(fingerprint->username));
        
        // 运行时信息
        SafeStrCopy(fingerprint->screen_resolution, cpp_fingerprint.runtime.screen_resolution, 
                   sizeof(fingerprint->screen_resolution));
        
        // 指纹和置信度
        SafeStrCopy(fingerprint->fingerprint_hash, cpp_fingerprint.fingerprint_hash, 
                   sizeof(fingerprint->fingerprint_hash));
        fingerprint->confidence_score = cpp_fingerprint.confidence_score;
        
        return C_SUCCESS;
        
    } catch (const std::exception& e) {
        SafeStrCopy(fingerprint->error_message, e.what(), sizeof(fingerprint->error_message));
        return C_ERROR_UNKNOWN;
    } catch (...) {
        SafeStrCopy(fingerprint->error_message, "未知异常", sizeof(fingerprint->error_message));
        return C_ERROR_UNKNOWN;
    }
}

/**
 * @brief 设置采集配置
 * @param config 配置结构体指针
 * @return 错误码
 */
CErrorCode DeviceFingerprint_SetConfiguration(const CConfiguration* config) {
    if (!config) {
        return C_ERROR_INVALID_PARAM;
    }
    
    if (!g_collector) {
        return C_ERROR_INIT_FAILED;
    }
    
    try {
        g_collector->SetConfiguration(
            config->collect_sensitive_info != 0,
            config->enable_encryption != 0,
            config->enable_signature != 0
        );
        
        return C_SUCCESS;
        
    } catch (...) {
        return C_ERROR_UNKNOWN;
    }
}

/**
 * @brief 生成设备指纹哈希
 * @param fingerprint 设备指纹结构体指针
 * @param hash_buffer 输出的哈希缓冲区
 * @param buffer_size 缓冲区大小
 * @return 错误码
 */
CErrorCode DeviceFingerprint_GenerateHash(const CDeviceFingerprint* fingerprint,
                                         char* hash_buffer,
                                         size_t buffer_size) {
    if (!fingerprint || !hash_buffer || buffer_size == 0) {
        return C_ERROR_INVALID_PARAM;
    }
    
    if (!g_collector) {
        return C_ERROR_INIT_FAILED;
    }
    
    try {
        // 构建C++设备指纹对象
        gaokao::device::DeviceFingerprint cpp_fingerprint;
        cpp_fingerprint.device_id = fingerprint->device_id;
        cpp_fingerprint.hardware.cpu_id = fingerprint->cpu_id;
        cpp_fingerprint.hardware.cpu_model = fingerprint->cpu_model;
        cpp_fingerprint.hardware.cpu_cores = fingerprint->cpu_cores;
        cpp_fingerprint.hardware.total_memory = fingerprint->total_memory;
        cpp_fingerprint.hardware.motherboard_serial = fingerprint->motherboard_serial;
        cpp_fingerprint.system.os_version = fingerprint->os_version;
        cpp_fingerprint.system.hostname = fingerprint->hostname;
        cpp_fingerprint.system.username = fingerprint->username;
        cpp_fingerprint.runtime.screen_resolution = fingerprint->screen_resolution;
        
        // 生成哈希
        std::string hash = g_collector->GenerateFingerprintHash(cpp_fingerprint);
        
        SafeStrCopy(hash_buffer, hash, buffer_size);
        return C_SUCCESS;
        
    } catch (...) {
        return C_ERROR_UNKNOWN;
    }
}

/**
 * @brief 比较两个设备指纹
 * @param fingerprint1 第一个设备指纹
 * @param fingerprint2 第二个设备指纹
 * @param similarity_score 输出的相似度分数(0.0-1.0)
 * @param is_same_device 输出是否为同一设备(1表示是，0表示否)
 * @return 错误码
 */
CErrorCode DeviceFingerprint_Compare(const CDeviceFingerprint* fingerprint1,
                                    const CDeviceFingerprint* fingerprint2,
                                    double* similarity_score,
                                    int* is_same_device) {
    if (!fingerprint1 || !fingerprint2 || !similarity_score || !is_same_device) {
        return C_ERROR_INVALID_PARAM;
    }
    
    if (!g_collector) {
        return C_ERROR_INIT_FAILED;
    }
    
    try {
        // 构建C++设备指纹对象
        gaokao::device::DeviceFingerprint cpp_fp1, cpp_fp2;
        
        // 转换第一个指纹
        cpp_fp1.device_id = fingerprint1->device_id;
        cpp_fp1.fingerprint_hash = fingerprint1->fingerprint_hash;
        
        // 转换第二个指纹
        cpp_fp2.device_id = fingerprint2->device_id;
        cpp_fp2.fingerprint_hash = fingerprint2->fingerprint_hash;
        
        // 比较指纹
        gaokao::device::ComparisonResult result;
        auto error = g_collector->CompareFingerprints(cpp_fp1, cpp_fp2, result);
        
        if (error != gaokao::device::ErrorCode::SUCCESS) {
            return ConvertErrorCode(error);
        }
        
        *similarity_score = result.similarity_score;
        *is_same_device = result.is_same_device ? 1 : 0;
        
        return C_SUCCESS;
        
    } catch (...) {
        return C_ERROR_UNKNOWN;
    }
}

/**
 * @brief 验证设备指纹
 * @param fingerprint 设备指纹结构体指针
 * @param reference_hash 参考哈希值
 * @param is_valid 输出验证结果(1表示有效，0表示无效)
 * @return 错误码
 */
CErrorCode DeviceFingerprint_Validate(const CDeviceFingerprint* fingerprint,
                                     const char* reference_hash,
                                     int* is_valid) {
    if (!fingerprint || !reference_hash || !is_valid) {
        return C_ERROR_INVALID_PARAM;
    }
    
    if (!g_collector) {
        return C_ERROR_INIT_FAILED;
    }
    
    try {
        // 构建C++设备指纹对象
        gaokao::device::DeviceFingerprint cpp_fingerprint;
        cpp_fingerprint.device_id = fingerprint->device_id;
        cpp_fingerprint.fingerprint_hash = fingerprint->fingerprint_hash;
        
        // 验证指纹
        bool valid = g_collector->ValidateFingerprint(cpp_fingerprint, reference_hash);
        *is_valid = valid ? 1 : 0;
        
        return C_SUCCESS;
        
    } catch (...) {
        return C_ERROR_UNKNOWN;
    }
}

/**
 * @brief 序列化设备指纹为JSON字符串
 * @param fingerprint 设备指纹结构体指针
 * @param json_buffer 输出的JSON缓冲区
 * @param buffer_size 缓冲区大小
 * @return 错误码
 */
CErrorCode DeviceFingerprint_SerializeToJson(const CDeviceFingerprint* fingerprint,
                                            char* json_buffer,
                                            size_t buffer_size) {
    if (!fingerprint || !json_buffer || buffer_size == 0) {
        return C_ERROR_INVALID_PARAM;
    }
    
    if (!g_collector) {
        return C_ERROR_INIT_FAILED;
    }
    
    try {
        // 构建C++设备指纹对象
        gaokao::device::DeviceFingerprint cpp_fingerprint;
        cpp_fingerprint.device_id = fingerprint->device_id;
        cpp_fingerprint.hardware.cpu_id = fingerprint->cpu_id;
        cpp_fingerprint.hardware.cpu_model = fingerprint->cpu_model;
        cpp_fingerprint.hardware.cpu_cores = fingerprint->cpu_cores;
        cpp_fingerprint.hardware.total_memory = fingerprint->total_memory;
        cpp_fingerprint.hardware.motherboard_serial = fingerprint->motherboard_serial;
        cpp_fingerprint.system.os_version = fingerprint->os_version;
        cpp_fingerprint.system.hostname = fingerprint->hostname;
        cpp_fingerprint.system.username = fingerprint->username;
        cpp_fingerprint.runtime.screen_resolution = fingerprint->screen_resolution;
        cpp_fingerprint.fingerprint_hash = fingerprint->fingerprint_hash;
        cpp_fingerprint.confidence_score = fingerprint->confidence_score;
        
        // 序列化为JSON
        std::string json = g_collector->SerializeToJson(cpp_fingerprint);
        
        SafeStrCopy(json_buffer, json, buffer_size);
        return C_SUCCESS;
        
    } catch (...) {
        return C_ERROR_UNKNOWN;
    }
}

/**
 * @brief 检查是否存在调试器
 * @param is_debugger_present 输出结果(1表示存在，0表示不存在)
 * @return 错误码
 */
CErrorCode DeviceFingerprint_IsDebuggerPresent(int* is_debugger_present) {
    if (!is_debugger_present) {
        return C_ERROR_INVALID_PARAM;
    }
    
    if (!g_collector) {
        return C_ERROR_INIT_FAILED;
    }
    
    try {
        bool present = g_collector->IsDebuggerPresent();
        *is_debugger_present = present ? 1 : 0;
        
        return C_SUCCESS;
        
    } catch (...) {
        return C_ERROR_UNKNOWN;
    }
}

/**
 * @brief 检查是否在虚拟机中运行
 * @param is_virtual_machine 输出结果(1表示是，0表示否)
 * @return 错误码
 */
CErrorCode DeviceFingerprint_IsVirtualMachine(int* is_virtual_machine) {
    if (!is_virtual_machine) {
        return C_ERROR_INVALID_PARAM;
    }
    
    if (!g_collector) {
        return C_ERROR_INIT_FAILED;
    }
    
    try {
        bool vm = g_collector->IsRunningInVirtualMachine();
        *is_virtual_machine = vm ? 1 : 0;
        
        return C_SUCCESS;
        
    } catch (...) {
        return C_ERROR_UNKNOWN;
    }
}

/**
 * @brief 获取库版本
 * @param version_buffer 输出的版本缓冲区
 * @param buffer_size 缓冲区大小
 * @return 错误码
 */
CErrorCode DeviceFingerprint_GetVersion(char* version_buffer, size_t buffer_size) {
    if (!version_buffer || buffer_size == 0) {
        return C_ERROR_INVALID_PARAM;
    }
    
    try {
        std::string version = gaokao::device::DeviceFingerprintCollector::GetVersion();
        SafeStrCopy(version_buffer, version, buffer_size);
        
        return C_SUCCESS;
        
    } catch (...) {
        return C_ERROR_UNKNOWN;
    }
}

/**
 * @brief 获取错误描述
 * @param error_code 错误码
 * @param error_buffer 输出的错误描述缓冲区
 * @param buffer_size 缓冲区大小
 * @return 错误码
 */
CErrorCode DeviceFingerprint_GetErrorDescription(CErrorCode error_code,
                                                char* error_buffer,
                                                size_t buffer_size) {
    if (!error_buffer || buffer_size == 0) {
        return C_ERROR_INVALID_PARAM;
    }
    
    try {
        const char* description = "";
        
        switch (error_code) {
            case C_SUCCESS:
                description = "操作成功";
                break;
            case C_ERROR_INIT_FAILED:
                description = "初始化失败";
                break;
            case C_ERROR_INVALID_PARAM:
                description = "无效参数";
                break;
            case C_ERROR_MEMORY_ALLOC:
                description = "内存分配失败";
                break;
            case C_ERROR_HARDWARE_ACCESS:
                description = "硬件访问失败";
                break;
            case C_ERROR_SYSTEM_INFO:
                description = "系统信息获取失败";
                break;
            case C_ERROR_ENCRYPTION:
                description = "加密操作失败";
                break;
            default:
                description = "未知错误";
                break;
        }
        
        SafeStrCopy(error_buffer, description, buffer_size);
        return C_SUCCESS;
        
    } catch (...) {
        return C_ERROR_UNKNOWN;
    }
}

/**
 * @brief 快速采集设备指纹(无需初始化)
 * @param fingerprint 输出的设备指纹结构体指针
 * @return 错误码
 */
CErrorCode DeviceFingerprint_QuickCollect(CDeviceFingerprint* fingerprint) {
    if (!fingerprint) {
        return C_ERROR_INVALID_PARAM;
    }
    
    try {
        // 清空结构体
        memset(fingerprint, 0, sizeof(CDeviceFingerprint));
        
        // 使用全局函数快速采集
        gaokao::device::DeviceFingerprint cpp_fingerprint;
        auto result = gaokao::device::QuickCollectFingerprint(cpp_fingerprint);
        
        if (result != gaokao::device::ErrorCode::SUCCESS) {
            SafeStrCopy(fingerprint->error_message, 
                       gaokao::device::DeviceFingerprintCollector::GetErrorDescription(result),
                       sizeof(fingerprint->error_message));
            return ConvertErrorCode(result);
        }
        
        // 转换数据到C结构体
        SafeStrCopy(fingerprint->device_id, cpp_fingerprint.device_id, sizeof(fingerprint->device_id));
        ConvertDeviceType(cpp_fingerprint.device_type, fingerprint->device_type, sizeof(fingerprint->device_type));
        SafeStrCopy(fingerprint->cpu_id, cpp_fingerprint.hardware.cpu_id, sizeof(fingerprint->cpu_id));
        SafeStrCopy(fingerprint->fingerprint_hash, cpp_fingerprint.fingerprint_hash, 
                   sizeof(fingerprint->fingerprint_hash));
        fingerprint->confidence_score = cpp_fingerprint.confidence_score;
        
        return C_SUCCESS;
        
    } catch (const std::exception& e) {
        SafeStrCopy(fingerprint->error_message, e.what(), sizeof(fingerprint->error_message));
        return C_ERROR_UNKNOWN;
    } catch (...) {
        SafeStrCopy(fingerprint->error_message, "未知异常", sizeof(fingerprint->error_message));
        return C_ERROR_UNKNOWN;
    }
}

} // extern "C"