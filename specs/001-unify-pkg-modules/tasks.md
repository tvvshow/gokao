---

description: "Task list for scoped pkg replace cleanup, reuse audit, dependency sync, verification, and quality-gate closure"

---

# Tasks: Shared Module Unification (Scoped, Quality-Gate Synced)

**Input**: Design documents from `/specs/001-unify-pkg-modules/`  
**Prerequisites**: `plan.md` (required), `spec.md` (required), `research.md`, `data-model.md`, `contracts/`, `quickstart.md`

**Tests**: 本特性为模块依赖治理与验证任务，不新增业务行为；不新增单元/集成测试文件，改为执行 `go mod tidy`、`go build`、运行检查并沉淀证据。  
**Organization**: Tasks are grouped by user story to enable independent implementation and validation.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: 建立执行基线与证据文件。

- [X] T001 Create or reset verification record file at `specs/001-unify-pkg-modules/verification-log.md`
- [X] T002 Capture pre-change snapshots of `services/api-gateway/go.mod` and `services/payment-service/go.mod` into `specs/001-unify-pkg-modules/verification-log.md`
- [X] T003 Capture baseline scope list for `services/*/go.mod` into `specs/001-unify-pkg-modules/verification-log.md`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: 固化范围约束、失败分类与质量门槛协议，阻断越界修改。

- [X] T004 Extract acceptance checkpoints from `specs/001-unify-pkg-modules/contracts/module-scope-contract.md` into `specs/001-unify-pkg-modules/verification-log.md`
- [X] T005 [P] Record command matrix (`go mod tidy`, `go build`, runtime checks) and gate criteria in `specs/001-unify-pkg-modules/verification-log.md`
- [X] T006 [P] Record environment limitation policy and waiver metadata fields (owner/deadline/remediation) in `specs/001-unify-pkg-modules/verification-log.md`
- [X] T007 Align execution notes in `specs/001-unify-pkg-modules/quickstart.md` with `docs/ARCHITECTURE_REPORT.md` and quality-gate definitions

**Checkpoint**: 范围和验证协议确定后，才能进入用户故事实施。

---

## Phase 3: User Story 1 - 清理冲突 replace 规则与复用审计 (Priority: P1) 🎯 MVP

**Goal**: 在两个目标服务删除冲突 `replace` 规则，并审计可复用共享实现。  
**Independent Test**: 两个目标 `go.mod` 不再含冲突 `replace`（或触发保留条件时有选择性回退证据），其他服务 `go.mod` 未修改；重复实现有明确审计结论。

### Implementation for User Story 1

- [X] T008 [US1] Remove candidate conflicting `replace` directives in `services/api-gateway/go.mod` per判定规则
- [X] T009 [US1] Remove candidate conflicting `replace` directives in `services/payment-service/go.mod` per判定规则
- [X] T010 [US1] Verify no other service module file changed by checking `services/*/go.mod` and append result to `specs/001-unify-pkg-modules/verification-log.md`
- [X] T011 [US1] Record post-change diffs of `services/api-gateway/go.mod` and `services/payment-service/go.mod` in `specs/001-unify-pkg-modules/verification-log.md`
- [X] T012 [P] [US1] Audit duplicate CORS/auth/error implementations under `services/api-gateway/` against `pkg/*` and append keep/deprecate/remove verdicts to `specs/001-unify-pkg-modules/verification-log.md`
- [X] T013 [P] [US1] Audit duplicate CORS/auth implementations under `services/payment-service/` against `pkg/*` and append keep/deprecate/remove verdicts to `specs/001-unify-pkg-modules/verification-log.md`
- [X] T014 [US1] Apply deprecate/remove actions for confirmed duplicates in `services/api-gateway/**` and `services/payment-service/**` (or record no-op with rationale) and append to `specs/001-unify-pkg-modules/verification-log.md`
- [X] T015 [US1] Run `go build ./...` in `services/api-gateway/` and `services/payment-service/` and append outputs to `specs/001-unify-pkg-modules/verification-log.md`
- [X] T016 [US1] If keep-condition is triggered, selectively restore only affected `replace` directives in `services/api-gateway/go.mod` or `services/payment-service/go.mod`, with rationale recorded in `specs/001-unify-pkg-modules/verification-log.md`
- [X] T017 [US1] Re-run `go build ./...` for affected target service(s) and append final keep/remove decision in `specs/001-unify-pkg-modules/verification-log.md`

**Checkpoint**: US1 完成时可独立交付"冲突 replace 清理 + 复用审计 + 条件回退闭环"结果。

---

## Phase 4: User Story 2 - 同步依赖并验证可运行性 (Priority: P2)

