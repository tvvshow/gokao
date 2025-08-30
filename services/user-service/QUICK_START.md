# Go-C++ 设备指纹模块快速开始指南

## 快速安装

### 1. 环境检查
```bash
# 检查必要工具
make check-cpp
```

### 2. 安装依赖
```bash
# 安装开发工具
make install-tools

# 安装Go依赖
make deps
```

### 3. 构建项目
```bash
# 构建C++库
make build-cpp

# 构建Go应用(含CGO)
make build-cgo
```

### 4. 运行测试
```bash
# 运行CGO集成测试
make test-cgo

# 运行性能测试
make bench-cgo
```

## 基础使用

### 1. 设备指纹采集

```go
package main

import (
    "context"
    "fmt"
    "user-service/internal/cpp"
)

func main() {
    // 快速采集设备指纹
    fingerprint, err := cpp.QuickCollectFingerprint()
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("设备ID: %s\n", fingerprint.DeviceID)
    fmt.Printf("设备类型: %s\n", fingerprint.DeviceType)
    fmt.Printf("置信度: %d%%\n", fingerprint.ConfidenceScore)
}
```

### 2. 高级使用（完整功能）

```go
package main

import (
    "context"
    "log"
    "user-service/internal/services"
)

func main() {
    // 创建设备服务
    config := &services.DeviceServiceConfig{
        EnableCache:      true,
        EnableEncryption: true,
        SecurityLevel:    80,
    }
    
    deviceService, err := services.NewDeviceService(db, logger, config)
    if err != nil {
        log.Fatal(err)
    }
    defer deviceService.Close()
    
    ctx := context.Background()
    userID := uint(1)
    
    // 采集设备指纹
    deviceInfo, err := deviceService.CollectDeviceFingerprint(ctx, userID)
    if err != nil {
        log.Fatal(err)
    }
    
    // 生成许可证
    license, err := deviceService.GenerateDeviceLicense(ctx, userID, deviceInfo.ID, "trial", 30)
    if err != nil {
        log.Fatal(err)
    }
    
    // 验证设备访问
    isValid, err := deviceService.ValidateDeviceAccess(ctx, userID, deviceInfo.ID)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("设备访问有效: %v\n", isValid)
}
```

## 主要API

### DeviceFingerprintCollector
```go
collector := cpp.NewDeviceFingerprintCollector()
collector.Initialize("")
defer collector.Uninitialize()

// 采集指纹
fingerprint, err := collector.CollectFingerprint()

// 生成哈希
hash, err := collector.GenerateHash(fingerprint)

// 比较指纹
comparison, err := collector.CompareFingerprints(fp1, fp2)
```

### CryptoService
```go
crypto := cpp.NewCryptoService()

// 加密数据
encrypted, err := crypto.EncryptData(data, key)

// 解密数据
decrypted, err := crypto.DecryptData(encrypted, key)

// 数字签名
signature, err := crypto.SignData(data, privateKey)

// 验证签名
valid, err := crypto.VerifySignature(data, signature, publicKey)
```

### LicenseService
```go
license := cpp.NewLicenseService()

// 生成许可证
licenseData, err := license.GenerateLicense(deviceID, expiresAt, privateKey, "trial", features)

// 验证许可证
licenseInfo, err := license.ValidateLicense(licenseData, deviceID)

// 检查过期
expired, remaining, err := license.CheckLicenseExpiry(licenseData)
```

## 常用命令

```bash
# 开发环境设置
make setup-dev

# 环境验证
make verify-env

# 完整CI流程
make ci-cgo

# 生产部署准备
make prod-ready-cgo

# 清理所有构建文件
make clean-all
```

## 故障排除

### 编译错误
```bash
# 检查C++环境
make check-cpp

# 重新构建C++库
make clean && make build-cpp
```

### 运行时错误
```bash
# 检查库路径
export LD_LIBRARY_PATH=/path/to/cpp-modules/device-fingerprint/lib

# 验证库依赖
ldd bin/user-service-cgo
```

### 测试失败
```bash
# 运行详细测试
make test-cgo -v

# 检查CGO环境变量
echo $CGO_ENABLED
echo $CGO_CPPFLAGS
echo $CGO_LDFLAGS
```

## 配置示例

### 基础配置
```json
{
  "enable_cache": true,
  "cache_ttl": "10m",
  "enable_encryption": true,
  "security_level": 80,
  "max_concurrent_tasks": 10
}
```

### 生产配置
```json
{
  "enable_cache": true,
  "cache_ttl": "5m",
  "enable_encryption": true,
  "enable_signature": true,
  "security_level": 90,
  "max_concurrent_tasks": 50,
  "enable_performance_log": true
}
```

## 性能基准

| 操作 | 延迟 | 吞吐量 |
|------|------|--------|
| 指纹采集 | ~5ms | 200 req/s |
| 指纹比较 | ~0.5ms | 2000 req/s |
| 数据加密 | ~2ms | 500 req/s |
| 许可证验证 | ~0.3ms | 3000 req/s |

## 获取帮助

- 完整文档: [CGO_INTEGRATION_GUIDE.md](./CGO_INTEGRATION_GUIDE.md)
- API文档: `make swagger`
- 问题反馈: 项目Issue页面