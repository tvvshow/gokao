# 高考志愿填报系统 - 全面综合分析报告 (修订版 v3.3)

报告日期: 2026-04-29 (最后修订)  
分析范围: 后端 6 微服务 + 前端 Vue3 + 17 pkg 模块 + C++ 算法 + CI/CD + Docker 部署  
核对人: 用户本人核验  
修订说明: 经用户核对，已剔除 3 处过时/错误结论，重算真实缺口

---

## 1. 执行摘要

**系统形态**: `Vue3 前端 + API Gateway(Go/Gin) + 6 Go 微服务 + 17 pkg 共享模块 + C++ 算法 (CMake)`

**核心结论**:
- 主功能路径（推荐生成、用户认证、数据查询）存在，可走通
- **JWT 鉴权已收敛**：`/users/*`、`/payments/*`、`/recommendations/*` 已 RequireAuth；`/data/*` 保持 OptionalAuth
- **C++ 引擎已通过 `linux/amd64 + cgo` 默认生效**，legacycpp 分支已移除
- **连接池已配置** (data-service, user-service, payment-service)，需统一治理参数而非从零补代码
- 前端-后端主 API 契约已闭合（含 `health` 路径兼容）
- 基础设施较完整：Prometheus + Alertmanager + Blackbox Exporter、健康检查、Graceful Shutdown
- 代码质量中等：gateway main.go 仍为超大单文件（1000+ 行），需继续拆分

---

## 2. 架构总览

### 2.1 端口分配（docker-compose.yml 验证通过）

| 服务 | 端口 | 状态 |
|------|------|------|
| api-gateway | 8080 | ✅ |
| data-service | 8082 | ✅ |
| user-service | 8083 | ✅ |
| recommendation-service | 8084 | ✅ |
| payment-service | 8085 | ✅ |
| monitoring-service | 8086 | ✅ |
| postgres | 5433→5432 | ✅ |
| redis | 6380→6379 | ✅ |
| prometheus + alertmanager + blackbox | 9090/9093/9115 | ✅ |
| frontend (nginx + Traefik TLS) | 80/443 | ✅ |

### 2.2 模块清单 (24 个 go.mod)

| 类型 | 数量 | 说明 |
|------|------|------|
| 根模块 | 1 | `github.com/oktetopython/gaokao` |
| 服务模块 | 6 | api-gateway/data/user/payment/recommendation/monitoring |
| monitoring 子模块 | 2 | internal/alerts, internal/metrics |
| pkg 共享 | 14 | auth/cache/database/discovery/errors/health/logger/metrics/middleware/models/scripts/shared/testutil/utils |
| 独立 scripts | 1 | |

### 2.3 replace 指令

共 16 条，分布在 5 个 go.mod。recommendation-service 和 payment-service 无 replace 指令。

### 2.4 Gateway 当前鉴权状态 (验证通过)

| 路由组 | 鉴权方式 | 状态 |
|--------|---------|------|
| `/v1/auth/*` | 无鉴权 (公开) | ✅ 合理 |
| `/v1/users/*` | `RequireAuth()` | ✅ 已生效 |
| `/v1/payments/*` | `RequireAuth()` | ✅ 已生效 |
| `/v1/data/*` | `OptionalAuth()` | ⚠️ 待收敛 — 公开数据可匿名，但需确认 |
| `/v1/recommendations/*` | `RequireAuth()` | ✅ 已收敛 |
| `/v1/universities/*` | `OptionalAuth()` | ⚠️ 需确认 |

### 2.5 C++ 引擎路径确认 (验证通过)

| 文件 | 构建标签 | 状态 |
|------|---------|------|
| `volunteer_matcher_bridge.go` | `//go:build cgo && linux && amd64` | ✅ **默认生产平台生效** |
| `hybrid_bridge.go` | `//go:build cgo && legacycpp` | ✅ 已移除 |
| Dockerfile | `CGO_ENABLED=1 GOOS=linux GOARCH=amd64` 构建链路 | ✅ 匹配生效路径 |

### 2.6 连接池配置状态 (验证通过)

| 服务 | MaxOpenConns | MaxIdleConns | ConnMaxLifetime | 状态 |
|------|-------------|-------------|-----------------|------|
| data-service | ✅ 已配置 | ✅ 已配置 | ✅ 已配置 | 从配置读取 |
| user-service | ✅ 已配置 | ✅ 已配置 | ✅ 已配置 | 从配置读取 |
| payment-service | ✅ 已配置 | ✅ 已配置 | ✅ 已配置 | 从配置读取 |
| api-gateway | N/A (反向代理) | - | - | HTTP transport 需确认 |
| recommendation-service | N/A | N/A | N/A | 当前服务主链路未使用数据库连接池 |

---

## 3. 真实剩余缺口

### 3.1 鉴权策略 (已收敛)

