# PostgreSQL 一键安装配置脚本
# 适用于高考志愿填报系统

param(
    [string]$Method = "docker",  # docker, local, or check
    [string]$Password = "postgres"
)

# 颜色输出函数
function Write-Info($message) {
    Write-Host "ℹ️  $message" -ForegroundColor Cyan
}

function Write-Success($message) {
    Write-Host "✅ $message" -ForegroundColor Green
}

function Write-Warning($message) {
    Write-Host "⚠️  $message" -ForegroundColor Yellow
}

function Write-Error($message) {
    Write-Host "❌ $message" -ForegroundColor Red
}

function Write-Step($step, $message) {
    Write-Host "`n🔸 步骤 $step : $message" -ForegroundColor Blue
}

# 检查系统环境
function Test-Environment {
    Write-Step "1" "检查系统环境"
    
    # 检查 PowerShell 版本
    $psVersion = $PSVersionTable.PSVersion.Major
    if ($psVersion -lt 5) {
        Write-Error "需要 PowerShell 5.0 或更高版本"
        exit 1
    }
    Write-Success "PowerShell 版本: $($PSVersionTable.PSVersion)"
    
    # 检查管理员权限
    $isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")
    if (-not $isAdmin) {
        Write-Warning "建议以管理员身份运行以避免权限问题"
    }
    
    # 检查网络连接
    try {
        Test-NetConnection -ComputerName "www.postgresql.org" -Port 443 -InformationLevel Quiet | Out-Null
        Write-Success "网络连接正常"
    }
    catch {
        Write-Warning "网络连接可能有问题，可能影响下载"
    }
}

# Docker 方式安装
function Install-Docker-PostgreSQL {
    Write-Step "2" "使用 Docker 安装 PostgreSQL"
    
    # 检查 Docker 是否安装
    try {
        $dockerVersion = docker --version
        Write-Success "Docker 已安装: $dockerVersion"
    }
    catch {
        Write-Error "Docker 未安装，请先安装 Docker Desktop"
        Write-Info "下载地址: https://www.docker.com/products/docker-desktop/"
        return $false
    }
    
    # 检查 Docker 是否运行
    try {
        docker ps | Out-Null
        Write-Success "Docker 服务正在运行"
    }
    catch {
        Write-Error "Docker 服务未运行，请启动 Docker Desktop"
        return $false
    }
    
    # 启动 PostgreSQL 容器
    Write-Info "启动 PostgreSQL 容器..."
    
    $projectRoot = Split-Path -Parent $PSScriptRoot
    Set-Location $projectRoot
    
    try {
        # 使用项目的 docker-compose.yml
        docker compose up -d postgres
        
        # 等待容器启动
        Write-Info "等待 PostgreSQL 启动..."
        Start-Sleep -Seconds 15
        
        # 检查容器状态
        $containerStatus = docker compose ps postgres
        if ($containerStatus -match "Up") {
            Write-Success "PostgreSQL 容器启动成功"
            return $true
        }
        else {
            Write-Error "PostgreSQL 容器启动失败"
            docker compose logs postgres
            return $false
        }
    }
    catch {
        Write-Error "启动 PostgreSQL 容器失败: $_"
        return $false
    }
}

# 本地安装 PostgreSQL
function Install-Local-PostgreSQL {
    Write-Step "2" "本地安装 PostgreSQL"
    
    # 检查是否已安装
    $pgPath = Get-Command psql -ErrorAction SilentlyContinue
    if ($pgPath) {
        Write-Success "PostgreSQL 已安装: $($pgPath.Source)"
        return $true
    }
    
    Write-Info "PostgreSQL 未安装，开始下载安装..."
    
    # 下载 PostgreSQL 安装包
    $downloadUrl = "https://get.enterprisedb.com/postgresql/postgresql-15.4-1-windows-x64.exe"
    $installerPath = "$env:TEMP\postgresql-installer.exe"
    
    try {
        Write-Info "下载 PostgreSQL 安装包..."
        Invoke-WebRequest -Uri $downloadUrl -OutFile $installerPath -UseBasicParsing
        Write-Success "下载完成"
    }
    catch {
        Write-Error "下载失败: $_"
        Write-Info "请手动下载: https://www.postgresql.org/download/windows/"
        return $false
    }
    
    # 静默安装
    Write-Info "开始安装 PostgreSQL..."
    try {
        $installArgs = @(
            "--mode", "unattended",
            "--superpassword", $Password,
            "--servicename", "postgresql",
            "--servicepassword", $Password,
            "--serverport", "5432"
        )
        
        Start-Process -FilePath $installerPath -ArgumentList $installArgs -Wait -NoNewWindow
        Write-Success "PostgreSQL 安装完成"
        
        # 添加到 PATH
        $pgBinPath = "C:\Program Files\PostgreSQL\15\bin"
        if (Test-Path $pgBinPath) {
            $env:PATH += ";$pgBinPath"
            Write-Success "已添加到 PATH: $pgBinPath"
        }
        
        return $true
    }
    catch {
        Write-Error "安装失败: $_"
        return $false
    }
    finally {
        # 清理安装包
        if (Test-Path $installerPath) {
            Remove-Item $installerPath -Force
        }
    }
}

