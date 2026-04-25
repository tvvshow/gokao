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
  - Conflicting `replace` directives related to `pkg` namespace must be removed.
  - Module dependency references must remain compatible with `github.com/oktetopython/gaokao/pkg/*`.
  - `go mod tidy` must be executed in module context.

## Verification Contract

- For each in-scope service:
  - Build check must be attempted and result recorded.
  - Basic runtime check must be attempted and result recorded.
  - If environment blocks execution, limitation must be documented with command evidence.

## Acceptance Contract

- Change accepted only when all are true:
  - Two target `go.mod` files updated as specified.
  - No out-of-scope service `go.mod` modified.
  - Tidy/build/runtime evidence available for both targets (or explicit environment limitation for failed attempts).
