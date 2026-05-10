# 高考志愿填报系统 - 全面综合分析报告 (最终上线审计版 v3.5)

报告日期: 2026-04-30 (最后修订)  
分析范围: 后端 6 微服务 + 前端 Vue3 + 17 pkg 模块 + C++ 算法 + CI/CD + Docker 部署  
核对人: 用户本人核验  
修订说明: v3.5 补充注册链路前后端契约漂移分析（密码验证规则、字段对齐、错误信息传递）

---

## 1. 执行摘要

**系统形态**: `Vue3 前端 + API Gateway(Go/Gin) + 6 Go 微服务 + 17 pkg 共享模块 + C++ 算法 (CMake)`

**核心结论**:
- 主功能路径（推荐生成、用户认证、数据查询）代码存在，**但注册链路存在前后端契约漂移导致 400 错误**
- **P0 阻塞项**: 前端密码验证规则（大写+小写+数字）与后端（大写+小写+数字+特殊字符）不一致；400 错误信息被前端吞掉
- **JWT 鉴权已收敛**：`/users/*`、`/payments/*`、`/recommendations/*` 已 RequireAuth；`/data/*` 保持 OptionalAuth
- **C++ 引擎已通过 `linux/amd64 + cgo` 默认生效**，legacycpp 分支已移除
- **连接池已配置** (data-service, user-service, payment-service)，需统一治理参数而非从零补代码
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
| 根模块 | 1 | `github.com/tvvshow/gokao` |
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

### 3.3 API 契约状态

| 前端调用 | 期望路径 | 后端状态 |
|----------|----------|----------|
| `api.health()` | `/api/v1/data/health` | ✅ 已补齐（data-service 同时保留 `/health`） |
| `userApi.register()` | `/api/v1/users/auth/register` | ❌ 密码验证规则不一致，导致 400 |

说明:
- `userApi.getMembershipInfo()` 已对齐 `/api/v1/users/membership`
- `recommendationApi.getRecommendTypes()/getRiskToleranceOptions()` 已改为 `/api/v1/data/algorithm/*` 并可用
- `saveScheme/getSchemes/exportReport` 当前为前端本地实现（localStorage/Blob），非后端阻塞项
- `popular/favorite/favorites/compare/admission-trend` 当前为前端派生/本地组合能力，非必需后端路由

#### 注册链路前后端契约漂移（2026-04-30 发现）

**请求链路**: `LoginPage.vue → userStore.register() → userApi.register() → POST /api/v1/users/auth/register → nginx → api-gateway (rewriteServicePath) → user-service /api/v1/auth/register → authHandler.Register`

| 问题 | 前端行为 | 后端行为 | 影响 |
|------|----------|----------|------|
| 密码验证规则不一致 | `validators.ts:92-122` 仅检查大写+小写+数字 | `auth_handler.go:435-500` 额外要求特殊字符 + 禁止重复模式 | **用户输入 `Abc12345` 前端通过，后端返回 400** |
| `confirmPassword` 字段漂移 | `LoginPage.vue:111` 发送 `confirmPassword` | `RegisterRequest` 结构体无此字段 | 无功能影响（Gin 忽略多余字段），但契约不对齐 |
| `phone` 字段必填/可选不一致 | `types/user.ts:22` 定义 `phone?: string`（可选），`LoginPage.vue:239` 验证规则设为必填 | `RegisterRequest.Phone` binding 为 `max=20`（可选） | 前端强制收集手机号，后端不强制 |
| 400 错误信息被吞掉 | `api-client.ts:244-261` 对 400 无特殊处理，显示通用"请求失败 (400)" | 返回 `{"error":"password_complexity_failed","details":"password must contain ..."}` | 用户看不到有意义的错误原因 |

**请求/响应格式对照**:

```
前端发送 (RegisterForm):
{ username, email, password, phone?, confirmPassword }

后端期望 (RegisterRequest):
{ username, password, email, nickname?, phone?, province?, city?, gender?, birthday? }

后端成功响应 (201):
{ message: "User registered successfully", user_id: "..." }

前端解析 (normalizeMessageResponse):
检查 isWrappedResponse → 否 → 返回 { success: true, message: raw.message }
```

