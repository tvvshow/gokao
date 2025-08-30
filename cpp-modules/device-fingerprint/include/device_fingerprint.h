/**
 * @file device_fingerprint.h
 * @brief 设备指纹采集模块头文件
 * @author 高考志愿填报系统开发团队
 * @version 1.0.0
 * @date 2025-01-18
 * 
 * 提供跨平台的设备指纹采集功能，支持硬件信息采集、系统信息获取、
 * 运行时环境检测等功能，用于实现设备唯一性识别和安全验证。
 */

#ifndef DEVICE_FINGERPRINT_H
#define DEVICE_FINGERPRINT_H

#include <string>
#include <vector>
#include <map>
#include <memory>
#include <chrono>

namespace gaokao {
namespace device {

/**
 * @brief 错误码定义
 */
enum class ErrorCode {
    SUCCESS = 0,                    ///< 成功
    INITIALIZATION_FAILED = 1001,   ///< 初始化失败
    HARDWARE_ACCESS_DENIED = 1002,  ///< 硬件访问被拒绝
    SYSTEM_INFO_UNAVAILABLE = 1003, ///< 系统信息不可用
    ENCRYPTION_FAILED = 1004,       ///< 加密失败
    INVALID_PARAMETER = 1005,       ///< 无效参数
    MEMORY_ALLOCATION_FAILED = 1006,///< 内存分配失败
    PLATFORM_NOT_SUPPORTED = 1007, ///< 平台不支持
    PERMISSION_DENIED = 1008        ///< 权限不足
};

/**
 * @brief 设备类型枚举
 */
enum class DeviceType {
    DESKTOP,    ///< 桌面设备
    LAPTOP,     ///< 笔记本设备
    TABLET,     ///< 平板设备
    MOBILE,     ///< 移动设备
    SERVER,     ///< 服务器设备
    UNKNOWN     ///< 未知设备
};

/**
 * @brief 操作系统类型枚举
 */
enum class OperatingSystem {
    WINDOWS,    ///< Windows系统
    LINUX,      ///< Linux系统
    MACOS,      ///< macOS系统
    ANDROID,    ///< Android系统
    IOS,        ///< iOS系统
    UNKNOWN     ///< 未知系统
};

/**
 * @brief 硬件信息结构体
 */
struct HardwareInfo {
    std::string cpu_id;                 ///< CPU标识符
    std::string cpu_model;              ///< CPU型号
    uint32_t cpu_cores;                 ///< CPU核心数
    uint64_t total_memory;              ///< 总内存大小(字节)
    std::string motherboard_serial;     ///< 主板序列号
    std::string motherboard_manufacturer; ///< 主板制造商
    std::vector<std::string> mac_addresses; ///< MAC地址列表
    std::vector<std::string> disk_serials;  ///< 硬盘序列号列表
    std::string gpu_info;               ///< GPU信息
    std::string bios_serial;            ///< BIOS序列号
    std::string bios_version;           ///< BIOS版本
    
    HardwareInfo() : cpu_cores(0), total_memory(0) {}
};

/**
 * @brief 系统信息结构体
 */
struct SystemInfo {
    OperatingSystem os_type;            ///< 操作系统类型
    std::string os_version;             ///< 操作系统版本
    std::string os_build;               ///< 系统构建号
    std::string hostname;               ///< 主机名
    std::string username;               ///< 用户名
    std::string domain;                 ///< 域名
    std::string architecture;           ///< 系统架构
    std::string timezone;               ///< 时区
    std::string locale;                 ///< 语言环境
    uint64_t uptime;                    ///< 系统运行时间(秒)
    
    SystemInfo() : os_type(OperatingSystem::UNKNOWN), uptime(0) {}
};

/**
 * @brief 运行时信息结构体
 */
struct RuntimeInfo {
    std::string screen_resolution;      ///< 屏幕分辨率
    uint32_t color_depth;               ///< 颜色深度
    std::vector<std::string> fonts;     ///< 字体列表
    std::string browser_info;           ///< 浏览器信息
    std::string java_version;           ///< Java版本
    std::string dotnet_version;         ///< .NET版本
    std::map<std::string, std::string> environment_vars; ///< 环境变量
    std::chrono::system_clock::time_point timestamp; ///< 采集时间戳
    
    RuntimeInfo() : color_depth(0), timestamp(std::chrono::system_clock::now()) {}
};

/**
 * @brief 设备指纹结构体
 */
struct DeviceFingerprint {
    std::string device_id;              ///< 设备唯一标识
    DeviceType device_type;             ///< 设备类型
    HardwareInfo hardware;              ///< 硬件信息
    SystemInfo system;                  ///< 系统信息
    RuntimeInfo runtime;                ///< 运行时信息
    std::string fingerprint_hash;       ///< 指纹哈希值
    std::string encrypted_data;         ///< 加密数据
    std::string signature;              ///< 数字签名
    uint32_t confidence_score;          ///< 置信度分数(0-100)
    std::chrono::system_clock::time_point created_at; ///< 创建时间
    
