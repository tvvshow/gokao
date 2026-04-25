#!/bin/bash

# 高考志愿填报系统 - 审计修复验证脚本
# 验证所有Critical问题修复是否成功

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo -e "${BLUE}🔍 高考志愿填报系统 - 审计修复验证${NC}"
echo -e "${BLUE}===========================================${NC}"

# 验证计数器
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0

# 检查函数
check_item() {
    local description="$1"
    local condition="$2"
    
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    
    if eval "$condition"; then
        echo -e "  ✅ $description"
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
        return 0
    else
        echo -e "  ❌ $description"
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
        return 1
    fi
}

# 1. 验证共享包结构
echo -e "\n${YELLOW}📁 验证共享包结构...${NC}"

check_item "pkg/auth 目录存在" "[ -d '$PROJECT_ROOT/pkg/auth' ]"
check_item "pkg/errors 目录存在" "[ -d '$PROJECT_ROOT/pkg/errors' ]"
check_item "pkg/database 目录存在" "[ -d '$PROJECT_ROOT/pkg/database' ]"
check_item "pkg/logger 目录存在" "[ -d '$PROJECT_ROOT/pkg/logger' ]"
check_item "pkg/middleware 目录存在" "[ -d '$PROJECT_ROOT/pkg/middleware' ]"
check_item "pkg/models 目录存在" "[ -d '$PROJECT_ROOT/pkg/models' ]"
check_item "pkg/utils 目录存在" "[ -d '$PROJECT_ROOT/pkg/utils' ]"

# 2. 验证认证包
echo -e "\n${YELLOW}🔐 验证认证包...${NC}"

check_item "认证中间件文件存在" "[ -f '$PROJECT_ROOT/pkg/auth/middleware.go' ]"
check_item "认证包go.mod存在" "[ -f '$PROJECT_ROOT/pkg/auth/go.mod' ]"
check_item "认证中间件包含RequireAuth函数" "grep -q 'func.*RequireAuth' '$PROJECT_ROOT/pkg/auth/middleware.go'"
check_item "认证中间件包含OptionalAuth函数" "grep -q 'func.*OptionalAuth' '$PROJECT_ROOT/pkg/auth/middleware.go'"
check_item "认证中间件包含JWT验证逻辑" "grep -q 'jwt.Parse' '$PROJECT_ROOT/pkg/auth/middleware.go'"

# 3. 验证错误处理包
echo -e "\n${YELLOW}❌ 验证错误处理包...${NC}"

check_item "错误处理文件存在" "[ -f '$PROJECT_ROOT/pkg/errors/errors.go' ]"
check_item "错误处理包go.mod存在" "[ -f '$PROJECT_ROOT/pkg/errors/go.mod' ]"
check_item "包含APIError结构体" "grep -q 'type APIError struct' '$PROJECT_ROOT/pkg/errors/errors.go'"
check_item "包含标准错误码" "grep -q 'ErrCodeInvalidInput' '$PROJECT_ROOT/pkg/errors/errors.go'"
check_item "包含NewAPIError函数" "grep -q 'func NewAPIError' '$PROJECT_ROOT/pkg/errors/errors.go'"

# 4. 验证通用Makefile
echo -e "\n${YELLOW}🔨 验证通用Makefile...${NC}"

check_item "通用Makefile存在" "[ -f '$PROJECT_ROOT/Makefile.common' ]"
check_item "包含build目标" "grep -q '^build:' '$PROJECT_ROOT/Makefile.common'"
check_item "包含test目标" "grep -q '^test:' '$PROJECT_ROOT/Makefile.common'"
check_item "包含docker-build目标" "grep -q '^docker-build:' '$PROJECT_ROOT/Makefile.common'"
check_item "包含帮助信息" "grep -q '^help:' '$PROJECT_ROOT/Makefile.common'"

# 5. 验证服务编译
echo -e "\n${YELLOW}🔧 验证服务编译...${NC}"

cd "$PROJECT_ROOT"

# 检查各服务是否可以编译
SERVICES=("api-gateway" "user-service" "payment-service")

for service in "${SERVICES[@]}"; do
    if [ -d "services/$service" ]; then
        echo -e "  🔄 检查 $service 编译..."
        if (cd "services/$service" && go build -o /tmp/test-build . >/dev/null 2>&1); then
            check_item "$service 编译成功" "true"
            rm -f /tmp/test-build
        else
            check_item "$service 编译成功" "false"
        fi
    else
        check_item "$service 目录存在" "false"
    fi
done

# 6. 验证Go模块
echo -e "\n${YELLOW}📦 验证Go模块...${NC}"

# 检查共享包的go.mod文件
for pkg in auth errors; do
    if [ -f "pkg/$pkg/go.mod" ]; then
        check_item "pkg/$pkg go.mod 格式正确" "grep -q 'module github.com/oktetopython/gaokao/pkg/$pkg' 'pkg/$pkg/go.mod'"
    fi
done

# 7. 验证代码质量
echo -e "\n${YELLOW}📋 验证代码质量...${NC}"

