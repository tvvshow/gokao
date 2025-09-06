#!/bin/bash

set -e

# Drone激活和测试脚本
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# 您的Drone服务器配置
DRONE_SERVER="http://166.108.233.220"
DRONE_TOKEN=""  # 将在激活后获取

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m'

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

log_info() {
    echo -e "${PURPLE}[$(date +'%H:%M:%S')] ℹ️  $1${NC}"
}

echo "=========================================="
echo "🚁 Drone CI/CD 激活和测试"
echo "🌐 服务器: $DRONE_SERVER"
echo "=========================================="

# 检测Drone服务器状态
check_drone_server() {
    log "检查Drone服务器状态..."
    
    if curl -s "$DRONE_SERVER/healthz" | grep -q "ok" 2>/dev/null; then
        log_success "Drone服务器运行正常"
        return 0
    elif curl -s -o /dev/null -w "%{http_code}" "$DRONE_SERVER" | grep -q "200\|302" 2>/dev/null; then
        log_success "Drone服务器可访问"
        return 0
    else
        log_error "无法连接到Drone服务器: $DRONE_SERVER"
        log_info "请确保："
        echo "  1. Drone服务器正在运行"
        echo "  2. 防火墙允许80端口访问"
        echo "  3. 服务器IP地址正确"
        return 1
    fi
}

# 检查Drone配置文件
validate_drone_config() {
    log "验证.drone.yml配置..."
    
    cd "$PROJECT_ROOT"
    
    if [ ! -f ".drone.yml" ]; then
        log_error "未找到.drone.yml文件"
        return 1
    fi
    
    # 检查配置文件大小和内容
    config_size=$(wc -l < .drone.yml)
    log_info "配置文件大小: $config_size 行"
    
    # 检查关键配置
    if grep -q "kind: pipeline" .drone.yml; then
        log_success "找到pipeline配置"
    else
        log_error "pipeline配置缺失"
        return 1
    fi
    
    if grep -q "gaokaohub-ci" .drone.yml; then
        log_success "主构建管道配置正确"
    else
        log_warning "主构建管道名称可能有问题"
    fi
    
    # 检查步骤数量
    step_count=$(grep -c "name:" .drone.yml)
    log_info "配置的构建步骤数: $step_count"
    
    log_success "Drone配置验证完成"
}

# 安装和配置Drone CLI
setup_drone_cli() {
    log "设置Drone CLI..."
    
    if command -v drone &> /dev/null; then
        log_success "Drone CLI已安装"
    else
        log "安装Drone CLI..."
        
        # 下载并安装drone CLI
        if [[ "$OSTYPE" == "linux-gnu"* ]]; then
            curl -L https://github.com/harness/drone-cli/releases/latest/download/drone_linux_amd64.tar.gz | tar zx
            chmod +x drone
            sudo mv drone /usr/local/bin/
        elif [[ "$OSTYPE" == "darwin"* ]]; then
            curl -L https://github.com/harness/drone-cli/releases/latest/download/drone_darwin_amd64.tar.gz | tar zx
            chmod +x drone
            sudo mv drone /usr/local/bin/
        else
            log_warning "请手动安装Drone CLI"
            log_info "Windows用户可以下载: https://github.com/harness/drone-cli/releases"
            log_info "或使用Docker运行: docker run --rm -v \$(pwd):/drone -w /drone drone/cli:latest"
            return 1
        fi
        
        if command -v drone &> /dev/null; then
            log_success "Drone CLI安装成功"
        else
            log_error "Drone CLI安装失败"
            return 1
        fi
    fi
    
    # 配置Drone CLI
    export DRONE_SERVER="$DRONE_SERVER"
    export DRONE_TOKEN="$DRONE_TOKEN"
    
    log_info "Drone CLI配置:"
    echo "  DRONE_SERVER=$DRONE_SERVER"
    echo "  DRONE_TOKEN=${DRONE_TOKEN:0:10}..." # 只显示前10个字符
}

