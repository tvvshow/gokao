# Data Model: Shared Module Unification (Scoped)

## Entity: TargetServiceModuleFile

- Description: A target service `go.mod` file included in this iteration.
- Fields:
  - `service_name` (enum): `api-gateway` | `payment-service`
  - `file_path` (string): absolute/project-relative path to `go.mod`
  - `conflicting_replace_present` (boolean)
  - `conflicting_replace_removed` (boolean)
  - `module_prefix_consistency` (enum): `consistent` | `mixed` | `unknown`
- Validation Rules:
  - Must map to one of the two allowed target services.
  - `conflicting_replace_removed` must be `true` before verification phase closes.

## Entity: DependencySyncResult

- Description: Outcome of running `go mod tidy` in a target service module.
- Fields:
  - `service_name` (enum): `api-gateway` | `payment-service`
  - `tidy_attempted` (boolean)
  - `tidy_success` (boolean)
  - `error_summary` (string, optional)
  - `timestamp` (datetime string)
- Validation Rules:
  - `tidy_attempted` must be `true` for both targets.
  - If `tidy_success` is `false`, `error_summary` is required.

## Entity: ReuseAuditItem

- Description: A local implementation in a target service that overlaps with a `pkg/*` shared library.
- Fields:
  - `service_name` (enum): `api-gateway` | `payment-service`
  - `file_path` (string): project-relative path to the local implementation
  - `pkg_equivalent` (string): the `pkg/*` module that provides equivalent functionality
  - `verdict` (enum): `keep` | `deprecate` | `remove`
  - `design_gap_flag` (boolean): `true` if removal blocked by pkg interface incompatibility
  - `rationale` (string): justification for the verdict
- Validation Rules:
  - Every identified duplicate must have a non-empty verdict.
  - If `verdict` is `remove` and `go build` fails after removal, `design_gap_flag` must be `true`.

## Entity: VerificationRecord

- Description: Build and basic runtime verification evidence for each target service.
- Fields:
  - `service_name` (enum): `api-gateway` | `payment-service`
  - `build_check_attempted` (boolean)
  - `build_check_success` (boolean)
  - `runtime_check_attempted` (boolean)
  - `runtime_check_success` (boolean)
  - `failure_classification` (enum, optional): `Code Issue` | `Environment Limitation` | `Design Gap`
  - `evidence` (string): command output summary or key lines
- Validation Rules:
  - Build and runtime checks must be attempted for both target services.
  - If any check fails, `failure_classification` must be set per FR-008/SC-007.

## Entity: QualityGateRecord

- Description: Merge-eligibility gate status for this feature iteration.
- Fields:
  - `gate_status` (enum): `Gate Passed` | `Gate Waived` | `Gate Blocked`
  - `waiver_reason` (string, optional)
  - `waiver_owner` (string, optional)
  - `waiver_deadline` (date, optional)
  - `remediation_plan` (string, optional)
- Validation Rules:
  - `Gate Passed` requires tidy/build/runtime required checks all passed.
  - `Gate Waived` requires `waiver_reason`, `waiver_owner`, `waiver_deadline`, `remediation_plan`.
  - `Gate Blocked` indicates not merge-eligible.

## State Transitions

1. `TargetServiceModuleFile`: `conflicting_replace_present=true` -> edit applied -> `conflicting_replace_removed=true`.
2. `ReuseAuditItem`: duplicate identified -> audit verdict assigned -> if `remove` then local code removed/deprecated and build verified; if build fails then `design_gap_flag=true`.
3. `DependencySyncResult`: `tidy_attempted=false` -> execute tidy -> `tidy_attempted=true` and success/failure captured.
4. `VerificationRecord`: `*_attempted=false` -> run checks -> attempted flags set and outcome recorded; if failed, `failure_classification` set per FR-008.
5. `QualityGateRecord`: initialize as `Gate Blocked` -> become `Gate Passed` when all checks pass, or `Gate Waived` when approved waiver metadata is complete.
