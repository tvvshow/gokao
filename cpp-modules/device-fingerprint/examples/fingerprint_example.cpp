#include <iostream>
#include <string>
#include "device_fingerprint.h"

<<<<<<< HEAD
using namespace gaokao::device;

=======
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
int main() {
    std::cout << "Device Fingerprint Example" << std::endl;
    std::cout << "=========================" << std::endl;
    
    try {
        // 创建设备指纹收集器
<<<<<<< HEAD
        DeviceFingerprintCollector collector;
        
        // 初始化收集器
        ErrorCode result = collector.Initialize();
        if (result != ErrorCode::SUCCESS) {
            std::cerr << "Failed to initialize collector" << std::endl;
            return 1;
        }
        
        // 收集完整设备指纹
        DeviceFingerprint fingerprint;
        result = collector.CollectFingerprint(fingerprint);
        
        if (result == ErrorCode::SUCCESS) {
            std::cout << "Device Fingerprint collected successfully!" << std::endl;
            std::cout << "Device ID: " << fingerprint.device_id << std::endl;
            std::cout << "Fingerprint Hash: " << fingerprint.fingerprint_hash << std::endl;
            std::cout << "Confidence Score: " << fingerprint.confidence_score << "%" << std::endl;
            std::cout << std::endl;
        }
        
        // 获取硬件信息
        std::cout << "Hardware Information:" << std::endl;
        HardwareInfo hwInfo;
        result = collector.CollectHardwareInfo(hwInfo);
        if (result == ErrorCode::SUCCESS) {
            std::cout << "CPU Model: " << hwInfo.cpu_model << std::endl;
            std::cout << "CPU Cores: " << hwInfo.cpu_cores << std::endl;
            std::cout << "Total Memory: " << hwInfo.total_memory << " bytes" << std::endl;
        }
=======
        DeviceFingerprint fingerprint;
        
        // 收集设备指纹
        std::string result = fingerprint.collectFingerprint();
        
        std::cout << "Device Fingerprint: " << result << std::endl;
        std::cout << std::endl;
        
        // 获取硬件信息
        std::cout << "Hardware Information:" << std::endl;
        std::string hwInfo = fingerprint.getHardwareInfo();
        std::cout << hwInfo << std::endl;
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
        std::cout << std::endl;
        
        // 获取系统信息
        std::cout << "System Information:" << std::endl;
<<<<<<< HEAD
        SystemInfo sysInfo;
        result = collector.CollectSystemInfo(sysInfo);
        if (result == ErrorCode::SUCCESS) {
            std::cout << "Hostname: " << sysInfo.hostname << std::endl;
            std::cout << "Username: " << sysInfo.username << std::endl;
        }
        std::cout << std::endl;
        
        // 获取运行时信息
        std::cout << "Runtime Information:" << std::endl;
        RuntimeInfo runtimeInfo;
        result = collector.CollectRuntimeInfo(runtimeInfo);
        if (result == ErrorCode::SUCCESS) {
            std::cout << "Screen Resolution: " << runtimeInfo.screen_resolution << std::endl;
            std::cout << "Color Depth: " << runtimeInfo.color_depth << std::endl;
            std::cout << "Browser Info: " << runtimeInfo.browser_info << std::endl;
        }
        
        // 清理资源
        collector.Uninitialize();
=======
        std::string sysInfo = fingerprint.getSystemInfo();
        std::cout << sysInfo << std::endl;
        std::cout << std::endl;
        
        // 获取网络信息
        std::cout << "Network Information:" << std::endl;
        std::string netInfo = fingerprint.getNetworkInfo();
        std::cout << netInfo << std::endl;
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
        
    } catch (const std::exception& e) {
        std::cerr << "Error: " << e.what() << std::endl;
        return 1;
    }
    
    return 0;
}