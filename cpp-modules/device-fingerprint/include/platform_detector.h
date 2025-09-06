/**
 * @file platform_detector.h
 * @brief 平台检测和系统信息采集头文件
 * @author 高考志愿填报系统开发团队
 * @version 1.0.0
 * @date 2025-01-18
 * 
 * 提供跨平台的系统信息检测功能，包括操作系统检测、硬件信息采集、
 * 安全环境检测等功能，支持Windows、Linux、macOS等主流平台。
 */

#ifndef PLATFORM_DETECTOR_H
#define PLATFORM_DETECTOR_H

#include <string>
#include <vector>
#include <map>
#include <memory>
#include <cstdint>

namespace gaokao {
namespace platform {

/**
 * @brief 平台类型枚举
 */
enum class PlatformType {
    WINDOWS,        ///< Windows平台
    LINUX,          ///< Linux平台
    MACOS,          ///< macOS平台
    FREEBSD,        ///< FreeBSD平台
    ANDROID,        ///< Android平台
    IOS,            ///< iOS平台
    UNKNOWN         ///< 未知平台
};

/**
 * @brief CPU架构枚举
 */
enum class Architecture {
    X86,            ///< x86 32位
    X64,            ///< x86 64位
    ARM32,          ///< ARM 32位
    ARM64,          ///< ARM 64位
    MIPS,           ///< MIPS架构
    POWERPC,        ///< PowerPC架构
    SPARC,          ///< SPARC架构
    UNKNOWN         ///< 未知架构
};

/**
 * @brief 检测错误码
 */
enum class DetectionError {
    SUCCESS = 0,                    ///< 成功
    PLATFORM_NOT_SUPPORTED = 3001, ///< 平台不支持
    ACCESS_DENIED = 3002,           ///< 访问被拒绝
    COMMAND_EXECUTION_FAILED = 3003,///< 命令执行失败
    REGISTRY_ACCESS_FAILED = 3004,  ///< 注册表访问失败
    SYSFS_ACCESS_FAILED = 3005,     ///< sysfs访问失败
    WMI_QUERY_FAILED = 3006,        ///< WMI查询失败
    IOCTL_FAILED = 3007,            ///< ioctl操作失败
    BUFFER_TOO_SMALL = 3008,        ///< 缓冲区太小
    INVALID_DATA = 3009             ///< 无效数据
};

/**
 * @brief CPU信息结构体
 */
struct CPUInfo {
    std::string vendor;             ///< CPU厂商
    std::string model_name;         ///< CPU型号名称
    std::string identifier;         ///< CPU标识符
    uint32_t physical_cores;        ///< 物理核心数
    uint32_t logical_cores;         ///< 逻辑核心数
    uint32_t cache_size;            ///< 缓存大小(KB)
    uint64_t frequency;             ///< 频率(Hz)
    std::vector<std::string> features; ///< CPU特性列表
    Architecture architecture;      ///< CPU架构
    std::string serial_number;      ///< 序列号(如果可用)
    
    CPUInfo() : physical_cores(0), logical_cores(0), cache_size(0), 
               frequency(0), architecture(Architecture::UNKNOWN) {}
};

/**
 * @brief 内存信息结构体
 */
struct MemoryInfo {
    uint64_t total_physical;        ///< 总物理内存(字节)
    uint64_t available_physical;    ///< 可用物理内存(字节)
    uint64_t total_virtual;         ///< 总虚拟内存(字节)
    uint64_t available_virtual;     ///< 可用虚拟内存(字节)
    uint32_t memory_load;           ///< 内存使用率(百分比)
    std::string memory_type;        ///< 内存类型(DDR3/DDR4等)
    uint32_t memory_speed;          ///< 内存频率(MHz)
    std::vector<std::string> memory_modules; ///< 内存模块信息
    
    MemoryInfo() : total_physical(0), available_physical(0), 
                  total_virtual(0), available_virtual(0), 
                  memory_load(0), memory_speed(0) {}
};

/**
 * @brief 主板信息结构体
 */
struct MotherboardInfo {
    std::string manufacturer;       ///< 制造商
    std::string product_name;       ///< 产品名称
    std::string version;            ///< 版本
    std::string serial_number;      ///< 序列号
    std::string asset_tag;          ///< 资产标签
    std::string uuid;               ///< UUID
    std::map<std::string, std::string> additional_info; ///< 附加信息
    
    MotherboardInfo() {}
};

/**
 * @brief BIOS信息结构体
 */
struct BIOSInfo {
    std::string vendor;             ///< BIOS厂商
    std::string version;            ///< BIOS版本
    std::string release_date;       ///< 发布日期
    std::string serial_number;      ///< 序列号
    std::string description;        ///< 描述
    uint32_t bios_size;             ///< BIOS大小(KB)
    std::vector<std::string> characteristics; ///< BIOS特性
    