**修复优先级**: High — 注册是用户首次接触的核心路径，400 错误直接阻断用户转化。

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
| 注册链路代码路径完整性 | ✅ 前端→网关→user-service 路由畅通 |
| 注册链路前后端契约对齐 | ❌ 未闭合（见 3.3 注册专项） |

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
| 硬编码密钥 | ✅ 已清理 | migrator 改为 `DATA_SERVICE_DATABASE_URL` / `DATABASE_URL` 环境变量 |
| LLM API Key | ✅ | 环境变量注入，非代码 |

### 5.1 最终上线前新增审计发现（2026-04-30）

| 严重级别 | 项目 | 现状 | 证据 |
|------|------|------|------|
| **High** | 注册链路密码验证规则前后端不一致 | 前端不检查特殊字符，后端强制要求 → 注册 400 | `validators.ts:92-122` vs `auth_handler.go:435-500` |
| **High** | 注册 400 错误信息被前端吞掉 | 后端返回具体错误详情，前端仅显示通用"请求失败 (400)" | `api-client.ts:244-261` |
| High | JWT 默认回退值风险 | `JWT_SECRET` 为空时仍可使用内置默认值启动 | `services/api-gateway/main.go:521-524` |
| High | Compose 默认弱口令风险 | `docker-compose.yml` 含数据库与密钥默认值，存在误用面 | `docker-compose.yml:147,151,215` |
| High | 公网 LLM 默认出口风险 | 推荐服务默认 `LLM_BASE_URL=https://api.openai.com/v1`，需按生产策略确认数据出境边界 | `docker-compose.yml:183` |
| Medium | Gateway InputValidationMiddleware 可能误杀合法密码 | 密码含 SQL 关键字或 HTML 标签时被拦截返回 400 | `pkg/middleware/security.go:394-413` |
| Medium | 前端 lint 警告残留 | 无 error，但存在 3 条 warning | `frontend/src/__tests__/properties/api-auth-token.test.ts:10`、`frontend/src/__tests__/properties/route-lazy-loading.test.ts:17`、`frontend/src/components/common/VirtualList.vue:31` |
| Medium | 归档文档含 token push 示例 | 历史文档含 token URL push 示例，建议继续脱敏或移除 | `docs/archive/misc/PUSH_GUIDE.md:44` |

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
| Docker 构建 + ghcr.io 推送 | ✅（流程具备；本地环境验证受网络条件影响） |
| 自动部署 | ❌ |
| 回滚 | ❌ |

---

## 8. 后续推进方向 (按顺序)

### 第 0 步: 注册链路修复 (High, 阻塞用户转化)
- 前端 `validators.ts` 密码验证增加特殊字符检查，与后端 `auth_handler.go:validatePassword` 对齐
- 前端 `api-client.ts` 对 400 响应解析后端 `details` 字段并展示给用户
- 前端 `LoginPage.vue` 注册提交前去掉 `confirmPassword` 字段（或后端接受并校验一致性）
- 统一 `phone` 字段必填/可选约定（前端类型定义 vs 表单验证 vs 后端 binding）
- 考虑 Gateway InputValidationMiddleware 对 `/auth/*` 路径的密码字段跳过 SQL/XSS 检测，或将检测限制在 URL 参数和表单参数（当前 body 未检测，实际风险低）

### 第 1 步: 基线对齐
- 将当前分支与 origin/main 对齐，确认哪些修复已在上游合并，避免重复工作。

### 第 2 步: 鉴权策略收敛 (High)
- ✅ 已完成：`/recommendations` 收敛为 `RequireAuth()`
- 持续项：`/data` 匿名访问策略文档化

### 第 3 步: 死代码清理 (Medium)
- ✅ 已完成：删除 `hybrid_bridge.go` (legacycpp)
- ✅ 已完成：删除 `memory_safe.go` (legacycpp)
- 持续项：清理 legacycpp 相关文档与构建说明，避免误用

### 第 4 步: API 契约维持 (Medium)
- 新增接口默认遵循 `/api/v1/{service}/...` 网关规范
- 前端本地能力（收藏/对比/本地方案）如改为后端持久化，需单独立项
- 注册响应格式：后端返回 `{ message, user_id }`，前端通过 `normalizeMessageResponse` 兼容处理，可工作但建议统一为 `{ success, data, message }` 包装格式

### 第 5 步: 数据库修复 (Medium)
- DeviceLicense AutoMigrate
- UniversityStatistics/MajorStatistics AutoMigrate
- 统一 admissions/admission_data 表名
- ✅ 已完成：`cmd/migrator/main.go` 移除硬编码 DSN

