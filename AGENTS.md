# Repository Guidelines

## Project Structure & Module Organization
This repository is a multi-service system with Go backends, a Vue frontend, and C++ modules.

- `services/`: Go microservices (`api-gateway`, `user-service`, `data-service`, `payment-service`, `recommendation-service`, `monitoring-service`).
- `frontend/`: Vue 3 + Vite + TypeScript app (`src/views`, `src/api`, `src/__tests__`).
- `cpp-modules/`: native C++ modules (for example `volunteer-matcher`) built via CMake.
- `pkg/`: shared Go packages used across services.
- `tests/integration/`: cross-service integration tests.
- `docker/`, `docker-compose.yml`, `docker-compose.prod.yml`: local/prod orchestration.
- `scripts/`: operational and validation scripts.

Avoid incidental edits in `vcpkg/` unless you are updating C++ dependency tooling.

## Build, Test, and Development Commands
Use these commands from repository root unless noted:

- `make deps`: install Go and frontend dependencies.
- `make build`: build Go services and frontend bundle.
- `make test`: run Go and frontend tests.
- `make test-go` / `make test-frontend`: run one test stack only.
- `docker-compose up -d`: start local services.
- `cd frontend && npm run dev`: run frontend in dev mode.
- `cd frontend && npm run lint && npm run type-check`: frontend quality checks.
- `cd services/api-gateway && go run github.com/swaggo/swag/cmd/swag@v1.8.12 init -g main.go -o docs --parseDependency --parseInternal`: regenerate Swagger docs before PR if API changes.

## Coding Style & Naming Conventions
Go code should be formatted with `gofmt`/`goimports`; keep package names lowercase and exported symbols `PascalCase`.  
Frontend code is enforced by ESLint + Prettier. Use `PascalCase` for Vue view/component files (for example `HomePageModern.vue`), `kebab-case` for API utility files (for example `api-client.ts`), and `camelCase` for functions/variables.

## Testing Guidelines
Go tests should live near implementation files and follow `*_test.go` naming. Run `go test ./... -race -cover`. CI enforces a **60% minimum coverage** gate for `services/api-gateway`.  
Frontend tests use Vitest (`happy-dom`), with tests under `frontend/src/__tests__` and `*.test.ts` suffixes.  
C++ modules should pass `ctest` in each module’s `build/` directory.

## Commit & Pull Request Guidelines
Recent history follows Conventional Commit style such as `fix(frontend): ...`, `build(r1): ...`, `chore(r0): ...`, `docs: ...`. Prefer:

- `type(scope): short imperative summary`
- Small, focused commits with clear scopes

PRs should include a concise summary, linked issue (if any), and verification evidence (commands run). For UI changes, attach screenshots. For API changes, include regenerated Swagger docs in the same PR.

## Security & Configuration Tips
Treat `.env*` and credentials as sensitive. Never commit real secrets. Keep environment-specific values in local or deployment-specific config, and validate production flags (for example Swagger disabled in prod compose).

<!-- SPECKIT START -->
For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan:
specs/001-unify-pkg-modules/plan.md
<!-- SPECKIT END -->