# 创建数据库和用户
function Initialize-Database {
    Write-Step "3" "创建数据库和用户"
    
    if ($Method -eq "docker") {
        $connectionString = "docker compose exec postgres psql -U gaokao_user -d gaokao_db"
        $createDbScript = @"
CREATE DATABASE gaokao_data;
CREATE DATABASE gaokao_users;
CREATE DATABASE gaokao_test;
\l
"@
    }
    else {
        $connectionString = "psql -U postgres -h localhost"
        $createDbScript = @"
CREATE USER gaokao_user WITH PASSWORD 'gaokao_pass';
CREATE DATABASE gaokao_data OWNER gaokao_user;
CREATE DATABASE gaokao_users OWNER gaokao_user;
CREATE DATABASE gaokao_test OWNER gaokao_user;
GRANT ALL PRIVILEGES ON DATABASE gaokao_data TO gaokao_user;
GRANT ALL PRIVILEGES ON DATABASE gaokao_users TO gaokao_user;
GRANT ALL PRIVILEGES ON DATABASE gaokao_test TO gaokao_user;
\l
"@
    }
    
    try {
        Write-Info "创建数据库..."
        
        # 将 SQL 脚本写入临时文件
        $sqlFile = "$env:TEMP\create_databases.sql"
        $createDbScript | Out-File -FilePath $sqlFile -Encoding UTF8
        
        if ($Method -eq "docker") {
            # Docker 方式
            docker compose exec -T postgres psql -U gaokao_user -d gaokao_db -f - < $sqlFile
        }
        else {
            # 本地方式
            $env:PGPASSWORD = $Password
            psql -U postgres -h localhost -f $sqlFile
        }
        
        Write-Success "数据库创建完成"
        Remove-Item $sqlFile -Force
        return $true
    }
    catch {
        Write-Error "创建数据库失败: $_"
        return $false
    }
}

# 初始化示例数据
function Initialize-SampleData {
    Write-Step "4" "初始化示例数据"
    
    $projectRoot = Split-Path -Parent $PSScriptRoot
    Set-Location $projectRoot
    
    # 设置环境变量
    if ($Method -eq "docker") {
        $env:DATABASE_URL = "postgres://gaokao_user:gaokao_pass@localhost:5432/gaokao_data?sslmode=disable"
    }
    else {
        $env:DATABASE_URL = "postgres://gaokao_user:gaokao_pass@localhost:5432/gaokao_data?sslmode=disable"
    }
    
    try {
        Write-Info "运行数据初始化脚本..."
        go run scripts/init-sample-data.go
        Write-Success "示例数据初始化完成"
        return $true
    }
    catch {
        Write-Error "数据初始化失败: $_"
        Write-Info "请检查 Go 环境和数据库连接"
        return $false
    }
}

