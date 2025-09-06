# 高考志愿填报系统 - Docker 环境

本目录包含高考志愿填报系统的完整Docker开发、测试和生产环境配置。

## 📁 目录结构

```
docker/
├── dev/                        # 开发环境
│   ├── Dockerfile.api-gateway   # API Gateway开发镜像
│   ├── Dockerfile.user-service  # User Service开发镜像
│   ├── Dockerfile.cpp-modules   # C++模块开发镜像
│   ├── docker-compose.dev.yml   # 开发环境编排
│   ├── .env.example             # 开发环境变量模板
│   └── monitoring/              # 监控配置
├── prod/                       # 生产环境
│   ├── Dockerfile.api-gateway   # API Gateway生产镜像
│   ├── Dockerfile.user-service  # User Service生产镜像
│   ├── Dockerfile.cpp-modules   # C++模块生产镜像
│   ├── docker-compose.prod.yml  # 生产环境编排
│   ├── .env.example             # 生产环境变量模板
│   ├── nginx/                   # Nginx配置
│   └── secrets/                 # 密钥文件
├── test/                       # 测试环境
│   ├── Dockerfile.test-runner   # 测试运行器镜像
│   ├── docker-compose.test.yml  # 测试环境编排
│   └── sql/                     # 测试数据
├── scripts/                    # 工具脚本
│   ├── build.sh                # 构建脚本
│   ├── deploy.sh               # 部署脚本
│   ├── test.sh                 # 测试脚本
│   └── cleanup.sh              # 清理脚本
└── README.md                   # 本文档
```

## 🚀 快速开始

### 1. 环境准备

确保您的系统已安装以下软件：

- Docker 20.10+
- Docker Compose 2.0+
- Git
- Bash (Windows用户可使用Git Bash)

### 2. 开发环境部署

```bash
# 克隆项目
git clone <repository-url>
cd gaokao

# 复制环境配置文件
cp docker/dev/.env.example docker/dev/.env

# 构建开发环境镜像
./docker/scripts/build.sh dev

# 部署开发环境
./docker/scripts/deploy.sh dev
```

### 3. 访问服务

开发环境部署完成后，您可以访问：

- **API Gateway**: http://localhost:8080
- **Swagger UI**: http://localhost:8080/swagger/index.html
- **User Service**: http://localhost:8081
- **pgAdmin**: http://localhost:5050 (admin@gaokao.dev / admin123)
- **Grafana**: http://localhost:3000 (admin / admin123)
- **Prometheus**: http://localhost:9090

## 🔧 详细使用说明

### 开发环境

开发环境提供完整的开发体验，包括：

- **热重载**: 代码变更自动重新编译和重启
- **调试支持**: 暴露调试端口用于IDE调试
- **开发工具**: 包含数据库管理、Redis管理、监控工具
- **测试数据**: 预置测试数据便于开发测试

#### 启动开发环境

```bash
# 完整部署
./docker/scripts/deploy.sh dev

# 仅启动特定服务
docker-compose -f docker/dev/docker-compose.dev.yml up api-gateway user-service

# 查看日志
docker-compose -f docker/dev/docker-compose.dev.yml logs -f api-gateway

# 进入容器调试
docker exec -it gaokao-api-gateway-dev bash
```

#### 开发工具

```bash
# 启动开发工具（可选）
docker-compose -f docker/dev/docker-compose.dev.yml --profile tools up -d

# 启动监控工具（可选）
docker-compose -f docker/dev/docker-compose.dev.yml --profile monitoring up -d
```

### 生产环境

生产环境经过优化，提供：

- **安全配置**: 最小权限、安全扫描、密钥管理
- **性能优化**: 多阶段构建、最小化镜像
- **高可用**: 健康检查、自动重启、负载均衡
- **监控告警**: 完整的监控和告警体系

#### 部署生产环境

```bash
# 1. 复制并配置环境变量
cp docker/prod/.env.example docker/prod/.env
# 编辑 docker/prod/.env，修改所有密码和密钥

# 2. 准备密钥文件
mkdir -p docker/prod/secrets
# 将证书、密钥等文件放入 secrets 目录

# 3. 构建生产镜像
./docker/scripts/build.sh prod

# 4. 部署生产环境
./docker/scripts/deploy.sh prod
```

#### 生产环境安全检查