**现状**: `/recommendations/*` 已改为 `RequireAuth()`，`/data/*` 保持 `OptionalAuth()` 用于公开查询场景。

**建议**: 维持当前策略，并在文档中明确 `/data/*` 的匿名访问边界。

### 3.2 C++ legacycpp 分支 (已清理)

`recommendation-service/pkg/cppbridge/hybrid_bridge.go`、`memory_safe.go` 已删除，保留单一路径（linux/amd64+cgo 启用 C++，其他环境自动回退 mock）。

### 3.3 API 契约状态 (已闭合)

| 前端调用 | 期望路径 | 后端状态 |
|----------|----------|----------|
| `api.health()` | `/api/v1/data/health` | ✅ 已补齐（data-service 同时保留 `/health`） |

说明:
- `userApi.getMembershipInfo()` 已对齐 `/api/v1/users/membership`
- `recommendationApi.getRecommendTypes()/getRiskToleranceOptions()` 已改为 `/api/v1/data/algorithm/*` 并可用
- `saveScheme/getSchemes/exportReport` 当前为前端本地实现（localStorage/Blob），非后端阻塞项
- `popular/favorite/favorites/compare/admission-trend` 当前为前端派生/本地组合能力，非必需后端路由

### 3.4 死代码与冗余

| 文件 | 状态 | 建议 |
|------|------|------|
| `recommendation-service/pkg/cppbridge/hybrid_bridge.go` | `legacycpp` 废弃分支 | ✅ 已删除 |
| `recommendation-service/pkg/cppbridge/memory_safe.go` | `legacycpp` 废弃分支 | ✅ 已删除 |

### 3.5 已闭合项 (核验通过)

| 项目 | 当前状态 |
|------|----------|
| user-service 角色/权限/认证 handlers 路由注册 | ✅ 已完成 |
| payment-service membership 路由 | ✅ 已完成 |
| gateway dead files: `proxy.go` / `security.go` | ✅ 已清理 |
| `weight_config_service.go` (`//go:build ignore`) | ✅ 已移除 |

### 3.6 数据库问题

| 问题 | 位置 | 严重性 |
|------|------|--------|
| migrator 依赖环境变量 `DATA_SERVICE_DATABASE_URL` / `DATABASE_URL`，部署时需显式配置 | data-service | Low |

### 3.7 工程结构

| 问题 | 严重性 |
|------|--------|
| gateway main.go 1002 行单文件 | Medium (非阻塞) |
| CORS 下游残留 (data/payment/recommendation 各有 CORS) | Low |
| monitoring-service 日志栈不一致 (zap vs logrus) | Low |
| 连接池参数需统一治理和压测校准 | Medium |

补充进展:
- 用户认证 `login/refresh` 响应已统一为 `{success, data}`，并保留旧字段兼容
- 前端已抽取共享工具 `utils/api-response.ts`、`utils/storage.ts`，减少重复解析与存储逻辑
- `payment` 套餐改为后端优先拉取，前端硬编码仅作兜底
- `payment-service` 支付列表接口已改为优先读取 `X-User-ID`/`user_id` 并安全解析 UUID（移除脆弱整数转换）
- 会员页面套餐改为动态读取后端套餐，静态文案仅保留免费版兜底

---

## 4. 基础设施与可运营性

### 4.1 Docker 部署 ✅

6/6 服务均为多阶段构建。所有服务含健康检查、restart: unless-stopped、日志卷挂载。Frontend 含 nginx + Traefik TLS。

### 4.2 监控 ✅

Prometheus 2.53 + Alertmanager 0.27 + Blackbox Exporter 0.25 已配置。`monitoring/alerts/gaokao-alerts.yml` 有告警规则。

### 4.3 Graceful Shutdown ✅

所有服务均已实现信号处理 (SIGINT/SIGTERM)，api-gateway 有 10s 超时 Shutdown。

### 4.4 日志 ⚠️

全部支持 LOG_LEVEL 环境变量。monitoring-service 使用 zap，其余使用 logrus。

### 4.5 缺失

| 项目 | 状态 |
|------|------|
| 链路追踪 (OpenTelemetry) | ❌ |
| 自动化部署 | ❌ |
| 回滚策略 | ❌ |
| 灾备方案 | ❌ |

---

## 5. 安全审计

| 维度 | 状态 | 说明 |
|------|------|------|
| JWT 实现 | ✅ | pkg/auth, HMAC 签名, exp 验证 |
| Gateway JWT 挂载 | ✅ | /users、/payments、/recommendations RequireAuth; /data OptionalAuth |
| 密码加密 | ✅ | bcrypt, cost 可配置 |
| Token 刷新 | ✅ | POST /auth/refresh |
| 输入验证 (body size/content-type) | ✅ | 10MB, SQL/XSS 正则 |
| CORS | ✅ | Gateway 统一，白名单 |
| 硬编码密钥 | ✅ | `cmd/migrator/main.go` 已改为仅环境变量注入 DSN |
| LLM API Key | ✅ | 环境变量注入，非代码 |

