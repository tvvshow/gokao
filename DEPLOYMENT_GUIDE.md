# 🚀 高考志愿填报助手 - 开发者部署指南

## 📋 目录

1. [快速部署](#快速部署)
2. [开发环境搭建](#开发环境搭建)
3. [生产环境部署](#生产环境部署)
4. [Docker容器化部署](#docker容器化部署)
5. [Kubernetes集群部署](#kubernetes集群部署)
6. [CI/CD流水线](#cicd流水线)
7. [监控和日志](#监控和日志)
8. [安全配置](#安全配置)
9. [性能优化](#性能优化)
10. [故障排除](#故障排除)

---

## ⚡ 快速部署

### 🛠️ 系统要求

| 组件 | 最小配置 | 推荐配置 | 生产配置 |
|------|----------|----------|----------|
| **CPU** | 2核心 | 4核心 | 8核心+ |
| **内存** | 4GB | 8GB | 16GB+ |
| **存储** | 20GB | 50GB | 100GB+ |
| **网络** | 100Mbps | 1Gbps | 10Gbps |

### 📦 依赖环境

```bash
# 必需软件
- Docker 20.0+
- Docker Compose 2.0+
- Git 2.30+

# 可选软件 (开发用)
- Go 1.21+
- Node.js 18+
- Redis 7+
- PostgreSQL 15+
```

### 🎯 一键部署

```bash
# 1. 克隆代码库
git clone https://github.com/oktetopython/gaokao.git
cd gaokao

# 2. 复制环境配置
cp .env.example .env

# 3. 启动所有服务
./scripts/build-all.sh --production
docker-compose up -d

# 4. 验证部署
./scripts/health-check.sh
```

---

## 💻 开发环境搭建

### 🔧 本地开发环境

#### 1. Go开发环境
```bash
# 安装Go 1.21+
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# 配置环境变量
export PATH=$PATH:/usr/local/go/bin
export GOPROXY=https://goproxy.cn,direct
export GOSUMDB=sum.golang.google.cn
```

#### 2. C++开发环境
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install -y build-essential cmake libssl-dev libsqlite3-dev libjsoncpp-dev

# CentOS/RHEL
sudo yum groupinstall -y "Development Tools"
sudo yum install -y cmake openssl-devel sqlite-devel jsoncpp-devel

# macOS
brew install cmake openssl sqlite jsoncpp
```

#### 3. 数据库环境
```bash
# 启动开发数据库
docker-compose -f docker-compose.dev.yml up -d postgres redis

# 初始化数据库
./scripts/db-init.sh --dev
```

### 🛠️ 开发工具配置

#### VS Code配置
```json
// .vscode/settings.json
{
  "go.toolsManagement.checkForUpdates": "local",
  "go.useLanguageServer": true,
  "go.gocodeAutoBuild": false,
  "go.lintTool": "golangci-lint",
  "go.formatTool": "goimports",
  "C_Cpp.default.configurationProvider": "ms-vscode.cmake-tools"
}
```

#### Git Hooks配置
```bash
# 安装pre-commit钩子
./scripts/install-hooks.sh

# 手动运行代码检查
./scripts/code-check.sh
```

---

## 🏭 生产环境部署

### 🔒 安全配置

#### 1. 环境变量配置
```bash
# .env.production
NODE_ENV=production
GIN_MODE=release
ENABLE_SWAGGER=false

# 数据库配置
DATABASE_URL=postgres://user:password@localhost:5432/gaokao
REDIS_URL=redis://localhost:6379

# 安全密钥
JWT_SECRET=your-super-secret-jwt-key
ENCRYPTION_KEY=your-32-char-encryption-key

# 外部服务
WECHAT_APP_ID=your-wechat-app-id
ALIPAY_APP_ID=your-alipay-app-id
```

#### 2. SSL证书配置
```bash
# 使用Let's Encrypt
sudo certbot --nginx -d your-domain.com

# 手动证书配置
sudo mkdir -p /etc/ssl/gaokao
sudo cp your-cert.pem /etc/ssl/gaokao/
sudo cp your-key.pem /etc/ssl/gaokao/
sudo chmod 600 /etc/ssl/gaokao/*
```

#### 3. 防火墙配置
```bash
# UFW (Ubuntu)
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable

# iptables
sudo iptables -A INPUT -p tcp --dport 22 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 80 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 443 -j ACCEPT
```

### 📊 负载均衡配置

#### Nginx配置
```nginx
# /etc/nginx/sites-available/gaokao
upstream api_gateway {
    server 127.0.0.1:8080 weight=3;
    server 127.0.0.1:8081 weight=2;
    server 127.0.0.1:8082 weight=1;
}

server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;
    
    ssl_certificate /etc/ssl/gaokao/cert.pem;
    ssl_certificate_key /etc/ssl/gaokao/key.pem;
    
    location / {
        proxy_pass http://api_gateway;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    location /health {
        access_log off;
        proxy_pass http://api_gateway;
    }
}
```

---

## 🐳 Docker容器化部署

### 📦 多阶段构建

#### Go服务Dockerfile
```dockerfile
# services/data-service/Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs

EXPOSE 8082
CMD ["./main"]
```

#### C++模块Dockerfile
```dockerfile
# cpp-modules/Dockerfile
FROM gcc:latest AS builder

RUN apt-get update && apt-get install -y \
    cmake \
    libssl-dev \
    libsqlite3-dev \
    libjsoncpp-dev

WORKDIR /app
COPY . .

RUN mkdir build && cd build && \
    cmake .. -DCMAKE_BUILD_TYPE=Release && \
    make -j$(nproc)

FROM ubuntu:22.04
RUN apt-get update && apt-get install -y \
    libssl3 \
    libsqlite3-0 \
    libjsoncpp25 \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/build/ /usr/local/bin/
EXPOSE 9000
CMD ["./recommendation-engine"]
```

### 🏗️ Docker Compose配置

#### 生产环境配置
```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  # Nginx代理
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/ssl/certs:ro
    depends_on:
      - api-gateway
    restart: unless-stopped

  # API网关 (多实例)
  api-gateway:
    image: gaokao/api-gateway:${VERSION:-latest}
    deploy:
      replicas: 3
    environment:
      - GIN_MODE=release
      - ENABLE_SWAGGER=false
    depends_on:
      - data-service
    restart: unless-stopped

  # 数据服务 (多实例)
  data-service:
    image: gaokao/data-service:${VERSION:-latest}
    deploy:
      replicas: 2
    environment:
      - GIN_MODE=release
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
    depends_on:
      - postgres
      - redis
    restart: unless-stopped

  # 数据库集群
  postgres:
    image: postgres:15
    environment:
      - POSTGRES_DB=${POSTGRES_DB}
      - POSTGRES_USER=${POSTGRES_USER}
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backups:/backups
    restart: unless-stopped

  # Redis集群
  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
```

---

## ☸️ Kubernetes集群部署

### 🏗️ 基础配置

#### Namespace
```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: gaokao
  labels:
    name: gaokao
```

#### ConfigMap
```yaml
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: gaokao-config
  namespace: gaokao
data:
  GIN_MODE: "release"
  ENABLE_SWAGGER: "false"
  DATABASE_HOST: "postgres-service"
  REDIS_HOST: "redis-service"
```

#### Secret
```yaml
# k8s/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: gaokao-secrets
  namespace: gaokao
type: Opaque
stringData:
  DATABASE_URL: "postgres://user:password@postgres:5432/gaokao"
  JWT_SECRET: "your-super-secret-jwt-key"
  REDIS_PASSWORD: "your-redis-password"
```

### 🚀 服务部署

#### API Gateway Deployment
```yaml
# k8s/api-gateway.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-gateway
  namespace: gaokao
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api-gateway
  template:
    metadata:
      labels:
        app: api-gateway
    spec:
      containers:
      - name: api-gateway
        image: gaokao/api-gateway:latest
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: gaokao-config
        - secretRef:
            name: gaokao-secrets
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /readyz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: api-gateway-service
  namespace: gaokao
spec:
  selector:
    app: api-gateway
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

#### 数据库StatefulSet
```yaml
# k8s/postgres.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: gaokao
spec:
  serviceName: postgres-service
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:15
        ports:
        - containerPort: 5432
        envFrom:
        - secretRef:
            name: gaokao-secrets
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
  volumeClaimTemplates:
  - metadata:
      name: postgres-storage
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 20Gi
```

### 🔄 自动扩容配置

#### HPA (水平自动扩容)
```yaml
# k8s/hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: api-gateway-hpa
  namespace: gaokao
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api-gateway
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

---

## 🔄 CI/CD流水线

### 🚁 Drone CI配置

项目已包含完整的Drone CI/CD配置 (`.drone.yml`)，支持：

- ✅ **多阶段构建** - Go服务 + C++模块
- ✅ **代码质量检查** - 静态分析 + 安全扫描
- ✅ **单元测试** - 覆盖率报告
- ✅ **Docker构建** - 多架构镜像
- ✅ **自动部署** - 测试/生产环境

#### 手动触发部署
```bash
# 推送到master分支 -> 生产镜像构建
git push origin master

# 推送到develop分支 -> 测试环境部署
git push origin develop

# 创建发布标签 -> 生产环境部署
git tag v1.0.0
git push origin v1.0.0
```

### 🎯 GitHub Actions配置

```yaml
# .github/workflows/ci.yml
name: CI/CD Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Setup Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Run tests
      run: |
        for service in services/*/; do
          cd $service
          go test -v -race -coverprofile=coverage.out ./...
          cd ../..
        done
    
    - name: Build Docker images
      run: docker-compose build
    
    - name: Security scan
      run: |
        docker run --rm -v $(pwd):/app securecodewarrior/docker-gosec /app

  deploy:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    
    steps:
    - name: Deploy to production
      run: |
        # 部署逻辑
        kubectl apply -f k8s/
```

---

## 📊 监控和日志

### 📈 Prometheus监控

#### 配置文件
```yaml
# monitoring/prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'gaokao-api-gateway'
    static_configs:
      - targets: ['api-gateway:8080']
    metrics_path: '/metrics'
    
  - job_name: 'gaokao-data-service'
    static_configs:
      - targets: ['data-service:8082']
    metrics_path: '/api/v1/performance/metrics'
```

#### 关键指标
```promql
# API请求量
rate(gaokao_http_requests_total[5m])

# 响应时间
histogram_quantile(0.95, gaokao_http_request_duration_seconds_bucket)

# 错误率
rate(gaokao_http_requests_total{status=~"5.."}[5m]) / 
rate(gaokao_http_requests_total[5m])

# 内存使用率
(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100
```

### 📊 Grafana仪表板

#### 系统概览面板
```json
{
  "dashboard": {
    "title": "高考志愿填报助手 - 系统监控",
    "panels": [
      {
        "title": "API请求量",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(gaokao_http_requests_total[5m])",
            "legendFormat": "{{method}} {{path}}"
          }
        ]
      },
      {
        "title": "响应时间分布",
        "type": "heatmap",
        "targets": [
          {
            "expr": "gaokao_http_request_duration_seconds_bucket",
            "legendFormat": "{{le}}"
          }
        ]
      }
    ]
  }
}
```

### 📝 日志聚合

#### ELK Stack配置
```yaml
# docker-compose.monitoring.yml
version: '3.8'

services:
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.8.0
    environment:
      - discovery.type=single-node
      - "ES_JAVA_OPTS=-Xms512m -Xmx512m"
    volumes:
      - elasticsearch_data:/usr/share/elasticsearch/data

  logstash:
    image: docker.elastic.co/logstash/logstash:8.8.0
    volumes:
      - ./logstash.conf:/usr/share/logstash/pipeline/logstash.conf:ro
    depends_on:
      - elasticsearch

  kibana:
    image: docker.elastic.co/kibana/kibana:8.8.0
    ports:
      - "5601:5601"
    depends_on:
      - elasticsearch
```

---

## 🔒 安全配置

### 🛡️ 应用层安全

#### 1. JWT安全配置
```go
// 安全的JWT配置
func setupJWT() {
    // 使用强密钥
    jwtKey := []byte(os.Getenv("JWT_SECRET")) // 至少32字符
    
    // 设置合理的过期时间
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "exp": time.Now().Add(time.Hour * 24).Unix(),
        "iat": time.Now().Unix(),
        "iss": "gaokao-hub",
    })
}
```

#### 2. 请求限流配置
```go
// 分级限流策略
rateLimits := map[string]int{
    "free":    100,  // 免费用户: 100次/小时
    "basic":   1000, // 基础会员: 1000次/小时
    "premium": 10000, // 高级会员: 10000次/小时
}
```

#### 3. 数据加密
```go
// 敏感数据加密存储
func encryptSensitiveData(data string) string {
    key := []byte(os.Getenv("ENCRYPTION_KEY"))
    encrypted, _ := encrypt([]byte(data), key)
    return base64.StdEncoding.EncodeToString(encrypted)
}
```

### 🔐 基础设施安全

#### 1. 网络安全
```bash
# 最小权限网络策略
iptables -P INPUT DROP
iptables -P FORWARD DROP
iptables -P OUTPUT ACCEPT

# 只允许必要端口
iptables -A INPUT -p tcp --dport 22 -j ACCEPT  # SSH
iptables -A INPUT -p tcp --dport 80 -j ACCEPT  # HTTP
iptables -A INPUT -p tcp --dport 443 -j ACCEPT # HTTPS
```

#### 2. 容器安全
```dockerfile
# 使用非root用户
FROM alpine:latest
RUN adduser -D -s /bin/sh appuser
USER appuser

# 最小化攻击面
RUN apk add --no-cache ca-certificates && \
    rm -rf /var/cache/apk/*
```

#### 3. 密钥管理
```bash
# 使用Kubernetes Secrets
kubectl create secret generic gaokao-secrets \
  --from-literal=jwt-secret="your-secret-key" \
  --from-literal=db-password="your-db-password"
```

---

## ⚡ 性能优化

### 🚀 应用层优化

#### 1. 数据库优化
```sql
-- 创建索引
CREATE INDEX CONCURRENTLY idx_universities_province ON universities(province);
CREATE INDEX CONCURRENTLY idx_admission_year_score ON admission_data(year, score);

-- 分区表
CREATE TABLE admission_data_2023 PARTITION OF admission_data 
FOR VALUES FROM ('2023-01-01') TO ('2024-01-01');
```

#### 2. 缓存策略
```go
// 多级缓存
type CacheConfig struct {
    RedisCache   *redis.Client // L1: Redis缓存
    MemoryCache  *bigcache.BigCache // L2: 内存缓存
    TTL          time.Duration
}

// 缓存预热
func (c *CacheService) WarmupCache() error {
    // 预加载热点数据
    universities, _ := c.db.GetTopUniversities()
    majors, _ := c.db.GetPopularMajors()
    
    // 缓存到Redis
    c.SetUniversities(universities)
    c.SetMajors(majors)
}
```

#### 3. 连接池优化
```go
// 数据库连接池配置
db.SetMaxOpenConns(25)      // 最大连接数
db.SetMaxIdleConns(10)      // 最大空闲连接数
db.SetConnMaxLifetime(300 * time.Second) // 连接最大生命周期
```

### 🔧 基础设施优化

#### 1. Docker优化
```dockerfile
# 多阶段构建减小镜像大小
FROM golang:1.21-alpine AS builder
# ... 构建阶段

FROM scratch
COPY --from=builder /app/main /main
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/main"]
```

#### 2. Nginx优化
```nginx
# gzip压缩
gzip on;
gzip_vary on;
gzip_min_length 1000;
gzip_types text/plain text/css application/json application/javascript;

# 静态文件缓存
location ~* \.(jpg|jpeg|png|gif|ico|css|js)$ {
    expires 1y;
    add_header Cache-Control "public, immutable";
}
```

#### 3. Kubernetes资源限制
```yaml
resources:
  requests:
    memory: "128Mi"
    cpu: "100m"
  limits:
    memory: "256Mi"
    cpu: "200m"
```

---

## 🔧 故障排除

### 🐛 常见问题诊断

#### 1. 服务启动失败
```bash
# 检查日志
docker-compose logs service-name

# 检查资源使用
docker stats

# 检查端口占用
netstat -tulpn | grep :8080

# 重启服务
docker-compose restart service-name
```

#### 2. 数据库连接问题
```bash
# 测试数据库连接
docker exec -it postgres-container psql -U username -d database

# 检查数据库状态
docker-compose ps postgres

# 查看数据库日志
docker-compose logs postgres
```

#### 3. 性能问题排查
```bash
# 检查系统资源
htop
iostat -x 1
free -h

# 检查应用性能
curl "http://localhost:8082/api/v1/performance/metrics"

# 数据库性能分析
docker exec postgres-container pg_stat_statements
```

### 🛠️ 调试工具

#### 1. 健康检查脚本
```bash
#!/bin/bash
# scripts/health-check.sh

echo "🔍 检查服务健康状态..."

services=("api-gateway:8080" "data-service:8082")
for service in "${services[@]}"; do
    IFS=':' read -r name port <<< "$service"
    if curl -f -s "http://localhost:$port/health" > /dev/null; then
        echo "✅ $name 服务正常"
    else
        echo "❌ $name 服务异常"
    fi
done
```

#### 2. 性能测试脚本
```bash
#!/bin/bash
# scripts/performance-test.sh

echo "🚀 开始性能测试..."

# API压力测试
ab -n 1000 -c 10 http://localhost:8080/api/v1/ping

# 数据库压力测试
pgbench -h localhost -U username -d database -c 10 -t 100
```

#### 3. 自动化故障恢复
```bash
#!/bin/bash
# scripts/auto-recovery.sh

while true; do
    if ! curl -f -s http://localhost:8080/health > /dev/null; then
        echo "⚠️ 检测到服务异常，尝试重启..."
        docker-compose restart api-gateway
        sleep 30
    fi
    sleep 60
done
```

---

## 📚 最佳实践

### 💡 开发最佳实践

1. **代码规范**
   - 使用 `golangci-lint` 进行代码检查
   - 遵循 Go 官方代码规范
   - 编写详细的注释和文档

2. **测试策略**
   - 单元测试覆盖率 > 80%
   - 集成测试覆盖关键路径
   - 性能测试验证SLA

3. **版本管理**
   - 使用语义化版本号
   - 详细的commit信息
   - 及时创建release

### 🚀 部署最佳实践

1. **环境管理**
   - 开发/测试/生产环境隔离
   - 配置文件版本化管理
   - 敏感信息加密存储

2. **监控告警**
   - 关键指标实时监控
   - 多级告警策略
   - 快速故障响应

3. **备份策略**
   - 数据库定期备份
   - 配置文件备份
   - 容灾恢复方案

---

## 📞 技术支持

### 🆘 获取帮助
- **文档中心**: https://docs.gaokaohub.com
- **GitHub Issues**: https://github.com/oktetopython/gaokao/issues
- **技术交流群**: QQ群 123456789
- **邮件支持**: devops@gaokaohub.com

### 📝 贡献指南
1. Fork 项目
2. 创建特性分支
3. 提交更改
4. 创建 Pull Request

---

**🚀 部署愉快！如有问题，随时联系技术团队！**