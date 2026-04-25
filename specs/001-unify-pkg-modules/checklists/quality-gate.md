# Quality-Gate Requirements Checklist: Shared Module Unification

**Purpose**: Validate requirement quality for scope control, conditional rollback, runtime criteria, and gate governance before implementation.
**Created**: 2026-04-25
**Feature**: `/mnt/d/mybitcoin/gaokao/specs/001-unify-pkg-modules/spec.md`

## Requirement Completeness

- [ ] CHK001 Are quality-gate outcome requirements explicitly defined with all three terminal states (`Gate Passed` / `Gate Waived` / `Gate Blocked`)? [Completeness, Spec §FR-009]
- [ ] CHK002 Are waiver metadata requirements complete (evidence, owner, deadline, remediation plan) for all waived outcomes? [Completeness, Spec §FR-009, Spec §SC-008]
- [ ] CHK003 Are scope-boundary requirements complete for both inclusion and exclusion sets of `go.mod` files? [Completeness, Spec §FR-001, Spec §FR-007]
- [ ] CHK004 Are fallback/rollback requirements explicitly documented when remove-then-build flow triggers keep conditions? [Completeness, Tasks T016-T017]

## Requirement Clarity

- [ ] CHK005 Is “basic runtime verification” quantified with concrete thresholds (time window, HTTP status, process state)? [Clarity, Spec §FR-006, Spec Assumptions]
- [ ] CHK006 Is the wording for “external limitation” precise enough to distinguish dependency-access blockage from product defects? [Clarity, Spec §FR-008, Spec Edge Cases]
- [ ] CHK007 Are the acceptance meanings of “candidate conflicting replace” and “keep-condition triggered” explicitly interpretable without implementation assumptions? [Clarity, Tasks T008-T009, Tasks T016]

## Requirement Consistency

- [ ] CHK008 Do quality-gate definitions in spec, plan, tasks, and contract use the same terminology and state set without drift? [Consistency, Spec §FR-009, Plan §Quality Gate Closure, Contract §Acceptance]
- [ ] CHK009 Are runtime success criteria consistent between spec assumptions and task execution language? [Consistency, Spec Assumptions, Tasks T022-T023]
- [ ] CHK010 Do rollback-related statements align between user-story independent test criteria and ordered task flow? [Consistency, Spec US1, Tasks §US1]

## Acceptance Criteria Quality

- [ ] CHK011 Are SC-004/SC-005 measurable independently from environment availability and clearly separable from SC-008 gate outcome semantics? [Measurability, Spec §SC-004, Spec §SC-005, Spec §SC-008]
- [ ] CHK012 Is SC-008 objectively auditable from artifacts (not subjective interpretation)? [Acceptance Criteria, Spec §SC-008, Tasks T027]
- [ ] CHK013 Do acceptance statements define what constitutes “not merge-eligible” in requirement terms, not only plan prose? [Gap, Plan §Quality Gate Closure, Spec §FR-009]

## Scenario Coverage

- [ ] CHK014 Are requirements present for Primary flow (all checks pass), Alternate flow (partial failures), and Recovery flow (conditional rollback)? [Coverage, Spec US1/US2, Tasks T016-T017]
- [ ] CHK015 Are exception-flow requirements defined for private repository access failures and external dependency outages? [Coverage, Spec Edge Cases, Plan §Risks]
- [ ] CHK016 Are requirements explicit on whether gate evaluation occurs per service and per iteration, including aggregation logic? [Gap, Spec §FR-009]

## Edge Case Coverage

- [ ] CHK017 Are requirements defined for mixed outcomes (one target service passes and the other is blocked) and their impact on final gate status? [Edge Case, Gap]
- [ ] CHK018 Is the requirement behavior defined when health endpoint is absent, renamed, or non-standard but service is otherwise healthy? [Edge Case, Spec Assumptions, Tasks T022-T023]
- [ ] CHK019 Are requirements explicit for stale evidence handling (e.g., evidence collected before latest task/spec revisions)? [Gap, Traceability]

## Non-Functional Requirements

- [ ] CHK020 Are performance non-functional requirements either measurable or explicitly waived with auditable evidence requirements? [Non-Functional, Constitution §I, Tasks T029]
- [ ] CHK021 Are auditability and traceability requirements sufficient to support governance review of gate decisions? [Non-Functional, Constitution §V, Tasks T026-T027]

## Dependencies & Assumptions

- [ ] CHK022 Are assumptions about toolchain, private module access, and runtime dependencies documented with requirement-level impact boundaries? [Assumption, Spec Assumptions, Plan §Technical Context]

## Notes

- Check items off as completed: `[x]`
- This checklist validates requirement quality only; it does not validate implementation behavior.
