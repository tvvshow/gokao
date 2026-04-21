
# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

---

## 0. Global Protocols

### 0.1 交互

- **语言**：与工具交互用英语；与用户交互用中文。
- **状态**：工具返回 `SESSION_ID` 时立即记录；后续调用需追加 `--SESSION_ID <ID>`。若工具输出被截断，自动续传直到输出完整。

### 0.2 代码风格

- 定位：精简、算法高效、零冗余，科研级。注释与文档**非必要不形成**。
- 最小作用域：仅对需求做针对性改动，严禁影响其他功能。

### 0.3 工作流程完整性

- **止损**：当前阶段未通过验证前，不进入下一阶段。
- **报告**：实时向用户报告当前阶段与下一阶段。
- **跳阶审批**：跳过任何 Phase 均属危险操作，需立即停止并说明理由等待用户批准。

---

## 1. Workflow

### Phase 1: 上下文检索

生成任何建议或代码前执行。

1. **历史会话检索**：优先使用 claude-mem 插件（`mem-search`、`timeline`、`get_observations`）复用已有解法。
2. **代码库检索**：历史不足时用 Grep/Glob/Read 递归读取，直到类/函数/变量的定义与签名完整。
3. **需求对齐**：模糊处必须向用户追问，直到边界清晰。

### Phase 2: 分析与规划

1. **方案推演**：基于上下文进行多角度推演、逻辑验证、优劣比较，产出无缺口的 Step-by-step 计划（含适度伪代码）。
2. **强制阻断 (Hard Stop)**：
   - 向用户展示计划；
   - 以加粗输出询问：**"Shall I proceed with this plan? (Y/N)"**；
   - 立即终止回复。未收到明确 "Y" 前禁止进入 Phase 3。

### Phase 3: 编码实施

1. **实现**：按计划落地为高可读、可维护、发布级代码；非必要不写注释。
2. **副作用审查**：强制审查变更是否越界，发现即修正。
3. **分步提交**：涉及多模块时分步骤落盘，每步均可独立验证。

### Phase 4: 自审计与交付

1. **自审**：变更完成后用项目原生测试/lint/type-check 回归；必要时用 `/review` 或 `/security-review` skill 复核逻辑正确性、需求覆盖、潜在 Bug。
2. **交付**：通过后向用户汇报变更清单、验证结果、残留风险。

---

## 2. 项目概述

高考志愿填报系统 (Gaokao College Application System) - An AI-powered college entrance exam application and career planning assistant for Chinese high school students. The system provides intelligent university recommendations using a "冲稳保" (reach-match-safety) strategy. 项目遵循严格的铁笼协议 v4.0 约束。

**实际技术栈**（优先级高于 `README.md` — 该文件内容过时）：Go 微服务 + Vue 3/TypeScript/Vite + C++ 算法模块，**非** FastAPI/Python/React。

---

## 3. 核心架构

### 模块结构

```
项目根目录/
├── go.work                # Go Workspace (Go 1.25.5), 17 个 module
├── services/              # 微服务后端
│   ├── api-gateway/       # API 网关，JWT 认证，限流，Swagger 文档
│   ├── user-service/      # 用户管理，认证，设备指纹 (CGO)
│   ├── data-service/      # 大学/专业数据，GORM + PostgreSQL/SQLite
│   ├── payment-service/   # 支付处理（微信支付、支付宝、银联）
│   ├── recommendation-service/  # AI 推荐引擎，CGO 桥接 C++
│   └── monitoring-service/      # 指标和告警
├── pkg/                   # 共享包（每个子包均为独立 module，见 go.work）
│   ├── auth/ errors/ middleware/ database/ cache/ discovery/ metrics/
│   ├── api/ response/ health/ logger/ models/ utils/
│   ├── shared/ testutil/ scripts/
├── cpp-modules/           # C++ 高性能原生模块（CMake 构建）
│   ├── device-fingerprint/ # 设备识别和加密
│   ├── volunteer-matcher/  # 录取预测算法
│   └── license/           # 许可证验证
├── frontend/              # Vue 3 + TypeScript + Vite + Pinia + Element Plus + Tailwind
│   ├── src/{views,components,stores,api,utils}
│   ├── CLAUDE.md          # 前端专属指引，编辑前端前必读
│   └── REFACTOR_PLAN.md / WORK_LOG.md
├── config/.env.{development,production}
├── .github/workflows/     # CI/CD
└── Makefile               # 顶层构建入口
```

### 子模块 CLAUDE.md 引用

改动特定模块前**必须先读**对应的子 CLAUDE.md（若存在），其中记录了模块级约束：
- `services/api-gateway/CLAUDE.md`
- `services/recommendation-service/CLAUDE.md`
- `frontend/CLAUDE.md`

### 微服务架构

- **API Gateway**: 统一入口，处理认证、限流、路由
- **User Service**: 用户管理、设备指纹、权限控制（CGO 依赖 OpenSSL）
- **Data Service**: 大学/专业数据管理，支持 PostgreSQL 和 SQLite
- **Payment Service**: 集成多种支付方式
- **Recommendation Service**: AI 推荐引擎，**强制** CGO 启用以调用 C++ 算法
- **Monitoring Service**: 系统监控和告警

---

## 4. 常用开发命令

### 构建关键约束

