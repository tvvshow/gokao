#!/bin/bash

# 修复所有go.mod文件的Go版本为1.22
# 确保与WSL中的Go版本兼容

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "🔧 修复Go版本兼容性问题..."

# 修复各个服务的go.mod
echo "修复服务go.mod文件..."
sed -i 's/go 1\.2[3-9]/go 1.22/g' "$PROJECT_ROOT/services/user-service/go.mod"
sed -i 's/go 1\.2[3-9]\.0/go 1.22/g' "$PROJECT_ROOT/services/payment-service/go.mod"
sed -i 's/go 1\.2[3-9]/go 1.22/g' "$PROJECT_ROOT/services/data-service/go.mod"
sed -i 's/go 1\.2[5-9]/go 1.22/g' "$PROJECT_ROOT/services/recommendation-service/go.mod"

# 修复pkg目录下的go.mod
echo "修复pkg go.mod文件..."
find "$PROJECT_ROOT/pkg" -name "go.mod" -exec sed -i 's/go 1\.2[1-9]/go 1.22/g' {} \;

echo "✅ Go版本修复完成"

# 验证修复结果
echo "📋 验证修复结果:"
echo "go.work:"
head -1 "$PROJECT_ROOT/go.work"

echo "服务go.mod版本:"
grep "^go " "$PROJECT_ROOT/services/*/go.mod" | head -5

echo "pkg go.mod版本:"
grep "^go " "$PROJECT_ROOT/pkg/*/go.mod" | head -5
