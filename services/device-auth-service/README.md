# Device Auth Service

设备认证服务，负责处理设备指纹采集、许可证验证等安全相关功能。

## 功能特性

- 设备指纹采集与验证
- 许可证生成与验证
- 安全环境检测
- 性能监控

## 技术架构

- Go语言实现服务框架
- C++实现核心算法和硬件访问
- 通过CGO集成C++库
- RESTful API接口

## 目录结构

```
device-auth-service/
├── internal/
│   ├── config/      # 配置管理
│   ├── handlers/    # HTTP处理器
│   ├── middleware/  # 中间件
│   ├── models/      # 数据模型
│   ├── services/    # 业务逻辑
│   └── cpp/         # C++集成代码
├── api/             # API定义
├── config/          # 配置文件
├── docs/            # 文档
├── scripts/         # 脚本
├── Dockerfile       # Docker配置
├── go.mod           # Go模块定义
├── main.go          # 主程序入口
└── Makefile         # 构建脚本
```

## 构建和运行

### 环境要求

- Go 1.22+
- GCC 9.0+ 或 Clang 10.0+
- CMake 3.15+

### 构建步骤

```bash
# 构建C++库
cd ../../cpp-modules/device-fingerprint
mkdir build && cd build
cmake ..
make

# 构建Go服务
cd ../../../services/device-auth-service
go build -o bin/device-auth-service .

# 运行服务
./bin/device-auth-service
```

### Docker构建

```bash
docker build -t device-auth-service .
docker run -p 8085:8085 device-auth-service
```