- **Go Workspace**：根目录 `go.work` 管理所有 module（Go 1.25.5）。跨 module 命令需在仓库根执行，单 module 命令 `cd` 进入后再运行。
- **CGO 强制启用**：`recommendation-service` 绑定 C++ 推荐引擎，构建全程保持 `CGO_ENABLED=1`（即使 `RELEASE=1`，见 `Makefile:40-46`）。
- **user-service 额外链接**：构建时需要 `-lssl -lcrypto`（OpenSSL），见 `Makefile:109-111`。独立 `go build` 时要自行带上 `-ldflags "-extldflags '-lssl -lcrypto'"`。

### 构建系统

```bash
# 完整构建
make all              # 清理 + 依赖 + 构建 + 测试

# Go服务
make build-go         # 构建所有Go服务到bin/
make test-go          # 运行Go测试，启用竞态检测
make deps-go          # go mod download && go mod tidy

# 前端
make build-frontend   # npm run build
make test-frontend    # npm run test:unit
make deps-frontend    # npm ci

# 单个服务构建
cd services/api-gateway && go build ./...
cd services/api-gateway && go test ./...

# C++模块
cd cpp-modules/device-fingerprint && mkdir -p build && cd build && cmake .. && make

# Docker
make docker           # 构建所有Docker镜像
docker-compose up -d  # 启动开发环境(postgres:5433, redis:6380)

# 代码检查
golangci-lint run --timeout=5m    # Go(使用.golangci.yml)
cd frontend && npm run lint       # 前端ESLint
cd frontend && npm run type-check # TypeScript检查
```

### 运行程序

```bash
# 启动所有服务
make dev

# 启动单个服务
cd services/api-gateway && go run main.go

# 前端开发服务器
cd frontend && npm run dev

# 数据库迁移
cd services/data-service && go run main.go migrate

# 生成API文档
cd services/api-gateway
go run github.com/swaggo/swag/cmd/swag@v1.8.12 init -g main.go -o docs --parseDependency --parseInternal
```

### 测试和调试

```bash
# Go测试
go test -v -run TestFunctionName ./path/to/package
cd services/payment-service && go test ./...
go test -v -race -coverprofile=coverage.out ./...

# 前端测试
cd frontend && npm run test:unit
cd frontend && npm run test:coverage

# 调试
dlv debug ./services/api-gateway/main.go
```

---

## 5. 关键技术约束

### 性能目标

- API Gateway: < 100ms 响应时间 (P95)
- Recommendation Service: < 500ms 推荐生成时间
- Frontend: < 3s 首屏加载时间
- 数据库查询: < 50ms (P95)

### 架构适配

- 开发环境: SQLite + 本地Redis
- 生产环境: PostgreSQL集群 + Redis集群
- 前端构建: Vite + TypeScript + Vue 3

### 重要参数说明

- 服务端口: API Gateway (8080), Data Service (8082), Frontend (3000)
- 数据库: PostgreSQL (5433), Redis (6380)
- 覆盖率要求: API Gateway测试覆盖率 ≥ 60%

---

## 6. 代码规范

### 命名约定 - 必须使用中性技术词汇

- ✅ 正确: `ApplicationProcessor`, `RecommendationEngine`, `UserProfileManager`, `DataValidator`
- ❌ 禁止: `StudentHack`, `ExamCracker`, `AdmissionBypass`, `ScoreManipulator`

### 注释语言

- 源码注释: English
- AI 回复、文档、日志: 简体中文
- 代码标识符: 中性技术词汇

---

## 7. 项目特定约束

### 高考志愿填报系统特有约束

1. **数据处理**：
   - 所有高考分数和排名数据必须经过验证
   - 录取分数线数据必须来自官方渠道
   - "冲稳保"策略必须严格实现，推荐结果必须可解释

2. **合规性**：
   - 遵循教育部门相关法规
   - 用户数据隐私保护符合 GDPR 和国内法规
   - 支付功能符合金融监管要求

3. **性能要求**：
   - 高峰期（高考成绩发布后）必须支持 10 万+ 并发用户
   - 推荐算法响应时间 < 500ms
   - 系统可用性 > 99.9%

---

## 8. 测试和性能验证

### 基准测试命令

```bash
# API性能测试
cd services/api-gateway && go test -bench=. -benchmem

# 推荐算法性能测试
cd services/recommendation-service && go test -bench=Recommendation -benchmem

# 前端性能测试
cd frontend && npm run test:e2e
```

### 回归测试

修改代码后必须：
1. 运行完整测试套件
2. 对比性能数据（±5% 容差）
3. 确认功能正确性

---

## 9. Claude-mem 使用指南

### 核心工作流程

1. **搜索阶段** - 始终从搜索开始：
   ```python
   # 搜索相关历史任务
   search(query="关键词", limit=20, project="GaokaoSystem")
   ```

2. **时间线阶段** - 理解任务演进：
   ```python
   # 获取搜索结果周围的时间线上下文
   timeline(query="关键词", depth_before=3, depth_after=3, project="GaokaoSystem")
   ```

3. **批量获取阶段** - 只获取需要的记录：
   ```python
   # 批量获取多个观察记录（2个或更多时必须使用）
   get_observations(ids=[11131, 10942, 10855])
   ```

### 关键原则

- **效率优先**：先搜索索引（~50-100 tokens），再获取详情（~500-1000 tokens）
- **批量操作**：获取多个记录时，始终使用 `get_observations` 而非单独调用
- **历史复用**：找到相似解决方案后，直接复用，避免重复工作
- **上下文完整**：确保有足够的历史上下文后再开始编码

### 搜索技巧

- 使用英文关键词搜索效果更佳
- 可以使用 `obs_type` 过滤：bugfix, feature, decision, discovery, change
- 使用日期过滤查找特定时间段的工作
- 优先查看最近的会话以获取最新进展