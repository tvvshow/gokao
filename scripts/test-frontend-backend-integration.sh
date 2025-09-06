#!/bin/bash

# 前后端集成测试脚本
# 测试前端和后端的API交互是否正常

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo -e "${BLUE}🔗 前后端集成测试${NC}"
echo -e "${BLUE}==================${NC}"

cd "$PROJECT_ROOT"

# 检查必要的工具
echo -e "\n${YELLOW}🔍 检查必要工具...${NC}"

if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go 未安装${NC}"
    exit 1
fi

if ! command -v node &> /dev/null; then
    echo -e "${RED}❌ Node.js 未安装${NC}"
    exit 1
fi

if ! command -v curl &> /dev/null; then
    echo -e "${RED}❌ curl 未安装${NC}"
    exit 1
fi

echo -e "${GREEN}✅ 所有必要工具已安装${NC}"

# 1. 初始化数据库数据
echo -e "\n${YELLOW}📊 初始化数据库数据...${NC}"

# 设置数据库环境变量
export DATABASE_URL="host=localhost user=postgres password=postgres dbname=gaokao_data port=5432 sslmode=disable"

# 运行数据初始化脚本
if go run scripts/init-data.go; then
    echo -e "${GREEN}✅ 数据库数据初始化成功${NC}"
else
    echo -e "${RED}❌ 数据库数据初始化失败${NC}"
    echo -e "${YELLOW}⚠️ 请确保PostgreSQL正在运行并且数据库gaokao_data已创建${NC}"
    exit 1
fi

# 2. 启动后端服务
echo -e "\n${YELLOW}🚀 启动后端服务...${NC}"

# 启动data-service
echo -e "${BLUE}启动 Data Service...${NC}"
cd services/data-service
go build -o ../../bin/data-service .
cd ../..

# 设置环境变量
export PORT=8082
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=gaokao_data
export DB_USER=postgres
export DB_PASSWORD=postgres

# 后台启动data-service
./bin/data-service > /tmp/data-service.log 2>&1 &
DATA_SERVICE_PID=$!
echo -e "Data Service PID: $DATA_SERVICE_PID"

# 等待服务启动
sleep 5

# 检查data-service是否启动成功
if kill -0 $DATA_SERVICE_PID 2>/dev/null; then
    echo -e "${GREEN}✅ Data Service 启动成功${NC}"
else
    echo -e "${RED}❌ Data Service 启动失败${NC}"
    cat /tmp/data-service.log
    exit 1
fi

# 启动api-gateway
echo -e "${BLUE}启动 API Gateway...${NC}"
cd services/api-gateway
go build -o ../../bin/api-gateway .
cd ../..

# 设置环境变量
export PORT=8080
export DATA_SERVICE_URL=http://localhost:8082

# 后台启动api-gateway
./bin/api-gateway > /tmp/api-gateway.log 2>&1 &
API_GATEWAY_PID=$!
echo -e "API Gateway PID: $API_GATEWAY_PID"

# 等待服务启动
sleep 5

# 检查api-gateway是否启动成功
if kill -0 $API_GATEWAY_PID 2>/dev/null; then
    echo -e "${GREEN}✅ API Gateway 启动成功${NC}"
else
    echo -e "${RED}❌ API Gateway 启动失败${NC}"
    cat /tmp/api-gateway.log
    kill $DATA_SERVICE_PID 2>/dev/null || true
    exit 1
fi

# 3. 测试后端API
echo -e "\n${YELLOW}🧪 测试后端API...${NC}"

# 测试健康检查
echo -e "${BLUE}测试健康检查...${NC}"
if curl -s http://localhost:8080/healthz > /dev/null; then
    echo -e "${GREEN}✅ 健康检查通过${NC}"
else
    echo -e "${RED}❌ 健康检查失败${NC}"
fi

