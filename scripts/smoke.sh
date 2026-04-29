#!/usr/bin/env bash
# gokao 冒烟测试脚本 — 验证所有服务健康
# 用法: ./scripts/smoke.sh [BASE_URL]
# 默认 BASE_URL=http://localhost:8080

set -euo pipefail

BASE="${1:-http://localhost:8080}"
PASS=0
FAIL=0

check() {
  local name="$1" url="$2"
  if curl -sf -o /dev/null -w "%{http_code}" "$url" | grep -q "200"; then
    echo "  ✅ $name"
    ((PASS++))
  else
    echo "  ❌ $name ($url)"
    ((FAIL++))
  fi
}

echo "=== gokao 冒烟测试 ==="
echo "Gateway: $BASE"
echo ""

echo "Gateway:"
check "healthz" "$BASE/healthz"

echo ""
echo "Data Service (via gateway):"
check "health" "$BASE/api/v1/data/health"

echo ""
echo "User Service (via gateway):"
check "health" "$BASE/api/v1/users/health"

echo ""
echo "Payment Service (via gateway):"
check "health" "$BASE/api/v1/payments/health"

echo ""
echo "Recommendation Service (via gateway):"
check "health" "$BASE/api/v1/recommendations/health"

echo ""
echo "Frontend:"
check "index" "http://localhost:3000/"

echo ""
echo "=== 结果: $PASS 通过, $FAIL 失败 ==="
exit $FAIL
