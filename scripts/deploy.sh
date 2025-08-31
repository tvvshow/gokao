#!/bin/bash

# 高考志愿填报系统部署脚本
# 支持本地开发、预发布、生产环境部署

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 显示帮助信息
show_help() {
    cat << EOF
高考志愿填报系统部署脚本

用法: $0 [选项] <环境>

环境:
    local       本地开发环境
    staging     预发布环境
    production  生产环境

选项:
    -h, --help              显示帮助信息
    -v, --version VERSION   指定部署版本 (默认: latest)
    -f, --force             强制部署，跳过确认
    -r, --rollback          回滚到上一个版本
    -s, --skip-tests        跳过测试
    -d, --dry-run           模拟运行，不实际部署
    --no-backup             不创建备份
    --scale REPLICAS        设置副本数量

示例:
    $0 local                    # 部署到本地环境
    $0 staging -v v1.2.3        # 部署指定版本到预发布环境
    $0 production --force       # 强制部署到生产环境
    $0 production --rollback    # 回滚生产环境

EOF
}

# 默认参数
ENVIRONMENT=""
VERSION="latest"
FORCE=false
ROLLBACK=false
SKIP_TESTS=false
DRY_RUN=false
NO_BACKUP=false
SCALE=""

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -f|--force)
            FORCE=true
            shift
            ;;
        -r|--rollback)
            ROLLBACK=true
            shift
            ;;
        -s|--skip-tests)
            SKIP_TESTS=true
            shift
            ;;
        -d|--dry-run)
            DRY_RUN=true
            shift
            ;;
        --no-backup)
            NO_BACKUP=true
            shift
            ;;
        --scale)
            SCALE="$2"
            shift 2
            ;;
        local|staging|production)
            ENVIRONMENT="$1"
            shift
            ;;
        *)
            log_error "未知参数: $1"
            show_help
            exit 1
            ;;
    esac
done

# 检查必需参数
if [[ -z "$ENVIRONMENT" ]]; then
    log_error "请指定部署环境"
    show_help
    exit 1
fi

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    local deps=("docker" "kubectl" "helm")
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            log_error "缺少依赖: $dep"
            exit 1
        fi
    done
    
    log_success "依赖检查通过"
}

# 设置环境变量
setup_environment() {
    log_info "设置环境变量..."
    
    case $ENVIRONMENT in
        local)
            export NAMESPACE="gaokao-local"
            export REGISTRY="localhost:5000"
            export REPLICAS="1"
            export RESOURCES_LIMITS_CPU="500m"
            export RESOURCES_LIMITS_MEMORY="512Mi"
            ;;
        staging)
            export NAMESPACE="gaokao-staging"
            export REGISTRY="ghcr.io/gaokao"
            export REPLICAS="2"
            export RESOURCES_LIMITS_CPU="1000m"
            export RESOURCES_LIMITS_MEMORY="1Gi"
            ;;
        production)
            export NAMESPACE="gaokao-production"
            export REGISTRY="ghcr.io/gaokao"
            export REPLICAS="3"
            export RESOURCES_LIMITS_CPU="2000m"
            export RESOURCES_LIMITS_MEMORY="2Gi"
            ;;
    esac
    
    if [[ -n "$SCALE" ]]; then
        export REPLICAS="$SCALE"
    fi
    
    export IMAGE_TAG="$VERSION"
    
    log_success "环境变量设置完成"
}

# 创建命名空间
create_namespace() {
    log_info "创建命名空间: $NAMESPACE"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] 将创建命名空间: $NAMESPACE"
        return
    fi
    
    kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
    log_success "命名空间创建完成"
}

# 备份当前部署
backup_deployment() {
    if [[ "$NO_BACKUP" == "true" ]]; then
        log_warning "跳过备份"
        return
    fi
    
    log_info "备份当前部署..."
    
    local backup_dir="backups/$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$backup_dir"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] 将备份到: $backup_dir"
        return
    fi
    
    kubectl get all -n "$NAMESPACE" -o yaml > "$backup_dir/resources.yaml"
    kubectl get configmaps -n "$NAMESPACE" -o yaml > "$backup_dir/configmaps.yaml"
    kubectl get secrets -n "$NAMESPACE" -o yaml > "$backup_dir/secrets.yaml"
    
    log_success "备份完成: $backup_dir"
}