    DeviceFingerprint() : device_type(DeviceType::UNKNOWN), 
                         confidence_score(0), 
                         created_at(std::chrono::system_clock::now()) {}
};

/**
 * @brief 指纹比较结果结构体
 */
struct ComparisonResult {
    double similarity_score;            ///< 相似度分数(0.0-1.0)
    bool is_same_device;               ///< 是否为同一设备
    std::vector<std::string> differences; ///< 差异列表
    uint32_t confidence_level;          ///< 置信度等级(0-100)
    
    ComparisonResult() : similarity_score(0.0), is_same_device(false), confidence_level(0) {}
};

/**
 * @brief 设备指纹采集器类
 * 
 * 主要功能：
 * - 跨平台硬件信息采集
 * - 系统信息获取
 * - 运行时环境检测
 * - 指纹生成和比较
 * - 数据加密和签名
 */
class DeviceFingerprintCollector {
public:
    /**
     * @brief 构造函数
     */
    DeviceFingerprintCollector();
    
    /**
     * @brief 析构函数
     */
    ~DeviceFingerprintCollector();
    
    /**
     * @brief 初始化采集器
     * @param config_path 配置文件路径(可选)
     * @return 错误码
     */
    ErrorCode Initialize(const std::string& config_path = "");
    
    /**
     * @brief 反初始化采集器
     */
    void Uninitialize();
    
    /**
     * @brief 采集完整设备指纹
     * @param fingerprint 输出的设备指纹信息
     * @return 错误码
     */
    ErrorCode CollectFingerprint(DeviceFingerprint& fingerprint);
    
    /**
     * @brief 采集硬件信息
     * @param hardware_info 输出的硬件信息
     * @return 错误码
     */
    ErrorCode CollectHardwareInfo(HardwareInfo& hardware_info);
    
    /**
     * @brief 采集系统信息
     * @param system_info 输出的系统信息
     * @return 错误码
     */
    ErrorCode CollectSystemInfo(SystemInfo& system_info);
    
    /**
     * @brief 采集运行时信息
     * @param runtime_info 输出的运行时信息
     * @return 错误码
     */
    ErrorCode CollectRuntimeInfo(RuntimeInfo& runtime_info);
    
    /**
     * @brief 生成设备指纹哈希
     * @param fingerprint 设备指纹信息
     * @return 指纹哈希字符串
     */
    std::string GenerateFingerprintHash(const DeviceFingerprint& fingerprint);
    
    /**
     * @brief 比较两个设备指纹
     * @param fingerprint1 第一个设备指纹
     * @param fingerprint2 第二个设备指纹
     * @param result 比较结果
     * @return 错误码
     */
    ErrorCode CompareFingerprints(const DeviceFingerprint& fingerprint1,
                                 const DeviceFingerprint& fingerprint2,
                                 ComparisonResult& result);
    
    /**
     * @brief 验证设备指纹
     * @param fingerprint 设备指纹信息
     * @param reference_hash 参考哈希值
     * @return 是否验证通过
     */
    bool ValidateFingerprint(const DeviceFingerprint& fingerprint,
                           const std::string& reference_hash);
    
    /**
     * @brief 序列化设备指纹为JSON
     * @param fingerprint 设备指纹信息
     * @return JSON字符串
     */
    std::string SerializeToJson(const DeviceFingerprint& fingerprint);
    
    /**
     * @brief 从JSON反序列化设备指纹
     * @param json_data JSON字符串
     * @param fingerprint 输出的设备指纹信息
     * @return 错误码
     */
    ErrorCode DeserializeFromJson(const std::string& json_data,
                                 DeviceFingerprint& fingerprint);
    
    /**
     * @brief 设置采集配置
     * @param collect_sensitive_info 是否采集敏感信息
     * @param enable_encryption 是否启用加密
     * @param enable_signature 是否启用签名
     */
    void SetConfiguration(bool collect_sensitive_info = true,
                         bool enable_encryption = true,
                         bool enable_signature = true);
    
    /**
     * @brief 获取错误描述
     * @param error_code 错误码
     * @return 错误描述字符串
     */
    static std::string GetErrorDescription(ErrorCode error_code);
    
    /**
     * @brief 获取当前采集器版本
     * @return 版本字符串
     */
    static std::string GetVersion();
    
    /**
     * @brief 检查反调试状态
     * @return 是否检测到调试器
     */
    bool IsDebuggerPresent();
    
    /**
     * @brief 检查虚拟机环境
     * @return 是否运行在虚拟机中
     */
    bool IsRunningInVirtualMachine();

private:
    class Impl;
    std::unique_ptr<Impl> pimpl_;       ///< PIMPL实现指针
    
    // 禁用拷贝构造和赋值
    DeviceFingerprintCollector(const DeviceFingerprintCollector&) = delete;
    DeviceFingerprintCollector& operator=(const DeviceFingerprintCollector&) = delete;
};

/**
 * @brief 全局函数：快速采集设备指纹
 * @param fingerprint 输出的设备指纹信息
 * @return 错误码
 */
ErrorCode QuickCollectFingerprint(DeviceFingerprint& fingerprint);

/**
 * @brief 全局函数：比较设备指纹相似度
 * @param hash1 第一个指纹哈希
 * @param hash2 第二个指纹哈希
 * @return 相似度分数(0.0-1.0)
 */
double CalculateFingerprintSimilarity(const std::string& hash1, 
                                     const std::string& hash2);

} // namespace device
} // namespace gaokao

#endif // DEVICE_FINGERPRINT_H