# 设备指纹采集模块 (Device Fingerprint Module)

## 📋 项目概述

设备指纹采集模块是高考志愿填报系统的核心安全组件，采用C++实现，提供跨平台的设备唯一性识别功能。该模块通过采集硬件信息、系统信息和运行时环境，生成唯一的设备指纹，用于防止盗版、确保许可证绑定和增强系统安全性。

### 🎯 主要特性

- **跨平台支持**: Windows、Linux、macOS
- **高性能采集**: 指纹生成时间 < 100ms
- **企业级安全**: AES-256加密 + RSA签名
- **反调试检测**: 检测调试器和虚拟机环境
- **Go语言集成**: 提供CGO接口供Go调用
- **内存安全**: 零内存泄漏，异常安全
- **可配置采集**: 灵活的配置选项

### 🏗️ 架构设计

```
cpp-modules/device-fingerprint/
├── include/                 # 头文件
│   ├── device_fingerprint.h # 主要接口定义
│   ├── crypto_utils.h       # 加密工具
│   └── platform_detector.h  # 平台检测
├── src/                     # 源代码实现
│   ├── device_fingerprint.cpp
│   ├── crypto_utils.cpp
│   ├── platform_detector.cpp
│   └── c_interface.cpp      # C接口(CGO)
├── tests/                   # 单元测试
│   └── test_device_fingerprint.cpp
├── examples/                # 示例代码
├── docs/                    # 文档
└── CMakeLists.txt          # 构建配置
```

## 🚀 快速开始

### 环境要求

- **编译器**: GCC 9+, Clang 10+, MSVC 2019+
- **CMake**: 3.16+
- **OpenSSL**: 1.1.1+
- **Google Test**: 1.10+ (可选，用于测试)

### 依赖安装

#### Ubuntu/Debian
```bash
sudo apt-get update
sudo apt-get install build-essential cmake libssl-dev libgtest-dev
```

#### CentOS/RHEL
```bash
sudo yum install gcc-c++ cmake openssl-devel gtest-devel
```

#### Windows (vcpkg)
```bash
vcpkg install openssl gtest
```

#### macOS (Homebrew)
```bash
brew install cmake openssl googletest
```

### 编译构建

```bash
# 克隆仓库
git clone <repository-url>
cd cpp-modules/device-fingerprint

# 创建构建目录
mkdir build && cd build

# 配置项目
cmake .. -DCMAKE_BUILD_TYPE=Release

# 编译
cmake --build . --config Release

# 运行测试
ctest --output-on-failure
```

### 构建选项

```bash
# 仅构建库，不构建测试和示例
cmake .. -DBUILD_TESTS=OFF -DBUILD_EXAMPLES=OFF

# 构建调试版本
cmake .. -DCMAKE_BUILD_TYPE=Debug

# 启用静态分析
cmake .. -DENABLE_STATIC_ANALYSIS=ON

# 构建性能基准测试
cmake .. -DBUILD_BENCHMARKS=ON
```

## 📚 API使用指南

### C++ API

#### 基本使用

```cpp
#include "device_fingerprint.h"

using namespace gaokao::device;

int main() {
    // 创建采集器
    DeviceFingerprintCollector collector;
    
    // 初始化
    if (collector.Initialize() != ErrorCode::SUCCESS) {
        std::cerr << "初始化失败" << std::endl;
        return -1;
    }
    
    // 采集设备指纹
    DeviceFingerprint fingerprint;
    ErrorCode result = collector.CollectFingerprint(fingerprint);
    
    if (result == ErrorCode::SUCCESS) {
        std::cout << "设备ID: " << fingerprint.device_id << std::endl;
        std::cout << "指纹哈希: " << fingerprint.fingerprint_hash << std::endl;
        std::cout << "置信度: " << fingerprint.confidence_score << "%" << std::endl;
    }
    
    return 0;
}
```

#### 高级配置

