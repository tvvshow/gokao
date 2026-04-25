# Quickstart: Shared Module Unification (Scoped)

## Preconditions

- Current branch: `001-unify-pkg-modules`
- Target files only:
  - `services/api-gateway/go.mod`
  - `services/payment-service/go.mod`

## Steps

1. Edit the two target `go.mod` files:
   - Remove conflicting `replace` directives tied to legacy `pkg` namespace mapping.
2. Run dependency sync:
   - `cd services/api-gateway && go mod tidy`
   - `cd services/payment-service && go mod tidy`
3. Run build checks:
   - `cd services/api-gateway && go build ./...`
   - `cd services/payment-service && go build ./...`
4. Run basic runtime checks:
   - Start `api-gateway` and verify `GET /healthz` responds.
   - Start `payment-service` and verify startup path reaches service initialization; if blocked by DB/network, record as environment limitation.
5. Confirm scope guard:
   - Verify no other `services/*/go.mod` was edited.

## Expected Results

- Both target `go.mod` files no longer contain conflicting `replace` directives.
- `go mod tidy` executed for both services with recorded outcomes.
- Build/runtime checks attempted for both services with clear pass/fail status.
- Out-of-scope service `go.mod` modifications remain zero.

## Actual Execution Outcomes (2026-04-24)

- Scope verification passed: only `services/api-gateway/go.mod` and `services/payment-service/go.mod` changed among `services/*/go.mod`.
- `go mod tidy` and `go build ./...` were attempted for both target services, but blocked by environment:
  - `/bin/bash: line 1: go: command not found`
- Runtime checks were attempted:
  - `api-gateway`: process started, `/healthz` reachable, returned `HTTP/1.1 503 Service Unavailable` (dependent services unavailable).
  - `payment-service`: startup failed at database initialization (`dial tcp [::1]:5432 ... connection refused`).
- All blocked verifications are classified as environment limitations and recorded in `verification-log.md`.
