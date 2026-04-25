# Implementation Plan: Shared Module Unification (Scoped)

**Branch**: `001-unify-pkg-modules` | **Date**: 2026-04-25 | **Spec**: [/mnt/d/mybitcoin/gaokao/specs/001-unify-pkg-modules/spec.md](/mnt/d/mybitcoin/gaokao/specs/001-unify-pkg-modules/spec.md)  
**Input**: Feature specification from `/specs/001-unify-pkg-modules/spec.md`

## Summary

Apply a strictly scoped dependency-governance fix to two services only:
1. Remove conflicting `replace` rules in `services/api-gateway/go.mod`.
2. Remove conflicting `replace` rules in `services/payment-service/go.mod`.
3. Run `go mod tidy` in both services.
4. Verify build and basic runtime behavior for both services.

No other service `go.mod` files are in scope.

## Technical Context

**Language/Version**: Go modules workflow (project standard targets Go 1.25)  
**Primary Dependencies**: Go module system, existing `github.com/oktetopython/gaokao/pkg/*` shared modules, Gin-based services  
**Storage**: N/A (no schema or persistence model changes)  
**Testing**: `go mod tidy`, service-level build checks, basic runtime checks (`/healthz` where applicable); when commands fail, persist command evidence and failure classification  
**Target Platform**: Linux dev/CI shell environment for Go commands and service startup  
**Environment Probe (2026-04-25)**: Current shell returned `go: command not found`; treat this as environment mismatch until PATH/toolchain is visible in this execution context  
**Project Type**: Go microservices maintenance task (dependency/config hygiene)  
**Performance Goals**: No startup regression for `api-gateway`; no dependency-resolution regression in both targets  
**Constraints**: Must only modify two target `go.mod` files; must not alter other service `go.mod` files  
**Execution Policy**: Required verification commands are always attempted. If blocked by environment (missing `go`, unavailable DB/Redis, missing container runtime), classify as environment limitation with stderr/exit evidence; otherwise classify as code issue  
**Scale/Scope**: 2 service modules + verification evidence + scoped documentation updates

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] 最优秀原则: Plan includes explicit verification evidence, not file-only edits.
- [x] 有即复用原则: Reuses existing `github.com/oktetopython/gaokao/pkg/*` namespace conventions.
- [x] 不允许简化原则: Includes dependency sync plus build/runtime verification with evidence.
- [x] 技术栈与架构边界原则: Stays within existing Go module boundary; no new stack.
- [x] 强制验证与质量闸门原则: Defines required checks and evidence-driven classification.

No constitutional gate violations detected.

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
└── tasks.md
```

### Source Code (repository root)

```text
services/
├── api-gateway/
│   └── go.mod
└── payment-service/
    └── go.mod
```

**Structure Decision**: Keep code-touch surface constrained to the two target modules and avoid cross-service dependency churn.

## Phase 0: Research Outline

Research items to resolve before implementation details:
- Confirm best-practice handling for removing conflicting `replace` rules in partially unified monorepos.
- Confirm safe `go mod tidy` sequencing in scoped service modules.
- Confirm minimum runtime/build verification set when environment may miss `go` or container tooling.
- Confirm standardized evidence format for blocked commands (exit code, stderr, and limitation classification).

Output artifact: `research.md`.

## Phase 1: Design & Contracts

Design outputs:
- `data-model.md`: defines entities for module file scope, dependency sync result, and verification record.
- `contracts/module-scope-contract.md`: defines scope contract and acceptance contract for this maintenance feature.
- `quickstart.md`: executable verification steps and expected outcomes.

Agent context:
- Keep `AGENTS.md` SPECKIT block pointing to `specs/001-unify-pkg-modules/plan.md`.

## Post-Design Constitution Check

- [x] Scope remains constrained to two service modules.
- [x] Validation evidence is explicit and testable.
- [x] No conflicting requirement remains with clarified spec scope.
- [x] Out-of-scope guardrails are documented.

## Complexity Tracking

No constitution violations requiring justification at planning stage.
