@echo off
echo 🛑 停止高考志愿填报系统...
echo.

docker-compose down

if %errorlevel% equ 0 (
    echo ✅ 服务已停止
) else (
    echo ❌ 停止服务时出现错误
)

pause