```bash
# 运行安全扫描
trivy image gaokao/api-gateway:latest
trivy image gaokao/user-service:latest
trivy image gaokao/cpp-modules:latest

# 检查配置
./docker/scripts/deploy.sh -c prod
```

### 测试环境

测试环境专门用于自动化测试：

- **隔离环境**: 独立的测试数据库和缓存
- **测试数据**: 预置测试数据和场景
- **自动化测试**: 单元测试、集成测试、API测试、性能测试
- **覆盖率收集**: 完整的代码覆盖率统计

#### 运行测试

```bash
# 运行所有测试
./docker/scripts/test.sh

# 运行特定类型测试
./docker/scripts/test.sh unit          # 单元测试
./docker/scripts/test.sh integration   # 集成测试
./docker/scripts/test.sh api          # API测试
./docker/scripts/test.sh performance  # 性能测试

# 只启动测试环境
./docker/scripts/test.sh -s

# 查看测试结果
open test-results/test-report.html
```

## 🛠️ 工具脚本详解

### build.sh - 构建脚本

```bash
# 构建所有环境镜像
./docker/scripts/build.sh

# 构建特定环境
./docker/scripts/build.sh dev
./docker/scripts/build.sh prod
./docker/scripts/build.sh test

# 清理后构建
./docker/scripts/build.sh -c prod

# 带版本标签构建
./docker/scripts/build.sh --version v1.0.0 prod
```

### deploy.sh - 部署脚本

```bash
# 部署开发环境
./docker/scripts/deploy.sh dev

# 强制重新部署
./docker/scripts/deploy.sh -f prod

# 跳过镜像拉取
./docker/scripts/deploy.sh -s dev

# 仅检查状态
./docker/scripts/deploy.sh -c prod
```

### test.sh - 测试脚本

```bash
# 完整测试流程
./docker/scripts/test.sh

# 分步执行
./docker/scripts/test.sh -s        # 启动测试环境
./docker/scripts/test.sh unit      # 运行单元测试
./docker/scripts/test.sh -e        # 停止测试环境

# 测试后清理
./docker/scripts/test.sh -c all
```

### cleanup.sh - 清理脚本

```bash
# 交互式清理
./docker/scripts/cleanup.sh

# 强制清理所有资源
./docker/scripts/cleanup.sh -f all

# 清理前备份
./docker/scripts/cleanup.sh -b volumes

# 重置开发环境
./docker/scripts/cleanup.sh dev-reset
```

## 🔧 配置详解

### 环境变量

每个环境都有对应的环境变量配置文件：

- `docker/dev/.env` - 开发环境配置
- `docker/prod/.env` - 生产环境配置
- `docker/test/.env` - 测试环境配置（通常不需要修改）

重要配置项：

```bash
# 数据库配置
POSTGRES_DB=gaokao_prod
POSTGRES_USER=gaokao_user
POSTGRES_PASSWORD=强密码        # 生产环境必须修改

# Redis配置
REDIS_PASSWORD=强密码            # 生产环境必须修改

# 应用配置
JWT_SECRET=强密钥               # 生产环境必须修改
DEBUG=false                     # 生产环境设为false
LOG_LEVEL=info                  # 生产环境建议info

# 安全配置
ENABLE_SWAGGER=0                # 生产环境禁用Swagger
CORS_ALLOWED_ORIGINS=https://yourdomain.com
```

### 密钥管理（生产环境）

生产环境使用Docker Secrets管理敏感信息：

```bash
# 创建密钥文件
mkdir -p docker/prod/secrets

# 数据库密码
echo "your-strong-password" > docker/prod/secrets/postgres_password.txt

# Redis密码
echo "your-redis-password" > docker/prod/secrets/redis_password.txt

# JWT密钥
openssl rand -base64 64 > docker/prod/secrets/jwt_secret.txt

# TLS证书
cp your-cert.crt docker/prod/secrets/tls.crt
cp your-key.key docker/prod/secrets/tls.key
```

### 网络配置

系统使用多个Docker网络进行隔离：

- `gaokao-dev-network` - 开发环境内部网络
- `gaokao-prod-frontend` - 生产环境前端网络
- `gaokao-prod-backend` - 生产环境后端网络（内部）
- `gaokao-test-network` - 测试环境网络

### 数据持久化

重要数据通过Docker volumes持久化：

