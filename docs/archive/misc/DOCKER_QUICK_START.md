# 高考志愿填报系统 - Docker 快速开始指南

## 🚀 5分钟快速部署

### 1. 前置要求

- Docker 20.10+
- Docker Compose 2.0+
- 至少4GB可用内存

### 2. 快速启动开发环境

```bash
# Windows (PowerShell 或 Git Bash)
cd D:\mybitcoin\gaokao

# Linux/macOS
cd /path/to/gaokao

# 一键启动开发环境
docker-compose -f docker/dev/docker-compose.dev.yml up -d
```

### 3. 访问服务

等待2-3分钟后，访问以下地址：

- **主服务**: http://localhost:8080
- **API文档**: http://localhost:8080/swagger/index.html
- **数据库管理**: http://localhost:5050
- **监控面板**: http://localhost:3000

### 4. 测试API

```bash
# 健康检查
curl http://localhost:8080/health

# API测试
curl http://localhost:8080/api/v1/users
```

## 🛠️ 高级操作

### 构建自定义镜像

```bash
# Windows
.\docker\scripts\build.sh dev

# Linux/macOS
./docker/scripts/build.sh dev
```

### 生产环境部署

```bash
# 1. 配置生产环境变量
cp docker/prod/.env.example docker/prod/.env
# 编辑 docker/prod/.env 修改密码

# 2. 部署生产环境
.\docker\scripts\deploy.sh prod  # Windows
./docker/scripts/deploy.sh prod  # Linux/macOS
```

### 运行测试

```bash
# 运行所有测试
.\docker\scripts\test.sh  # Windows
./docker/scripts/test.sh  # Linux/macOS
```

### 清理环境

```bash
# 停止所有服务
docker-compose -f docker/dev/docker-compose.dev.yml down

# 完整清理
.\docker\scripts\cleanup.sh  # Windows
./docker/scripts/cleanup.sh  # Linux/macOS
```

## 🔧 故障排除

### 端口冲突

如果8080端口被占用，修改 `docker/dev/.env` 文件：

```bash
API_GATEWAY_PORT=8081  # 改为其他端口
```

### 内存不足

最少需要4GB内存，如果内存不足：

```bash
# 清理无用资源
docker system prune -f
```

### 权限问题 (Linux/macOS)

```bash
# 给脚本执行权限
chmod +x docker/scripts/*.sh
```

## 📋 默认账号信息

### 开发环境

- **数据库**: 
  - Host: localhost:5432
  - User: gaokao_user
  - Password: gaokao_pass
  - Database: gaokao_dev

- **pgAdmin**: 
  - URL: http://localhost:5050
  - Email: admin@gaokao.dev
  - Password: admin123

- **Grafana**: 
  - URL: http://localhost:3000
  - Username: admin
  - Password: admin123

- **Redis**: 
  - Host: localhost:6379
  - Password: (无密码)

## 📞 获取帮助

- 查看完整文档: `docker/README.md`
- 查看配置说明: `docker/dev/.env.example`
- 问题反馈: 创建 GitHub Issue