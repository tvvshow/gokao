#include <iostream>
#include <string>
#include "device_fingerprint.h"

int main() {
    std::cout << "Device Fingerprint Example" << std::endl;
    std::cout << "=========================" << std::endl;
    
    try {
        // 创建设备指纹收集器
        DeviceFingerprint fingerprint;
        
        // 收集设备指纹
        std::string result = fingerprint.collectFingerprint();
        
        std::cout << "Device Fingerprint: " << result << std::endl;
        std::cout << std::endl;
        
        // 获取硬件信息
        std::cout << "Hardware Information:" << std::endl;
        std::string hwInfo = fingerprint.getHardwareInfo();
        std::cout << hwInfo << std::endl;
        std::cout << std::endl;
        
        // 获取系统信息
        std::cout << "System Information:" << std::endl;
        std::string sysInfo = fingerprint.getSystemInfo();
        std::cout << sysInfo << std::endl;
        std::cout << std::endl;
        
        // 获取网络信息
        std::cout << "Network Information:" << std::endl;
        std::string netInfo = fingerprint.getNetworkInfo();
        std::cout << netInfo << std::endl;
        
    } catch (const std::exception& e) {
        std::cerr << "Error: " << e.what() << std::endl;
        return 1;
    }
    
    return 0;
}