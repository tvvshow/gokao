---

description: "Task list for scoped go.mod replace cleanup and verification"
---

# Tasks: Shared Module Unification (Scoped)

**Input**: Design documents from `/specs/001-unify-pkg-modules/`  
**Prerequisites**: `plan.md` (required), `spec.md` (required), `research.md`, `data-model.md`, `contracts/`, `quickstart.md`

**Tests**: Verification tasks are required by spec (dependency sync, build checks, runtime checks).  
**Organization**: Tasks are grouped by user story to preserve independent delivery and validation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Task can run in parallel (different files/no blocking dependency)
- **[Story]**: `[US1]`/`[US2]` map to user stories in `spec.md`
- Every task includes an exact file path

## Path Conventions

- Target module files: `services/api-gateway/go.mod`, `services/payment-service/go.mod`
- Feature artifacts: `specs/001-unify-pkg-modules/*`
- Scope guard checks: repository-wide `services/*/go.mod`

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Establish execution baseline and evidence locations.

- [X] T001 Create verification evidence file at `specs/001-unify-pkg-modules/verification-log.md`
- [X] T002 Capture pre-change target module snapshots from `services/api-gateway/go.mod` and `services/payment-service/go.mod` into `specs/001-unify-pkg-modules/verification-log.md`
- [X] T003 Capture pre-change scope baseline for all `services/*/go.mod` paths in `specs/001-unify-pkg-modules/verification-log.md`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Lock scope and acceptance criteria before editing module files.

- [X] T004 Record scope contract checkpoints from `specs/001-unify-pkg-modules/contracts/module-scope-contract.md` in `specs/001-unify-pkg-modules/verification-log.md`
- [X] T005 Define command matrix for tidy/build/runtime checks in `specs/001-unify-pkg-modules/verification-log.md`
- [X] T006 Document environment limitation handling policy (missing go/container/DB) in `specs/001-unify-pkg-modules/verification-log.md`

**Checkpoint**: Scope guard and verification protocol are fixed; story work can begin.

---

## Phase 3: User Story 1 - 清理冲突 replace 规则 (Priority: P1) 🎯 MVP

**Goal**: Remove conflicting `replace` directives from exactly two target service module files.

**Independent Test**: `services/api-gateway/go.mod` and `services/payment-service/go.mod` contain no conflicting `replace` directives, and no other service `go.mod` changed.

### Implementation for User Story 1

- [X] T007 [US1] Remove conflicting `replace` directives in `services/api-gateway/go.mod`
- [X] T008 [US1] Remove conflicting `replace` directives in `services/payment-service/go.mod`
- [X] T009 [US1] Validate only two target files changed among `services/*/go.mod` and record result in `specs/001-unify-pkg-modules/verification-log.md`
- [X] T010 [US1] Record post-change diff evidence for `services/api-gateway/go.mod` and `services/payment-service/go.mod` in `specs/001-unify-pkg-modules/verification-log.md`

**Checkpoint**: US1 is complete when replace cleanup is done with scope guard evidence.

---

## Phase 4: User Story 2 - 同步依赖并验证可运行性 (Priority: P2)

**Goal**: Synchronize dependencies and provide build/runtime verification evidence for both target services.

**Independent Test**: `go mod tidy` attempted for both services, build/runtime checks attempted for both services, and evidence captured with explicit success/failure/environment classification.

### Implementation for User Story 2

- [X] T011 [P] [US2] Run dependency sync for `services/api-gateway/go.mod` via `go mod tidy` in `services/api-gateway/` and record output in `specs/001-unify-pkg-modules/verification-log.md`
- [X] T012 [P] [US2] Run dependency sync for `services/payment-service/go.mod` via `go mod tidy` in `services/payment-service/` and record output in `specs/001-unify-pkg-modules/verification-log.md`
- [X] T013 [P] [US2] Run build check for `services/api-gateway/` and append result to `specs/001-unify-pkg-modules/verification-log.md`
- [X] T014 [P] [US2] Run build check for `services/payment-service/` and append result to `specs/001-unify-pkg-modules/verification-log.md`
- [X] T015 [US2] Run basic runtime check for `services/api-gateway/` (`/healthz` or equivalent) and append result to `specs/001-unify-pkg-modules/verification-log.md`
- [X] T016 [US2] Run basic runtime check for `services/payment-service/` startup path and append result to `specs/001-unify-pkg-modules/verification-log.md`
- [X] T017 [US2] Classify all failures in `specs/001-unify-pkg-modules/verification-log.md` as code issue vs environment limitation

**Checkpoint**: US2 is complete when tidy/build/runtime evidence exists for both services.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Final consistency and handoff readiness.

- [X] T018 Ensure `specs/001-unify-pkg-modules/quickstart.md` reflects actual executed command outcomes
- [X] T019 Validate success criteria SC-001..SC-005 against `specs/001-unify-pkg-modules/verification-log.md` and summarize in that same file
- [X] T020 Produce final implementation summary in `specs/001-unify-pkg-modules/verification-log.md` with scope compliance statement

---

## Dependencies & Execution Order

### Phase Dependencies

- Setup (Phase 1) -> Foundational (Phase 2) -> US1 (Phase 3) -> US2 (Phase 4) -> Polish (Phase 5)

### User Story Dependencies

- **US1 (P1)**: Starts after Phase 2; no dependency on US2.
- **US2 (P2)**: Depends on US1 completion because tidy/build/runtime checks require finalized `go.mod` edits.

### Within Each User Story

- US1: edit target module files -> scope guard check -> evidence capture.
- US2: run tidy -> run build checks -> run runtime checks -> classify outcomes.

## Parallel Opportunities

- T011 and T012 can run in parallel.
- T013 and T014 can run in parallel.
- Runtime checks (T015, T016) can run in parallel if execution environment supports both services concurrently.

## Parallel Example: User Story 2

```bash
# Dependency sync in parallel:
Task: "T011 [US2] go mod tidy in services/api-gateway/"
Task: "T012 [US2] go mod tidy in services/payment-service/"

# Build checks in parallel:
Task: "T013 [US2] build check in services/api-gateway/"
Task: "T014 [US2] build check in services/payment-service/"
```

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 and Phase 2.
2. Complete US1 (Phase 3).
3. Validate that only target `go.mod` files changed and conflicting `replace` directives are removed.

### Incremental Delivery

1. Deliver US1 cleanup with scope evidence.
2. Deliver US2 validation evidence (tidy/build/runtime).
3. Finish polish tasks for handoff-ready traceability.

### Team Strategy

- Developer A: US1 file edits + scope guard tasks.
- Developer B: US2 verification execution + evidence capture.
- Final cross-check together in Phase 5.

## Notes

- Do not modify any `services/*/go.mod` outside the two target files.
- Keep all command outputs and conclusions in `specs/001-unify-pkg-modules/verification-log.md`.
- If environment limitations block execution, record exact limitation and retain command evidence.
