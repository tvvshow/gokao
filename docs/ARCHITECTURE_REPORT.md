# 高考志愿填报系统 架构分析报告（现状修订版）

修订日期: 2026-04-25  
修订范围: 基于当前仓库代码和配置文件的静态核查（README、go.work、docker-compose、services/*、frontend/src/api）

## 1. 执行摘要

你感受到的“混乱”来自两类问题叠加：

1. 原推进计划要解决的架构治理问题  
   例如 `replace` 治理、workspace 一致性、跨服务依赖收敛。
2. 推进过程中暴露的实现一致性问题  
   例如网关历史路径兼容逻辑膨胀、测试断言与实现偏移、服务行为与文档不一致。

结论：当前并非“计划失焦”，而是“治理主线 + 实施期偏差”同时存在，需要分层治理。

## 2. 架构总览与拓扑

系统形态为：`Vue3 前端 + API Gateway + 多 Go 微服务 + pkg 共享库 + C++ 能力模块`。

当前端口（以 `docker-compose.yml` 为准）：

- api-gateway: `8080`
- data-service: `8082`
- user-service: `8083`
- recommendation-service: `8084`
- payment-service: `8085`
- monitoring-service: `8086`
- postgres: `5433 -> 5432`
- redis: `6380 -> 6379`

说明：旧报告中的 payment/recommendation 端口冲突、monitoring/gateway 端口冲突已不成立。

## 3. 主要模块职责

### 3.1 frontend

- 基于 Vue 3 + Vite + TypeScript。
- 通过 Axios 封装统一请求，具备 token 自动刷新与并发 401 排队机制。
- API 路径主要走 `/api/v1/*`（见 `frontend/src/api/*.ts`）。

### 3.2 api-gateway

- 统一入口，承担反向代理、限流、指标、缓存、中间件编排。
- 维护多套历史别名路径兼容，并最终重写到后端服务路径。
- 问题：`main.go` 体量大，网关策略与路由重写耦合较高。

### 3.3 data-service

- 数据域主服务，承载院校、专业、录取等核心查询能力。
- 具备数据库与缓存集成能力，是最核心的业务服务之一。

### 3.4 user-service

- 认证、用户、权限相关能力。
- 与 `pkg/auth` 在职责上存在潜在重叠，需要边界澄清。

### 3.5 payment-service

- 支付域能力（订单、回调等），部分扩展能力存在阶段性未启用/注释代码。

### 3.6 recommendation-service

- 推荐域能力，包含 C++ 桥接与回退策略链。
- 在外部依赖异常时可降级，具备一定韧性。

### 3.7 monitoring-service

- 指标和告警服务，已纳入 compose 编排，端口为 `8086`。
- 日志技术栈与其他服务存在不一致（zap vs logrus）。

### 3.8 pkg/*

- 共享基础能力层，当前目录下模块已普遍具备 `go.mod`。
- `go.work` 已注册主要 `pkg/*` 与服务模块，基础可达性较旧版本明显改善。

## 4. 数据流向

主链路：

`Frontend -> API Gateway -> Domain Service -> Database/Redis`

推荐链路：

`Frontend -> Gateway -> recommendation-service -> C++ bridge -> fallback path`

网关同时承担横切关注点：

- 认证上下文透传
- 限流
- 指标采集
- 路径重写与兼容

## 5. 依赖关系与模块治理现状

### 5.1 已完成或明显改善

1. `go.work` 已注册 `pkg/cache`、`pkg/discovery`、`pkg/health`、`pkg/metrics`、`pkg/shared`、`pkg/testutil` 等模块。  
2. `pkg` 下模块化完整度较之前提升，多数功能性子目录已有 `go.mod`；但 `pkg/api/` 为空目录、`pkg/response/` 仅有占位文件，尚不构成可复用功能包。  
3. `docker-compose` 服务端口冲突已解除。

### 5.2 仍待统一

1. 当前仓库正确命名空间为 `github.com/oktetopython/gaokao/*`，不需要迁移到其他前缀。  
2. `replace` 指令并非仅出现在两个服务：当前共 16 条，分布在 6 个 `go.mod` 中（root、api-gateway、data-service、user-service、payment-service、monitoring-service）。  
3. `services/recommendation-service/go.mod` 当前未使用 `replace` 指令；服务间治理状态不一致，增加维护复杂度。

### 5.3 CORS 分布现状（冗余实现）

1. 网关层已启用统一 CORS，并在代理层清理后端返回的 CORS 头。  
2. 3/5 个后端业务服务仍各自启用 CORS（`data-service`、`payment-service`、`recommendation-service`）。  
3. `user-service` 已注释 CORS 中间件。  
4. 现状属于“网关统一治理 + 下游重复实现并存”，是客观设计冗余，建议收敛。

## 6. 设计模式使用评估

1. API Gateway 模式  
   网关职责明确，但实现集中在单文件，维护成本高。
2. Middleware Chain  
   鉴权、限流、日志、指标等横切逻辑通过中间件组织，方向正确。
3. Adapter/Bridge  
   支付适配器与推荐 C++ 桥接体现适配思想。
4. Fallback/降级策略  
   推荐服务具备回退路径，提升可用性。

总体评价：模式选型合理，主要短板在“实现结构化程度”而非“模式缺失”。

## 7. 当前架构问题（按优先级）

### P0（必须优先处理）

1. `replace` 指令在多模块分散存在（16 条 / 6 个 go.mod），影响构建一致性与可移植性。  
2. 网关历史路径兼容逻辑较重，路由规则复杂，回归风险高。  
3. CORS 策略在网关与下游服务重复实现，治理边界不清晰。

### P1（应在近期迭代处理）

1. 网关单体 `main.go` 过大，建议拆分为 router/rewriter/proxy/middleware 子模块。  
2. 认证与错误处理在服务间仍有重复实现倾向，需要统一契约。  
3. 监控服务日志栈与其他服务不一致，运维观察面不统一。

### P2（中期优化）

1. 建立跨服务 API 契约测试，避免“实现变化先于测试更新”。  
2. 梳理配置加载模式，减少服务各自维护配置模板带来的漂移。

## 8. 计划相关 vs 推进暴露问题

### 8.1 计划主线问题（架构治理债务）

- `replace` 指令清理
- workspace/模块依赖一致性
- CORS 责任边界收敛（网关统一，下游去冗余）

### 8.2 推进暴露问题（实施一致性债务）

- 网关路径兼容逻辑持续堆叠
- 测试断言与当前实现不同步
- 局部功能“半启用”导致行为预期不稳定

## 9. 优化方向与落地建议

### 短期（1-2 个迭代）

1. 冻结并遵循现有正确模块前缀 `github.com/oktetopython/gaokao/*`。  
2. 分批清理 6 个模块中的 `replace` 指令，优先处理服务模块，再处理 root/monitoring 内部替换。  
3. 将 CORS 策略收敛到网关，逐步移除下游服务重复 CORS 中间件。  
4. 先补网关路由契约测试，再重构路由重写逻辑。

### 中期（2-4 个迭代）

1. 拆分网关核心模块，降低单文件耦合。  
2. 统一错误码、认证上下文、日志字段规范。  
3. 建立跨服务接口变更检查（OpenAPI diff/契约测试）。

### 长期

1. 引入更完善可观测性（trace + 指标关联）。  
2. 推进配置与治理标准化，减少新服务接入成本。

## 10. 本次修订对旧报告的处理说明

已删除或修正的过时结论：

1. payment/recommendation 端口冲突。  
2. monitoring 与 gateway 端口冲突。  
3. `go.work` 缺失多个 pkg 模块。  
4. `pkg/metrics` 中文字符导致语法错误、`pkg/utils` 缺少 `fmt` 等历史编译问题（当前代码中已修复）。

保留并强化的核心问题：

1. 多模块 `replace` 依赖未完全清理。  
2. 网关路径兼容逻辑过重和结构性拆分不足。  
3. CORS 重复实现带来的治理边界冗余。

## 11. 验证说明

本次为静态架构核查。尝试执行 Go 构建时，当前执行环境返回 `go: command not found`，因此未在本环境完成动态编译验证。  
建议在已安装 Go 工具链的终端执行：

```bash
go version
go work sync
cd services/api-gateway && go build ./...
cd ../payment-service && go build ./...
```