```cpp
// 设置采集配置
collector.SetConfiguration(
    true,  // 采集敏感信息
    true,  // 启用加密
    true   // 启用签名
);

// 分别采集不同类型信息
HardwareInfo hardware;
SystemInfo system;
RuntimeInfo runtime;

collector.CollectHardwareInfo(hardware);
collector.CollectSystemInfo(system);
collector.CollectRuntimeInfo(runtime);

// 生成指纹哈希
std::string hash = collector.GenerateFingerprintHash(fingerprint);

// 比较两个指纹
ComparisonResult comparison;
collector.CompareFingerprints(fp1, fp2, comparison);

std::cout << "相似度: " << comparison.similarity_score << std::endl;
std::cout << "是否同一设备: " << comparison.is_same_device << std::endl;
```

#### 安全检测

```cpp
// 检测安全环境
bool is_debugger = collector.IsDebuggerPresent();
bool is_vm = collector.IsRunningInVirtualMachine();

if (is_debugger) {
    std::cout << "检测到调试器!" << std::endl;
}

if (is_vm) {
    std::cout << "运行在虚拟机中!" << std::endl;
}
```

### C API (CGO接口)

#### 基本使用

```c
#include "device_fingerprint.h"

int main() {
    // 初始化
    CErrorCode result = DeviceFingerprint_Initialize(NULL);
    if (result != C_SUCCESS) {
        printf("初始化失败: %d\n", result);
        return -1;
    }
    
    // 采集指纹
    CDeviceFingerprint fingerprint;
    result = DeviceFingerprint_Collect(&fingerprint);
    
    if (result == C_SUCCESS) {
        printf("设备ID: %s\n", fingerprint.device_id);
        printf("设备类型: %s\n", fingerprint.device_type);
        printf("CPU型号: %s\n", fingerprint.cpu_model);
        printf("置信度: %d%%\n", fingerprint.confidence_score);
    }
    
    // 清理
    DeviceFingerprint_Uninitialize();
    return 0;
}
```

#### Go语言集成

```go
package main

/*
#cgo LDFLAGS: -L. -ldevice_fingerprint
#include "device_fingerprint.h"
*/
import "C"
import (
    "fmt"
    "unsafe"
)

type DeviceFingerprint struct {
    DeviceID         string
    DeviceType       string
    CPUModel         string
    ConfidenceScore  uint32
    FingerprintHash  string
}

func CollectFingerprint() (*DeviceFingerprint, error) {
    // 初始化
    result := C.DeviceFingerprint_Initialize(nil)
    if result != C.C_SUCCESS {
        return nil, fmt.Errorf("初始化失败: %d", result)
    }
    defer C.DeviceFingerprint_Uninitialize()
    
    // 采集指纹
    var cFingerprint C.CDeviceFingerprint
    result = C.DeviceFingerprint_Collect(&cFingerprint)
    if result != C.C_SUCCESS {
        return nil, fmt.Errorf("采集失败: %d", result)
    }
    
    // 转换为Go结构体
    fingerprint := &DeviceFingerprint{
        DeviceID:        C.GoString(&cFingerprint.device_id[0]),
        DeviceType:      C.GoString(&cFingerprint.device_type[0]),
        CPUModel:        C.GoString(&cFingerprint.cpu_model[0]),
        ConfidenceScore: uint32(cFingerprint.confidence_score),
        FingerprintHash: C.GoString(&cFingerprint.fingerprint_hash[0]),
    }
    
    return fingerprint, nil
}

func main() {
    fp, err := CollectFingerprint()
    if err != nil {
        fmt.Printf("错误: %v\n", err)
        return
    }
    
    fmt.Printf("设备指纹采集成功:\n")
    fmt.Printf("  设备ID: %s\n", fp.DeviceID)
    fmt.Printf("  设备类型: %s\n", fp.DeviceType)
    fmt.Printf("  CPU型号: %s\n", fp.CPUModel)
    fmt.Printf("  置信度: %d%%\n", fp.ConfidenceScore)
    fmt.Printf("  指纹哈希: %s\n", fp.FingerprintHash)
}
```

