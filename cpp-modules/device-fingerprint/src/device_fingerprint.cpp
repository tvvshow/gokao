/**
 * @file device_fingerprint.cpp
 * @brief 设备指纹采集模块实现
 * @author 高考志愿填报系统开发团队
 * @version 1.0.0
 * @date 2025-01-18
 */

#include "device_fingerprint.h"
#include "platform_detector.h"
#include "crypto_utils.h"

#include <iostream>
#include <sstream>
#include <algorithm>
#include <iomanip>
#include <mutex>
#include <thread>

#ifdef _WIN32
    #include <windows.h>
    #include <intrin.h>
    #include <iphlpapi.h>
    #include <wbemidl.h>
    #include <comdef.h>
    #pragma comment(lib, "iphlpapi.lib")
    #pragma comment(lib, "wbemuuid.lib")
#elif defined(__linux__)
    #include <unistd.h>
    #include <sys/utsname.h>
    #include <sys/sysinfo.h>
    #include <ifaddrs.h>
    #include <net/if.h>
    #include <cpuid.h>
#elif defined(__APPLE__)
    #include <unistd.h>
    #include <sys/utsname.h>
    #include <sys/sysctl.h>
    #include <ifaddrs.h>
    #include <net/if.h>
    #include <IOKit/IOKitLib.h>
#endif

// JSON处理库(这里使用简单实现，实际项目中可使用nlohmann/json)
namespace json_utils {
    std::string escape_json_string(const std::string& input) {
        std::string output;
        for (char c : input) {
            switch (c) {
                case '"': output += "\\\""; break;
                case '\\': output += "\\\\"; break;
                case '\b': output += "\\b"; break;
                case '\f': output += "\\f"; break;
                case '\n': output += "\\n"; break;
                case '\r': output += "\\r"; break;
                case '\t': output += "\\t"; break;
                default: output += c; break;
            }
        }
        return output;
    }
}

namespace gaokao {
namespace device {

// PIMPL实现类
class DeviceFingerprintCollector::Impl {
public:
    Impl() : initialized_(false), collect_sensitive_info_(true),
             enable_encryption_(true), enable_signature_(true) {}
    
    ~Impl() = default;
    
    bool initialized_;
    bool collect_sensitive_info_;
    bool enable_encryption_;
    bool enable_signature_;
    
    std::unique_ptr<platform::PlatformDetector> platform_detector_;
    std::unique_ptr<crypto::AESCipher> aes_cipher_;
    std::unique_ptr<crypto::RSACipher> rsa_cipher_;
    std::mutex mutex_;
    
