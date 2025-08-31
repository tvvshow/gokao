#!/bin/bash

# 高考志愿填报助手 - 前端启动脚本
# 用于快速启动前端开发环境

echo "🎓 高考志愿填报助手 - 前端启动"
echo "=================================="

# 检查Node.js版本
if ! command -v node &> /dev/null; then
    echo "❌ Node.js 未安装，请先安装 Node.js 18+"
    exit 1
fi

NODE_VERSION=$(node -v | cut -d'v' -f2 | cut -d'.' -f1)
if [ "$NODE_VERSION" -lt 16 ]; then
    echo "❌ Node.js 版本过低，需要 16+，当前版本: $(node -v)"
    exit 1
fi

echo "✅ Node.js 版本检查通过: $(node -v)"

# 进入前端目录
cd "$(dirname "$0")/../frontend" || {
    echo "❌ 无法进入前端目录"
    exit 1
}

echo "📁 当前目录: $(pwd)"

# 检查是否已安装依赖
if [ ! -d "node_modules" ]; then
    echo "📦 安装前端依赖..."
    npm install || {
        echo "❌ 依赖安装失败"
        exit 1
    }
    echo "✅ 依赖安装完成"
else
    echo "✅ 依赖已存在"
fi

# 检查后端是否运行
echo "🔍 检查后端服务状态..."
if curl -s http://localhost:8080/healthz > /dev/null 2>&1; then
    echo "✅ 后端服务运行正常"
else
    echo "⚠️  后端服务未运行，请先启动后端服务："
    echo "   cd ../scripts && ./start-services.sh"
    echo ""
    echo "🚀 继续启动前端开发服务器..."
fi

# 启动开发服务器
echo "🚀 启动前端开发服务器..."
echo "📍 访问地址: http://localhost:3000"
echo "🔧 开发模式，支持热更新"
echo ""
echo "按 Ctrl+C 停止服务"
echo "=================================="

npm run dev