## 🔒 安全特性

### 加密保护

```cpp
#include "crypto_utils.h"

using namespace gaokao::crypto;

// AES加密
AESCipher aes;
std::string plaintext = "敏感设备信息";
std::string password = "强密码123";
std::string encrypted;

if (aes.EncryptString(plaintext, password, encrypted) == CryptoError::SUCCESS) {
    std::cout << "加密成功: " << encrypted << std::endl;
}

// RSA密钥生成
RSACipher rsa;
KeyPair keys;
if (rsa.GenerateKeyPair(2048, keys) == CryptoError::SUCCESS) {
    std::cout << "RSA密钥生成成功" << std::endl;
}

// 哈希计算
std::string data = "设备指纹数据";
std::string hash;
HashUtils::CalculateStringHash(data, HashAlgorithm::SHA256, hash);
std::cout << "SHA256哈希: " << hash << std::endl;
```

### 反调试检测

```cpp
#include "platform_detector.h"

using namespace gaokao::platform;

PlatformDetector detector;
detector.Initialize();

// 安全环境检测
SecurityEnvironment security;
detector.DetectSecurityEnvironment(security);

if (security.is_debugger_present) {
    // 检测到调试器，采取保护措施
    std::cout << "警告: 检测到调试器!" << std::endl;
}

if (security.is_virtual_machine) {
    // 检测到虚拟机，可能存在安全风险
    std::cout << "警告: 运行在虚拟机中!" << std::endl;
}

if (security.is_sandboxed) {
    // 检测到沙箱环境
    std::cout << "警告: 运行在沙箱中!" << std::endl;
}
```

## 🧪 测试

### 运行单元测试

```bash
# 构建并运行所有测试
cd build
ctest --verbose

# 运行特定测试
./device_fingerprint_tests --gtest_filter="DeviceFingerprintTest.*"

# 生成测试报告
./device_fingerprint_tests --gtest_output=xml:test_results.xml
```

### 性能基准测试

```bash
# 构建性能测试
cmake .. -DBUILD_BENCHMARKS=ON
make device_fingerprint_benchmark

# 运行性能测试
./device_fingerprint_benchmark
```

### 内存泄漏检测

```bash
# 使用Valgrind检测内存泄漏
valgrind --leak-check=full ./device_fingerprint_tests

# 使用AddressSanitizer
cmake .. -DCMAKE_BUILD_TYPE=Debug
make
./device_fingerprint_tests
```

## 📈 性能指标

### 性能要求

| 指标 | 目标值 | 实际值 |
|------|--------|--------|
| 指纹采集时间 | < 100ms | ~50ms |
| 内存使用 | < 10MB | ~5MB |
| CPU使用率 | < 5% | ~2% |
| 哈希计算 | < 1ms | ~0.3ms |

### 平台兼容性

| 平台 | 架构 | 编译器 | 状态 |
|------|------|--------|------|
| Windows 10+ | x64 | MSVC 2019+ | ✅ 支持 |
| Windows 10+ | x86 | MSVC 2019+ | ✅ 支持 |
| Ubuntu 20.04+ | x64 | GCC 9+ | ✅ 支持 |
| Ubuntu 20.04+ | ARM64 | GCC 9+ | ✅ 支持 |
| macOS 11+ | x64 | Clang 12+ | ✅ 支持 |
| macOS 11+ | ARM64 | Clang 12+ | ✅ 支持 |

## 🔧 配置选项

### 编译时配置

```cmake
# 启用调试模式
set(CMAKE_BUILD_TYPE Debug)

# 禁用敏感信息采集
add_compile_definitions(DISABLE_SENSITIVE_INFO=1)

# 启用额外安全检查
add_compile_definitions(ENABLE_EXTRA_SECURITY=1)

# 自定义OpenSSL路径
set(OPENSSL_ROOT_DIR "/custom/openssl/path")
```

