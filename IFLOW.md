# 高考志愿填报系统 - iFlow CLI 上下文

## 项目概述

这是一个基于Go和C++混合架构的高考志愿填报系统，旨在为中国高中学生、家长和老师提供智能化的大学申请解决方案。系统采用微服务架构，包含用户服务、数据服务、支付服务、推荐服务等多个模块，并通过API网关统一对外提供服务。

### 核心技术栈

- **后端**: Go (主框架) + C++ (核心算法)
- **数据库**: PostgreSQL + Redis
- **前端**: Vue 3 + TypeScript + Tailwind CSS
- **基础设施**: Docker + Makefile构建系统
- **安全防护**: C++模块使用VMProtect加密保护，Go使用garble混淆

### 系统架构

系统采用Go+C++混合架构：
- **Go作为主框架**（70%代码量）：负责API服务、业务逻辑、数据库操作、用户管理
- **C++核心模块**（30%代码量）：负责算法引擎、AI推理、安全验证、付费功能核心
- **语言间通信**：通过CGO或gRPC实现高效通信

## 项目结构

```
gaokao/
├── services/                 # 后端微服务
│   ├── api-gateway/         # API网关
│   ├── user-service/        # 用户服务
│   ├── data-service/        # 数据服务
│   ├── payment-service/     # 支付服务
│   ├── recommendation-service/ # 推荐服务
│   └── monitoring-service/  # 监控服务
├── frontend/                # 前端应用 (Vue 3)
├── pkg/                     # 共享包
├── cpp-modules/             # C++核心模块
├── scripts/                 # 脚本工具
├── docker/                  # Docker配置
├── config/                  # 配置文件
└── docs/                    # 文档
```

## 构建和运行

### 环境要求

- Go 1.22+
- Node.js 16+
- Docker & Docker Compose
- PostgreSQL 15
- Redis 7

### 开发环境启动

```bash
# 克隆项目后，启动开发环境
docker-compose up -d

# 或使用Makefile
make dev
```

### 构建项目

```bash
# 构建所有组件
make build

# 仅构建Go服务
make build-go

# 仅构建前端
make build-frontend
```

### 运行测试

```bash
# 运行所有测试
make test

# 仅运行Go测试
make test-go

# 仅运行前端测试
make test-frontend
```

## 开发约定

### Go代码规范

- 使用Gin作为Web框架
- 使用GORM作为ORM库
- 使用logrus进行日志记录
- 使用Viper进行配置管理
- 遵循标准Go项目结构

### 前端开发

- 使用Vue 3 + TypeScript
- 使用Tailwind CSS进行样式设计
- 使用Pinia进行状态管理
- 使用Vue Router进行路由管理

### C++模块

- 使用Eigen进行数学计算
- 使用ONNX Runtime进行AI推理
- 使用nlohmann/json处理JSON
- 核心模块使用VMProtect进行保护

### 安全实践

- C++代码使用VMProtect/Themida加壳保护
- Go代码使用garble混淆标识符
- 使用Docker Content Trust进行镜像签名
- 使用Istio mTLS加密通信
- 敏感数据使用AES-256加密存储

## 服务说明

### API网关 (api-gateway)
- 端口: 8080
- 负责请求路由、认证、限流等

### 用户服务 (user-service)
- 端口: 8081
- 处理用户注册、登录、权限管理等

### 数据服务 (data-service)
- 端口: 8082
- 提供院校、专业等数据查询接口

### 支付服务 (payment-service)
- 端口: 8083
- 集成微信支付、支付宝等支付方式

### 推荐服务 (recommendation-service)
- 端口: 8084
- 提供AI志愿推荐功能

### 前端应用 (frontend)
- 端口: 3000
- 提供用户界面