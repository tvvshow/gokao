@echo off
echo 🚀 高考志愿填报系统 - 简单构建脚本

REM 检查Go是否安装
where go >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Go未安装，请先安装Go语言
    pause
    exit /b 1
)

REM 检查Node.js是否安装
where node >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Node.js未安装，请先安装Node.js
    pause
    exit /b 1
)

echo ✅ 依赖检查通过

REM 创建bin目录
if not exist bin mkdir bin

echo 🔨 构建Go后端服务...

REM 构建API Gateway
echo   构建 api-gateway...
cd services\api-gateway
go build -o ..\..\bin\api-gateway.exe .
if %errorlevel% neq 0 (
    echo ❌ api-gateway构建失败
    cd ..\..
    pause
    exit /b 1
)
cd ..\..

REM 构建User Service
echo   构建 user-service...
cd services\user-service
go build -o ..\..\bin\user-service.exe .
if %errorlevel% neq 0 (
    echo ❌ user-service构建失败
    cd ..\..
    pause
    exit /b 1
)
cd ..\..

REM 构建Data Service
echo   构建 data-service...
cd services\data-service
go build -o ..\..\bin\data-service.exe .
if %errorlevel% neq 0 (
    echo ❌ data-service构建失败
    cd ..\..
    pause
    exit /b 1
)
cd ..\..

REM 构建Payment Service
echo   构建 payment-service...
cd services\payment-service
go build -o ..\..\bin\payment-service.exe .
if %errorlevel% neq 0 (
    echo ❌ payment-service构建失败
    cd ..\..
    pause
    exit /b 1
)
cd ..\..

REM 构建Recommendation Service
echo   构建 recommendation-service...
cd services\recommendation-service
go build -o ..\..\bin\recommendation-service.exe .
if %errorlevel% neq 0 (
    echo ❌ recommendation-service构建失败
    cd ..\..
    pause
    exit /b 1
)
cd ..\..

echo ✅ Go服务构建完成

echo 🎨 构建前端应用...
cd frontend
call npm ci
if %errorlevel% neq 0 (
    echo ❌ 前端依赖安装失败
    cd ..
    pause
    exit /b 1
)

call npm run build
if %errorlevel% neq 0 (
    echo ❌ 前端构建失败
    cd ..
    pause
    exit /b 1
)
cd ..

echo ✅ 前端构建完成

echo 🎉 构建完成！

echo.
echo 📋 构建产物：
dir /b bin
echo   - frontend\dist\

echo.
echo 🚀 启动开发环境：
echo   - 运行 start-dev.bat 启动所有服务
echo   - 或手动启动各个服务

pause
