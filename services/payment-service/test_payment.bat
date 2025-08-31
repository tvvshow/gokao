@echo off
echo === 支付服务测试 ===

REM 设置环境变量
for /f "usebackq tokens=*" %%i in (`.env.test`) do (
    for /f "tokens=1,* delims==" %%a in ("%%i") do (
        if not "%%a"=="" if not "%%a"=="#" set "%%a=%%b"
    )
)

echo 构建支付服务...
go build -o payment-service-test.exe main.go

if %errorlevel% neq 0 (
    echo 构建失败
    pause
    exit /b 1
)

echo 支付服务构建成功

echo 启动支付服务...
start /b payment-service-test.exe

echo 等待服务启动...
timeout /t 3 /nobreak >nul

echo 测试健康检查...
curl -f http://localhost:8084/health

if %errorlevel% equ 0 (
    echo ✓ 健康检查通过
) else (
    echo ✗ 健康检查失败
    pause
    exit /b 1
)

echo 测试支付渠道接口...
curl -f http://localhost:8084/api/v1/payments/channels

if %errorlevel% equ 0 (
    echo ✓ 支付渠道接口正常
) else (
    echo ✗ 支付渠道接口异常
)

echo.
echo 支付服务测试完成
echo 按任意键退出...
pause >nul