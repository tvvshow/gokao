<!--
Sync Impact Report
- Version change: template -> 1.0.0
- Modified principles:
  - Template Principle 1 -> I. 最优秀原则
  - Template Principle 2 -> II. 有即复用原则
  - Template Principle 3 -> III. 不允许简化原则
  - Template Principle 4 -> IV. 技术栈与架构边界原则
  - Template Principle 5 -> V. 强制验证与质量闸门原则
- Added sections:
  - 技术与架构硬约束
  - 研发流程与质量闸门
- Removed sections:
  - 无（保留模板原有章节结构，仅完成具象化）
- Templates requiring updates:
  - ✅ updated: /mnt/d/mybitcoin/gaokao/.specify/templates/plan-template.md
  - ✅ updated: /mnt/d/mybitcoin/gaokao/.specify/templates/spec-template.md
  - ✅ updated: /mnt/d/mybitcoin/gaokao/.specify/templates/tasks-template.md
  - ⚠ pending: /mnt/d/mybitcoin/gaokao/.specify/templates/commands/*.md (目录不存在，无法校验)
  - ✅ updated: /mnt/d/mybitcoin/gaokao/README.md
- Follow-up TODOs:
  - 无
-->
# Gaokao Constitution

## Core Principles

### I. 最优秀原则
所有设计与实现 MUST 以可验证的优秀标准交付，不接受“够用即可”。后端关键接口性能目标
MUST 明确且可测（默认非推荐链路 P99 <= 200ms）；任何“性能优化”声明 MUST 提供基准测试
或压测证据。该原则用于确保系统在正确性、性能、安全性上持续达到生产级水平。

### II. 有即复用原则
开发前 MUST 优先检索并复用现有实现（`pkg/`、`services/`、`frontend/src/components/`）。
已存在的通用能力 MUST NOT 被重复实现；跨服务共性逻辑 MUST 上提到共享层。引入外部实现
MUST 标注来源。该原则用于控制重复建设、减少分叉和维护成本。

### III. 不允许简化原则
需求一旦进入实现范围，核心业务路径、边界条件、错误处理、超时与重试 MUST 完整实现。
MUST NOT 用 `TODO`、占位返回、假数据、注释掉主流程来替代交付。该原则用于防止“伪完成”
并确保行为与规格一致。

### IV. 技术栈与架构边界原则
本项目技术栈 MUST 锁定为 Go 1.25（后端）与 Vue 3（前端）；新增模块不得偏离该主栈。
微服务边界 MUST 清晰：前端仅经网关访问后端，服务间职责不可交叉。涉及 CGO/C++ 的能力
MUST 明确接口与生命周期管理。该原则用于保证架构一致性与长期演进稳定性。

### V. 强制验证与质量闸门原则
所有变更 MUST 通过与变更范围匹配的自动化验证：Go 测试、前端单测、lint/type-check、
必要的集成/契约测试、Swagger 一致性检查。未通过质量闸门的变更 MUST NOT 合并。
该原则用于让“优秀、复用、完整实现”具备可执行与可审计的证据链。

## 技术与架构硬约束

- 约束协议基线 MUST 遵循 `docs/Gaokao_Constraint_Protocol_v1.0.md`。
- 后端 MUST 使用 Go 1.25，前端 MUST 使用 Vue 3 + TypeScript。
- 对外 API MUST 由 API Gateway 暴露并统一治理认证、限流、观测与路由策略。
- 关键业务（认证、支付、推荐）MUST 具备可观测性指标与错误追踪。
- 生产路径 MUST NOT 使用长期 mock、占位实现或未落地的“临时分支逻辑”。

## 研发流程与质量闸门

- 规格阶段 MUST 在 `spec.md` 中显式说明如何满足五项核心原则。
- 计划阶段 MUST 在 `plan.md` 中完成 Constitution Check，并给出复用扫描结果与性能目标。
- 任务阶段 MUST 将“复用改造”“完整异常路径”“验证任务”显式拆分到 `tasks.md`。
- 实施阶段 MUST 执行最小可行增量，但不得以“简化实现”牺牲需求完整性。
- 评审阶段 MUST 提供验证证据（测试/检查/基准）并说明对约束协议的符合性。

## Governance

本宪章高于项目内其他工作习惯与临时约定。任何修订 MUST 通过 Pull Request 提交，且包含：
变更动机、影响范围、模板同步说明、迁移计划（如适用）。

版本采用语义化规则：
- MAJOR：原则被重定义、移除或引入不兼容治理要求。
- MINOR：新增原则/章节，或对现有要求做实质性扩展。
- PATCH：措辞澄清、错别字修正、无语义变化的编辑。

每个 PR 与发布前检查 MUST 进行宪章合规复核；不合规项要么在本次修复，要么记录有时限的
整改计划并由负责人确认。

**Version**: 1.0.0 | **Ratified**: 2026-04-24 | **Last Amended**: 2026-04-24