    BIOSInfo() : bios_size(0) {}
};

/**
 * @brief 网络适配器信息结构体
 */
struct NetworkAdapterInfo {
    std::string name;               ///< 适配器名称
    std::string description;        ///< 描述
    std::string mac_address;        ///< MAC地址
    std::string adapter_type;       ///< 适配器类型
    bool is_physical;               ///< 是否为物理适配器
    bool is_connected;              ///< 是否已连接
    uint64_t speed;                 ///< 连接速度(bps)
    std::vector<std::string> ip_addresses; ///< IP地址列表
    
    NetworkAdapterInfo() : is_physical(false), is_connected(false), speed(0) {}
};

/**
 * @brief 存储设备信息结构体
 */
struct StorageDeviceInfo {
    std::string model;              ///< 设备型号
    std::string serial_number;      ///< 序列号
    std::string interface_type;     ///< 接口类型(SATA/NVMe等)
    std::string device_type;        ///< 设备类型(HDD/SSD)
    uint64_t total_size;            ///< 总容量(字节)
    uint64_t free_space;            ///< 可用空间(字节)
    std::string firmware_version;   ///< 固件版本
    bool is_removable;              ///< 是否为可移动设备
    uint32_t rotation_rate;         ///< 转速(RPM，仅HDD)
    
    StorageDeviceInfo() : total_size(0), free_space(0), 
                         is_removable(false), rotation_rate(0) {}
};

/**
 * @brief 显示设备信息结构体
 */
struct DisplayDeviceInfo {
    std::string name;               ///< 显示设备名称
    std::string description;        ///< 描述
    std::string driver_version;     ///< 驱动版本
    uint32_t width;                 ///< 屏幕宽度
    uint32_t height;                ///< 屏幕高度
    uint32_t color_depth;           ///< 颜色深度
    uint32_t refresh_rate;          ///< 刷新率(Hz)
    bool is_primary;                ///< 是否为主显示器
    
    DisplayDeviceInfo() : width(0), height(0), color_depth(0), 
                         refresh_rate(0), is_primary(false) {}
};

/**
 * @brief 操作系统信息结构体
 */
struct OSInfo {
    PlatformType platform;          ///< 平台类型
    std::string name;               ///< 操作系统名称
    std::string version;            ///< 版本
    std::string build_number;       ///< 构建号
    std::string service_pack;       ///< 服务包
    std::string edition;            ///< 版本(家庭版/专业版等)
    Architecture architecture;      ///< 系统架构
    std::string kernel_version;     ///< 内核版本
    std::string install_date;       ///< 安装日期
    uint64_t uptime;                ///< 运行时间(秒)
    
    OSInfo() : platform(PlatformType::UNKNOWN), 
              architecture(Architecture::UNKNOWN), uptime(0) {}
};

/**
 * @brief 用户信息结构体
 */
struct UserInfo {
    std::string username;           ///< 用户名
    std::string full_name;          ///< 全名
    std::string domain;             ///< 域名
    std::string home_directory;     ///< 主目录
    std::string shell;              ///< Shell
    std::vector<std::string> groups; ///< 用户组列表
    bool is_admin;                  ///< 是否为管理员
    std::string last_login;         ///< 最后登录时间
    
    UserInfo() : is_admin(false) {}
};

/**
 * @brief 环境信息结构体
 */
struct EnvironmentInfo {
    std::string hostname;           ///< 主机名
    std::string timezone;           ///< 时区
    std::string locale;             ///< 语言环境
    std::string keyboard_layout;    ///< 键盘布局
    std::map<std::string, std::string> environment_variables; ///< 环境变量
    std::vector<std::string> installed_software; ///< 已安装软件列表
    
    EnvironmentInfo() {}
};

/**
 * @brief 安全环境检测结果结构体
 */
struct SecurityEnvironment {
    bool is_debugger_present;       ///< 是否存在调试器
    bool is_virtual_machine;        ///< 是否为虚拟机
    bool is_sandboxed;              ///< 是否在沙箱中运行
    bool has_antivirus;             ///< 是否有杀毒软件
    std::vector<std::string> security_products; ///< 安全产品列表
    std::vector<std::string> suspicious_processes; ///< 可疑进程列表
    
    SecurityEnvironment() : is_debugger_present(false), is_virtual_machine(false),
                           is_sandboxed(false), has_antivirus(false) {}
};

/**
 * @brief 平台检测器类
 */
class PlatformDetector {
public:
    /**
     * @brief 构造函数
     */
    PlatformDetector();
    
    /**
     * @brief 析构函数
     */
    ~PlatformDetector();
    
