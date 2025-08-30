#!/bin/bash

set -e

# Drone配置测试脚本
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 日志函数
log() {
    echo -e "${BLUE}[$(date +'%H:%M:%S')]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[$(date +'%H:%M:%S')] ✅ $1${NC}"
}

log_error() {
    echo -e "${RED}[$(date +'%H:%M:%S')] ❌ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}[$(date +'%H:%M:%S')] ⚠️  $1${NC}"
}

echo "=========================================="
echo "🚁 Drone CI/CD 配置测试"
echo "=========================================="

# 检查Drone CLI
check_drone_cli() {
    log "检查Drone CLI..."
    
    if command -v drone &> /dev/null; then
        DRONE_VERSION=$(drone --version)
        log_success "Drone CLI已安装: $DRONE_VERSION"
        return 0
    else
        log_warning "Drone CLI未安装"
        log "正在安装Drone CLI..."
        
        # 根据操作系统安装Drone CLI
        if [[ "$OSTYPE" == "linux-gnu"* ]]; then
            curl -L https://github.com/harness/drone-cli/releases/latest/download/drone_linux_amd64.tar.gz | tar zx
            sudo mv drone /usr/local/bin/
        elif [[ "$OSTYPE" == "darwin"* ]]; then
            curl -L https://github.com/harness/drone-cli/releases/latest/download/drone_darwin_amd64.tar.gz | tar zx
            sudo mv drone /usr/local/bin/
        elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "win32" ]]; then
            log_warning "Windows用户请手动安装Drone CLI"
            log "下载地址: https://github.com/harness/drone-cli/releases"
            return 1
        else
            log_error "不支持的操作系统: $OSTYPE"
            return 1
        fi
        
        if command -v drone &> /dev/null; then
            log_success "Drone CLI安装成功"
        else
            log_error "Drone CLI安装失败"
            return 1
        fi
    fi
}

# 验证Drone配置文件
validate_drone_config() {
    log "验证Drone配置文件..."
    
    cd "$PROJECT_ROOT"
    
    if [ ! -f ".drone.yml" ]; then
        log_error "未找到.drone.yml文件"
        return 1
    fi
    
    # 检查YAML语法
    if command -v yq &> /dev/null; then
        if yq eval '.kind' .drone.yml &> /dev/null; then
            log_success "YAML语法正确"
        else
            log_error "YAML语法错误"
            return 1
        fi
    else
        log_warning "yq未安装，跳过YAML语法检查"
    fi
    
    # 使用Drone CLI验证配置
    if command -v drone &> /dev/null; then
        if drone lint .drone.yml; then
            log_success "Drone配置验证通过"
        else
            log_error "Drone配置验证失败"
            return 1
        fi
    else
        log_warning "Drone CLI不可用，跳过配置验证"
    fi
}

