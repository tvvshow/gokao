# 高考志愿填报系统 - Windows开发环境设置脚本
# PowerShell脚本，支持本机Windows和WSL环境

param(
    [Parameter(Mandatory=$false)]
    [ValidateSet("windows", "wsl", "both")]
    [string]$Environment = "both",
    
    [Parameter(Mandatory=$false)]
    [switch]$SkipDependencies,
    
    [Parameter(Mandatory=$false)]
    [switch]$Force,
    
    [Parameter(Mandatory=$false)]
    [switch]$Verbose
)

# 设置错误处理
$ErrorActionPreference = "Stop"

# 颜色输出函数
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

function Write-Info { param([string]$Message) Write-ColorOutput "ℹ️  [INFO] $Message" "Cyan" }
function Write-Success { param([string]$Message) Write-ColorOutput "✅ [SUCCESS] $Message" "Green" }
function Write-Warning { param([string]$Message) Write-ColorOutput "⚠️  [WARNING] $Message" "Yellow" }
function Write-Error { param([string]$Message) Write-ColorOutput "❌ [ERROR] $Message" "Red" }

# 检查管理员权限
function Test-Administrator {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

# 检查WSL是否可用
function Test-WSLAvailable {
    try {
        $wslOutput = wsl --list --quiet 2>$null
        return $LASTEXITCODE -eq 0
    }
    catch {
        return $false
    }
}

# 安装Chocolatey
function Install-Chocolatey {
    if (Get-Command choco -ErrorAction SilentlyContinue) {
        Write-Info "Chocolatey 已安装"
        return
    }
    
    Write-Info "安装 Chocolatey..."
    Set-ExecutionPolicy Bypass -Scope Process -Force
    [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072
    iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))
    Write-Success "Chocolatey 安装完成"
}

# 安装Windows依赖
function Install-WindowsDependencies {
    Write-Info "安装Windows开发依赖..."
    
    if (-not $SkipDependencies) {
        Install-Chocolatey
        
        # 安装基础工具
        $tools = @(
            "git",
            "nodejs",
            "golang",
            "docker-desktop",
            "vscode",
            "postman",
            "redis-64"
        )
        
        foreach ($tool in $tools) {
            Write-Info "安装 $tool..."
            try {
                choco install $tool -y --force:$Force
                Write-Success "$tool 安装完成"
            }
            catch {
                Write-Warning "$tool 安装失败: $_"
            }
        }
        
        # 安装PostgreSQL
        Write-Info "安装 PostgreSQL..."
        try {
            choco install postgresql15 --params '/Password:postgres' -y --force:$Force
            Write-Success "PostgreSQL 安装完成"
        }
        catch {
            Write-Warning "PostgreSQL 安装失败: $_"
        }
    }
    
    Write-Success "Windows依赖安装完成"
}

# 设置WSL环境
function Setup-WSLEnvironment {
    Write-Info "设置WSL开发环境..."
    
    if (-not (Test-WSLAvailable)) {
        Write-Warning "WSL不可用，跳过WSL环境设置"
        return
    }
    
    # 在WSL中安装依赖
    $wslScript = @'
#!/bin/bash
set -e

echo "🔧 设置WSL开发环境..."

# 更新包管理器
sudo apt update

# 安装基础依赖
sudo apt install -y curl wget git build-essential

# 安装Go
if ! command -v go &> /dev/null; then
    echo "📦 安装Go..."
    wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
    sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    rm go1.21.5.linux-amd64.tar.gz
fi

# 安装Node.js
if ! command -v node &> /dev/null; then
    echo "📦 安装Node.js..."
    curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
    sudo apt-get install -y nodejs
fi

# 安装Docker
if ! command -v docker &> /dev/null; then
    echo "📦 安装Docker..."
    curl -fsSL https://get.docker.com -o get-docker.sh
    sudo sh get-docker.sh
    sudo usermod -aG docker $USER
    rm get-docker.sh
fi

# 安装PostgreSQL
if ! command -v psql &> /dev/null; then
    echo "📦 安装PostgreSQL..."
    sudo apt install -y postgresql postgresql-contrib
    sudo systemctl start postgresql
    sudo systemctl enable postgresql
    sudo -u postgres psql -c "ALTER USER postgres PASSWORD 'postgres';"
fi

# 安装Redis
if ! command -v redis-server &> /dev/null; then
    echo "📦 安装Redis..."
    sudo apt install -y redis-server
    sudo systemctl start redis-server
    sudo systemctl enable redis-server
fi

echo "✅ WSL环境设置完成"
'@
    
    # 将脚本写入临时文件并执行
    $tempScript = [System.IO.Path]::GetTempFileName() + ".sh"
    $wslScript | Out-File -FilePath $tempScript -Encoding UTF8
    
    try {
        wsl bash $tempScript
        Write-Success "WSL环境设置完成"
    }
    catch {
        Write-Error "WSL环境设置失败: $_"
    }
    finally {
        Remove-Item $tempScript -ErrorAction SilentlyContinue
    }
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

# 初始化数据库
function Initialize-Database {
    Write-Info "初始化数据库..."
    
    try {
        # 创建开发数据库
        $createDbScript = @"
CREATE DATABASE gaokao_dev;
CREATE DATABASE gaokao_test;
"@
        
        # 尝试连接PostgreSQL并创建数据库
        $env:PGPASSWORD = "postgres"
        echo $createDbScript | psql -h localhost -U postgres -d postgres
        
        Write-Success "数据库初始化完成"
    }
    catch {
        Write-Warning "数据库初始化失败，请手动创建数据库: $_"
    }
}

# 创建开发脚本
function Create-DevScripts {
    Write-Info "创建开发脚本..."
    
    # 创建启动脚本
    $startScript = @"
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
"@
    
    $startScript | Out-File -FilePath "start-dev.bat" -Encoding UTF8
    
    # 创建停止脚本
    $stopScript = @"
@echo off
echo 🛑 停止高考志愿填报系统开发环境...

echo 关闭Node.js进程...
taskkill /f /im node.exe 2>nul

echo 关闭Go进程...
taskkill /f /im main.exe 2>nul

echo ✅ 开发环境已停止
pause
"@
    
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
    Write-Info "环境: $Environment"
    
    # 检查管理员权限
    if (-not (Test-Administrator) -and -not $SkipDependencies) {
        Write-Warning "建议以管理员身份运行以安装系统依赖"
    }
    
    # 根据选择的环境进行设置
    switch ($Environment) {
        "windows" {
            Install-WindowsDependencies
        }
        "wsl" {
            Setup-WSLEnvironment
        }
        "both" {
            Install-WindowsDependencies
            Setup-WSLEnvironment
        }
    }
    
    # 设置项目环境
    Setup-ProjectEnvironment
    Initialize-Database
    Create-DevScripts
    
    # 运行测试
    if ($Verbose) {
        Run-Tests
    }
    
    Write-Success "🎉 开发环境设置完成！"
    Write-Info ""
    Write-Info "📋 下一步操作："
    Write-Info "1. 运行 start-dev.bat 启动开发环境"
    Write-Info "2. 访问 http://localhost:5173 查看前端"
    Write-Info "3. 访问 http://localhost:8080 查看API"
    Write-Info "4. 运行 stop-dev.bat 停止开发环境"
    Write-Info ""
    Write-Info "🔧 开发工具："
    Write-Info "- VS Code: 代码编辑"
    Write-Info "- Postman: API测试"
    Write-Info "- pgAdmin: 数据库管理"
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
