# Quickstart: Shared Module Unification (Scoped)

## Preconditions

- Current branch: `001-unify-pkg-modules`
- Target files only:
  - `services/api-gateway/go.mod`
  - `services/payment-service/go.mod`
- Module prefix rule: `github.com/oktetopython/gaokao/*` is the correct namespace; do not migrate to any other prefix.

## Steps

1. Edit the two target `go.mod` files:
   - Remove conflicting local `replace` directives per T008/T009 判定规则 (see `tasks.md`).
   - Keep module prefix `github.com/oktetopython/gaokao/*` unchanged.
2. Audit duplicate implementations in both target services (T012/T013):
   - Identify local CORS, auth wrapper, error handler implementations that duplicate `pkg/*` equivalents.
   - Mark as deprecated or remove per复用改造判定规则 (T014).
   - Verify build passes after changes (T015); failures classified as Design Gap.
3. Run dependency sync:
   - `cd services/api-gateway && go mod tidy`
   - `cd services/payment-service && go mod tidy`
4. Run build checks:
   - `cd services/api-gateway && go build ./...`
   - `cd services/payment-service && go build ./...`
5. Run basic runtime checks:
   - Start `api-gateway` and verify `GET /healthz` (or equivalent health endpoint) returns HTTP 200 within 15 seconds.
   - Start `payment-service` and verify process remains alive for 15 seconds without fatal/panic; if health endpoint exists, prefer HTTP 200 check.
   - If blocked by DB/network/private dependency access, record as Environment Limitation with command evidence.
6. Confirm scope guard:
   - Verify no other `services/*/go.mod` was edited.
7. Classify failures (T022):
   - Code Issue: module path/version mismatch exposed by replace removal.
   - Environment Limitation: toolchain/DB/Redis unavailable.
   - Design Gap: pkg interface incompatibility exposed by replace removal (requires follow-up iteration).
8. Determine quality gate status:
   - `Gate Passed`: tidy/build/runtime required checks all pass.
   - `Gate Waived`: external limitation with recorded evidence + owner + deadline + remediation plan.
   - `Gate Blocked`: neither passed nor validly waived.
9. Cross-check against architecture baseline:
   - Ensure outcomes align with `docs/ARCHITECTURE_REPORT.md`.
   - Note: the architecture report's claim that `oktetopython` namespace needs "unification to gaokaohub" is incorrect — `oktetopython/gaokao` is the correct repo namespace. Do not act on that recommendation.
   - Note: the architecture report understates replace directive coverage — `data-service` (4) and `user-service` (4) also have replace directives, but these are out of scope for this feature.

## Expected Results

- Both target `go.mod` files no longer contain conflicting `replace` directives.
- Duplicate implementations in target services audited with clear verdicts (keep/deprecate/remove).
- `go mod tidy` executed for both services with recorded outcomes.
- Build/runtime checks attempted for both services with clear pass/fail status and failure classification.
- Out-of-scope service `go.mod` modifications remain zero.
- Quality gate status is explicit (`Gate Passed` / `Gate Waived` / `Gate Blocked`).

## Execution Status

- Execution completed for T001-T029 in current iteration.
- `services/api-gateway/go.mod` and `services/payment-service/go.mod` conflicting `replace` directives were removed.
- `go mod tidy` / `go build ./...` / `go run .` were executed for both target services and failed before startup due to module resolution to private remote (`github.com/oktetopython/gaokao/pkg/*@v0.0.0`).
- Failures were classified as Environment Limitation + Design Gap in `verification-log.md`.
- Scope guard passed: only the two in-scope `go.mod` files changed.
- Current quality gate status: `Gate Blocked` (required checks failed and no approved waiver metadata recorded).