- `gaokao-postgres-*-data` - PostgreSQL数据
- `gaokao-redis-*-data` - Redis数据
- `gaokao-grafana-*-data` - Grafana配置
- `gaokao-prometheus-*-data` - Prometheus数据

## 🐛 故障排除

### 常见问题

#### 1. 端口冲突

```bash
# 检查端口占用
netstat -tulpn | grep :8080

# 修改端口映射
vim docker/dev/.env
# 修改 API_GATEWAY_PORT=8080 为其他端口
```

#### 2. 权限问题

```bash
# Linux/macOS 赋予脚本执行权限
chmod +x docker/scripts/*.sh

# Windows 使用Git Bash运行脚本
```

#### 3. 内存不足

```bash
# 检查Docker资源使用
docker system df

# 清理未使用资源
./docker/scripts/cleanup.sh system
```

#### 4. 数据库连接失败

```bash
# 检查数据库状态
docker-compose -f docker/dev/docker-compose.dev.yml ps postgres

# 查看数据库日志
docker logs gaokao-postgres-dev

# 重启数据库
docker-compose -f docker/dev/docker-compose.dev.yml restart postgres
```

#### 5. 热重载不工作

```bash
# 检查文件挂载
docker inspect gaokao-api-gateway-dev | grep Mounts -A 20

# 重新构建开发镜像
./docker/scripts/build.sh dev
```

### 日志查看

```bash
# 查看所有服务日志
docker-compose -f docker/dev/docker-compose.dev.yml logs

# 查看特定服务日志
docker-compose -f docker/dev/docker-compose.dev.yml logs -f api-gateway

# 查看最近的错误日志
docker-compose -f docker/dev/docker-compose.dev.yml logs --tail=100 | grep ERROR
```

### 性能监控

```bash
# 查看容器资源使用
docker stats

# 查看镜像大小
docker images | grep gaokao

# 系统资源使用
docker system df
```

## 🔒 安全最佳实践

### 生产环境安全检查清单

- [ ] 修改所有默认密码
- [ ] 使用强密钥和证书
- [ ] 禁用Swagger UI
- [ ] 配置CORS白名单
- [ ] 启用HTTPS和HSTS
- [ ] 配置防火墙规则
- [ ] 定期更新镜像
- [ ] 运行安全扫描
- [ ] 配置日志监控
- [ ] 设置备份策略

### 开发环境安全注意事项

- 不要在开发环境使用生产数据
- 定期清理开发环境数据
- 不要提交密钥到版本控制
- 使用.env文件管理配置

## 📊 监控和维护

### 监控指标

系统提供以下监控指标：

- **应用指标**: QPS、响应时间、错误率
- **系统指标**: CPU、内存、磁盘、网络
- **数据库指标**: 连接数、查询性能、锁等待
- **缓存指标**: 命中率、内存使用、网络IO

### 日志管理

- **开发环境**: 文本格式，输出到stdout
- **生产环境**: JSON格式，输出到文件
- **日志轮转**: 自动清理过期日志
- **错误聚合**: 重要错误自动告警

### 备份策略

```bash
# 数据库备份
docker exec gaokao-postgres-prod pg_dump -U gaokao_user gaokao_prod > backup.sql

# Redis备份
docker exec gaokao-redis-prod redis-cli SAVE
docker cp gaokao-redis-prod:/data/dump.rdb ./redis-backup.rdb

# 配置备份
tar -czf config-backup.tar.gz docker/
```

## 🎯 性能优化

### 镜像优化

- 使用多阶段构建减小镜像大小
- 使用Alpine Linux作为基础镜像
- 合并RUN指令减少镜像层数
- 使用.dockerignore排除不必要文件

### 运行时优化

- 配置合适的资源限制
- 使用健康检查确保服务可用
- 配置合适的重启策略
- 使用缓存减少数据库查询

### 网络优化

- 使用内部网络减少延迟
- 配置连接池优化连接使用
- 使用Nginx负载均衡
- 启用HTTP/2和压缩

## 📞 技术支持

如果您在使用过程中遇到问题，请：

1. 查看本文档的故障排除章节
2. 检查相关日志文件
3. 在项目仓库提交Issue
4. 联系技术支持团队

## 📄 许可证

本项目采用 MIT 许可证，详见 LICENSE 文件。