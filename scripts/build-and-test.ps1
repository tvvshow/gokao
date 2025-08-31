# 高考志愿填报系统 - 构建和测试脚本
# 支持Windows本机和WSL环境的编译测试

param(
    [Parameter(Mandatory=$false)]
    [ValidateSet("windows", "wsl", "both")]
    [string]$Environment = "windows",
    
    [Parameter(Mandatory=$false)]
    [ValidateSet("all", "backend", "frontend", "cpp")]
    [string]$Target = "all",
    
    [Parameter(Mandatory=$false)]
    [switch]$SkipTests,
    
    [Parameter(Mandatory=$false)]
    [switch]$Release,
    
    [Parameter(Mandatory=$false)]
    [switch]$Verbose,
    
    [Parameter(Mandatory=$false)]
    [switch]$Clean
)

$ErrorActionPreference = "Stop"

# 颜色输出函数
function Write-ColorOutput {
    param([string]$Message, [string]$Color = "White")
    Write-Host $Message -ForegroundColor $Color
}

function Write-Info { param([string]$Message) Write-ColorOutput "ℹ️  [INFO] $Message" "Cyan" }
function Write-Success { param([string]$Message) Write-ColorOutput "✅ [SUCCESS] $Message" "Green" }
function Write-Warning { param([string]$Message) Write-ColorOutput "⚠️  [WARNING] $Message" "Yellow" }
function Write-Error { param([string]$Message) Write-ColorOutput "❌ [ERROR] $Message" "Red" }

# 检查依赖
function Test-Dependencies {
    Write-Info "检查构建依赖..."
    
    $dependencies = @{
        "go" = "Go语言编译器"
        "node" = "Node.js运行时"
        "npm" = "NPM包管理器"
    }
    
    if ($Target -eq "cpp" -or $Target -eq "all") {
        $dependencies["gcc"] = "GCC编译器"
        $dependencies["cmake"] = "CMake构建工具"
    }
    
    $missing = @()
    foreach ($dep in $dependencies.Keys) {
        if (-not (Get-Command $dep -ErrorAction SilentlyContinue)) {
            $missing += "$dep ($($dependencies[$dep]))"
        }
    }
    
    if ($missing.Count -gt 0) {
        Write-Error "缺少以下依赖: $($missing -join ', ')"
        Write-Info "请运行 scripts/dev-setup.ps1 安装依赖"
        exit 1
    }
    
    Write-Success "依赖检查通过"
}