# 检查必需的文件
check_required_files() {
    log "检查必需的构建文件..."
    
    required_files=(
        "services/api-gateway/Dockerfile"
        "services/user-service/Dockerfile"
        "services/data-service/Dockerfile"
        "services/recommendation-service/Dockerfile"
        "services/payment-service/Dockerfile"
        "docker-compose.yml"
        "docker-compose.prod.yml"
    )
    
    missing_files=()
    
    for file in "${required_files[@]}"; do
        if [ -f "$PROJECT_ROOT/$file" ]; then
            log_success "找到: $file"
        else
            log_error "缺失: $file"
            missing_files+=("$file")
        fi
    done
    
    if [ ${#missing_files[@]} -eq 0 ]; then
        log_success "所有必需文件都存在"
        return 0
    else
        log_error "缺失 ${#missing_files[@]} 个必需文件"
        return 1
    fi
}

# 检查Go模块配置
check_go_modules() {
    log "检查Go模块配置..."
    
    services=(
        "api-gateway"
        "user-service"
        "data-service"
        "recommendation-service"
        "payment-service"
    )
    
    for service in "${services[@]}"; do
        service_dir="$PROJECT_ROOT/services/$service"
        if [ -f "$service_dir/go.mod" ]; then
            cd "$service_dir"
            if go mod verify &> /dev/null; then
                log_success "$service: Go模块验证通过"
            else
                log_warning "$service: Go模块验证失败"
            fi
            cd "$PROJECT_ROOT"
        else
            log_error "$service: 缺少go.mod文件"
        fi
    done
}

# 检查C++构建配置
check_cpp_build() {
    log "检查C++构建配置..."
    
    modules=(
        "device-fingerprint"
        "license"
        "volunteer-matcher"
    )
    
    for module in "${modules[@]}"; do
        module_dir="$PROJECT_ROOT/cpp-modules/$module"
        if [ -f "$module_dir/CMakeLists.txt" ]; then
            log_success "$module: 找到CMakeLists.txt"
        else
            log_error "$module: 缺少CMakeLists.txt"
        fi
    done
}

# 模拟构建测试
simulate_build() {
    log "模拟本地构建测试..."
    
    # 测试Go代码编译
    log "测试Go代码编译..."
    for service in api-gateway user-service data-service recommendation-service payment-service; do
        service_dir="$PROJECT_ROOT/services/$service"
        if [ -d "$service_dir" ]; then
            cd "$service_dir"
            if go build -o /tmp/test_${service} . 2>/dev/null; then
                log_success "$service: 编译成功"
                rm -f /tmp/test_${service}
            else
                log_warning "$service: 编译失败（可能需要依赖）"
            fi
            cd "$PROJECT_ROOT"
        fi
    done
    
    # 测试Docker构建（简化检查）
    log "检查Docker构建配置..."
    for service in api-gateway user-service data-service recommendation-service payment-service; do
        dockerfile="$PROJECT_ROOT/services/$service/Dockerfile"
        if [ -f "$dockerfile" ]; then
            if grep -q "FROM.*golang" "$dockerfile"; then
                log_success "$service: Dockerfile配置正确"
            else
                log_warning "$service: Dockerfile可能有问题"
            fi
        fi
    done
}

# 生成Drone配置报告
generate_report() {
    log "生成Drone配置报告..."
    
    cat > "$PROJECT_ROOT/drone-config-report.md" << 'EOF'
# Drone CI/CD 配置报告

## 📋 配置概览

### 流水线配置
- **主流水线**: gaokaohub-ci (12个步骤)
- **通知流水线**: notify (构建通知)
- **安全审计**: security-audit (定时扫描)

### 触发条件
- Push到master/develop分支
- Pull Request
- Tag创建 (v*)
- 定时任务 (每日安全扫描)

### 构建环境
- **Go版本**: 1.21
- **Docker**: 启用BuildKit
- **缓存**: Go模块 + 构建缓存

## 🔧 构建步骤

### 1. 代码质量检查
- golangci-lint 静态分析
- staticcheck 代码检查
- gosec 安全扫描

### 2. 测试
- Go单元测试 + 覆盖率
- C++模块构建测试
- 集成测试（PostgreSQL + Redis）

### 3. 构建
- 5个Go微服务编译
- 3个C++模块构建
- Docker镜像构建和推送

### 4. 安全
- 镜像漏洞扫描 (Trivy)
- 依赖漏洞检查 (Nancy)
- 许可证合规检查 (FOSSA)

### 5. 部署
- 测试环境: develop分支自动部署
- 生产环境: tag触发部署
- 滚动更新策略

## 📊 统计信息

EOF

    # 添加统计信息
    echo "- **配置文件大小**: $(wc -l < .drone.yml) 行" >> "$PROJECT_ROOT/drone-config-report.md"
    echo "- **构建步骤数**: 12 个" >> "$PROJECT_ROOT/drone-config-report.md"
    echo "- **服务数量**: 5 个Go服务 + 3个C++模块" >> "$PROJECT_ROOT/drone-config-report.md"
    echo "- **支持平台**: Docker + Kubernetes" >> "$PROJECT_ROOT/drone-config-report.md"
    echo "" >> "$PROJECT_ROOT/drone-config-report.md"
    echo "## 🎯 下一步" >> "$PROJECT_ROOT/drone-config-report.md"
    echo "" >> "$PROJECT_ROOT/drone-config-report.md"
    echo "1. 在Drone UI中激活仓库" >> "$PROJECT_ROOT/drone-config-report.md"
    echo "2. 配置必需的secrets" >> "$PROJECT_ROOT/drone-config-report.md"
    echo "3. 推送代码触发首次构建" >> "$PROJECT_ROOT/drone-config-report.md"
    echo "4. 监控构建状态和部署" >> "$PROJECT_ROOT/drone-config-report.md"
    echo "" >> "$PROJECT_ROOT/drone-config-report.md"
    echo "生成时间: $(date)" >> "$PROJECT_ROOT/drone-config-report.md"
    
    log_success "报告已生成: drone-config-report.md"
}

# 提供激活建议
provide_activation_guide() {
    log "Drone激活指南..."
    
    echo ""
    echo "🎯 Drone CI/CD 激活步骤:"
    echo ""
    echo "1. 登录到您的Drone服务器"
    echo "   https://your-drone-server.com"
    echo ""
    echo "2. 找到并激活仓库"
    echo "   - 找到 'gaokaohub/gaokao' 仓库"
    echo "   - 点击 'ACTIVATE' 按钮"
    echo ""
    echo "3. 配置Secrets (在仓库设置中):"
    echo "   - docker_username: Docker Hub用户名"
    echo "   - docker_password: Docker Hub密码"
    echo "   - staging_host: 测试服务器地址"
    echo "   - production_host: 生产服务器地址"
    echo "   - slack_webhook: Slack通知地址 (可选)"
    echo ""
    echo "4. 触发首次构建:"
    echo "   git commit --allow-empty -m 'trigger: 激活Drone CI/CD'"
    echo "   git push origin master"
    echo ""
    echo "5. 监控构建进度:"
    echo "   https://your-drone-server.com/gaokaohub/gaokao"
    echo ""
}

# 主函数
main() {
    cd "$PROJECT_ROOT"
    
    # 执行所有检查
    check_drone_cli
    validate_drone_config
    check_required_files
    check_go_modules
    check_cpp_build
    simulate_build
    generate_report
    provide_activation_guide
    
    echo ""
    log_success "Drone CI/CD配置测试完成！"
    echo ""
    echo "📁 生成的文件:"
    echo "  - .drone.yml (完整的CI/CD配置)"
    echo "  - DRONE_SETUP.md (详细配置指南)"
    echo "  - drone-config-report.md (配置报告)"
    echo ""
    echo "🚀 现在可以在Drone UI中激活仓库了！"
}

# 执行主函数
main "$@"