# 运行测试
run_tests() {
    if [[ "$SKIP_TESTS" == "true" ]]; then
        log_warning "跳过测试"
        return
    fi
    
    log_info "运行测试..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] 将运行测试"
        return
    fi
    
    # 运行单元测试
    log_info "运行单元测试..."
    go test -v ./...
    
    # 运行前端测试
    log_info "运行前端测试..."
    cd frontend && npm test && cd ..
    
    log_success "测试通过"
}

# 构建镜像
build_images() {
    log_info "构建镜像..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] 将构建镜像"
        return
    fi
    
    local services=("api-gateway" "user-service" "data-service" "payment-service" "recommendation-service")
    
    for service in "${services[@]}"; do
        log_info "构建 $service 镜像..."
        docker build -t "$REGISTRY/$service:$VERSION" "./services/$service"
        
        if [[ "$ENVIRONMENT" != "local" ]]; then
            docker push "$REGISTRY/$service:$VERSION"
        fi
    done
    
    # 构建前端镜像
    log_info "构建前端镜像..."
    cd frontend && npm run build && cd ..
    docker build -t "$REGISTRY/frontend:$VERSION" "./frontend"
    
    if [[ "$ENVIRONMENT" != "local" ]]; then
        docker push "$REGISTRY/frontend:$VERSION"
    fi
    
    log_success "镜像构建完成"
}

# 部署应用
deploy_application() {
    log_info "部署应用..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] 将部署应用到 $NAMESPACE"
        return
    fi
    
    # 使用 Helm 部署
    helm upgrade --install gaokao-system ./helm/gaokao \
        --namespace "$NAMESPACE" \
        --set image.tag="$VERSION" \
        --set image.registry="$REGISTRY" \
        --set replicaCount="$REPLICAS" \
        --set resources.limits.cpu="$RESOURCES_LIMITS_CPU" \
        --set resources.limits.memory="$RESOURCES_LIMITS_MEMORY" \
        --set environment="$ENVIRONMENT" \
        --wait --timeout=600s
    
    log_success "应用部署完成"
}

# 健康检查
health_check() {
    log_info "执行健康检查..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] 将执行健康检查"
        return
    fi
    
    # 等待所有 Pod 就绪
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=gaokao -n "$NAMESPACE" --timeout=300s
    
    # 检查服务状态
    local services=("api-gateway" "user-service" "data-service" "payment-service" "recommendation-service" "frontend")
    
    for service in "${services[@]}"; do
        local ready=$(kubectl get deployment "$service" -n "$NAMESPACE" -o jsonpath='{.status.readyReplicas}')
        local desired=$(kubectl get deployment "$service" -n "$NAMESPACE" -o jsonpath='{.spec.replicas}')
        
        if [[ "$ready" == "$desired" ]]; then
            log_success "$service 服务健康"
        else
            log_error "$service 服务不健康 ($ready/$desired)"
            exit 1
        fi
    done
    
    log_success "健康检查通过"
}

# 回滚部署
rollback_deployment() {
    log_info "回滚部署..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] 将回滚部署"
        return
    fi
    
    helm rollback gaokao-system -n "$NAMESPACE"
    
    log_success "回滚完成"
}

# 确认部署
confirm_deployment() {
    if [[ "$FORCE" == "true" ]]; then
        return
    fi
    
    echo
    log_warning "即将部署到 $ENVIRONMENT 环境"
    log_warning "版本: $VERSION"
    log_warning "副本数: $REPLICAS"
    echo
    
    read -p "确认继续? (y/N): " -n 1 -r
    echo
    
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "部署已取消"
        exit 0
    fi
}

# 主函数
main() {
    log_info "开始部署高考志愿填报系统"
    log_info "环境: $ENVIRONMENT"
    log_info "版本: $VERSION"
    
    check_dependencies
    setup_environment
    
    if [[ "$ROLLBACK" == "true" ]]; then
        confirm_deployment
        rollback_deployment
        health_check
        log_success "回滚完成"
        exit 0
    fi
    
    confirm_deployment
    create_namespace
    backup_deployment
    run_tests
    
    if [[ "$ENVIRONMENT" == "local" ]]; then
        build_images
    fi
    
    deploy_application
    health_check
    
    log_success "部署完成！"
    
    # 显示访问信息
    if [[ "$ENVIRONMENT" == "local" ]]; then
        log_info "本地访问地址:"
        log_info "  前端: http://localhost:3000"
        log_info "  API: http://localhost:8080"
    else
        local ingress_ip=$(kubectl get ingress gaokao-ingress -n "$NAMESPACE" -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
        if [[ -n "$ingress_ip" ]]; then
            log_info "访问地址: http://$ingress_ip"
        fi
    fi
}

# 执行主函数
main "$@"
