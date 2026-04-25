# 高考志愿填报系统 (Gaokao)

一个基于微服务的高考志愿推荐与数据查询系统，包含 Go 后端服务、Vue 3 前端，以及 C++ 推荐模块。

## 技术栈

- 后端: Go 1.25 + Gin（多服务）
- 前端: Vue 3 + TypeScript + Vite
- 数据: PostgreSQL + Redis（data-service 集成 Elasticsearch）
- 推荐: Go + CGO + C++ (`cpp-modules/`)
- 运维: Docker Compose + Nginx + Prometheus

## 仓库结构

```text
gaokao/
├── services/                  # Go 微服务
│   ├── api-gateway
│   ├── data-service
│   ├── user-service
│   ├── payment-service
│   ├── recommendation-service
│   └── monitoring-service
├── frontend/                  # Vue 3 前端
├── cpp-modules/               # C++ 模块（设备指纹/推荐等）
├── pkg/                       # 共享 Go 包
├── docker/                    # Docker 相关文件
├── docs/                      # 架构与审计文档
├── Makefile
└── docker-compose.yml
```

## 快速开始

### 1) 安装依赖

```bash
make deps
```

### 2) 启动基础设施与服务（Docker）

```bash
docker-compose up -d
```

默认端口（见 `docker-compose.yml`）:
- 前端: `80`
- API Gateway: `8080`
- data-service: `8082`
- user-service: `8083`
- recommendation-service: `8084`
- payment-service: `8085`
- monitoring-service: `8086`

### 3) 本地开发（可选）

```bash
# 后端构建
make build-go

# 前端开发模式
cd frontend && npm run dev
```

## 构建与测试

```bash
# 全量构建（Go + Frontend）
make build

# 全量测试（Go + Frontend）
make test

# 仅 Go 测试
make test-go

# 仅前端测试
make test-frontend
```

前端质量检查:

```bash
cd frontend
npm run lint
npm run type-check
```

## API 与文档

- API 网关入口: `http://localhost:8080`
- 若修改了 `services/api-gateway` 接口注释，需更新 Swagger 文档（CI 会校验）:

```bash
cd services/api-gateway
go run github.com/swaggo/swag/cmd/swag@v1.8.12 init -g main.go -o docs --parseDependency --parseInternal
```

## 治理与约束

- 项目宪章: `.specify/memory/constitution.md`
- 约束协议: `docs/Gaokao_Constraint_Protocol_v1.0.md`
- 所有变更需满足“最优秀原则 / 有即复用原则 / 不允许简化原则”，并通过质量闸门后合并。

## 贡献说明

提交前建议至少执行:

```bash
make test
cd frontend && npm run lint && npm run type-check
```

提交信息建议采用 Conventional Commits，例如:
- `fix(api-gateway): correct proxy prefix handling`
- `feat(frontend): add recommendation filters`
- `chore(ci): tighten swagger check`
