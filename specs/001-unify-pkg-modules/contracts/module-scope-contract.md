# Contract: Module Scope & Validation

## Scope Contract

- In scope files:
  - `services/api-gateway/go.mod`
  - `services/payment-service/go.mod`
- Out of scope:
  - Any other `services/*/go.mod`
  - Non-essential cross-service dependency refactors

## Dependency Contract

- For each in-scope module file:
  - Conflicting local `replace` directives must be removed (per spec US1 判定规则).
  - Module dependency references must remain compatible with `github.com/oktetopython/gaokao/pkg/*`.
  - `go mod tidy` must be executed in module context.

## Reuse Audit Contract

- For each in-scope service:
  - Local implementations duplicating `pkg/*` shared libraries must be audited (spec FR-003).
  - Each duplicate must receive a keep/deprecate/remove verdict (spec SC-002).
  - Items marked remove must be verified via `go build`; failures classified as Design Gap.

## Verification Contract

- For each in-scope service:
  - Build check must be attempted and result recorded.
  - Basic runtime check must be attempted and result recorded.
  - If environment blocks execution, limitation must be documented with command evidence.
  - All failures must be classified as Code Issue / Environment Limitation / Design Gap (spec FR-008, SC-007).

## Acceptance Contract

- Change accepted only when all are true:
  - Two target `go.mod` files updated as specified.
  - No out-of-scope service `go.mod` modified (spec FR-007).
  - Reuse audit completed with verdicts for all identified duplicates (spec SC-002).
  - Tidy/build/runtime evidence available for both targets.
  - Quality gate status is explicit:
    - `Gate Passed`: all required checks pass in available verification environment
    - `Gate Waived`: blocked by external limitation with recorded evidence + owner + deadline + remediation plan
    - `Gate Blocked`: neither passed nor validly waived (not merge-eligible)
  - All failures classified per spec FR-008.