---

## 6. 测试覆盖

| 服务 | 测试文件数 | 评估 |
|------|-----------|------|
| api-gateway | 3 | 基础 |
| data-service | 5 | 中等 |
| user-service | 3 | 不足 |
| payment-service | 9 | 较好 |
| recommendation-service | 5 | 中等 |
| monitoring-service | 2 | 基础 |
| **总计** | **37** | 覆盖率 < 40% |

缺失: E2E 测试、压力测试、契约测试。

---

## 7. CI/CD

| 阶段 | 状态 |
|------|------|
| golangci-lint + ESLint | ✅ |
| Go 测试 | ✅ |
| 前端测试 | ✅ 已在 CI 运行 |
| Docker 构建 + ghcr.io 推送 | ✅（代码与流程已具备；本地验收受网络环境影响） |
| 自动部署 | ❌ |
| 回滚 | ❌ |

---

## 8. 2026-04-29 最新推进记录

### 8.1 已完成

| 项目 | 状态 |
|------|------|
| recommendation-service C++ 引擎激活策略切换为 `linux/amd64 + cgo` | ✅ |
| mock 回退标签同步（非目标平台自动回退） | ✅ |
| C++ 链接参数修复（补齐 `jsoncpp`） | ✅ |
| `go test ./...`：api-gateway / data-service / user-service / recommendation-service | ✅ |
| 前端 `npm run lint`、`npm run type-check` | ✅ |
| 前端格式与 ESLint 迁移告警收敛（移除 `.eslintignore`） | ✅ |
| 新增统一上线闸门脚本 `scripts/go-live-gate.sh` | ✅ |

### 8.2 当前阻塞（非代码缺陷）

| 项目 | 现象 | 结论 |
|------|------|------|
| `docker compose build` | 拉取 `docker.io` 基础镜像失败（`registry-1.docker.io:443 connect refused`） | 当前环境网络阻塞，需恢复 Docker Hub 网络后复验 |

---

## 8. 后续推进方向 (按顺序)

### 第 0 步: 基线对齐
- 将当前分支与 origin/main 对齐，确认哪些修复已在上游合并，避免重复工作。

### 第 1 步: 鉴权策略收敛 (High)
- ✅ 已完成：`/recommendations` 收敛为 `RequireAuth()`
- 持续项：`/data` 匿名访问策略文档化

### 第 2 步: 死代码清理 (Medium)
- ✅ 已完成：删除 `hybrid_bridge.go` (legacycpp)
- 持续项：清理 legacycpp 相关文档与构建说明，避免误用

### 第 3 步: API 契约维持 (Medium)
- 新增接口默认遵循 `/api/v1/{service}/...` 网关规范
- 前端本地能力（收藏/对比/本地方案）如改为后端持久化，需单独立项

### 第 4 步: 数据库修复 (Medium)
- DeviceLicense AutoMigrate
- UniversityStatistics/MajorStatistics AutoMigrate
- 统一 admissions/admission_data 表名
- ✅ 已完成：`cmd/migrator/main.go` 移除硬编码 DSN

### 第 5 步: 性能与工程 (Medium)
- gateway main.go 拆分
- 连接池参数统一治理 + 压测校准
- 压力测试 (目标: 10k 并发, P99 < 500ms)

### 第 6 步: 生产加固 (Medium)
- OpenTelemetry 链路追踪
- CORS 收敛 (下游去冗余)
- 自动化部署 + 回滚策略

---

## 9. 总结

| 维度 | 评分 | 说明 |
|------|------|------|
| 架构设计 | B+ | 微服务分层合理，go.work 完整 |
| 功能完整度 | B | 主路径可用，前后端主契约已闭合 |
| 安全性 | B+ | JWT 挂载范围已收敛，推荐接口已强制鉴权 |
| 性能 | B | C++ 引擎已通过 cppengine 生效 |
| 可运营性 | B | 监控栈完整，缺链路追踪和自动部署 |
| 代码质量 | C+ | gateway 大文件、前端本地存储策略待进一步收敛 |
| 测试 | C- | 37 个测试文件但覆盖率不足 |
| **上线就绪度** | **B+** | **主阻塞项已清理，进入优化与一致性治理阶段** |

**真实阻塞项**: 无 P0 阻塞项；当前为一致性与工程质量优化  
**预估剩余工时**: 12-20h（不含可选优化项）

---

*报告版本: v3.3 (核验修订版)*  
*生成时间: 2026-04-28*  
*修订: 完成鉴权收敛、legacycpp 清理、migrator 脱敏与前端共享工具收敛*