    /**
     * @brief 初始化检测器
     * @return 错误码
     */
    DetectionError Initialize();
    
    /**
     * @brief 反初始化检测器
     */
    void Uninitialize();
    
    /**
     * @brief 检测当前平台类型
     * @return 平台类型
     */
    PlatformType DetectPlatform();
    
    /**
     * @brief 检测CPU架构
     * @return CPU架构
     */
    Architecture DetectArchitecture();
    
    /**
     * @brief 获取CPU信息
     * @param cpu_info 输出的CPU信息
     * @return 错误码
     */
    DetectionError GetCPUInfo(CPUInfo& cpu_info);
    
    /**
     * @brief 获取内存信息
     * @param memory_info 输出的内存信息
     * @return 错误码
     */
    DetectionError GetMemoryInfo(MemoryInfo& memory_info);
    
    /**
     * @brief 获取主板信息
     * @param motherboard_info 输出的主板信息
     * @return 错误码
     */
    DetectionError GetMotherboardInfo(MotherboardInfo& motherboard_info);
    
    /**
     * @brief 获取BIOS信息
     * @param bios_info 输出的BIOS信息
     * @return 错误码
     */
    DetectionError GetBIOSInfo(BIOSInfo& bios_info);
    
    /**
     * @brief 获取网络适配器信息
     * @param adapters 输出的网络适配器信息列表
     * @return 错误码
     */
    DetectionError GetNetworkAdapters(std::vector<NetworkAdapterInfo>& adapters);
    
    /**
     * @brief 获取存储设备信息
     * @param devices 输出的存储设备信息列表
     * @return 错误码
     */
    DetectionError GetStorageDevices(std::vector<StorageDeviceInfo>& devices);
    
    /**
     * @brief 获取显示设备信息
     * @param displays 输出的显示设备信息列表
     * @return 错误码
     */
    DetectionError GetDisplayDevices(std::vector<DisplayDeviceInfo>& displays);
    
    /**
     * @brief 获取操作系统信息
     * @param os_info 输出的操作系统信息
     * @return 错误码
     */
    DetectionError GetOSInfo(OSInfo& os_info);
    
    /**
     * @brief 获取用户信息
     * @param user_info 输出的用户信息
     * @return 错误码
     */
    DetectionError GetUserInfo(UserInfo& user_info);
    
    /**
     * @brief 获取环境信息
     * @param env_info 输出的环境信息
     * @return 错误码
     */
    DetectionError GetEnvironmentInfo(EnvironmentInfo& env_info);
    
    /**
     * @brief 检测安全环境
     * @param security_env 输出的安全环境信息
     * @return 错误码
     */
    DetectionError DetectSecurityEnvironment(SecurityEnvironment& security_env);
    
    /**
     * @brief 检测虚拟机环境
     * @return 是否为虚拟机环境
     */
    bool IsVirtualMachine();
    
    /**
     * @brief 检测调试器
     * @return 是否存在调试器
     */
    bool IsDebuggerPresent();
    
    /**
     * @brief 检测沙箱环境
     * @return 是否在沙箱中运行
     */
    bool IsSandboxed();
    
    /**
     * @brief 获取系统运行时间
     * @return 运行时间(秒)
     */
    uint64_t GetSystemUptime();
    
    /**
     * @brief 获取系统安装日期
     * @return 安装日期字符串
     */
    std::string GetSystemInstallDate();
    
    /**
     * @brief 执行系统命令并获取输出
     * @param command 命令
     * @param output 输出结果
     * @return 是否执行成功
     */
    bool ExecuteCommand(const std::string& command, std::string& output);
    
    /**
     * @brief 获取错误描述
     * @param error 错误码
     * @return 错误描述字符串
     */
    static std::string GetErrorDescription(DetectionError error);

private:
    class Impl;
    std::unique_ptr<Impl> pimpl_;
    
    // 禁用拷贝构造和赋值
    PlatformDetector(const PlatformDetector&) = delete;
    PlatformDetector& operator=(const PlatformDetector&) = delete;
};

/**
 * @brief 全局函数：快速检测平台类型
 * @return 平台类型
 */
PlatformType QuickDetectPlatform();

/**
 * @brief 全局函数：快速检测CPU架构
 * @return CPU架构
 */
Architecture QuickDetectArchitecture();

/**
 * @brief 全局函数：检查是否为虚拟机环境
 * @return 是否为虚拟机环境
 */
bool QuickCheckVirtualMachine();

/**
 * @brief 全局函数：检查是否存在调试器
 * @return 是否存在调试器
 */
bool QuickCheckDebugger();

} // namespace platform
} // namespace gaokao

#endif // PLATFORM_DETECTOR_H