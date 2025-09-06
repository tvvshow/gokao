@echo off
chcp 65001 >nul
title 高考志愿填报系统 - 快速启动

echo.
echo 🚀 高考志愿填报系统 - 快速启动脚本
echo =====================================
echo.

:: 检查是否在项目根目录
if not exist "go.mod" (
    echo ❌ 请在项目根目录运行此脚本
    pause
    exit /b 1
)

:: 设置颜色
for /f %%A in ('echo prompt $E ^| cmd') do set "ESC=%%A"
set "GREEN=%ESC%[32m"
set "YELLOW=%ESC%[33m"
set "BLUE=%ESC%[34m"
set "RED=%ESC%[31m"
set "NC=%ESC%[0m"

echo %BLUE%选择数据库安装方式:%NC%
echo 1. Docker 方式 (推荐)
echo 2. 本地安装方式
echo 3. 跳过数据库安装 (已安装)
echo 4. 仅检查环境
echo.
set /p choice="请选择 (1-4): "

if "%choice%"=="1" goto docker_install
if "%choice%"=="2" goto local_install
if "%choice%"=="3" goto skip_db
if "%choice%"=="4" goto check_env
goto invalid_choice

:docker_install
echo.
echo %BLUE%🐳 使用 Docker 安装 PostgreSQL...%NC%
echo.

:: 检查 Docker
docker --version >nul 2>&1
if errorlevel 1 (
    echo %RED%❌ Docker 未安装或未启动%NC%
    echo %YELLOW%请先安装 Docker Desktop: https://www.docker.com/products/docker-desktop/%NC%
    pause
    exit /b 1
)

echo %GREEN%✅ Docker 已安装%NC%

:: 启动 PostgreSQL 容器
echo %BLUE%启动 PostgreSQL 容器...%NC%
docker-compose up -d postgres

if errorlevel 1 (
    echo %RED%❌ 启动 PostgreSQL 容器失败%NC%
    pause
    exit /b 1
)

echo %GREEN%✅ PostgreSQL 容器启动成功%NC%

:: 等待容器启动
echo %YELLOW%等待数据库启动 (15秒)...%NC%
timeout /t 15 /nobreak >nul

:: 创建数据库
echo %BLUE%创建数据库...%NC%
docker-compose exec -T postgres psql -U gaokao_user -d gaokao_db -c "CREATE DATABASE IF NOT EXISTS gaokao_data;"
docker-compose exec -T postgres psql -U gaokao_user -d gaokao_db -c "CREATE DATABASE IF NOT EXISTS gaokao_users;"

goto init_data

:local_install
echo.
echo %BLUE%💻 本地安装 PostgreSQL...%NC%
echo.

:: 检查是否已安装
psql --version >nul 2>&1
if not errorlevel 1 (
    echo %GREEN%✅ PostgreSQL 已安装%NC%
    goto create_local_db
)

echo %YELLOW%PostgreSQL 未安装，请手动安装:%NC%
echo 1. 访问: https://www.postgresql.org/download/windows/
echo 2. 下载并安装 PostgreSQL 15+
echo 3. 设置超级用户密码为: postgres
echo 4. 安装完成后重新运行此脚本
echo.
pause
exit /b 1

:create_local_db
echo %BLUE%创建数据库和用户...%NC%

:: 设置密码环境变量
set PGPASSWORD=postgres

:: 创建用户和数据库
psql -U postgres -h localhost -c "CREATE USER IF NOT EXISTS gaokao_user WITH PASSWORD 'gaokao_pass';"
psql -U postgres -h localhost -c "CREATE DATABASE gaokao_data OWNER gaokao_user;"
psql -U postgres -h localhost -c "CREATE DATABASE gaokao_users OWNER gaokao_user;"
psql -U postgres -h localhost -c "GRANT ALL PRIVILEGES ON DATABASE gaokao_data TO gaokao_user;"
psql -U postgres -h localhost -c "GRANT ALL PRIVILEGES ON DATABASE gaokao_users TO gaokao_user;"

if errorlevel 1 (
    echo %RED%❌ 创建数据库失败%NC%
    echo %YELLOW%请检查 PostgreSQL 服务是否运行%NC%
    pause
    exit /b 1
)

echo %GREEN%✅ 数据库创建成功%NC%
goto init_data

:skip_db
echo.
echo %YELLOW%⏭️ 跳过数据库安装%NC%
goto init_data

:check_env
echo.
echo %BLUE%🔍 检查环境...%NC%
echo.

:: 检查 Go
go version >nul 2>&1
if errorlevel 1 (
    echo %RED%❌ Go 未安装%NC%
) else (
    echo %GREEN%✅ Go 已安装%NC%
    go version
)

