@echo off
echo 🚀 启动高考志愿填报系统...
echo.

REM 检查Docker是否运行
docker --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Docker未安装或未启动，请先安装并启动Docker
    pause
    exit /b 1
)

echo 📦 启动所有服务...
docker-compose up -d

if %errorlevel% equ 0 (
    echo.
    echo ✅ 服务启动完成!
    echo 📖 API文档: http://localhost/docs
    echo 🔍 健康检查: http://localhost/health
    echo 📊 服务状态: docker-compose ps
    echo.
    echo 💡 其他命令:
    echo   查看日志: docker-compose logs -f
    echo   停止服务: docker-compose down
    echo   重启服务: docker-compose restart
) else (
    echo ❌ 服务启动失败，请检查错误信息
)

pause