# 清理构建产物
function Clear-BuildArtifacts {
    if (-not $Clean) { return }
    
    Write-Info "清理构建产物..."
    
    # 清理Go构建产物
    if (Test-Path "bin") {
        Remove-Item -Recurse -Force "bin"
    }
    
    # 清理前端构建产物
    if (Test-Path "frontend/dist") {
        Remove-Item -Recurse -Force "frontend/dist"
    }
    if (Test-Path "frontend/node_modules") {
        Remove-Item -Recurse -Force "frontend/node_modules"
    }
    
    # 清理C++构建产物
    if (Test-Path "cpp-modules/build") {
        Remove-Item -Recurse -Force "cpp-modules/build"
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
    
    $buildMode = if ($Release) { "release" } else { "debug" }
    $ldflags = if ($Release) { "-s -w" } else { "" }
    
    New-Item -ItemType Directory -Force -Path "bin" | Out-Null
    
    foreach ($service in $services) {
        Write-Info "构建 $service..."
        
        $servicePath = "services/$service"
        if (-not (Test-Path "$servicePath/main.go")) {
            Write-Warning "$service 主文件不存在，跳过"
            continue
        }
        
        try {
            Push-Location $servicePath
            
            $outputPath = "../../bin/$service"
            if ($IsWindows) {
                $outputPath += ".exe"
            }
            
            $env:CGO_ENABLED = "1"
            if ($Release) {
                $env:CGO_ENABLED = "0"
            }
            
            $buildArgs = @(
                "build",
                "-o", $outputPath
            )
            
            if ($ldflags) {
                $buildArgs += "-ldflags", $ldflags
            }
            
            if ($Verbose) {
                $buildArgs += "-v"
            }
            
            $buildArgs += "."
            
            & go @buildArgs
            
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
        
        # 类型检查
        Write-Info "执行TypeScript类型检查..."
        npm run type-check
        
        # 代码检查
        Write-Info "执行ESLint代码检查..."
        npm run lint
        
        # 构建
        Write-Info "构建前端应用..."
        if ($Release) {
            npm run build
        } else {
            npm run build:dev
        }
        
        Write-Success "前端构建完成"
    }
    catch {
        Write-Error "前端构建失败: $_"
    }
    finally {
        Pop-Location
    }
}

# 构建C++模块
function Build-CppModules {
    Write-Info "构建C++模块..."
    
    $modules = @(
        "device-fingerprint",
        "license", 
        "volunteer-matcher"
    )
    
    foreach ($module in $modules) {
        $modulePath = "cpp-modules/$module"
        if (-not (Test-Path "$modulePath/CMakeLists.txt")) {
            Write-Warning "$module CMakeLists.txt不存在，跳过"
            continue
        }
        
        Write-Info "构建 $module..."
        
        try {
            Push-Location $modulePath
            
            # 创建构建目录
            New-Item -ItemType Directory -Force -Path "build" | Out-Null
            Push-Location "build"
            
            # CMake配置
            $cmakeArgs = @(
                "..",
                "-DCMAKE_BUILD_TYPE=$(if ($Release) { 'Release' } else { 'Debug' })"
            )
            
            if ($IsWindows) {
                $cmakeArgs += "-G", "MinGW Makefiles"
            }
            
            & cmake @cmakeArgs
            
            # 构建
            $buildArgs = @("--build", ".")
            if ($Verbose) {
                $buildArgs += "--verbose"
            }
            
            & cmake @buildArgs
            
            if ($LASTEXITCODE -eq 0) {
                Write-Success "$module 构建完成"
            } else {
                Write-Error "$module 构建失败"
            }
        }
        catch {
            Write-Error "$module 构建异常: $_"
        }
        finally {
            Pop-Location
            Pop-Location
        }
    }
    
    Write-Success "C++模块构建完成"
}

# 运行Go测试
function Test-GoServices {
    if ($SkipTests) { return }
    
    Write-Info "运行Go单元测试..."
    
    try {
        # 设置测试环境变量
        $env:GO_ENV = "test"
        $env:DATABASE_URL = "postgres://postgres:postgres@localhost:5432/gaokao_test?sslmode=disable"
        
        # 运行测试
        $testArgs = @(
            "test",
            "-v",
            "-race",
            "-coverprofile=coverage.out",
            "./..."
        )
        
        if ($Verbose) {
            $testArgs += "-covermode=atomic"
        }
        
        & go @testArgs
        
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Go测试通过"
            
            # 生成覆盖率报告
            if (Test-Path "coverage.out") {
                go tool cover -html=coverage.out -o coverage.html
                Write-Info "覆盖率报告生成: coverage.html"
            }
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
        
        # E2E测试（如果存在）
        if (Get-Content "package.json" | Select-String "test:e2e") {
            Write-Info "运行E2E测试..."
            npm run test:e2e
        }
        
        Write-Success "前端测试通过"
    }
    catch {
        Write-Error "前端测试失败: $_"
    }
    finally {
        Pop-Location
    }
}

# 运行C++测试
function Test-CppModules {
    if ($SkipTests) { return }
    
    Write-Info "运行C++测试..."
    
    $modules = @(
        "device-fingerprint",
        "license",
        "volunteer-matcher"
    )
    
    foreach ($module in $modules) {
        $testPath = "cpp-modules/$module/build"
        if (-not (Test-Path $testPath)) {
            Write-Warning "$module 构建目录不存在，跳过测试"
            continue
        }
        
        Write-Info "测试 $module..."
        
        try {
            Push-Location $testPath
            
            # 查找测试可执行文件
            $testExe = Get-ChildItem -Filter "*test*" -File | Where-Object { $_.Extension -eq ".exe" -or $_.Extension -eq "" }
            
            if ($testExe) {
                & $testExe.FullName
                if ($LASTEXITCODE -eq 0) {
                    Write-Success "$module 测试通过"
                } else {
                    Write-Error "$module 测试失败"
                }
            } else {
                Write-Warning "$module 没有找到测试可执行文件"
            }
        }
        catch {
            Write-Error "$module 测试异常: $_"
        }
        finally {
            Pop-Location
        }
    }
}

# WSL环境构建
function Build-InWSL {
    Write-Info "在WSL环境中构建..."
    
    if (-not (Get-Command wsl -ErrorAction SilentlyContinue)) {
        Write-Error "WSL不可用"
        exit 1
    }
    
    $wslScript = @"
#!/bin/bash
set -e

echo "🔧 在WSL中构建项目..."

# 设置Go环境
export PATH=/usr/local/go/bin:$PATH
export CGO_ENABLED=1

# 构建Go服务
if [ "$1" = "all" ] || [ "$1" = "backend" ]; then
    echo "📦 构建Go服务..."
    make build-go || go build -o bin/ ./services/...
fi

# 构建前端
if [ "$1" = "all" ] || [ "$1" = "frontend" ]; then
    echo "🎨 构建前端..."
    cd frontend
    npm ci
    npm run build
    cd ..
fi

# 构建C++模块
if [ "$1" = "all" ] || [ "$1" = "cpp" ]; then
    echo "⚙️ 构建C++模块..."
    cd cpp-modules
    for module in device-fingerprint license volunteer-matcher; do
        if [ -d "$module" ]; then
            echo "构建 $module..."
            cd $module
            mkdir -p build && cd build
            cmake .. -DCMAKE_BUILD_TYPE=Release
            cmake --build .
            cd ../..
        fi
    done
    cd ..
fi

echo "✅ WSL构建完成"
"@
    
    $tempScript = [System.IO.Path]::GetTempFileName() + ".sh"
    $wslScript | Out-File -FilePath $tempScript -Encoding UTF8
    
    try {
        wsl bash $tempScript $Target
        Write-Success "WSL构建完成"
    }
    catch {
        Write-Error "WSL构建失败: $_"
    }
    finally {
        Remove-Item $tempScript -ErrorAction SilentlyContinue
    }
}

# 主函数
function Main {
    Write-Info "🚀 高考志愿填报系统 - 构建和测试"
    Write-Info "环境: $Environment | 目标: $Target | 模式: $(if ($Release) { 'Release' } else { 'Debug' })"
    
    Clear-BuildArtifacts
    
    if ($Environment -eq "wsl") {
        Build-InWSL
        return
    }
    
    Test-Dependencies
    
    # 根据目标进行构建
    switch ($Target) {
        "all" {
            Build-GoServices
            Build-Frontend
            Build-CppModules
            Test-GoServices
            Test-Frontend
            Test-CppModules
        }
        "backend" {
            Build-GoServices
            Test-GoServices
        }
        "frontend" {
            Build-Frontend
            Test-Frontend
        }
        "cpp" {
            Build-CppModules
            Test-CppModules
        }
    }
    
    if ($Environment -eq "both") {
        Write-Info "同时在WSL中构建..."
        Build-InWSL
    }
    
    Write-Success "🎉 构建和测试完成！"
    
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
}

# 执行主函数
try {
    Main
}
catch {
    Write-Error "构建过程中发生错误: $_"
    exit 1
}
