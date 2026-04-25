# 高考志愿填报系统 (Gaokao)

高考志愿填报系统是一个多服务架构项目，提供志愿推荐、数据查询、用户与支付能力。后端基于 Go 微服务，前端基于 Vue 3，推荐能力包含 C++ 模块并通过 CGO 集成。

## 技术栈

- 后端: Go 1.25 + Gin
- 前端: Vue 3 + TypeScript + Vite
- 数据层: PostgreSQL + Redis
- 推荐模块: C++ (`cpp-modules/`) + CGO
- 运维: Docker Compose + Nginx

## 目录结构

```text
services/                  # 业务微服务
  api-gateway              # 统一入口与路由聚合
  data-service             # 数据服务
  user-service             # 用户服务
  payment-service          # 支付服务
  recommendation-service   # 推荐服务（含 C++ 能力集成）
  monitoring-service       # 监控服务
frontend/                  # Vue 3 前端
pkg/                       # 共享 Go 模块
cpp-modules/               # C++ 模块源码
docker/                    # Docker 相关配置
docs/                      # 架构与说明文档
```

## 快速开始（推荐 Docker）

1. 安装依赖
```bash
make deps
```

2. 启动服务
```bash
docker compose up -d
```

3. 查看状态
```bash
docker compose ps
```

默认端口:
- Frontend: `80`
- API Gateway: `8080`
- Data Service: `8082`
- User Service: `8083`
- Recommendation Service: `8084`
- Payment Service: `8085`
- Monitoring Service: `8086`
- PostgreSQL: `5433`
- Redis: `6380`

## 本地开发

```bash
# 构建后端服务
make build-go

# 前端开发模式
cd frontend && npm run dev
```

`go.work` 已注册共享包与服务模块，建议在仓库根目录执行 Go 命令以获得一致依赖解析行为。

## 构建、测试与质量检查

```bash
make build           # Go + Frontend 全量构建
make test            # Go + Frontend 全量测试
make test-go         # 仅 Go 测试
make test-frontend   # 仅前端测试
```

前端质量检查:
```bash
cd frontend
npm run lint
npm run type-check
```

## API 与 Swagger

- 网关入口: `http://localhost:8080`
- 当修改 `services/api-gateway` 接口注释后，需同步 Swagger:

```bash
cd services/api-gateway
go run github.com/swaggo/swag/cmd/swag@v1.8.12 init -g main.go -o docs --parseDependency --parseInternal
```

## 贡献规范

提交前建议至少执行:

```bash
make test
cd frontend && npm run lint && npm run type-check
```

提交信息建议采用 Conventional Commits，例如:
- `fix(api-gateway): correct proxy prefix handling`
- `feat(frontend): add recommendation filters`
- `chore(ci): tighten swagger check`
