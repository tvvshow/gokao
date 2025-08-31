# 高考志愿填报系统 - 快速构建脚本

param(
    [Parameter(Mandatory=$false)]
    [ValidateSet("all", "backend", "frontend")]
    [string]$Target = "all",
    
    [Parameter(Mandatory=$false)]
    [switch]$SkipTests,
    
    [Parameter(Mandatory=$false)]
    [switch]$Clean
)

$ErrorActionPreference = "Stop"

function Write-Info { param([string]$Message) Write-Host "ℹ️  [INFO] $Message" -ForegroundColor Cyan }
function Write-Success { param([string]$Message) Write-Host "✅ [SUCCESS] $Message" -ForegroundColor Green }
function Write-Warning { param([string]$Message) Write-Host "⚠️  [WARNING] $Message" -ForegroundColor Yellow }
function Write-Error { param([string]$Message) Write-Host "❌ [ERROR] $Message" -ForegroundColor Red }

# 检查依赖
function Test-Dependencies {
    Write-Info "检查构建依赖..."
    
    $dependencies = @("go", "node", "npm")
    $missing = @()
    
    foreach ($dep in $dependencies) {
        if (-not (Get-Command $dep -ErrorAction SilentlyContinue)) {
            $missing += $dep
        }
    }
    
    if ($missing.Count -gt 0) {
        Write-Error "缺少以下依赖: $($missing -join ', ')"
        Write-Info "请先安装这些依赖"
        exit 1
    }
    
    Write-Success "依赖检查通过"
}

# 清理构建产物
function Clear-BuildArtifacts {
    if (-not $Clean) { return }
    
    Write-Info "清理构建产物..."
    
    if (Test-Path "bin") {
        Remove-Item -Recurse -Force "bin"
    }
    
    if (Test-Path "frontend/dist") {
        Remove-Item -Recurse -Force "frontend/dist"
    }
    
    Write-Success "构建产物清理完成"
}

# 构建Go后端服务
function Build-GoServices {
    Write-Info "构建Go后端服务..."
    
    $services = @(
        "api-gateway",
        "user-service", 
        "data-service",
        "payment-service",
        "recommendation-service"
    )
    
    if (-not (Test-Path "bin")) {
        New-Item -ItemType Directory -Force -Path "bin" | Out-Null
    }
    
    foreach ($service in $services) {
        Write-Info "构建 $service..."
        
        $servicePath = "services/$service"
        if (-not (Test-Path "$servicePath/main.go")) {
            Write-Warning "$service 主文件不存在，跳过"
            continue
        }
        
        try {
            Push-Location $servicePath
            
            $outputPath = "../../bin/$service.exe"
            
            $env:CGO_ENABLED = "1"
            
            go build -o $outputPath .
            
            if ($LASTEXITCODE -eq 0) {
                Write-Success "$service 构建完成"
            } else {
                Write-Error "$service 构建失败"
            }
        }
        catch {
            Write-Error "$service 构建异常: $_"
        }
        finally {
            Pop-Location
        }
    }
    
    Write-Success "Go服务构建完成"
}

# 构建前端应用
function Build-Frontend {
    Write-Info "构建前端应用..."
    
    if (-not (Test-Path "frontend/package.json")) {
        Write-Warning "前端项目不存在，跳过"
        return
    }
    
    try {
        Push-Location "frontend"
        
        # 安装依赖
        Write-Info "安装前端依赖..."
        npm ci
        
        # 构建
        Write-Info "构建前端应用..."
        npm run build
        
        Write-Success "前端构建完成"
    }
    catch {
        Write-Error "前端构建失败: $_"
    }
    finally {
        Pop-Location
    }
}

# 运行Go测试
function Test-GoServices {
    if ($SkipTests) { return }
    
    Write-Info "运行Go单元测试..."
    
    try {
        $env:GO_ENV = "test"
        
        go test -v ./...
        
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Go测试通过"
        } else {
            Write-Error "Go测试失败"
        }
    }
    catch {
        Write-Error "Go测试异常: $_"
    }
}

# 运行前端测试
function Test-Frontend {
    if ($SkipTests) { return }
    
    Write-Info "运行前端测试..."
    
    if (-not (Test-Path "frontend/package.json")) {
        Write-Warning "前端项目不存在，跳过测试"
        return
    }
    
    try {
        Push-Location "frontend"
        
        # 单元测试
        Write-Info "运行前端单元测试..."
        npm run test:unit
        
        Write-Success "前端测试通过"
    }
    catch {
        Write-Error "前端测试失败: $_"
    }
    finally {
        Pop-Location
    }
}

# 主函数
function Main {
    Write-Info "🚀 高考志愿填报系统 - 快速构建"
    Write-Info "目标: $Target"
    
    Clear-BuildArtifacts
    Test-Dependencies
    
    # 根据目标进行构建
    switch ($Target) {
        "all" {
            Build-GoServices
            Build-Frontend
            Test-GoServices
            Test-Frontend
        }
        "backend" {
            Build-GoServices
            Test-GoServices
        }
        "frontend" {
            Build-Frontend
            Test-Frontend
        }
    }
    
    Write-Success "🎉 构建完成！"
    
    # 显示构建结果
    Write-Info ""
    Write-Info "📋 构建产物："
    if (Test-Path "bin") {
        Get-ChildItem "bin" | ForEach-Object { Write-Info "  - $($_.Name)" }
    }
    if (Test-Path "frontend/dist") {
        Write-Info "  - frontend/dist/"
    }
    Write-Info ""
    Write-Info "🚀 启动开发环境："
    Write-Info "  - 运行 start-dev.bat 启动所有服务"
    Write-Info "  - 或手动启动各个服务"
}

# 执行主函数
try {
    Main
}
catch {
    Write-Error "构建过程中发生错误: $_"
    exit 1
}