:: 检查 Node.js
node --version >nul 2>&1
if errorlevel 1 (
    echo %RED%❌ Node.js 未安装%NC%
) else (
    echo %GREEN%✅ Node.js 已安装%NC%
    node --version
)

:: 检查 Docker
docker --version >nul 2>&1
if errorlevel 1 (
    echo %RED%❌ Docker 未安装%NC%
) else (
    echo %GREEN%✅ Docker 已安装%NC%
    docker --version
)

:: 检查 PostgreSQL
psql --version >nul 2>&1
if errorlevel 1 (
    echo %YELLOW%⚠️ PostgreSQL 未安装 (可使用 Docker)%NC%
) else (
    echo %GREEN%✅ PostgreSQL 已安装%NC%
    psql --version
)

echo.
pause
exit /b 0

:init_data
echo.
echo %BLUE%🗄️ 初始化示例数据...%NC%

:: 设置环境变量
set DATABASE_URL=postgres://gaokao_user:gaokao_pass@localhost:5432/gaokao_data?sslmode=disable

:: 运行数据初始化
go run scripts/init-sample-data.go

if errorlevel 1 (
    echo %RED%❌ 数据初始化失败%NC%
    echo %YELLOW%请检查数据库连接和 Go 环境%NC%
    pause
    exit /b 1
)

echo %GREEN%✅ 示例数据初始化完成%NC%

:build_services
echo.
echo %BLUE%🔧 编译后端服务...%NC%

:: 创建 bin 目录
if not exist "bin" mkdir bin

:: 编译 data-service
echo 编译 data-service...
go build -o bin/data-service.exe ./services/data-service
if errorlevel 1 (
    echo %RED%❌ data-service 编译失败%NC%
    pause
    exit /b 1
)
echo %GREEN%✅ data-service 编译成功%NC%

:: 编译 api-gateway
echo 编译 api-gateway...
go build -o bin/api-gateway.exe ./services/api-gateway
if errorlevel 1 (
    echo %RED%❌ api-gateway 编译失败%NC%
    pause
    exit /b 1
)
echo %GREEN%✅ api-gateway 编译成功%NC%

:start_services
echo.
echo %BLUE%🚀 启动服务...%NC%

:: 启动 data-service
echo 启动 data-service (端口 8082)...
start "Data Service" bin/data-service.exe

:: 等待服务启动
timeout /t 3 /nobreak >nul

:: 启动 api-gateway
echo 启动 api-gateway (端口 8080)...
start "API Gateway" bin/api-gateway.exe

:: 等待服务启动
timeout /t 3 /nobreak >nul

:test_api
echo.
echo %BLUE%🧪 测试 API...%NC%

:: 测试健康检查
curl -s http://localhost:8080/healthz >nul 2>&1
if errorlevel 1 (
    echo %YELLOW%⚠️ API 可能还在启动中...%NC%
) else (
    echo %GREEN%✅ API Gateway 响应正常%NC%
)

:: 测试数据 API
curl -s http://localhost:8080/v1/universities >nul 2>&1
if errorlevel 1 (
    echo %YELLOW%⚠️ 数据 API 可能还在启动中...%NC%
) else (
    echo %GREEN%✅ 数据 API 响应正常%NC%
)

:start_frontend
echo.
echo %BLUE%🌐 启动前端 (可选)...%NC%
set /p start_frontend="是否启动前端? (y/n): "

if /i "%start_frontend%"=="y" (
    echo 检查前端依赖...
    cd frontend
    
    if not exist "node_modules" (
        echo 安装前端依赖...
        npm install
    )
    
    echo 启动前端开发服务器...
    start "Frontend" npm run dev
    
    cd ..
    echo %GREEN%✅ 前端启动中...%NC%
)

:success
echo.
echo %GREEN%🎉 系统启动完成！%NC%
echo ==========================================
echo.
echo %BLUE%📋 服务信息:%NC%
echo   🔧 API Gateway: http://localhost:8080
echo   📊 Data Service: http://localhost:8082
if /i "%start_frontend%"=="y" (
    echo   🌐 前端界面: http://localhost:3000
)
echo.
echo %BLUE%🧪 测试命令:%NC%
echo   curl http://localhost:8080/healthz
echo   curl http://localhost:8080/v1/universities
echo   curl http://localhost:8080/v1/universities/statistics
echo.
echo %BLUE%📚 API 文档:%NC%
echo   http://localhost:8080/swagger/index.html
echo.
echo %YELLOW%💡 提示:%NC%
echo   - 服务已在后台运行
echo   - 关闭命令行窗口将停止服务
echo   - 数据库数据已初始化完成
echo.
pause
goto :eof

:invalid_choice
echo %RED%❌ 无效选择，请重新运行脚本%NC%
pause
exit /b 1
