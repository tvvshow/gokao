# 高考志愿填报系统 - 用户服务

用户服务是高考志愿填报系统的核心服务之一，负责用户认证、授权、用户管理和权限控制。

## 🚀 功能特性

### 核心功能
- **用户认证**: 用户注册、登录、JWT令牌管理
- **权限管理**: 基于RBAC的角色权限控制
- **用户管理**: 用户信息CRUD操作
- **安全防护**: 密码加密、登录限制、审计日志
- **API文档**: 集成Swagger自动生成API文档

### 技术特性
- **高性能**: 基于Gin框架，支持高并发
- **数据持久化**: PostgreSQL + Redis缓存
- **容器化**: Docker + Docker Compose部署
- **监控**: 健康检查、性能指标
- **安全**: JWT认证、CORS、请求限制

## 📋 系统要求

- Go 1.21+
- PostgreSQL 15+
- Redis 7+
- Docker & Docker Compose (可选)

## 🛠️ 快速开始

### 1. 克隆项目

```bash
cd services/user-service
```

### 2. 环境配置

```bash
# 复制环境配置文件
cp .env.example .env

# 编辑配置文件
vim .env
```

### 3. 使用Docker启动（推荐）

```bash
# 启动所有服务
make docker-run

# 或者使用docker-compose
docker-compose up -d
```

### 4. 本地开发启动

```bash
# 安装依赖
make deps

# 启动开发环境
make dev

# 或者直接运行
make run
```

## 📚 API 文档

启动服务后，访问以下地址查看API文档：

- **Swagger UI**: http://localhost:8080/swagger/index.html
- **健康检查**: http://localhost:8080/health

## 🔧 开发命令

项目提供了完整的Makefile命令，方便开发和部署：

```bash
# 查看所有可用命令
make help

# 开发相关
make dev          # 启动开发环境（热重载）
make build        # 构建应用
make test         # 运行测试
make lint         # 代码检查
make format       # 格式化代码
make swagger      # 生成API文档

# Docker相关
make docker-build # 构建Docker镜像
make docker-run   # 启动Docker服务
make docker-stop  # 停止Docker服务
make logs         # 查看日志

# 数据库相关
make db-shell     # 进入数据库
make db-backup    # 备份数据库
```

## 🏗️ 项目结构

```
user-service/
├── cmd/                    # 应用入口
├── internal/              # 内部代码
│   ├── config/           # 配置管理
│   ├── database/         # 数据库连接
│   ├── handlers/         # HTTP处理器
│   ├── middleware/       # 中间件
│   ├── models/          # 数据模型
│   └── services/        # 业务逻辑
├── docs/                 # API文档
├── bin/                  # 编译输出
├── .env.example         # 环境配置示例
├── docker-compose.yml   # Docker编排
├── Dockerfile          # Docker镜像
├── Makefile           # 构建脚本
├── go.mod             # Go模块
└── README.md          # 项目说明
```

## 🔐 API 接口

### 认证接口

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/api/v1/auth/register` | 用户注册 |
| POST | `/api/v1/auth/login` | 用户登录 |
| POST | `/api/v1/auth/refresh` | 刷新令牌 |
| POST | `/api/v1/auth/logout` | 用户登出 |
| GET | `/api/v1/auth/profile` | 获取用户信息 |
| PUT | `/api/v1/auth/password` | 修改密码 |

### 用户管理接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/v1/users` | 获取用户列表 |
| GET | `/api/v1/users/{id}` | 获取用户详情 |
| PUT | `/api/v1/users/{id}` | 更新用户信息 |
| DELETE | `/api/v1/users/{id}` | 删除用户 |
| POST | `/api/v1/users/{id}/roles` | 分配角色 |
| DELETE | `/api/v1/users/{id}/roles/{role_id}` | 撤销角色 |

### 角色权限接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/v1/roles` | 获取角色列表 |
| POST | `/api/v1/roles` | 创建角色 |
| GET | `/api/v1/roles/{id}` | 获取角色详情 |
| PUT | `/api/v1/roles/{id}` | 更新角色 |
| DELETE | `/api/v1/roles/{id}` | 删除角色 |
| GET | `/api/v1/permissions` | 获取权限列表 |
| POST | `/api/v1/permissions` | 创建权限 |

## 🔒 安全特性

### 认证授权
- JWT令牌认证
- 刷新令牌机制
- 角色基础访问控制(RBAC)
- 权限细粒度控制

### 安全防护
- 密码BCrypt加密
- 登录失败限制
- CORS跨域保护
- 请求速率限制
- SQL注入防护

### 审计日志
- 用户操作记录
- 登录尝试记录
- 权限变更记录
- 系统事件记录

## 🗄️ 数据库设计

### 核心表结构

- **users**: 用户基础信息
- **roles**: 角色定义
- **permissions**: 权限定义
- **user_roles**: 用户角色关联
- **role_permissions**: 角色权限关联
- **audit_logs**: 审计日志
- **login_attempts**: 登录尝试记录
- **refresh_tokens**: 刷新令牌

### 默认角色

- **超级管理员**: 系统最高权限
- **管理员**: 用户管理权限
- **普通用户**: 基础功能权限

## 🚀 部署指南

### Docker部署（推荐）

```bash
# 1. 构建镜像
make docker-build

# 2. 启动服务
make docker-run

# 3. 检查状态
make status
```

### 生产环境部署

```bash
# 1. 设置生产环境变量
export GIN_MODE=release
export JWT_SECRET=your-production-secret

# 2. 构建生产镜像
docker build -t gaokao/user-service:prod .

# 3. 运行生产容器
docker run -d \
  --name user-service-prod \
  -p 8080:8080 \
  --env-file .env.prod \
  gaokao/user-service:prod
```

## 🧪 测试

```bash
# 运行所有测试
make test

# 运行性能测试
make bench

# 运行安全扫描
make security
```

## 📊 监控

### 健康检查

```bash
# 检查服务健康状态
curl http://localhost:8080/health
```

### 性能指标

- 响应时间监控
- 内存使用监控
- 数据库连接池监控
- Redis连接监控

## 🔧 配置说明

### 环境变量

| 变量名 | 描述 | 默认值 |
|--------|------|--------|
| `SERVER_PORT` | 服务端口 | `8080` |
| `DB_HOST` | 数据库主机 | `localhost` |
| `DB_PORT` | 数据库端口 | `5432` |
| `REDIS_HOST` | Redis主机 | `localhost` |
| `JWT_SECRET` | JWT密钥 | - |
| `JWT_EXPIRES_IN` | JWT过期时间 | `24h` |

### 数据库配置

```yaml
# PostgreSQL
DB_HOST: localhost
DB_PORT: 5432
DB_USER: gaokao_user
DB_PASSWORD: gaokao_password
DB_NAME: gaokao_user_db

# Redis
REDIS_HOST: localhost
REDIS_PORT: 6379
REDIS_DB: 0
```

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📝 更新日志

### v1.0.0 (2024-01-20)
- ✨ 初始版本发布
- 🔐 用户认证系统
- 👥 RBAC权限管理
- 📚 API文档集成
- 🐳 Docker容器化

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 📞 联系我们

- 项目地址: [GitHub](https://github.com/gaokao/user-service)
- 问题反馈: [Issues](https://github.com/gaokao/user-service/issues)
- 邮箱: dev@gaokao.com

---

**高考志愿填报系统用户服务** - 为每个学生的未来保驾护航 🎓