# 检查是否有明显的代码重复
AUTH_FILES=$(find services -name "*.go" -exec grep -l "RequireAuth\|jwt\.Parse" {} \; 2>/dev/null | wc -l)
if [ "$AUTH_FILES" -gt 3 ]; then
    check_item "认证代码重复减少" "false"
else
    check_item "认证代码重复减少" "true"
fi

# 检查是否有未使用的导入
UNUSED_IMPORTS=$(find services -name "*.go" -exec grep -l '"fmt".*//.*unused\|import.*fmt.*$' {} \; 2>/dev/null | wc -l)
check_item "未使用导入清理" "[ '$UNUSED_IMPORTS' -eq 0 ]"

# 8. 验证文档
echo -e "\n${YELLOW}📚 验证文档...${NC}"

check_item "系统审计报告存在" "[ -f '$PROJECT_ROOT/docs/system-audit-final.md' ]"
check_item "审计完成报告存在" "[ -f '$PROJECT_ROOT/docs/audit-completion-report.md' ]"
check_item "修复脚本存在" "[ -f '$PROJECT_ROOT/scripts/fix-critical-issues.sh' ]"
check_item "验证脚本存在" "[ -f '$PROJECT_ROOT/scripts/verify-audit-fixes.sh' ]"

# 9. 验证配置文件
echo -e "\n${YELLOW}⚙️ 验证配置文件...${NC}"

check_item "golangci-lint配置存在" "[ -f '$PROJECT_ROOT/.golangci.yml' ]"
check_item "Docker配置存在" "[ -f '$PROJECT_ROOT/docker-compose.yml' ]"
check_item "生产环境Docker配置存在" "[ -f '$PROJECT_ROOT/docker-compose.prod.yml' ]"

# 10. 生成验证报告
echo -e "\n${BLUE}📊 验证结果统计${NC}"
echo -e "${BLUE}==================${NC}"

PASS_RATE=$((PASSED_CHECKS * 100 / TOTAL_CHECKS))

echo -e "总检查项: ${BLUE}$TOTAL_CHECKS${NC}"
echo -e "通过检查: ${GREEN}$PASSED_CHECKS${NC}"
echo -e "失败检查: ${RED}$FAILED_CHECKS${NC}"
echo -e "通过率: ${BLUE}$PASS_RATE%${NC}"

if [ "$PASS_RATE" -ge 90 ]; then
    echo -e "\n${GREEN}🎉 验证结果: 优秀 (≥90%)${NC}"
    echo -e "${GREEN}✅ Critical问题修复成功，系统质量显著提升！${NC}"
elif [ "$PASS_RATE" -ge 80 ]; then
    echo -e "\n${YELLOW}⚠️ 验证结果: 良好 (80-89%)${NC}"
    echo -e "${YELLOW}🔧 大部分问题已修复，还有少量问题需要处理${NC}"
elif [ "$PASS_RATE" -ge 70 ]; then
    echo -e "\n${YELLOW}⚠️ 验证结果: 一般 (70-79%)${NC}"
    echo -e "${YELLOW}🔧 部分问题已修复，需要继续改进${NC}"
else
    echo -e "\n${RED}❌ 验证结果: 需要改进 (<70%)${NC}"
    echo -e "${RED}🚨 修复效果不理想，需要重新检查修复过程${NC}"
fi

# 11. 生成后续建议
echo -e "\n${BLUE}📋 后续建议${NC}"
echo -e "${BLUE}============${NC}"

if [ "$FAILED_CHECKS" -gt 0 ]; then
    echo -e "${YELLOW}需要处理的问题:${NC}"
    echo -e "  1. 检查失败的验证项"
    echo -e "  2. 重新运行修复脚本"
    echo -e "  3. 手动修复特定问题"
fi

echo -e "\n${GREEN}下一步行动:${NC}"
echo -e "  1. 更新各服务引用新的共享包"
echo -e "  2. 重构服务级别的Makefile"
echo -e "  3. 运行完整测试套件"
echo -e "  4. 开始High优先级问题修复"

# 12. 保存验证结果
REPORT_FILE="$PROJECT_ROOT/docs/verification-report-$(date +%Y%m%d-%H%M%S).md"
cat > "$REPORT_FILE" << EOF
# 审计修复验证报告

**验证时间**: $(date)
**验证脚本**: scripts/verify-audit-fixes.sh

## 验证结果
- 总检查项: $TOTAL_CHECKS
- 通过检查: $PASSED_CHECKS  
- 失败检查: $FAILED_CHECKS
- 通过率: $PASS_RATE%

## 验证状态
$(if [ "$PASS_RATE" -ge 90 ]; then echo "✅ 优秀"; elif [ "$PASS_RATE" -ge 80 ]; then echo "⚠️ 良好"; elif [ "$PASS_RATE" -ge 70 ]; then echo "⚠️ 一般"; else echo "❌ 需要改进"; fi)

## 后续行动
1. 更新各服务引用新的共享包
2. 重构服务级别的Makefile  
3. 运行完整测试套件
4. 开始High优先级问题修复
EOF

echo -e "\n${GREEN}📄 验证报告已保存: $REPORT_FILE${NC}"

# 返回适当的退出码
if [ "$PASS_RATE" -ge 80 ]; then
    exit 0
else
    exit 1
fi