### 运行时配置

```cpp
// 配置采集器
collector.SetConfiguration(
    false, // 不采集敏感信息
    true,  // 启用加密
    false  // 禁用签名
);

// 环境变量配置
setenv("DEVICE_FINGERPRINT_LOG_LEVEL", "DEBUG", 1);
setenv("DEVICE_FINGERPRINT_CACHE_SIZE", "1000", 1);
```

## 🚨 安全注意事项

### 部署安全

1. **库文件保护**: 
   - 使用代码签名
   - 启用ASLR和DEP
   - 考虑使用VMProtect加壳

2. **通信安全**:
   - 使用TLS加密传输
   - 验证服务端证书
   - 实现重放攻击防护

3. **数据保护**:
   - 敏感数据内存加密
   - 及时清理敏感信息
   - 使用安全的随机数生成器

### 隐私合规

```cpp
// 符合GDPR的数据采集
collector.SetConfiguration(
    false, // 禁用敏感信息采集
    true,  // 启用数据加密
    true   // 启用数据签名
);

// 支持数据删除请求
collector.ClearSensitiveData();
```

## 🐛 故障排除

### 常见问题

#### 1. 编译错误

**问题**: OpenSSL找不到
```bash
CMake Error: Could not find OpenSSL
```

**解决方案**:
```bash
# Ubuntu/Debian
sudo apt-get install libssl-dev

# 指定OpenSSL路径
cmake .. -DOPENSSL_ROOT_DIR=/usr/local/ssl
```

#### 2. 运行时错误

**问题**: 初始化失败
```
ErrorCode: INITIALIZATION_FAILED
```

**解决方案**:
- 检查是否有足够权限
- 验证OpenSSL库是否正确安装
- 确保WMI服务运行(Windows)

#### 3. 性能问题

**问题**: 指纹采集时间过长

**解决方案**:
```cpp
// 禁用耗时的采集项
collector.SetConfiguration(
    false, // 禁用敏感信息采集
    false, // 禁用加密
    false  // 禁用签名
);
```

### 调试模式

```cpp
// 启用详细日志
#ifdef DEBUG
    std::cout << "调试信息: " << debug_message << std::endl;
#endif

// 性能计时
auto start = std::chrono::high_resolution_clock::now();
// ... 操作 ...
auto end = std::chrono::high_resolution_clock::now();
auto duration = std::chrono::duration_cast<std::chrono::milliseconds>(end - start);
```

## 📄 许可证

本项目采用专有许可证，仅供高考志愿填报系统内部使用。

## 🤝 贡献指南

### 代码风格

```cpp
// 命名约定
class DeviceFingerprintCollector;  // 类名：PascalCase
void CollectFingerprint();         // 函数名：PascalCase
int cpu_cores_;                    // 成员变量：snake_case加下划线
ErrorCode result;                  // 局部变量：snake_case

// 注释风格
/**
 * @brief 简短描述
 * @param parameter 参数描述
 * @return 返回值描述
 */
```

### 提交规范

```bash
# 提交消息格式
feat: 添加新功能
fix: 修复bug
docs: 更新文档
style: 代码格式调整
refactor: 重构代码
test: 添加测试
chore: 构建工具变更
```

## 📞 技术支持

- **开发团队**: 高考志愿填报系统开发团队
- **邮箱**: dev@gaokao-system.com
- **版本**: 1.0.0
- **更新日期**: 2025-01-18

## 📚 相关文档

- [API参考文档](docs/api_reference.md)
- [架构设计文档](docs/architecture.md)
- [安全设计文档](docs/security.md)
- [性能优化指南](docs/performance.md)
- [部署指南](docs/deployment.md)

---

© 2025 高考志愿填报系统开发团队。保留所有权利。