# 验证安装
function Test-Installation {
    Write-Step "5" "验证安装"
    
    try {
        if ($Method -eq "docker") {
            # 测试 Docker 容器
            $containerStatus = docker compose ps postgres
            if ($containerStatus -match "Up") {
                Write-Success "PostgreSQL 容器运行正常"
            }
            else {
                Write-Error "PostgreSQL 容器未运行"
                return $false
            }
            
            # 测试数据库连接
            $testResult = docker compose exec -T postgres psql -U gaokao_user -d gaokao_data -c "SELECT 1;"
            if ($testResult -match "1 row") {
                Write-Success "数据库连接测试成功"
            }
        }
        else {
            # 测试本地安装
            $env:PGPASSWORD = "gaokao_pass"
            $testResult = psql -U gaokao_user -h localhost -d gaokao_data -c "SELECT 1;"
            if ($testResult -match "1 row") {
                Write-Success "数据库连接测试成功"
            }
        }
        
        # 测试数据
        Write-Info "检查示例数据..."
        if ($Method -eq "docker") {
            $dataCount = docker compose exec -T postgres psql -U gaokao_user -d gaokao_data -c "SELECT COUNT(*) FROM universities;"
        }
        else {
            $env:PGPASSWORD = "gaokao_pass"
            $dataCount = psql -U gaokao_user -h localhost -d gaokao_data -c "SELECT COUNT(*) FROM universities;"
        }
        
        if ($dataCount -match "\d+") {
            Write-Success "示例数据验证成功"
        }
        
        return $true
    }
    catch {
        Write-Error "验证失败: $_"
        return $false
    }
}

# 显示连接信息
function Show-ConnectionInfo {
    Write-Host "`n🎉 PostgreSQL 安装配置完成！" -ForegroundColor Green
    Write-Host "===========================================" -ForegroundColor Green
    
    if ($Method -eq "docker") {
        Write-Host "📊 Docker 容器信息:" -ForegroundColor Cyan
        Write-Host "  容器名称: gaokao_postgres" -ForegroundColor White
        Write-Host "  主机: localhost" -ForegroundColor White
        Write-Host "  端口: 5432" -ForegroundColor White
        Write-Host "  用户: gaokao_user" -ForegroundColor White
        Write-Host "  密码: gaokao_pass" -ForegroundColor White
    }
    else {
        Write-Host "💻 本地安装信息:" -ForegroundColor Cyan
        Write-Host "  主机: localhost" -ForegroundColor White
        Write-Host "  端口: 5432" -ForegroundColor White
        Write-Host "  超级用户: postgres" -ForegroundColor White
        Write-Host "  应用用户: gaokao_user" -ForegroundColor White
        Write-Host "  密码: gaokao_pass" -ForegroundColor White
    }
    
    Write-Host "`n🗄️ 数据库列表:" -ForegroundColor Cyan
    Write-Host "  gaokao_data  - 主数据库（高校、专业、录取数据）" -ForegroundColor White
    Write-Host "  gaokao_users - 用户数据库（用户、角色、权限）" -ForegroundColor White
    Write-Host "  gaokao_test  - 测试数据库" -ForegroundColor White
    
    Write-Host "`n🔗 连接字符串:" -ForegroundColor Cyan
    Write-Host "  postgres://gaokao_user:gaokao_pass@localhost:5432/gaokao_data?sslmode=disable" -ForegroundColor Yellow
    
    Write-Host "`n🚀 下一步操作:" -ForegroundColor Cyan
    Write-Host "  1. 编译后端服务: go build -o bin/data-service ./services/data-service" -ForegroundColor White
    Write-Host "  2. 启动服务: ./bin/data-service" -ForegroundColor White
    Write-Host "  3. 测试API: curl http://localhost:8082/v1/universities" -ForegroundColor White
}

# 主函数
function Main {
    Write-Host "🚀 高考志愿填报系统 - PostgreSQL 安装配置脚本" -ForegroundColor Green
    Write-Host "=================================================" -ForegroundColor Green
    
    # 检查参数
    if ($Method -notin @("docker", "local", "check")) {
        Write-Error "无效的安装方式。请使用: docker, local, 或 check"
        Write-Info "用法: .\setup-database.ps1 -Method docker"
        exit 1
    }
    
    Write-Info "安装方式: $Method"
    Write-Info "数据库密码: $Password"
    
    # 执行安装步骤
    if (-not (Test-Environment)) { exit 1 }
    
    if ($Method -eq "check") {
        Test-Installation
        return
    }
    
    $success = $false
    if ($Method -eq "docker") {
        $success = Install-Docker-PostgreSQL
    }
    elseif ($Method -eq "local") {
        $success = Install-Local-PostgreSQL
    }
    
    if (-not $success) { exit 1 }
    
    if (-not (Initialize-Database)) { exit 1 }
    if (-not (Initialize-SampleData)) { exit 1 }
    if (-not (Test-Installation)) { exit 1 }
    
    Show-ConnectionInfo
}

# 运行主函数
Main
