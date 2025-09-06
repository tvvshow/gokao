#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "c_interface.h"

int main() {
    printf("Device Fingerprint C Interface Example\n");
    printf("=====================================\n");
    
    // 创建设备指纹收集器
    DeviceFingerprintHandle* handle = device_fingerprint_create();
    if (!handle) {
        fprintf(stderr, "Failed to create device fingerprint handle\n");
        return 1;
    }
    
    // 收集设备指纹
    char* fingerprint = device_fingerprint_collect(handle);
    if (fingerprint) {
        printf("Device Fingerprint: %s\n", fingerprint);
        free(fingerprint);
    } else {
        fprintf(stderr, "Failed to collect device fingerprint\n");
    }
    printf("\n");
    
    // 获取硬件信息
    printf("Hardware Information:\n");
    char* hw_info = device_fingerprint_get_hardware_info(handle);
    if (hw_info) {
        printf("%s\n", hw_info);
        free(hw_info);
    } else {
        fprintf(stderr, "Failed to get hardware information\n");
    }
    printf("\n");
    
    // 获取系统信息
    printf("System Information:\n");
    char* sys_info = device_fingerprint_get_system_info(handle);
    if (sys_info) {
        printf("%s\n", sys_info);
        free(sys_info);
    } else {
        fprintf(stderr, "Failed to get system information\n");
    }
    printf("\n");
    
    // 获取网络信息
    printf("Network Information:\n");
    char* net_info = device_fingerprint_get_network_info(handle);
    if (net_info) {
        printf("%s\n", net_info);
        free(net_info);
    } else {
        fprintf(stderr, "Failed to get network information\n");
    }
    
    // 清理资源
    device_fingerprint_destroy(handle);
    
    return 0;
}