# 测试高校列表API
echo -e "${BLUE}测试高校列表API...${NC}"
UNIVERSITIES_RESPONSE=$(curl -s http://localhost:8080/v1/universities)
if echo "$UNIVERSITIES_RESPONSE" | grep -q "success"; then
    echo -e "${GREEN}✅ 高校列表API正常${NC}"
    echo -e "响应示例: $(echo "$UNIVERSITIES_RESPONSE" | head -c 100)..."
else
    echo -e "${RED}❌ 高校列表API失败${NC}"
    echo -e "响应: $UNIVERSITIES_RESPONSE"
fi

# 测试高校统计API
echo -e "${BLUE}测试高校统计API...${NC}"
STATS_RESPONSE=$(curl -s http://localhost:8080/v1/universities/statistics)
if echo "$STATS_RESPONSE" | grep -q "success"; then
    echo -e "${GREEN}✅ 高校统计API正常${NC}"
    echo -e "响应示例: $(echo "$STATS_RESPONSE" | head -c 100)..."
else
    echo -e "${RED}❌ 高校统计API失败${NC}"
    echo -e "响应: $STATS_RESPONSE"
fi

# 4. 启动前端服务
echo -e "\n${YELLOW}🌐 启动前端服务...${NC}"

cd frontend

# 检查依赖
if [ ! -d "node_modules" ]; then
    echo -e "${BLUE}安装前端依赖...${NC}"
    npm install
fi

# 启动前端开发服务器
echo -e "${BLUE}启动前端开发服务器...${NC}"
npm run dev > /tmp/frontend.log 2>&1 &
FRONTEND_PID=$!
echo -e "Frontend PID: $FRONTEND_PID"

cd ..

# 等待前端启动
sleep 10

# 检查前端是否启动成功
if kill -0 $FRONTEND_PID 2>/dev/null; then
    echo -e "${GREEN}✅ 前端服务启动成功${NC}"
else
    echo -e "${RED}❌ 前端服务启动失败${NC}"
    cat /tmp/frontend.log
    kill $DATA_SERVICE_PID $API_GATEWAY_PID 2>/dev/null || true
    exit 1
fi

# 5. 测试前端页面
echo -e "\n${YELLOW}🧪 测试前端页面...${NC}"

# 测试前端首页
echo -e "${BLUE}测试前端首页...${NC}"
if curl -s http://localhost:3000 > /dev/null; then
    echo -e "${GREEN}✅ 前端首页可访问${NC}"
else
    echo -e "${RED}❌ 前端首页无法访问${NC}"
fi

# 6. 测试前后端API交互
echo -e "\n${YELLOW}🔗 测试前后端API交互...${NC}"

# 通过前端代理测试API
echo -e "${BLUE}通过前端代理测试高校API...${NC}"
PROXY_RESPONSE=$(curl -s http://localhost:3000/v1/universities)
if echo "$PROXY_RESPONSE" | grep -q "success"; then
    echo -e "${GREEN}✅ 前端代理API正常${NC}"
    echo -e "响应示例: $(echo "$PROXY_RESPONSE" | head -c 100)..."
else
    echo -e "${RED}❌ 前端代理API失败${NC}"
    echo -e "响应: $PROXY_RESPONSE"
fi

# 7. 显示访问信息
echo -e "\n${GREEN}🎉 集成测试完成！${NC}"
echo -e "${GREEN}==================${NC}"

echo -e "\n${BLUE}📋 服务访问信息:${NC}"
echo -e "  🌐 前端界面: ${GREEN}http://localhost:3000${NC}"
echo -e "  🔧 API网关: ${GREEN}http://localhost:8080${NC}"
echo -e "  📊 数据服务: ${GREEN}http://localhost:8082${NC}"

echo -e "\n${BLUE}🧪 API测试示例:${NC}"
echo -e "  高校列表: ${GREEN}curl http://localhost:8080/v1/universities${NC}"
echo -e "  高校统计: ${GREEN}curl http://localhost:8080/v1/universities/statistics${NC}"
echo -e "  健康检查: ${GREEN}curl http://localhost:8080/healthz${NC}"

echo -e "\n${YELLOW}📋 注意事项:${NC}"
echo -e "  - 所有服务正在后台运行"
echo -e "  - 使用 Ctrl+C 停止此脚本并清理进程"
echo -e "  - 前端通过代理访问后端API"
echo -e "  - 数据库已初始化基础数据"

# 等待用户中断
echo -e "\n${BLUE}按 Ctrl+C 停止所有服务...${NC}"

# 设置信号处理
cleanup() {
    echo -e "\n${YELLOW}🧹 清理进程...${NC}"
    kill $FRONTEND_PID $API_GATEWAY_PID $DATA_SERVICE_PID 2>/dev/null || true
    wait $FRONTEND_PID $API_GATEWAY_PID $DATA_SERVICE_PID 2>/dev/null || true
    echo -e "${GREEN}✅ 所有服务已停止${NC}"
    exit 0
}

trap cleanup SIGINT SIGTERM

# 保持脚本运行
while true; do
    sleep 1
done
