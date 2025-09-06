# 高考志愿填报系统 - 简化Windows开发环境设置脚本

param(
    [Parameter(Mandatory=$false)]
    [switch]$SkipDependencies,
    
    [Parameter(Mandatory=$false)]
    [switch]$Force
)

$ErrorActionPreference = "Stop"

function Write-Info { param([string]$Message) Write-Host "ℹ️  [INFO] $Message" -ForegroundColor Cyan }
function Write-Success { param([string]$Message) Write-Host "✅ [SUCCESS] $Message" -ForegroundColor Green }
function Write-Warning { param([string]$Message) Write-Host "⚠️  [WARNING] $Message" -ForegroundColor Yellow }
function Write-Error { param([string]$Message) Write-Host "❌ [ERROR] $Message" -ForegroundColor Red }

# 检查管理员权限
function Test-Administrator {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

# 设置项目环境
function Setup-ProjectEnvironment {
    Write-Info "设置项目开发环境..."
    
    # 检查项目目录
    if (-not (Test-Path "go.mod")) {
        Write-Error "请在项目根目录运行此脚本"
        exit 1
    }
    
    # 创建环境配置文件
    $envContent = @"
# 高考志愿填报系统 - 开发环境配置

# 数据库配置
DATABASE_URL=postgres://postgres:postgres@localhost:5432/gaokao_dev?sslmode=disable
DATABASE_TEST_URL=postgres://postgres:postgres@localhost:5432/gaokao_test?sslmode=disable

# Redis配置
REDIS_URL=redis://localhost:6379

# JWT密钥
JWT_SECRET=your-super-secret-jwt-key-for-development

# 服务端口
API_GATEWAY_PORT=8080
USER_SERVICE_PORT=8081
DATA_SERVICE_PORT=8082
PAYMENT_SERVICE_PORT=8083
RECOMMENDATION_SERVICE_PORT=8084

# 前端配置
VITE_API_BASE_URL=http://localhost:8080
VITE_APP_TITLE=高考志愿填报助手

# 开发模式
NODE_ENV=development
GO_ENV=development
"@
    
    if (-not (Test-Path ".env") -or $Force) {
        $envContent | Out-File -FilePath ".env" -Encoding UTF8
        Write-Success "环境配置文件 .env 创建完成"
    }
    
    # 安装Go依赖
    Write-Info "安装Go依赖..."
    try {
        go mod download
        go mod tidy
        Write-Success "Go依赖安装完成"
    }
    catch {
        Write-Warning "Go依赖安装失败: $_"
    }
    
    # 安装前端依赖
    if (Test-Path "frontend/package.json") {
        Write-Info "安装前端依赖..."
        try {
            Push-Location "frontend"
            npm install
            Write-Success "前端依赖安装完成"
        }
        catch {
            Write-Warning "前端依赖安装失败: $_"
        }
        finally {
            Pop-Location
        }
    }
    
    Write-Success "项目环境设置完成"
}

# 创建开发脚本
function Create-DevScripts {
    Write-Info "创建开发脚本..."
    
    # 创建启动脚本
    $startScript = @'
@echo off
echo 🚀 启动高考志愿填报系统开发环境...

echo 📊 启动数据库服务...
net start postgresql-x64-15 2>nul
net start Redis 2>nul

echo 🔧 启动后端服务...
start "API Gateway" cmd /k "cd services\api-gateway && go run main.go"
timeout /t 2 /nobreak >nul
start "User Service" cmd /k "cd services\user-service && go run main.go"
timeout /t 2 /nobreak >nul
start "Data Service" cmd /k "cd services\data-service && go run main.go"
timeout /t 2 /nobreak >nul
start "Payment Service" cmd /k "cd services\payment-service && go run main.go"
timeout /t 2 /nobreak >nul
start "Recommendation Service" cmd /k "cd services\recommendation-service && go run main.go"

echo 🎨 启动前端服务...
timeout /t 5 /nobreak >nul
start "Frontend" cmd /k "cd frontend && npm run dev"

echo ✅ 开发环境启动完成！
echo 🌐 前端地址: http://localhost:5173
echo 🔌 API地址: http://localhost:8080
pause
'@
    
    $startScript | Out-File -FilePath "start-dev.bat" -Encoding UTF8
    
    # 创建停止脚本
    $stopScript = @'
@echo off
echo 🛑 停止高考志愿填报系统开发环境...

echo 关闭Node.js进程...
taskkill /f /im node.exe 2>nul

echo 关闭Go进程...
taskkill /f /im main.exe 2>nul

echo ✅ 开发环境已停止
pause
'@
    
    $stopScript | Out-File -FilePath "stop-dev.bat" -Encoding UTF8
    
    Write-Success "开发脚本创建完成"
}

# 运行测试
function Run-Tests {
    Write-Info "运行测试..."
    
    # Go测试
    Write-Info "运行Go单元测试..."
    try {
        go test -v ./...
        Write-Success "Go测试通过"
    }
    catch {
        Write-Warning "Go测试失败: $_"
    }
    
    # 前端测试
    if (Test-Path "frontend/package.json") {
        Write-Info "运行前端测试..."
        try {
            Push-Location "frontend"
            npm run test:unit
            Write-Success "前端测试通过"
        }
        catch {
            Write-Warning "前端测试失败: $_"
        }
        finally {
            Pop-Location
        }
    }
}

# 主函数
function Main {
    Write-Info "🚀 高考志愿填报系统 - Windows开发环境设置"
    
    # 检查基本依赖
    $dependencies = @("go", "node", "npm")
    $missing = @()
    
    foreach ($dep in $dependencies) {
        if (-not (Get-Command $dep -ErrorAction SilentlyContinue)) {
            $missing += $dep
        }
    }
    
    if ($missing.Count -gt 0) {
        Write-Warning "缺少以下依赖: $($missing -join ', ')"
        Write-Info "请先安装这些依赖，或运行 scripts/install-deps.ps1"
    }
    
    # 设置项目环境
    Setup-ProjectEnvironment
    Create-DevScripts
    
    Write-Success "🎉 开发环境设置完成！"
    Write-Info ""
    Write-Info "📋 下一步操作："
    Write-Info "1. 运行 start-dev.bat 启动开发环境"
    Write-Info "2. 访问 http://localhost:5173 查看前端"
    Write-Info "3. 访问 http://localhost:8080 查看API"
    Write-Info "4. 运行 stop-dev.bat 停止开发环境"
    Write-Info ""
    Write-Info "🔧 使用Makefile构建："
    Write-Info "- make help     查看所有可用命令"
    Write-Info "- make build    构建所有组件"
    Write-Info "- make test     运行所有测试"
    Write-Info "- make clean    清理构建产物"
    Write-Info ""
}

# 执行主函数
try {
    Main
}
catch {
    Write-Error "设置过程中发生错误: $_"
    exit 1
}
