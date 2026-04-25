# Research: Shared Module Unification (Scoped)

## Decision 1: Keep correct module prefix and restrict file scope

- Decision: In this iteration, keep the existing correct module prefix `github.com/oktetopython/gaokao/*`, and remove conflicting local `replace` directives only in `api-gateway` and `payment-service`.
- Rationale: The clarified spec explicitly limits scope to two service modules and forbids editing other service `go.mod` files.
- Alternatives considered:
  - Expand to all services now: rejected due to explicit out-of-scope constraint and larger regression surface.
  - Keep conflicting `replace` directives: rejected because it preserves path ambiguity and violates acceptance criteria.

## Decision 2: Run `go mod tidy` per target service

- Decision: Execute `go mod tidy` separately in each target service directory after `go.mod` edits.
- Rationale: Service-local tidy minimizes blast radius and keeps dependency changes attributable to each module.
- Alternatives considered:
  - Run tidy at repo root only: rejected because this repo is multi-module and root tidy may not update target modules correctly.
  - Skip tidy: rejected because dependency sync is an explicit feature requirement.

## Decision 3: Verification contract under environment constraints

- Decision: Verification requires build + basic runtime checks for both target services; when `go`/container tools are unavailable, record limitation explicitly and provide best available runtime evidence.
- Rationale: The spec requires verification output and explicit separation of code issues vs environment limitations.
- Alternatives considered:
  - Report success without command evidence: rejected as non-compliant with constitution quality gates.
  - Block completion until perfect environment exists: rejected for planning; execution can continue with transparent limitation reporting.

## Decision 4: Use architecture report as current-state gate

- Decision: Use `docs/ARCHITECTURE_REPORT.md` as baseline to exclude already-resolved issues (port conflicts, incomplete go.work registration) from this iteration scope. Identify and correct report inaccuracies (wrong namespace migration claim, understated replace coverage) rather than acting on them.
- Rationale: Planning must target unresolved items only; otherwise task set is noisy and causes false blockers. The report's claim that `oktetopython` needs migration to `gaokaohub` is based on incorrect premise and must not be executed.
- Alternatives considered:
  - Re-open all historical issues from old reports: rejected because it mixes closed and open items, reducing execution focus.
  - Ignore architecture report: rejected because this feature is explicitly requested to be updated based on the latest report.

## Decision 5: Integrate reuse audit into US1 with three-class failure taxonomy

- Decision: Merge the reuse audit (T012-T015) into User Story 1 alongside replace cleanup, using keep/deprecate/remove verdicts and the Design Gap classification for removals blocked by pkg interface incompatibility.
- Rationale: The audit is a direct consequence of replace removal — removing `replace` directives may expose that local implementations duplicate pkg modules. Auditing and classifying within the same user story keeps the dependency governance loop closed. The Design Gap category (per FR-008) captures the intermediate state where removal is blocked by interface incompatibility, avoiding forced migration or silent suppression.
- Alternatives considered:
  - Separate audit into its own user story: rejected because audit findings directly inform whether replace removal is safe, making the dependency bidirectional.
  - Two-class taxonomy (Code Issue / Environment Limitation) only: rejected because it cannot express "pkg interface incompatibility exposed by replace removal" — neither a code bug nor an environment problem.
