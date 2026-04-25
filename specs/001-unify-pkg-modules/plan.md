# Implementation Plan: Shared Module Unification (Scoped, Quality-Gate Updated)

**Branch**: `001-unify-pkg-modules` | **Date**: 2026-04-25 | **Spec**: [spec.md](./spec.md)  
**Input**: Feature specification from `/specs/001-unify-pkg-modules/spec.md`  
**Architecture Baseline**: [docs/ARCHITECTURE_REPORT.md](../../docs/ARCHITECTURE_REPORT.md)

## Summary

本次计划保持既定范围：仅处理 `api-gateway` 与 `payment-service` 两个目标模块的冲突 `replace` 治理、复用审计、依赖同步与验证收口。  
本版重点更新质量门槛：验证不再只要求“执行并分类”，而是要求在可用验证环境中达到“可合并通过”标准；若受外部限制，必须进入受控豁免流程并记录 owner/deadline。

## Technical Context

**Language/Version**: Go 1.25（后端）  
**Primary Dependencies**: Gin, logrus, go-redis, golang.org/x/time（按服务差异）  
**Storage**: PostgreSQL, Redis（运行期依赖；本次不改存储模型）  
**Testing**: `go mod tidy`, `go build ./...`, 基本运行检查（启动/健康检查）  
**Target Platform**: Linux shell / 容器环境  
**Project Type**: Go 多模块微服务仓库  
**Performance Goals**: 本次不改请求处理路径；默认以“无新增中间件/无路由行为变更”为性能豁免依据。若请求链路发生变化，需补充 P99 基线证据。  
**Constraints**: 必须遵循 `docs/Gaokao_Constraint_Protocol_v1.0.md`，且仅修改两个目标服务 `go.mod`  
**Scale/Scope**: 仅影响 `services/api-gateway` 与 `services/payment-service` 的模块解析与验证闭环

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] 最优秀原则: 已定义“可合并”质量门槛，不接受仅记录失败即交付
- [x] 有即复用原则: 复用 `go.work` 与现有 `pkg/*` 结构，不引入平行依赖机制
- [x] 不允许简化原则: 覆盖 replace 清理、复用审计、tidy/build/runtime、失败归因与豁免治理
- [x] 技术栈与架构边界原则: 仅触及 Go 模块依赖治理边界，不扩展栈
- [x] 强制验证与质量闸门原则: 明确“通过方可合并；受限则受控豁免”规则

## Project Structure

### Documentation (this feature)

```text
specs/001-unify-pkg-modules/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── module-scope-contract.md
├── tasks.md
└── verification-log.md
```

### Source Code (repository root)

```text
services/
├── api-gateway/
│   └── go.mod
└── payment-service/
    └── go.mod

pkg/
go.work
```

**Structure Decision**: 保持现有多模块结构，不引入新目录或替代依赖机制。

## Phase 0: Research Output

参见 [research.md](./research.md)。本轮无新增 NEEDS CLARIFICATION。  
补充质量门槛结论：

1. “验证任务已执行”不等于“质量门槛通过”；两者需分离记录。
2. 可合并条件：目标服务在可用环境内满足 `tidy/build/runtime` 全通过。
3. 若环境阻塞（私有仓库鉴权、外部依赖不可达等），必须产生豁免记录（阻塞证据、owner、deadline、回补计划）。

## Phase 1: Design & Contracts

### Data Model

参见 [data-model.md](./data-model.md)。  
质量闸门实体字段已补充并落地（`QualityGateRecord`，含 `gate_status`、`waiver_owner`、`waiver_deadline` 等）。

### Contracts

参见 [contracts/module-scope-contract.md](./contracts/module-scope-contract.md)。  
建议将 Acceptance Contract 补充为：
- “仅当门槛通过或已批准豁免时才可视为可合并完成”。

### Quickstart

参见 [quickstart.md](./quickstart.md)。  
建议在执行状态区分两层状态：
- 执行完成（Execution Complete）
- 质量门槛状态（Gate Passed / Waived / Blocked）

### Agent Context Update

`AGENTS.md` 的 SPECKIT plan 引用已保持为：

`specs/001-unify-pkg-modules/plan.md`

## Post-Design Constitution Check

- [x] 最优秀原则: 质量目标从“执行覆盖”提升为“可合并门槛”
- [x] 有即复用原则: 依赖治理沿用 `go.work` 与现有模块体系
- [x] 不允许简化原则: 明确禁止“仅分类失败即收口”
- [x] 技术栈与架构边界原则: 无跨栈改造
- [x] 强制验证与质量闸门原则: 门槛和豁免路径均可审计

## Implementation Strategy (Phase 2 Plan Basis)

### US1: Replace Cleanup + Reuse Audit

1. 清理 `api-gateway` 与 `payment-service` 冲突 `replace`。  
2. 执行范围守卫，确保非目标 `go.mod` 零改动。  
3. 完成复用审计（CORS/auth wrapper/error handler）并给出 keep/deprecate/remove 结论。  
4. 对可安全项执行去重或保留 deprecated 标记。  

### US2: Dependency Sync + Build/Run Verification

1. 在两个目标服务分别执行 `go mod tidy`。  
2. 分别执行 `go build ./...`。  
3. 分别执行基本运行检查。  
4. 对失败项进行 Code Issue / Environment Limitation / Design Gap 分类。  

### Quality Gate Closure (Updated)

1. Gate-Pass 条件：`tidy/build/runtime` 在可用环境中全通过。  
2. Gate-Waive 条件：存在外部阻塞且已记录证据、owner、deadline、回补动作。  
3. Gate-Blocked：既未通过也无有效豁免，不可判定可合并。  

## Replace Directive Full Inventory

| go.mod | Count | In Scope? |
|--------|-------|-----------|
| root `go.mod` | 2 | No |
| api-gateway | 3 | **Yes** |
| payment-service | 1 | **Yes** |
| data-service | 4 | No |
| user-service | 4 | No |
| monitoring-service | 2 | No |
| recommendation-service | 0 | — |

## Risks & Mitigations

1. 移除 replace 后触发私有模块解析失败  
缓解: 记录为 Environment Limitation/Design Gap，并要求 Gate-Waive 记录而非直接收口。  
2. 运行检查受 DB/Redis/网络影响  
缓解: 明确外部阻塞证据与回补计划。  
3. 复用审计误删服务特有逻辑  
缓解: 仅在“完全重叠 + 验证通过”时移除，其余保留并标注。  

## Complexity Tracking

无宪章豁免项；当前方案不新增架构复杂度，仅提升交付质量门槛定义。