### 第 6 步: 性能与工程 (Medium)
- gateway main.go 拆分
- 连接池参数统一治理 + 压测校准
- 压力测试 (目标: 10k 并发, P99 < 500ms)

### 第 7 步: 生产加固 (Medium)
- OpenTelemetry 链路追踪
- CORS 收敛 (下游去冗余)
- 自动化部署 + 回滚策略
- 收敛默认弱口令/默认密钥回退（仅允许强制环境变量注入）
- 明确 LLM 出网与数据脱敏策略

---

## 9. 总结

| 维度 | 评分 | 说明 |
|------|------|------|
| 架构设计 | B+ | 微服务分层合理，go.work 完整 |
| 功能完整度 | B- | 主路径代码存在，但注册链路存在前后端契约漂移导致 400 |
| 安全性 | B | JWT 挂载范围已收敛；密码验证前后端不一致为新增风险 |
| 性能 | B | C++ 引擎已通过 linux/amd64+cgo 生效 |
| 可运营性 | B | 监控栈完整，缺链路追踪和自动部署 |
| 代码质量 | C+ | gateway 大文件、前后端契约漂移、前端错误处理不完善 |
| 测试 | C- | 37 个测试文件但覆盖率不足，无注册 E2E 测试 |
| **上线就绪度** | **B** | **注册链路 400 为 P0 阻塞项，需修复后方可上线** |

**P0 阻塞项**: 注册链路密码验证规则前后端不一致 + 400 错误信息被吞掉（阻断用户首次注册）  
**预估剩余工时**: 14-22h（含注册修复 2-4h）

---

*报告版本: v3.5 (注册链路专项审计)*  
*生成时间: 2026-04-30*  
*修订: 补充注册链路前后端契约漂移分析，新增 P0 阻塞项（密码验证规则不一致 + 错误信息吞没），调整上线就绪度 B+ → B*

---

## 10. 本轮最终审计结论

### 10.1 已验证通过

| 项目 | 结果 |
|------|------|
| Go 核心服务测试（api-gateway/data-service/user-service/recommendation-service） | ✅ 通过 |
| Gateway 鉴权挂载（users/payments/recommendations） | ✅ RequireAuth 生效 |
| Swagger 生成一致性（api-gateway） | ✅ 可生成且当前一致 |
| 注册链路代码路径完整性 | ✅ 前端→网关→user-service 路由畅通，handler/service/model 层均有实现 |
| user-service 数据库自动迁移 | ✅ AutoMigrate 覆盖 12 个模型，seedDefaultData 初始化角色权限 |

### 10.2 注册链路专项审计（2026-04-30 新增）

| 检查项 | 结果 | 详情 |
|------|------|------|
| 前端注册表单 → API 调用 | ✅ | `LoginPage.vue:289-316` → `userApi.register()` → `POST /api/v1/users/auth/register` |
| Gateway 路由转发 | ✅ | `createUserProxy` 识别 `/users/auth/` 前缀，rewrite 为 `/api/v1/auth/register` |
| user-service handler 绑定 | ✅ | `auth_handler.go:56-64` ShouldBindJSON 正确绑定 username/password/email/phone |
| user-service 密码验证 | ⚠️ | 要求特殊字符，前端未检查 → 400 |
| user-service CreateUser | ✅ | 检查重复→bcrypt 加密→创建记录→分配默认角色→审计日志 |
| 前端错误处理 | ❌ | 400 响应的 `details` 字段未解析展示 |
| 前后端字段对齐 | ⚠️ | confirmPassword 多余字段、phone 必填/可选不一致 |

### 10.3 当前剩余风险（按优先级）

| 项目 | 优先级 | 说明 |
|------|------|------|
| **注册密码验证规则前后端不一致** | **P0** | 阻断用户注册，必须在上线前修复 |
| **注册 400 错误信息被吞** | **P0** | 用户无法理解失败原因，影响转化率 |
| JWT/数据库默认回退值收敛 | P1 | 避免配置缺失时以弱默认值启动 |
| LLM 出网与脱敏策略固化 | P1 | 明确生产数据边界与审计策略 |
| 前后端契约统一（响应格式包装） | P2 | 当前 `normalizeMessageResponse` 兼容处理可工作，但非理想状态 |
| 前端 lint warnings 清零 | P2 | 不阻塞上线，但建议保持质量门零警告 |
| 归档文档敏感示例清理 | P2 | 继续减少历史文档误导风险 |