**Goal**: 在两目标服务完成依赖同步与构建/运行验证，并形成可复核证据。  
**Independent Test**: 两服务都执行 tidy/build/runtime 检查，结果与失败归因明确。

### Implementation for User Story 2

- [X] T018 [P] [US2] Run `go mod tidy` in `services/api-gateway/` and append outputs to `specs/001-unify-pkg-modules/verification-log.md`
- [X] T019 [P] [US2] Run `go mod tidy` in `services/payment-service/` and append outputs to `specs/001-unify-pkg-modules/verification-log.md`
- [X] T020 [P] [US2] Run `go build ./...` in `services/api-gateway/` and append outputs to `specs/001-unify-pkg-modules/verification-log.md`
- [X] T021 [P] [US2] Run `go build ./...` in `services/payment-service/` and append outputs to `specs/001-unify-pkg-modules/verification-log.md`
- [X] T022 [US2] Run runtime check for `services/api-gateway/` and verify health endpoint returns HTTP 200 within 15 seconds, then append result to `specs/001-unify-pkg-modules/verification-log.md`
- [X] T023 [US2] Run runtime check for `services/payment-service/` and verify process survives 15 seconds without fatal/panic (or health endpoint HTTP 200 when available), then append result to `specs/001-unify-pkg-modules/verification-log.md`
- [X] T024 [US2] Classify all failed checks as Code Issue / Environment Limitation / Design Gap in `specs/001-unify-pkg-modules/verification-log.md`

**Checkpoint**: US2 完成时可独立交付"依赖同步 + 构建/运行验证 + 异常归因"结果。

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: 收敛交付证据、校验成功标准并完成质量门槛判定。

- [X] T025 Update execution outcomes section in `specs/001-unify-pkg-modules/quickstart.md` from `specs/001-unify-pkg-modules/verification-log.md`
- [X] T026 Validate SC-001..SC-008 and FR-001..FR-009 from `specs/001-unify-pkg-modules/spec.md` against `specs/001-unify-pkg-modules/verification-log.md`
- [X] T027 Determine and record quality gate status (`Gate Passed` / `Gate Waived` / `Gate Blocked`) in `specs/001-unify-pkg-modules/verification-log.md`, including waiver owner/deadline/remediation when waived
- [X] T028 Add final scope compliance statement in `specs/001-unify-pkg-modules/verification-log.md` (only target files modified in `services/*/go.mod`)
- [X] T029 Record performance exemption evidence (no request-path behavior change) or benchmark evidence in `specs/001-unify-pkg-modules/verification-log.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- Phase 1 (Setup): no dependency
- Phase 2 (Foundational): depends on Phase 1
- Phase 3 (US1): depends on Phase 2
- Phase 4 (US2): depends on Phase 3
- Phase 5 (Polish): depends on Phase 4

### User Story Dependencies

- **US1 (P1)**: starts after Foundational, no dependency on US2
- **US2 (P2)**: depends on US1 (`go.mod` cleanup + reuse decisions + conditional rollback must be finalized first)

### Within Each User Story

- US1: candidate replace removal -> scope guard -> diff evidence -> reuse audit -> deprecate/remove -> build verification -> conditional rollback -> final build verification
- US2: tidy -> build -> runtime checks -> failure classification

## Parallel Opportunities

- T005 and T006 can run in parallel.
- T008 and T009 can run in parallel.
- T012 and T013 can run in parallel.
- T018 and T019 can run in parallel.
- T020 and T021 can run in parallel.
- T022 and T023 can run in parallel if environment supports concurrent service runs.

## Parallel Example: User Story 2

```bash
# Run dependency sync in parallel:
Task: "Run go mod tidy in services/api-gateway and record output"
Task: "Run go mod tidy in services/payment-service and record output"

# Then run build checks in parallel:
Task: "Run go build ./... in services/api-gateway and record output"
Task: "Run go build ./... in services/payment-service and record output"
```

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 and Phase 2.
2. Complete Phase 3 (US1).
3. Validate US1 independent test criteria.

### Incremental Delivery

1. Deliver US1 (`replace` cleanup + reuse audit + conditional rollback evidence).
2. Deliver US2 (tidy/build/runtime + failure classification).
3. Complete Phase 5 (SC/FR alignment + quality gate + scope compliance statement).

## Notes

- Do not modify any `services/*/go.mod` outside target files.
- Use the existing correct module prefix `github.com/oktetopython/gaokao/*`; do not introduce module-prefix migration tasks.
- Keep raw command outputs and conclusions in `specs/001-unify-pkg-modules/verification-log.md`.
- If blocked by environment, record exact command and reason; gate can be waived only with owner/deadline/remediation metadata.
