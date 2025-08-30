/**
 * @file platform_detector.cpp
 * @brief 平台检测实现
 * @author 高考志愿填报系统开发团队
 * @version 1.0.0
 * @date 2025-01-18
 */

#include "platform_detector.h"
#include <iostream>
#include <sstream>
#include <fstream>
#include <cstring>
#include <algorithm>

// 平台特定头文件
#ifdef _WIN32
    #include <windows.h>
    #include <wbemidl.h>
    #include <comdef.h>
    #include <iphlpapi.h>
    #include <psapi.h>
    #include <tlhelp32.h>
    #include <intrin.h>
    #pragma comment(lib, "wbemuuid.lib")
    #pragma comment(lib, "iphlpapi.lib")
    #pragma comment(lib, "psapi.lib")
#elif defined(__linux__)
    #include <unistd.h>
    #include <sys/utsname.h>
    #include <sys/sysinfo.h>
    #include <sys/types.h>
    #include <sys/stat.h>
    #include <ifaddrs.h>
    #include <net/if.h>
    #include <netinet/in.h>
    #include <arpa/inet.h>
    #include <dirent.h>
    #include <pwd.h>
    #include <grp.h>
    #include <cpuid.h>
#elif defined(__APPLE__)
    #include <unistd.h>
    #include <sys/utsname.h>
    #include <sys/sysctl.h>
    #include <ifaddrs.h>
    #include <net/if.h>
    #include <IOKit/IOKitLib.h>
    #include <CoreFoundation/CoreFoundation.h>
#endif

namespace gaokao {
namespace platform {

// PIMPL实现类
class PlatformDetector::Impl {
public:
    Impl() : initialized_(false) {}
    ~Impl() = default;
    
    bool initialized_;
    
