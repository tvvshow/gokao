# Research: Shared Module Unification (Scoped)

## Decision 1: Keep namespace convergence but restrict file scope

- Decision: In this iteration, remove conflicting `replace` directives only in `api-gateway` and `payment-service`, while aligning references to `github.com/oktetopython/gaokao/pkg/*`.
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
