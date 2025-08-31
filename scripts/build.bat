@echo off
chcp 65001 >nul
echo 🚀 Gaokao System - Build Script

REM Check Go installation
where go >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Go not installed, please install Go first
    pause
    exit /b 1
)

REM Check Node.js installation
where node >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Node.js not installed, please install Node.js first
    pause
    exit /b 1
)

echo ✅ Dependencies check passed

REM Create bin directory
if not exist bin mkdir bin

echo 🔨 Building Go backend services...

REM Build API Gateway
echo   Building api-gateway...
cd services\api-gateway
go build -o ..\..\bin\api-gateway.exe .
if %errorlevel% neq 0 (
    echo ❌ api-gateway build failed
    cd ..\..
    pause
    exit /b 1
)
cd ..\..

REM Build User Service
echo   Building user-service...
cd services\user-service
go build -o ..\..\bin\user-service.exe .
if %errorlevel% neq 0 (
    echo ❌ user-service build failed
    cd ..\..
    pause
    exit /b 1
)
cd ..\..

REM Build Data Service
echo   Building data-service...
cd services\data-service
go build -o ..\..\bin\data-service.exe .
if %errorlevel% neq 0 (
    echo ❌ data-service build failed
    cd ..\..
    pause
    exit /b 1
)
cd ..\..

REM Build Payment Service
echo   Building payment-service...
cd services\payment-service
go build -o ..\..\bin\payment-service.exe .
if %errorlevel% neq 0 (
    echo ❌ payment-service build failed
    cd ..\..
    pause
    exit /b 1
)
cd ..\..

REM Build Recommendation Service
echo   Building recommendation-service...
cd services\recommendation-service
go build -o ..\..\bin\recommendation-service.exe .
if %errorlevel% neq 0 (
    echo ❌ recommendation-service build failed
    cd ..\..
    pause
    exit /b 1
)
cd ..\..

echo ✅ Go services build completed

echo 🎨 Building frontend application...
cd frontend
call npm ci
if %errorlevel% neq 0 (
    echo ❌ Frontend dependencies installation failed
    cd ..
    pause
    exit /b 1
)

call npm run build
if %errorlevel% neq 0 (
    echo ❌ Frontend build failed
    cd ..
    pause
    exit /b 1
)
cd ..

echo ✅ Frontend build completed

echo 🎉 Build completed successfully!

echo.
echo 📋 Build artifacts:
dir /b bin
echo   - frontend\dist\

echo.
echo 🚀 Start development environment:
echo   - Run start-dev.bat to start all services
echo   - Or manually start individual services

pause
