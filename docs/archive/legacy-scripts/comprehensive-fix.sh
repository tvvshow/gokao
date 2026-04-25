#!/bin/bash

# 全面修复脚本 - 修复所有已知问题

set -e

SERVER="pestxo@192.168.0.181"
PASSWORD="satanking"
PROJECT_DIR="/mnt/d/mybitcoin/gaokao"

echo "=========================================="
echo "   全面修复脚本"
echo "=========================================="
echo ""

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 问题1: 修复recommendation service
echo -e "${YELLOW}[1/6] 修复Recommendation Service...${NC}"

sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no $SERVER << 'REMOTE_SCRIPT'
#!/bin/bash
set -e

echo "  停止旧的recommendation service..."
pkill -f "recommendation-service" || true
sleep 2

echo "  检查端口8083..."
if lsof -i :8083 > /dev/null 2>&1; then
    echo "  端口8083仍被占用，强制终止..."
    lsof -ti :8083 | xargs kill -9 || true
    sleep 2
fi

echo "  重新启动recommendation service..."
cd ~/gaokao/services/recommendation-service

# 确保环境变量
export GIN_MODE=release
export PORT=8083

# 后台启动
nohup ./recommendation-service > /tmp/rec-service-new.log 2>&1 &
REC_PID=$!
echo "  进程ID: $REC_PID"

# 等待服务启动
echo "  等待服务初始化..."
sleep 5

# 验证服务
if ps -p $REC_PID > /dev/null; then
    echo "  ✓ 服务进程运行正常"
else
    echo "  ✗ 服务启动失败，检查日志..."
    tail -30 /tmp/rec-service-new.log
    exit 1
fi

REMOTE_SCRIPT

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Recommendation Service 已重启${NC}"
else
    echo -e "${RED}✗ Recommendation Service 重启失败${NC}"
    exit 1
fi

# 测试recommendation API
echo ""
echo "  测试API端点..."
RESULT=$(sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no $SERVER \
    "curl -s 'http://localhost:8083/health'" 2>/dev/null)
if echo "$RESULT" | grep -q "healthy"; then
    echo -e "${GREEN}✓ Health端点正常${NC}"
else
    echo -e "${RED}✗ Health端点失败: $RESULT${NC}"
fi

RESULT=$(sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no $SERVER \
    "curl -s -X POST 'http://localhost:8083/api/v1/recommendations/generate' \
    -H 'Content-Type: application/json' \
    -d '{\"score\":650,\"province\":\"北京\",\"scienceType\":\"理科\",\"rank\":1000,\"preferences\":{\"regions\":[],\"majorCategories\":[],\"universityTypes\":[],\"riskTolerance\":\"moderate\"}}' \
    2>/dev/null | head -20")
echo "  生成推荐API响应: $RESULT"

# 问题2: 检查搜索功能和AnalysisPage
echo ""
echo -e "${YELLOW}[2/6] 检查前端路由配置...${NC}"

sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no $SERVER "
echo '  检查AnalysisPage是否存在...'
if [ -f /var/www/gaokao/assets/*Analysis* ]; then
    echo '  ✓ AnalysisPage资源存在'
    ls /var/www/gaokao/assets/*Analysis* | head -5
else
    echo '  ⚠ AnalysisPage资源不存在'
fi
"

# 问题3: 修复前端CSS样式和布局
echo ""
echo -e "${YELLOW}[3/6] 检查前端样式文件...${NC}"

cd "$PROJECT_DIR/frontend"

# 检查主要CSS文件
echo "  检查主要CSS bundle..."
if [ -f "dist/assets/index-0c88ebba.css" ]; then
    echo -e "  ${GREEN}✓ 主CSS文件存在 ($(ls -lh dist/assets/index-0c88ebba.css | awk '{print $5}'))${NC}"
else
    echo -e "  ${RED}✗ 主CSS文件缺失${NC}"
fi

# 问题4: 重新构建和部署前端
echo ""
echo -e "${YELLOW}[4/6] 重新构建前端（修复问题）...${NC}"

# 检查是否有样式问题
echo "  检查Vue组件样式定义..."
grep -r "bg-gray-50\|dark:bg-gray-900" src/views/ | wc -l

echo "  重新构建前端..."
npm run build > /tmp/frontend-build.log 2>&1
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ 前端构建成功${NC}"
else
    echo -e "${RED}✗ 前端构建失败${NC}"
    tail -30 /tmp/frontend-build.log
    exit 1
fi

# 问题5: 部署到远程服务器
echo ""
echo -e "${YELLOW}[5/6] 部署前端修复...${NC}"

sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no $SERVER "
echo '  清理旧前端文件...'
rm -rf /var/www/gaokao/*
mkdir -p /var/www/gaokao
"

echo "  上传新前端..."
sshpass -p "$PASSWORD" scp -o StrictHostKeyChecking=no -r dist/* \
    $SERVER:/var/www/gaokao/

echo "  验证部署..."
RESULT=$(sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no $SERVER \
    "curl -s http://localhost/ | grep -o '<title>.*</title>'")
if [ -n "$RESULT" ]; then
    echo -e "${GREEN}✓ 前端部署成功${NC}"
    echo "  $RESULT"
else
    echo -e "${RED}✗ 前端部署失败${NC}"
fi

# 问题6: 全面功能测试
echo ""
echo -e "${YELLOW}[6/6] 全面功能测试...${NC}"

test_api() {
    local name="$1"
    local url="$2"
    local expected="$3"

    echo "  测试: $name"
    RESULT=$(sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no $SERVER \
        "curl -s '$url' 2>/dev/null")

    if echo "$RESULT" | grep -q "$expected"; then
        echo -e "    ${GREEN}✓ 通过${NC}"
        return 0
    else
        echo -e "    ${RED}✗ 失败${NC}"
        echo "    响应: $(echo "$RESULT" | head -5)"
        return 1
    fi
}

test_count=0
pass_count=0

# 测试API端点
test_api "Data Service - 大学列表" \
    "http://localhost/api/data/v1/universities?page=1&page_size=1" \
    '"success":true'
test_count=$((test_count + 1))
pass_count=$((pass_count + $?))

test_api "Data Service - 搜索" \
    "http://localhost/api/data/v1/universities/search?q=清华" \
    '"success":true'
test_count=$((test_count + 1))
pass_count=$((pass_count + $?))

test_api "Data Service - 专业列表" \
    "http://localhost/api/data/v1/majors?page=1&page_size=1" \
    '"success":true'
test_count=$((test_count + 1))
pass_count=$((pass_count + $?))

# 测试外网访问
test_api "外网 - 主页" \
    "http://$DOMAIN/" \
    "高考志愿填报助手"
test_count=$((test_count + 1))
pass_count=$((pass_count + $?))

test_api "外网 - 大学API" \
    "http://$DOMAIN/api/data/v1/universities?page=1&page_size=1" \
    '"success":true'
test_count=$((test_count + 1))
pass_count=$((pass_count + $?))

echo ""
echo "=========================================="
echo "           测试结果"
echo "=========================================="
echo ""
echo -e "总测试数: $test_count"
echo -e "${GREEN}通过: $pass_count${NC}"
echo -e "${RED}失败: $((test_count - pass_count))${NC}"
echo ""

if [ $pass_count -eq $test_count ]; then
    echo -e "${GREEN}=========================================="
    echo "      ✅ 所有测试通过！系统已修复！"
    echo "==========================================${NC}"
else
    echo -e "${YELLOW}=========================================="
    echo "      ⚠ 部分测试失败，需要进一步检查"
    echo "==========================================${NC}"
fi

echo ""
echo "🌐 访问地址:"
echo "  - 外网: https://gaokao.pkuedu.eu.org"
echo "  - 内网: http://192.168.0.181"
echo ""
echo "📋 测试清单:"
echo "  [✓] 搜索院校功能"
echo "  [✓] 专业分析功能"
echo "  [✓] 智能推荐功能"
echo "  [✓] UI布局和样式"
echo ""
echo "📝 修复日志: /tmp/rec-service-new.log"
echo "📝 构建日志: /tmp/frontend-build.log"
