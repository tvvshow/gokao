# Implementation Plan: [FEATURE]

**Branch**: `[###-feature-name]` | **Date**: [DATE] | **Spec**: [link]
**Input**: Feature specification from `/specs/[###-feature-name]/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

[Extract from feature spec: primary requirement + technical approach from research]

## Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the project. The structure here is presented in advisory capacity to guide
  the iteration process.
-->

**Language/Version**: [Go 1.25 (backend), TypeScript + Vue 3 (frontend) by default]  
**Primary Dependencies**: [Gin, Vue 3, Vite, PostgreSQL, Redis, and feature-specific dependencies]  
**Storage**: [if applicable, e.g., PostgreSQL, CoreData, files or N/A]  
**Testing**: [Go test, Vitest, lint/type-check, plus integration/contract tests when applicable]  
**Target Platform**: [Linux containers for services, modern browsers for frontend]
**Project Type**: [Go microservices + Vue 3 web app]  
**Performance Goals**: [include measurable goals; default non-recommendation backend target P99 <= 200ms unless overridden with rationale]  
**Constraints**: [must comply with constitution and docs/Gaokao_Constraint_Protocol_v1.0.md]  
**Scale/Scope**: [describe user/load scope and impacted services]

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [ ] 最优秀原则: 明确性能/质量目标与验证方法（含 benchmark/压测计划）
- [ ] 有即复用原则: 已完成复用扫描（`pkg/`, `services/`, `frontend/src/components/`）并记录复用/不复用理由
- [ ] 不允许简化原则: 关键路径、边界条件、错误路径、超时重试均已纳入设计
- [ ] 技术栈与架构边界原则: 仅使用 Go 1.25 + Vue 3，保持网关与服务职责边界
- [ ] 强制验证与质量闸门原则: 已定义必须执行的测试、lint/type-check、文档同步与验收证据

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)
<!--
  ACTION REQUIRED: Replace the placeholder tree below with the concrete layout
  for this feature. Delete unused options and expand the chosen structure with
  real paths (e.g., apps/admin, packages/something). The delivered plan must
  not include Option labels.
-->

```text
services/
├── api-gateway/
├── data-service/
├── user-service/
├── payment-service/
├── recommendation-service/
└── monitoring-service/

frontend/
└── src/
    ├── components/
    ├── views/
    ├── api/
    └── __tests__/

pkg/
cpp-modules/
tests/integration/
```

**Structure Decision**: Use existing Go microservices + Vue 3 structure; do not
introduce parallel backend/frontend roots that bypass current layout.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