# 提供手动激活指南
provide_manual_activation() {
    log_info "手动激活指南:"
    echo ""
    echo "🔗 请按以下步骤激活Drone CI/CD:"
    echo ""
    echo "1️⃣ 访问Drone服务器:"
    echo "   👉 http://166.108.233.220"
    echo ""
    echo "2️⃣ 使用GitHub账号登录"
    echo "   - 点击 'Continue with GitHub'"
    echo "   - 授权Drone访问您的GitHub账号"
    echo ""
    echo "3️⃣ 激活仓库:"
    echo "   - 在仓库列表中找到 'gaokao' 仓库"
    echo "   - 点击仓库右侧的 'ACTIVATE' 按钮"
    echo "   - 确认激活"
    echo ""
    echo "4️⃣ 配置仓库设置:"
    echo "   - 进入仓库设置页面"
    echo "   - 配置必需的Secrets (见下方)"
    echo ""
    echo "5️⃣ 触发首次构建:"
    echo "   - 推送任何更改到master分支"
    echo "   - 或手动触发构建"
    echo ""
}

# 生成Secrets配置指南
generate_secrets_guide() {
    log "生成Secrets配置指南..."
    
    cat > "$PROJECT_ROOT/drone-secrets.md" << 'EOF'
# Drone Secrets 配置指南

在Drone UI中配置以下Secrets (仓库设置 → Secrets)：

## 🐳 Docker Registry (必需)
```
docker_username: your_docker_hub_username
docker_password: your_docker_hub_password_or_token
```

## 🚀 部署配置 (可选，用于自动部署)
```
# 测试环境
staging_host: your_staging_server_ip
staging_user: deploy
staging_key: |
  -----BEGIN OPENSSH PRIVATE KEY-----
  your_private_key_for_staging_server
  -----END OPENSSH PRIVATE KEY-----

# 生产环境  
production_host: your_production_server_ip
production_user: deploy
production_key: |
  -----BEGIN OPENSSH PRIVATE KEY-----
  your_private_key_for_production_server
  -----END OPENSSH PRIVATE KEY-----
```

## 📢 通知配置 (可选)
```
# Slack通知
slack_webhook: https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK

# 钉钉通知
dingtalk_webhook: https://oapi.dingtalk.com/robot/send?access_token=YOUR_TOKEN
```

## 🔍 安全扫描 (可选)
```
# FOSSA许可证检查
fossa_api_key: your_fossa_api_key
```

## 💡 配置说明

### 最小配置
如果只想启用基本的CI功能，只需配置：
- `docker_username`
- `docker_password`

### 自动部署
如果需要自动部署到服务器，还需配置：
- 部署服务器的SSH配置
- 确保目标服务器安装了Docker和Docker Compose

### 通知
配置Webhook可以接收构建状态通知

---
配置完成后，推送代码即可触发构建！
EOF

    log_success "Secrets配置指南已生成: drone-secrets.md"
}