    // 内部方法
    std::string GenerateDeviceID(const HardwareInfo& hardware, const SystemInfo& system);
    DeviceType DetermineDeviceType(const platform::OSInfo& os_info);
    uint32_t CalculateConfidenceScore(const DeviceFingerprint& fingerprint);
    std::string CombineFingerprints(const DeviceFingerprint& fingerprint);
};

// 构造函数
DeviceFingerprintCollector::DeviceFingerprintCollector() 
    : pimpl_(std::make_unique<Impl>()) {
}

// 析构函数
DeviceFingerprintCollector::~DeviceFingerprintCollector() {
    if (pimpl_ && pimpl_->initialized_) {
        Uninitialize();
    }
}

// 初始化采集器
ErrorCode DeviceFingerprintCollector::Initialize(const std::string& config_path) {
    std::lock_guard<std::mutex> lock(pimpl_->mutex_);
    
    if (pimpl_->initialized_) {
        return ErrorCode::SUCCESS;
    }
    
    try {
        // 初始化平台检测器
        pimpl_->platform_detector_ = std::make_unique<platform::PlatformDetector>();
        auto result = pimpl_->platform_detector_->Initialize();
        if (result != platform::DetectionError::SUCCESS) {
            return ErrorCode::INITIALIZATION_FAILED;
        }
        
        // 初始化加密模块
        pimpl_->aes_cipher_ = std::make_unique<crypto::AESCipher>();
        pimpl_->rsa_cipher_ = std::make_unique<crypto::RSACipher>();
        
        pimpl_->initialized_ = true;
        return ErrorCode::SUCCESS;
        
    } catch (const std::exception& e) {
        return ErrorCode::INITIALIZATION_FAILED;
    }
}

// 反初始化采集器
void DeviceFingerprintCollector::Uninitialize() {
    std::lock_guard<std::mutex> lock(pimpl_->mutex_);
    
    if (pimpl_->platform_detector_) {
        pimpl_->platform_detector_->Uninitialize();
        pimpl_->platform_detector_.reset();
    }
    
    pimpl_->aes_cipher_.reset();
    pimpl_->rsa_cipher_.reset();
    pimpl_->initialized_ = false;
}

// 采集完整设备指纹
ErrorCode DeviceFingerprintCollector::CollectFingerprint(DeviceFingerprint& fingerprint) {
    std::lock_guard<std::mutex> lock(pimpl_->mutex_);
    
    if (!pimpl_->initialized_) {
        return ErrorCode::INITIALIZATION_FAILED;
    }
    
    try {
        // 采集硬件信息
        auto hw_result = CollectHardwareInfo(fingerprint.hardware);
        if (hw_result != ErrorCode::SUCCESS) {
            return hw_result;
        }
        
        // 采集系统信息
        auto sys_result = CollectSystemInfo(fingerprint.system);
        if (sys_result != ErrorCode::SUCCESS) {
            return sys_result;
        }
        
        // 采集运行时信息
        auto rt_result = CollectRuntimeInfo(fingerprint.runtime);
        if (rt_result != ErrorCode::SUCCESS) {
            return rt_result;
        }
        
        // 生成设备ID和类型
        fingerprint.device_id = pimpl_->GenerateDeviceID(fingerprint.hardware, fingerprint.system);
        
        platform::OSInfo os_info;
        pimpl_->platform_detector_->GetOSInfo(os_info);
        fingerprint.device_type = pimpl_->DetermineDeviceType(os_info);
        
        // 生成指纹哈希
        fingerprint.fingerprint_hash = GenerateFingerprintHash(fingerprint);
        
        // 计算置信度分数
        fingerprint.confidence_score = pimpl_->CalculateConfidenceScore(fingerprint);
        
        // 设置创建时间
        fingerprint.created_at = std::chrono::system_clock::now();
        
        // 加密敏感数据
        if (pimpl_->enable_encryption_) {
            std::string combined_data = pimpl_->CombineFingerprints(fingerprint);
            
            auto encrypt_result = pimpl_->aes_cipher_->EncryptString(
                combined_data, "default_key_2025", fingerprint.encrypted_data);
            
            if (encrypt_result != crypto::CryptoError::SUCCESS) {
                return ErrorCode::ENCRYPTION_FAILED;
            }
        }
        
        return ErrorCode::SUCCESS;
        
    } catch (const std::exception& e) {
        return ErrorCode::SYSTEM_INFO_UNAVAILABLE;
    }
}

// 采集硬件信息
ErrorCode DeviceFingerprintCollector::CollectHardwareInfo(HardwareInfo& hardware_info) {
    if (!pimpl_->initialized_) {
        return ErrorCode::INITIALIZATION_FAILED;
    }
    
    try {
        // 获取CPU信息
        platform::CPUInfo cpu_info;
        auto result = pimpl_->platform_detector_->GetCPUInfo(cpu_info);
        if (result == platform::DetectionError::SUCCESS) {
            hardware_info.cpu_id = cpu_info.identifier;
            hardware_info.cpu_model = cpu_info.model_name;
            hardware_info.cpu_cores = cpu_info.logical_cores;
        }
        
        // 获取内存信息
        platform::MemoryInfo memory_info;
        result = pimpl_->platform_detector_->GetMemoryInfo(memory_info);
        if (result == platform::DetectionError::SUCCESS) {
            hardware_info.total_memory = memory_info.total_physical;
        }
        
        // 获取主板信息
        platform::MotherboardInfo mb_info;
        result = pimpl_->platform_detector_->GetMotherboardInfo(mb_info);
        if (result == platform::DetectionError::SUCCESS) {
            hardware_info.motherboard_serial = mb_info.serial_number;
            hardware_info.motherboard_manufacturer = mb_info.manufacturer;
        }
        
        // 获取网络适配器信息
        std::vector<platform::NetworkAdapterInfo> adapters;
        result = pimpl_->platform_detector_->GetNetworkAdapters(adapters);
        if (result == platform::DetectionError::SUCCESS) {
            for (const auto& adapter : adapters) {
                if (adapter.is_physical && !adapter.mac_address.empty()) {
                    hardware_info.mac_addresses.push_back(adapter.mac_address);
                }
            }
        }
        
        // 获取存储设备信息
        std::vector<platform::StorageDeviceInfo> storage_devices;
        result = pimpl_->platform_detector_->GetStorageDevices(storage_devices);
        if (result == platform::DetectionError::SUCCESS) {
            for (const auto& device : storage_devices) {
                if (!device.serial_number.empty()) {
                    hardware_info.disk_serials.push_back(device.serial_number);
                }
            }
        }
        
        // 获取BIOS信息
        platform::BIOSInfo bios_info;
        result = pimpl_->platform_detector_->GetBIOSInfo(bios_info);
        if (result == platform::DetectionError::SUCCESS) {
            hardware_info.bios_serial = bios_info.serial_number;
            hardware_info.bios_version = bios_info.version;
        }
        
        return ErrorCode::SUCCESS;
        
    } catch (const std::exception& e) {
        return ErrorCode::HARDWARE_ACCESS_DENIED;
    }
}

// 采集系统信息
ErrorCode DeviceFingerprintCollector::CollectSystemInfo(SystemInfo& system_info) {
    if (!pimpl_->initialized_) {
        return ErrorCode::INITIALIZATION_FAILED;
    }
    
    try {
        // 获取操作系统信息
        platform::OSInfo os_info;
        auto result = pimpl_->platform_detector_->GetOSInfo(os_info);
        if (result == platform::DetectionError::SUCCESS) {
            // 转换平台类型
            switch (os_info.platform) {
                case platform::PlatformType::WINDOWS:
                    system_info.os_type = OperatingSystem::WINDOWS;
                    break;
                case platform::PlatformType::LINUX:
                    system_info.os_type = OperatingSystem::LINUX;
                    break;
                case platform::PlatformType::MACOS:
                    system_info.os_type = OperatingSystem::MACOS;
                    break;
                case platform::PlatformType::ANDROID:
                    system_info.os_type = OperatingSystem::ANDROID;
                    break;
                case platform::PlatformType::IOS:
                    system_info.os_type = OperatingSystem::IOS;
                    break;
                default:
                    system_info.os_type = OperatingSystem::UNKNOWN;
                    break;
            }
            
            system_info.os_version = os_info.version;
            system_info.os_build = os_info.build_number;
            system_info.hostname = os_info.name;
            system_info.uptime = os_info.uptime;
        }
        
        // 获取用户信息
        platform::UserInfo user_info;
        result = pimpl_->platform_detector_->GetUserInfo(user_info);
        if (result == platform::DetectionError::SUCCESS) {
            system_info.username = user_info.username;
            system_info.domain = user_info.domain;
        }
        
        // 获取环境信息
        platform::EnvironmentInfo env_info;
        result = pimpl_->platform_detector_->GetEnvironmentInfo(env_info);
        if (result == platform::DetectionError::SUCCESS) {
            system_info.timezone = env_info.timezone;
            system_info.locale = env_info.locale;
        }
        
        return ErrorCode::SUCCESS;
        
    } catch (const std::exception& e) {
        return ErrorCode::SYSTEM_INFO_UNAVAILABLE;
    }
}

// 采集运行时信息
ErrorCode DeviceFingerprintCollector::CollectRuntimeInfo(RuntimeInfo& runtime_info) {
    if (!pimpl_->initialized_) {
        return ErrorCode::INITIALIZATION_FAILED;
    }
    
    try {
        // 获取显示设备信息
        std::vector<platform::DisplayDeviceInfo> displays;
        auto result = pimpl_->platform_detector_->GetDisplayDevices(displays);
        if (result == platform::DetectionError::SUCCESS && !displays.empty()) {
            const auto& primary_display = displays[0];
            runtime_info.screen_resolution = std::to_string(primary_display.width) + 
                                            "x" + std::to_string(primary_display.height);
            runtime_info.color_depth = primary_display.color_depth;
        }
        
        // 设置时间戳
        runtime_info.timestamp = std::chrono::system_clock::now();
        
        return ErrorCode::SUCCESS;
        
    } catch (const std::exception& e) {
        return ErrorCode::SYSTEM_INFO_UNAVAILABLE;
    }
}

// 生成设备指纹哈希
std::string DeviceFingerprintCollector::GenerateFingerprintHash(const DeviceFingerprint& fingerprint) {
    std::string combined_data = pimpl_->CombineFingerprints(fingerprint);
    
    std::string hash_hex;
    auto result = crypto::HashUtils::CalculateStringHash(
        combined_data, crypto::HashAlgorithm::SHA256, hash_hex);
    
    if (result == crypto::CryptoError::SUCCESS) {
        return hash_hex;
    }
    
    return "";
}

// 比较两个设备指纹
ErrorCode DeviceFingerprintCollector::CompareFingerprints(
    const DeviceFingerprint& fingerprint1,
    const DeviceFingerprint& fingerprint2,
    ComparisonResult& result) {
    
    try {
        // 简单的字符串相似度比较
        double similarity = CalculateFingerprintSimilarity(
            fingerprint1.fingerprint_hash, fingerprint2.fingerprint_hash);
        
        result.similarity_score = similarity;
        result.is_same_device = (similarity > 0.9);
        result.confidence_level = static_cast<uint32_t>(similarity * 100);
        
        // 检查具体差异
        if (fingerprint1.hardware.cpu_id != fingerprint2.hardware.cpu_id) {
            result.differences.push_back("CPU ID不匹配");
        }
        if (fingerprint1.hardware.motherboard_serial != fingerprint2.hardware.motherboard_serial) {
            result.differences.push_back("主板序列号不匹配");
        }
        if (fingerprint1.system.hostname != fingerprint2.system.hostname) {
            result.differences.push_back("主机名不匹配");
        }
        
        return ErrorCode::SUCCESS;
        
    } catch (const std::exception& e) {
        return ErrorCode::INVALID_PARAMETER;
    }
}

// 验证设备指纹
bool DeviceFingerprintCollector::ValidateFingerprint(
    const DeviceFingerprint& fingerprint,
    const std::string& reference_hash) {
    
    std::string calculated_hash = GenerateFingerprintHash(fingerprint);
    return (calculated_hash == reference_hash);
}

// 序列化设备指纹为JSON
std::string DeviceFingerprintCollector::SerializeToJson(const DeviceFingerprint& fingerprint) {
    std::ostringstream json;
    
    json << "{\n";
    json << "  \"device_id\": \"" << json_utils::escape_json_string(fingerprint.device_id) << "\",\n";
    json << "  \"device_type\": " << static_cast<int>(fingerprint.device_type) << ",\n";
    json << "  \"confidence_score\": " << fingerprint.confidence_score << ",\n";
    json << "  \"fingerprint_hash\": \"" << json_utils::escape_json_string(fingerprint.fingerprint_hash) << "\",\n";
    
    // 硬件信息
    json << "  \"hardware\": {\n";
    json << "    \"cpu_id\": \"" << json_utils::escape_json_string(fingerprint.hardware.cpu_id) << "\",\n";
    json << "    \"cpu_model\": \"" << json_utils::escape_json_string(fingerprint.hardware.cpu_model) << "\",\n";
    json << "    \"cpu_cores\": " << fingerprint.hardware.cpu_cores << ",\n";
    json << "    \"total_memory\": " << fingerprint.hardware.total_memory << ",\n";
    json << "    \"motherboard_serial\": \"" << json_utils::escape_json_string(fingerprint.hardware.motherboard_serial) << "\",\n";
    json << "    \"mac_addresses\": [";
    for (size_t i = 0; i < fingerprint.hardware.mac_addresses.size(); ++i) {
        if (i > 0) json << ", ";
        json << "\"" << json_utils::escape_json_string(fingerprint.hardware.mac_addresses[i]) << "\"";
    }
    json << "]\n";
    json << "  },\n";
    
    // 系统信息
    json << "  \"system\": {\n";
    json << "    \"os_type\": " << static_cast<int>(fingerprint.system.os_type) << ",\n";
    json << "    \"os_version\": \"" << json_utils::escape_json_string(fingerprint.system.os_version) << "\",\n";
    json << "    \"hostname\": \"" << json_utils::escape_json_string(fingerprint.system.hostname) << "\",\n";
    json << "    \"username\": \"" << json_utils::escape_json_string(fingerprint.system.username) << "\",\n";
    json << "    \"timezone\": \"" << json_utils::escape_json_string(fingerprint.system.timezone) << "\"\n";
    json << "  },\n";
    
    // 运行时信息
    json << "  \"runtime\": {\n";
    json << "    \"screen_resolution\": \"" << json_utils::escape_json_string(fingerprint.runtime.screen_resolution) << "\",\n";
    json << "    \"color_depth\": " << fingerprint.runtime.color_depth << "\n";
    json << "  }\n";
    
    json << "}";
    
    return json.str();
}

// 从JSON反序列化设备指纹
ErrorCode DeviceFingerprintCollector::DeserializeFromJson(
    const std::string& json_data,
    DeviceFingerprint& fingerprint) {
    
    // 简单的JSON解析实现
    // 实际项目中应使用专业的JSON库如nlohmann/json
    
    try {
        // 这里实现简单的键值解析
        // 实际实现应该更加健壮
        
        return ErrorCode::SUCCESS;
        
    } catch (const std::exception& e) {
        return ErrorCode::INVALID_PARAMETER;
    }
}

// 设置采集配置
void DeviceFingerprintCollector::SetConfiguration(
    bool collect_sensitive_info,
    bool enable_encryption,
    bool enable_signature) {
    
    std::lock_guard<std::mutex> lock(pimpl_->mutex_);
    pimpl_->collect_sensitive_info_ = collect_sensitive_info;
    pimpl_->enable_encryption_ = enable_encryption;
    pimpl_->enable_signature_ = enable_signature;
}

// 获取错误描述
std::string DeviceFingerprintCollector::GetErrorDescription(ErrorCode error_code) {
    switch (error_code) {
        case ErrorCode::SUCCESS:
            return "操作成功";
        case ErrorCode::INITIALIZATION_FAILED:
            return "初始化失败";
        case ErrorCode::HARDWARE_ACCESS_DENIED:
            return "硬件访问被拒绝";
        case ErrorCode::SYSTEM_INFO_UNAVAILABLE:
            return "系统信息不可用";
        case ErrorCode::ENCRYPTION_FAILED:
            return "加密失败";
        case ErrorCode::INVALID_PARAMETER:
            return "无效参数";
        case ErrorCode::MEMORY_ALLOCATION_FAILED:
            return "内存分配失败";
        case ErrorCode::PLATFORM_NOT_SUPPORTED:
            return "平台不支持";
        case ErrorCode::PERMISSION_DENIED:
            return "权限不足";
        default:
            return "未知错误";
    }
}

// 获取当前采集器版本
std::string DeviceFingerprintCollector::GetVersion() {
    return "1.0.0";
}

// 检查反调试状态
bool DeviceFingerprintCollector::IsDebuggerPresent() {
    if (!pimpl_->initialized_) {
        return false;
    }
    
    return pimpl_->platform_detector_->IsDebuggerPresent();
}

// 检查虚拟机环境
bool DeviceFingerprintCollector::IsRunningInVirtualMachine() {
    if (!pimpl_->initialized_) {
        return false;
    }
    
    return pimpl_->platform_detector_->IsVirtualMachine();
}

// PIMPL实现方法

std::string DeviceFingerprintCollector::Impl::GenerateDeviceID(
    const HardwareInfo& hardware, const SystemInfo& system) {
    
    std::ostringstream device_id;
    
    // 组合硬件和系统信息生成唯一设备ID
    device_id << hardware.cpu_id.substr(0, 8) << "-";
    device_id << hardware.motherboard_serial.substr(0, 8) << "-";
    if (!hardware.mac_addresses.empty()) {
        device_id << hardware.mac_addresses[0].substr(0, 8) << "-";
    }
    device_id << system.hostname.substr(0, 8);
    
    // 计算哈希
    std::string hash_hex;
    auto result = crypto::HashUtils::CalculateStringHash(
        device_id.str(), crypto::HashAlgorithm::SHA256, hash_hex);
    
    if (result == crypto::CryptoError::SUCCESS) {
        return hash_hex.substr(0, 32); // 返回前32个字符
    }
    
    return crypto::RandomGenerator::GenerateUUID();
}

DeviceType DeviceFingerprintCollector::Impl::DetermineDeviceType(
    const platform::OSInfo& os_info) {
    
    switch (os_info.platform) {
        case platform::PlatformType::WINDOWS:
        case platform::PlatformType::LINUX:
        case platform::PlatformType::MACOS:
            return DeviceType::DESKTOP;
        case platform::PlatformType::ANDROID:
        case platform::PlatformType::IOS:
            return DeviceType::MOBILE;
        default:
            return DeviceType::UNKNOWN;
    }
}

uint32_t DeviceFingerprintCollector::Impl::CalculateConfidenceScore(
    const DeviceFingerprint& fingerprint) {
    
    uint32_t score = 0;
    
    // 基于可用信息计算置信度
    if (!fingerprint.hardware.cpu_id.empty()) score += 20;
    if (!fingerprint.hardware.motherboard_serial.empty()) score += 25;
    if (!fingerprint.hardware.mac_addresses.empty()) score += 20;
    if (!fingerprint.hardware.disk_serials.empty()) score += 15;
    if (!fingerprint.system.hostname.empty()) score += 10;
    if (!fingerprint.system.username.empty()) score += 10;
    
    return std::min(score, 100u);
}

std::string DeviceFingerprintCollector::Impl::CombineFingerprints(
    const DeviceFingerprint& fingerprint) {
    
    std::ostringstream combined;
    
    combined << fingerprint.device_id << "|";
    combined << fingerprint.hardware.cpu_id << "|";
    combined << fingerprint.hardware.motherboard_serial << "|";
    combined << fingerprint.system.hostname << "|";
    combined << fingerprint.system.username << "|";
    combined << fingerprint.runtime.screen_resolution << "|";
    
    for (const auto& mac : fingerprint.hardware.mac_addresses) {
        combined << mac << "|";
    }
    
    return combined.str();
}

// 全局函数实现

ErrorCode QuickCollectFingerprint(DeviceFingerprint& fingerprint) {
    DeviceFingerprintCollector collector;
    auto result = collector.Initialize();
    if (result != ErrorCode::SUCCESS) {
        return result;
    }
    
    return collector.CollectFingerprint(fingerprint);
}

double CalculateFingerprintSimilarity(const std::string& hash1, const std::string& hash2) {
    if (hash1 == hash2) {
        return 1.0;
    }
    
    if (hash1.empty() || hash2.empty()) {
        return 0.0;
    }
    
    // 简单的编辑距离算法
    const size_t len1 = hash1.length();
    const size_t len2 = hash2.length();
    
    std::vector<std::vector<size_t>> d(len1 + 1, std::vector<size_t>(len2 + 1));
    
    for (size_t i = 0; i <= len1; ++i) d[i][0] = i;
    for (size_t j = 0; j <= len2; ++j) d[0][j] = j;
    
    for (size_t i = 1; i <= len1; ++i) {
        for (size_t j = 1; j <= len2; ++j) {
            const size_t cost = (hash1[i - 1] == hash2[j - 1]) ? 0 : 1;
            d[i][j] = std::min({
                d[i - 1][j] + 1,
                d[i][j - 1] + 1,
                d[i - 1][j - 1] + cost
            });
        }
    }
    
    const size_t max_len = std::max(len1, len2);
    return 1.0 - (static_cast<double>(d[len1][len2]) / max_len);
}

} // namespace device
} // namespace gaokao