#!/bin/bash

# 验证所有服务编译状态
# 执行完整的编译验证

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo -e "${BLUE}🔍 验证所有服务编译状态${NC}"
echo -e "${BLUE}=========================${NC}"

# 服务列表
SERVICES=("api-gateway" "user-service" "data-service" "payment-service")
SUCCESS_COUNT=0
TOTAL_COUNT=${#SERVICES[@]}

# 创建bin目录
mkdir -p "$PROJECT_ROOT/bin"

echo -e "\n${YELLOW}📦 编译所有服务...${NC}"

for service in "${SERVICES[@]}"; do
    echo -e "\n${BLUE}🔄 编译 $service...${NC}"
    
    if [ -d "$PROJECT_ROOT/services/$service" ]; then
        cd "$PROJECT_ROOT/services/$service"
        
        # 清理依赖
        echo "  📋 清理依赖..."
        go mod tidy > /dev/null 2>&1
        
        # 编译服务
        echo "  🔨 编译中..."
        if go build -o "../../bin/$service" . > /dev/null 2>&1; then
            echo -e "  ✅ ${GREEN}$service 编译成功${NC}"
            SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
        else
            echo -e "  ❌ ${RED}$service 编译失败${NC}"
            echo "  🔍 错误详情:"
            go build -o "../../bin/$service" . 2>&1 | sed 's/^/    /'
        fi
    else
        echo -e "  ❌ ${RED}服务目录不存在: $service${NC}"
    fi
done

cd "$PROJECT_ROOT"

echo -e "\n${BLUE}📊 编译结果统计${NC}"
echo -e "${BLUE}===============${NC}"

echo -e "总服务数: ${BLUE}$TOTAL_COUNT${NC}"
echo -e "编译成功: ${GREEN}$SUCCESS_COUNT${NC}"
echo -e "编译失败: ${RED}$((TOTAL_COUNT - SUCCESS_COUNT))${NC}"

SUCCESS_RATE=$((SUCCESS_COUNT * 100 / TOTAL_COUNT))
echo -e "成功率: ${BLUE}$SUCCESS_RATE%${NC}"

# 检查生成的二进制文件
echo -e "\n${YELLOW}📁 检查生成的二进制文件...${NC}"
for service in "${SERVICES[@]}"; do
    if [ -f "bin/$service" ] || [ -f "bin/$service.exe" ]; then
        echo -e "  ✅ ${GREEN}bin/$service${NC}"
    else
        echo -e "  ❌ ${RED}bin/$service 未找到${NC}"
    fi
done

# 最终结果
echo -e "\n${BLUE}🎯 最终结果${NC}"
echo -e "${BLUE}==========${NC}"

if [ "$SUCCESS_COUNT" -eq "$TOTAL_COUNT" ]; then
    echo -e "${GREEN}🎉 所有服务编译成功！${NC}"
    echo -e "${GREEN}✅ 系统已达到生产部署标准${NC}"
    exit 0
elif [ "$SUCCESS_COUNT" -gt 0 ]; then
    echo -e "${YELLOW}⚠️ 部分服务编译成功${NC}"
    echo -e "${YELLOW}🔧 需要修复剩余编译问题${NC}"
    exit 1
else
    echo -e "${RED}❌ 所有服务编译失败${NC}"
    echo -e "${RED}🚨 需要重新检查修复过程${NC}"
    exit 2
fi
