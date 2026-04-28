#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

SKIP_DOCKER="${SKIP_DOCKER:-0}"

section() {
  echo -e "\n${BLUE}== $1 ==${NC}"
}

pass() {
  echo -e "${GREEN}✅ $1${NC}"
}

warn() {
  echo -e "${YELLOW}⚠️  $1${NC}"
}

section "Go Tests (Core Services)"
for svc in api-gateway data-service user-service recommendation-service; do
  echo "running: services/$svc"
  (cd "services/$svc" && go test ./...)
done
pass "Core Go service tests passed"

section "Frontend Quality"
(cd frontend && npm run lint)
(cd frontend && npm run type-check)
pass "Frontend lint/type-check passed"

section "Swagger Up-To-Date"
(cd services/api-gateway && go run github.com/swaggo/swag/cmd/swag@v1.8.12 init -g main.go -o docs --parseDependency --parseInternal)
if ! git diff --quiet -- services/api-gateway/docs; then
  echo -e "${RED}Swagger docs are out-of-date under services/api-gateway/docs${NC}"
  echo "Please commit regenerated docs before release."
  exit 1
fi
pass "Swagger docs are up-to-date"

section "Docker Build Gate"
if [[ "$SKIP_DOCKER" == "1" ]]; then
  warn "Docker build checks skipped (SKIP_DOCKER=1)"
else
  if ! command -v docker >/dev/null 2>&1; then
    warn "Docker not installed in current environment; skipping docker build gate"
  else
    if ! docker compose config >/dev/null; then
      echo -e "${RED}docker compose config failed${NC}"
      exit 1
    fi
    docker compose build api-gateway recommendation-service
    pass "Docker build gate passed"
  fi
fi

section "Final Verdict"
pass "Go-live gate checks completed"

