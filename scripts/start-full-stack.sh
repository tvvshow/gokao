#!/bin/bash

# 高考志愿填报助手 - 全栈启动脚本
# 用于一键启动整个系统（后端 + 前端）

echo "🎓 高考志愿填报助手 - 全栈启动"
echo "====================================="

# 获取脚本目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "📁 项目目录: $PROJECT_DIR"

# 检查必要工具
echo "🔍 检查系统环境..."

# 检查Docker
if ! command -v docker &> /dev/null; then
    echo "❌ Docker 未安装，请先安装 Docker"
    exit 1
fi

# 检查Docker Compose
if ! docker compose version &> /dev/null; then
    echo "❌ Docker Compose V2 未安装，请先安装 Docker Compose V2 (docker compose plugin)"
    exit 1
fi

# 检查Node.js（用于前端开发）
if ! command -v node &> /dev/null; then
    echo "❌ Node.js 未安装，请先安装 Node.js 18+"
    exit 1
fi

echo "✅ 环境检查通过"

# 停止可能存在的服务
echo "🛑 停止现有服务..."
cd "$PROJECT_DIR"
docker compose down 2>/dev/null || true

# 启动后端服务
echo "🚀 启动后端服务..."
docker compose up -d postgres redis data-service api-gateway

# 等待服务启动
echo "⏳ 等待后端服务启动..."
sleep 10

# 检查后端健康状态
MAX_RETRIES=30
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -s http://localhost:8080/healthz > /dev/null 2>&1; then
        echo "✅ 后端服务启动成功"
        break
    fi
    
    echo "⏳ 等待后端服务启动... ($((RETRY_COUNT + 1))/$MAX_RETRIES)"
    sleep 2
    RETRY_COUNT=$((RETRY_COUNT + 1))
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo "❌ 后端服务启动超时，请检查日志"
    docker compose logs
    exit 1
fi

# 安装前端依赖（如果需要）
echo "📦 检查前端依赖..."
cd "$PROJECT_DIR/frontend"

if [ ! -d "node_modules" ]; then
    echo "📦 安装前端依赖..."
    npm install || {
        echo "❌ 前端依赖安装失败"
        exit 1
    }
    echo "✅ 前端依赖安装完成"
fi

# 启动前端开发服务器
echo "🚀 启动前端开发服务器..."
echo ""
echo "🎉 系统启动完成！"
echo "====================================="
echo "🌐 前端界面: http://localhost:3000"
echo "🔧 API文档: http://localhost:8080/swagger/index.html"
echo "🗄️ 数据服务: http://localhost:8082"
echo ""
echo "📊 服务状态："
echo "   - PostgreSQL: localhost:5432"
echo "   - Redis: localhost:6379"
echo "   - API网关: localhost:8080"
echo "   - 数据服务: localhost:8082"
echo "   - 前端界面: localhost:3000"
echo ""
echo "🛑 按 Ctrl+C 停止前端服务"
echo "🛑 停止所有服务: docker compose down"
echo "====================================="

# 启动前端开发服务器
npm run dev