    // 内部辅助方法
    std::string ReadFile(const std::string& path);
    std::string ExecuteCommand(const std::string& command);
    std::vector<std::string> SplitString(const std::string& str, char delimiter);
    std::string TrimString(const std::string& str);
    
#ifdef _WIN32
    std::string GetWMIProperty(const std::string& query, const std::string& property);
    bool IsUserAdmin();
#endif

#ifdef __linux__
    std::string GetProcInfo(const std::string& file);
    std::string GetSysfsInfo(const std::string& path);
#endif
};

// 构造函数
PlatformDetector::PlatformDetector() : pimpl_(std::make_unique<Impl>()) {}

// 析构函数
PlatformDetector::~PlatformDetector() {
    if (pimpl_ && pimpl_->initialized_) {
        Uninitialize();
    }
}

// 初始化检测器
DetectionError PlatformDetector::Initialize() {
    if (pimpl_->initialized_) {
        return DetectionError::SUCCESS;
    }
    
    try {
#ifdef _WIN32
        // 初始化COM
        HRESULT hr = CoInitializeEx(0, COINIT_MULTITHREADED);
        if (FAILED(hr) && hr != RPC_E_CHANGED_MODE) {
            return DetectionError::PLATFORM_NOT_SUPPORTED;
        }
#endif
        
        pimpl_->initialized_ = true;
        return DetectionError::SUCCESS;
        
    } catch (const std::exception& e) {
        return DetectionError::PLATFORM_NOT_SUPPORTED;
    }
}

// 反初始化检测器
void PlatformDetector::Uninitialize() {
#ifdef _WIN32
    if (pimpl_->initialized_) {
        CoUninitialize();
    }
#endif
    pimpl_->initialized_ = false;
}

// 检测当前平台类型
PlatformType PlatformDetector::DetectPlatform() {
#ifdef _WIN32
    return PlatformType::WINDOWS;
#elif defined(__linux__)
    return PlatformType::LINUX;
#elif defined(__APPLE__)
    return PlatformType::MACOS;
#elif defined(__FreeBSD__)
    return PlatformType::FREEBSD;
#else
    return PlatformType::UNKNOWN;
#endif
}

// 检测CPU架构
Architecture PlatformDetector::DetectArchitecture() {
#ifdef _WIN32
    SYSTEM_INFO sysinfo;
    GetSystemInfo(&sysinfo);
    
    switch (sysinfo.wProcessorArchitecture) {
        case PROCESSOR_ARCHITECTURE_AMD64:
            return Architecture::X64;
        case PROCESSOR_ARCHITECTURE_INTEL:
            return Architecture::X86;
        case PROCESSOR_ARCHITECTURE_ARM:
            return Architecture::ARM32;
        case PROCESSOR_ARCHITECTURE_ARM64:
            return Architecture::ARM64;
        default:
            return Architecture::UNKNOWN;
    }
#elif defined(__linux__) || defined(__APPLE__)
    struct utsname info;
    if (uname(&info) == 0) {
        std::string machine = info.machine;
        if (machine == "x86_64" || machine == "amd64") {
            return Architecture::X64;
        } else if (machine == "i386" || machine == "i686") {
            return Architecture::X86;
        } else if (machine == "armv7l" || machine == "armv6l") {
            return Architecture::ARM32;
        } else if (machine == "aarch64" || machine == "arm64") {
            return Architecture::ARM64;
        }
    }
#endif
    return Architecture::UNKNOWN;
}

// 获取CPU信息
DetectionError PlatformDetector::GetCPUInfo(CPUInfo& cpu_info) {
    if (!pimpl_->initialized_) {
        return DetectionError::PLATFORM_NOT_SUPPORTED;
    }
    
    try {
#ifdef _WIN32
        // 使用WMI获取CPU信息
        cpu_info.vendor = pimpl_->GetWMIProperty("SELECT * FROM Win32_Processor", "Manufacturer");
        cpu_info.model_name = pimpl_->GetWMIProperty("SELECT * FROM Win32_Processor", "Name");
        
        // 获取CPU核心数
        SYSTEM_INFO sysinfo;
        GetSystemInfo(&sysinfo);
        cpu_info.logical_cores = sysinfo.dwNumberOfProcessors;
        
        // 获取物理核心数(简化实现)
        cpu_info.physical_cores = cpu_info.logical_cores / 2; // 假设支持超线程
        
        // 获取CPU标识符
        int cpuInfo[4];
        __cpuid(cpuInfo, 0);
        std::ostringstream oss;
        oss << std::hex << cpuInfo[1] << cpuInfo[3] << cpuInfo[2];
        cpu_info.identifier = oss.str();
        
#elif defined(__linux__)
        // 读取/proc/cpuinfo
        std::string cpuinfo = pimpl_->GetProcInfo("/proc/cpuinfo");
        
        // 解析CPU信息
        std::istringstream iss(cpuinfo);
        std::string line;
        
        while (std::getline(iss, line)) {
            if (line.find("vendor_id") != std::string::npos) {
                size_t pos = line.find(':');
                if (pos != std::string::npos) {
                    cpu_info.vendor = pimpl_->TrimString(line.substr(pos + 1));
                }
            } else if (line.find("model name") != std::string::npos) {
                size_t pos = line.find(':');
                if (pos != std::string::npos) {
                    cpu_info.model_name = pimpl_->TrimString(line.substr(pos + 1));
                }
            } else if (line.find("processor") != std::string::npos) {
                cpu_info.logical_cores++;
            }
        }
        
        // 获取物理核心数
        std::string core_id_file = "/sys/devices/system/cpu/cpu0/topology/core_id";
        if (access(core_id_file.c_str(), R_OK) == 0) {
            // 简化实现：物理核心数 = 逻辑核心数 / 2
            cpu_info.physical_cores = cpu_info.logical_cores / 2;
        }
        
#elif defined(__APPLE__)
        // 使用sysctl获取CPU信息
        size_t size = 0;
        
        // 获取CPU品牌
        sysctlbyname("machdep.cpu.brand_string", NULL, &size, NULL, 0);
        if (size > 0) {
            char* brand = new char[size];
            sysctlbyname("machdep.cpu.brand_string", brand, &size, NULL, 0);
            cpu_info.model_name = brand;
            delete[] brand;
        }
        
        // 获取CPU厂商
        sysctlbyname("machdep.cpu.vendor", NULL, &size, NULL, 0);
        if (size > 0) {
            char* vendor = new char[size];
            sysctlbyname("machdep.cpu.vendor", vendor, &size, NULL, 0);
            cpu_info.vendor = vendor;
            delete[] vendor;
        }
        
        // 获取核心数
        int logical_cores = 0;
        size = sizeof(logical_cores);
        sysctlbyname("hw.logicalcpu", &logical_cores, &size, NULL, 0);
        cpu_info.logical_cores = logical_cores;
        
        int physical_cores = 0;
        size = sizeof(physical_cores);
        sysctlbyname("hw.physicalcpu", &physical_cores, &size, NULL, 0);
        cpu_info.physical_cores = physical_cores;
#endif
        
        cpu_info.architecture = DetectArchitecture();
        return DetectionError::SUCCESS;
        
    } catch (const std::exception& e) {
        return DetectionError::COMMAND_EXECUTION_FAILED;
    }
}

// 获取内存信息
DetectionError PlatformDetector::GetMemoryInfo(MemoryInfo& memory_info) {
    if (!pimpl_->initialized_) {
        return DetectionError::PLATFORM_NOT_SUPPORTED;
    }
    
    try {
#ifdef _WIN32
        MEMORYSTATUSEX statex;
        statex.dwLength = sizeof(statex);
        
        if (GlobalMemoryStatusEx(&statex)) {
            memory_info.total_physical = statex.ullTotalPhys;
            memory_info.available_physical = statex.ullAvailPhys;
            memory_info.total_virtual = statex.ullTotalVirtual;
            memory_info.available_virtual = statex.ullAvailVirtual;
            memory_info.memory_load = statex.dwMemoryLoad;
        }
        
#elif defined(__linux__)
        struct sysinfo si;
        if (sysinfo(&si) == 0) {
            memory_info.total_physical = si.totalram * si.mem_unit;
            memory_info.available_physical = si.freeram * si.mem_unit;
            memory_info.total_virtual = si.totalswap * si.mem_unit;
            memory_info.available_virtual = si.freeswap * si.mem_unit;
            
            if (memory_info.total_physical > 0) {
                memory_info.memory_load = 
                    ((memory_info.total_physical - memory_info.available_physical) * 100) 
                    / memory_info.total_physical;
            }
        }
        
#elif defined(__APPLE__)
        // 获取物理内存
        int64_t total_memory = 0;
        size_t size = sizeof(total_memory);
        sysctlbyname("hw.memsize", &total_memory, &size, NULL, 0);
        memory_info.total_physical = total_memory;
        
        // 获取可用内存(简化实现)
        vm_size_t page_size;
        vm_statistics64_data_t vm_stat;
        mach_msg_type_number_t host_size = sizeof(vm_stat) / sizeof(uint32_t);
        
        host_page_size(mach_host_self(), &page_size);
        host_statistics64(mach_host_self(), HOST_VM_INFO64, 
                         (host_info64_t)&vm_stat, &host_size);
        
        memory_info.available_physical = 
            (vm_stat.free_count + vm_stat.inactive_count) * page_size;
#endif
        
        return DetectionError::SUCCESS;
        
    } catch (const std::exception& e) {
        return DetectionError::COMMAND_EXECUTION_FAILED;
    }
}

// 获取操作系统信息
DetectionError PlatformDetector::GetOSInfo(OSInfo& os_info) {
    if (!pimpl_->initialized_) {
        return DetectionError::PLATFORM_NOT_SUPPORTED;
    }
    
    try {
        os_info.platform = DetectPlatform();
        os_info.architecture = DetectArchitecture();
        
#ifdef _WIN32
        // 获取Windows版本信息
        OSVERSIONINFOEX osvi;
        ZeroMemory(&osvi, sizeof(OSVERSIONINFOEX));
        osvi.dwOSVersionInfoSize = sizeof(OSVERSIONINFOEX);
        
        if (GetVersionEx((OSVERSIONINFO*)&osvi)) {
            std::ostringstream version;
            version << osvi.dwMajorVersion << "." << osvi.dwMinorVersion;
            os_info.version = version.str();
            
            std::ostringstream build;
            build << osvi.dwBuildNumber;
            os_info.build_number = build.str();
        }
        
        os_info.name = "Windows";
        
        // 获取主机名
        char hostname[MAX_COMPUTERNAME_LENGTH + 1];
        DWORD size = sizeof(hostname);
        if (GetComputerNameA(hostname, &size)) {
            os_info.name = hostname;
        }
        
#elif defined(__linux__)
        struct utsname info;
        if (uname(&info) == 0) {
            os_info.name = info.sysname;
            os_info.version = info.release;
            os_info.kernel_version = info.version;
        }
        
        // 尝试读取发行版信息
        std::string release_info = pimpl_->ReadFile("/etc/os-release");
        if (!release_info.empty()) {
            // 解析os-release文件
            std::istringstream iss(release_info);
            std::string line;
            while (std::getline(iss, line)) {
                if (line.find("PRETTY_NAME=") == 0) {
                    os_info.name = line.substr(12);
                    // 移除引号
                    if (os_info.name.front() == '"' && os_info.name.back() == '"') {
                        os_info.name = os_info.name.substr(1, os_info.name.length() - 2);
                    }
                    break;
                }
            }
        }
        
#elif defined(__APPLE__)
        struct utsname info;
        if (uname(&info) == 0) {
            os_info.name = "macOS";
            os_info.kernel_version = info.version;
        }
        
        // 获取macOS版本
        int major = 0, minor = 0, patch = 0;
        size_t size = sizeof(major);
        
        sysctlbyname("kern.osproductversion", NULL, &size, NULL, 0);
        if (size > 0) {
            char* version = new char[size];
            sysctlbyname("kern.osproductversion", version, &size, NULL, 0);
            os_info.version = version;
            delete[] version;
        }
#endif
        
        // 获取系统运行时间
        os_info.uptime = GetSystemUptime();
        
        return DetectionError::SUCCESS;
        
    } catch (const std::exception& e) {
        return DetectionError::COMMAND_EXECUTION_FAILED;
    }
}

// 检测虚拟机环境
bool PlatformDetector::IsVirtualMachine() {
    try {
#ifdef _WIN32
        // 检查常见虚拟机厂商
        std::string manufacturer = pimpl_->GetWMIProperty("SELECT * FROM Win32_ComputerSystem", "Manufacturer");
        std::string model = pimpl_->GetWMIProperty("SELECT * FROM Win32_ComputerSystem", "Model");
        
        std::vector<std::string> vm_indicators = {
            "VMware", "VirtualBox", "Microsoft Corporation", "QEMU",
            "Xen", "innotek", "Parallels", "Red Hat"
        };
        
        for (const auto& indicator : vm_indicators) {
            if (manufacturer.find(indicator) != std::string::npos ||
                model.find(indicator) != std::string::npos) {
                return true;
            }
        }
        
        // 检查BIOS信息
        std::string bios_version = pimpl_->GetWMIProperty("SELECT * FROM Win32_BIOS", "Version");
        if (bios_version.find("VBOX") != std::string::npos ||
            bios_version.find("VMware") != std::string::npos) {
            return true;
        }
        
#elif defined(__linux__)
        // 检查DMI信息
        std::string dmi_sys_vendor = pimpl_->GetSysfsInfo("/sys/class/dmi/id/sys_vendor");
        std::string dmi_product_name = pimpl_->GetSysfsInfo("/sys/class/dmi/id/product_name");
        
        std::vector<std::string> vm_indicators = {
            "VMware", "VirtualBox", "QEMU", "Microsoft Corporation",
            "Xen", "innotek", "Parallels", "Red Hat"
        };
        
        for (const auto& indicator : vm_indicators) {
            if (dmi_sys_vendor.find(indicator) != std::string::npos ||
                dmi_product_name.find(indicator) != std::string::npos) {
                return true;
            }
        }
        
        // 检查hypervisor cpuid位
        if (access("/proc/cpuinfo", R_OK) == 0) {
            std::string cpuinfo = pimpl_->GetProcInfo("/proc/cpuinfo");
            if (cpuinfo.find("hypervisor") != std::string::npos) {
                return true;
            }
        }
        
#elif defined(__APPLE__)
        // macOS上的虚拟机检测
        std::string model;
        size_t size = 0;
        
        sysctlbyname("hw.model", NULL, &size, NULL, 0);
        if (size > 0) {
            char* model_str = new char[size];
            sysctlbyname("hw.model", model_str, &size, NULL, 0);
            model = model_str;
            delete[] model_str;
        }
        
        if (model.find("VMware") != std::string::npos ||
            model.find("VirtualBox") != std::string::npos ||
            model.find("Parallels") != std::string::npos) {
            return true;
        }
#endif
        
        return false;
        
    } catch (const std::exception& e) {
        return false;
    }
}

// 检测调试器
bool PlatformDetector::IsDebuggerPresent() {
    try {
#ifdef _WIN32
        // Windows调试器检测
        if (::IsDebuggerPresent()) {
            return true;
        }
        
        // 检查远程调试器
        BOOL remote_debugger = FALSE;
        CheckRemoteDebuggerPresent(GetCurrentProcess(), &remote_debugger);
        if (remote_debugger) {
            return true;
        }
        
        // 检查调试标志
        PPEB peb = (PPEB)__readgsqword(0x60);
        if (peb->BeingDebugged) {
            return true;
        }
        
#elif defined(__linux__)
        // Linux调试器检测
        // 检查TracerPid
        std::string status = pimpl_->GetProcInfo("/proc/self/status");
        if (status.find("TracerPid:\t0") == std::string::npos) {
            return true;
        }
        
        // 检查ptrace
        if (ptrace(PTRACE_TRACEME, 0, 1, 0) == -1) {
            return true;
        }
        ptrace(PTRACE_DETACH, 0, 1, 0);
        
#elif defined(__APPLE__)
        // macOS调试器检测
        int mib[4];
        struct kinfo_proc info;
        size_t size = sizeof(info);
        
        mib[0] = CTL_KERN;
        mib[1] = KERN_PROC;
        mib[2] = KERN_PROC_PID;
        mib[3] = getpid();
        
        if (sysctl(mib, 4, &info, &size, NULL, 0) == 0) {
            return (info.kp_proc.p_flag & P_TRACED) != 0;
        }
#endif
        
        return false;
        
    } catch (const std::exception& e) {
        return false;
    }
}

// 获取系统运行时间
uint64_t PlatformDetector::GetSystemUptime() {
    try {
#ifdef _WIN32
        return GetTickCount64() / 1000; // 转换为秒
        
#elif defined(__linux__)
        struct sysinfo si;
        if (sysinfo(&si) == 0) {
            return si.uptime;
        }
        
#elif defined(__APPLE__)
        struct timeval boottime;
        size_t size = sizeof(boottime);
        
        if (sysctlbyname("kern.boottime", &boottime, &size, NULL, 0) == 0) {
            time_t now;
            time(&now);
            return now - boottime.tv_sec;
        }
#endif
        
        return 0;
        
    } catch (const std::exception& e) {
        return 0;
    }
}

// 辅助方法实现

std::string PlatformDetector::Impl::ReadFile(const std::string& path) {
    std::ifstream file(path);
    if (!file.is_open()) {
        return "";
    }
    
    std::ostringstream content;
    content << file.rdbuf();
    return content.str();
}

std::string PlatformDetector::Impl::TrimString(const std::string& str) {
    size_t start = str.find_first_not_of(" \t\n\r");
    if (start == std::string::npos) {
        return "";
    }
    
    size_t end = str.find_last_not_of(" \t\n\r");
    return str.substr(start, end - start + 1);
}

#ifdef _WIN32
std::string PlatformDetector::Impl::GetWMIProperty(const std::string& query, 
                                                   const std::string& property) {
    // WMI查询实现(简化版)
    // 实际项目中需要完整的WMI实现
    return "";
}
#endif

#ifdef __linux__
std::string PlatformDetector::Impl::GetProcInfo(const std::string& file) {
    return ReadFile(file);
}

std::string PlatformDetector::Impl::GetSysfsInfo(const std::string& path) {
    return TrimString(ReadFile(path));
}
#endif

// 全局函数实现
PlatformType QuickDetectPlatform() {
    PlatformDetector detector;
    return detector.DetectPlatform();
}

Architecture QuickDetectArchitecture() {
    PlatformDetector detector;
    return detector.DetectArchitecture();
}

bool QuickCheckVirtualMachine() {
    PlatformDetector detector;
    detector.Initialize();
    return detector.IsVirtualMachine();
}

bool QuickCheckDebugger() {
    PlatformDetector detector;
    detector.Initialize();
    return detector.IsDebuggerPresent();
}

} // namespace platform
} // namespace gaokao