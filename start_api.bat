@echo off
echo 🚀 启动高考志愿填报系统 API...
echo.

REM 检查Python是否安装
python --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Python未安装，请先安装Python 3.11+
    pause
    exit /b 1
)

echo 📦 安装基础依赖...
pip install fastapi uvicorn

if %errorlevel% neq 0 (
    echo ❌ 依赖安装失败
    pause
    exit /b 1
)

echo 🌐 启动API服务器...
echo 📖 API文档将在 http://localhost:8000/docs 可用
echo 🔍 健康检查: http://localhost:8000/health
echo 按 Ctrl+C 停止服务器
echo.

cd backend
python simple_main.py