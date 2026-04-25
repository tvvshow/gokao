# Data Model: Shared Module Unification (Scoped)

## Entity: TargetServiceModuleFile

- Description: A target service `go.mod` file included in this iteration.
- Fields:
  - `service_name` (enum): `api-gateway` | `payment-service`
  - `file_path` (string): absolute/project-relative path to `go.mod`
  - `conflicting_replace_present` (boolean)
  - `conflicting_replace_removed` (boolean)
  - `namespace_reference_state` (enum): `aligned` | `mixed` | `unknown`
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

## Entity: VerificationRecord

- Description: Build and basic runtime verification evidence for each target service.
- Fields:
  - `service_name` (enum): `api-gateway` | `payment-service`
  - `build_check_attempted` (boolean)
  - `build_check_success` (boolean)
  - `runtime_check_attempted` (boolean)
  - `runtime_check_success` (boolean)
  - `environment_limitation` (string, optional)
  - `evidence` (string): command output summary or key lines
- Validation Rules:
  - Build and runtime checks must be attempted for both target services.
  - If checks fail due to environment, `environment_limitation` must be explicit.

## State Transitions

1. `TargetServiceModuleFile`: `conflicting_replace_present=true` -> edit applied -> `conflicting_replace_removed=true`.
2. `DependencySyncResult`: `tidy_attempted=false` -> execute tidy -> `tidy_attempted=true` and success/failure captured.
3. `VerificationRecord`: `*_attempted=false` -> run checks -> attempted flags set and outcome recorded.