# 检查项目构建就绪性
check_build_readiness() {
    log "检查项目构建就绪性..."
    
    cd "$PROJECT_ROOT"
    
    # 检查Go服务
    services=(
        "api-gateway"
        "user-service"
        "data-service"
        "recommendation-service"
        "payment-service"
    )
    
    go_services_ready=0
    for service in "${services[@]}"; do
        if [ -f "services/$service/main.go" ] && [ -f "services/$service/go.mod" ]; then
            log_success "Go服务就绪: $service"
            ((go_services_ready++))
        else
            log_warning "Go服务未就绪: $service"
        fi
    done
    
    # 检查C++模块
    cpp_modules=(
        "device-fingerprint"
        "license"
        "volunteer-matcher"
    )
    
    cpp_modules_ready=0
    for module in "${cpp_modules[@]}"; do
        if [ -f "cpp-modules/$module/CMakeLists.txt" ]; then
            log_success "C++模块就绪: $module"
            ((cpp_modules_ready++))
        else
            log_warning "C++模块未就绪: $module"
        fi
    done
    
    # 检查Docker配置
    docker_configs=0
    for service in "${services[@]}"; do
        if [ -f "services/$service/Dockerfile" ]; then
            ((docker_configs++))
        fi
    done
    
    log_info "构建就绪性统计:"
    echo "  📦 Go服务: $go_services_ready/${#services[@]}"
    echo "  🔧 C++模块: $cpp_modules_ready/${#cpp_modules[@]}"
    echo "  🐳 Docker配置: $docker_configs/${#services[@]}"
    
    if [ $go_services_ready -eq ${#services[@]} ] && [ $cpp_modules_ready -eq ${#cpp_modules[@]} ] && [ $docker_configs -eq ${#services[@]} ]; then
        log_success "项目完全构建就绪！"
        return 0
    else
        log_warning "项目部分就绪，但仍可进行CI/CD测试"
        return 0
    fi
}

# 生成测试推送命令
generate_test_commands() {
    log "生成测试命令..."
    
    cat > "$PROJECT_ROOT/test-drone-build.sh" << 'EOF'
#!/bin/bash

# Drone CI/CD 测试命令

echo "🚀 触发Drone构建测试..."

# 方法1: 空提交触发构建
echo "📝 创建测试提交..."
git add .
git commit --allow-empty -m "test: 触发Drone CI/CD构建

🔧 测试内容:
- Go微服务构建
- C++模块编译
- Docker镜像构建
- 安全扫描
- 代码质量检查

Generated by Drone activation script"

echo "📤 推送到GitHub..."
git push origin master

echo ""
echo "✅ 构建已触发！"
echo "📊 查看构建状态: http://166.108.233.220/pestxo/gaokao"
echo ""
echo "💡 如果这是首次构建，可能需要几分钟来拉取依赖和镜像"
EOF

    chmod +x "$PROJECT_ROOT/test-drone-build.sh"
    log_success "测试脚本已生成: test-drone-build.sh"
}

# 显示最终指南
show_final_guide() {
    echo ""
    echo "🎉 Drone CI/CD 配置和测试完成！"
    echo ""
    echo "📋 接下来的步骤:"
    echo ""
    echo "1️⃣ 手动激活 (必需):"
    echo "   👉 访问: http://166.108.233.220"
    echo "   👉 登录并激活 'gaokao' 仓库"
    echo ""
    echo "2️⃣ 配置Secrets (推荐):"
    echo "   👉 查看: drone-secrets.md"
    echo "   👉 至少配置Docker Hub凭据"
    echo ""
    echo "3️⃣ 触发首次构建:"
    echo "   👉 运行: ./test-drone-build.sh"
    echo "   👉 或推送任何代码更改"
    echo ""
    echo "4️⃣ 监控构建:"
    echo "   👉 构建状态: http://166.108.233.220/pestxo/gaokao"
    echo "   👉 预计构建时间: 5-10分钟"
    echo ""
    echo "📁 生成的文件:"
    echo "  - drone-secrets.md (Secrets配置指南)"
    echo "  - test-drone-build.sh (测试构建脚本)"
    echo "  - .drone.yml (完整CI/CD配置)"
    echo ""
    echo "🆘 如有问题，检查："
    echo "  - Drone服务器是否正常运行"
    echo "  - GitHub OAuth配置是否正确"
    echo "  - 网络连接是否正常"
    echo ""
}

# 主函数
main() {
    cd "$PROJECT_ROOT"
    
    # 执行检查和配置
    check_drone_server
    validate_drone_config
    setup_drone_cli || log_warning "Drone CLI设置跳过"
    check_build_readiness
    provide_manual_activation
    generate_secrets_guide
    generate_test_commands
    show_final_guide
    
    log_success "Drone激活准备完成！"
}

# 执行主函数
main "$@"