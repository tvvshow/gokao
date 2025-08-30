#!/bin/bash

# 高考志愿填报系统 - 构建脚本
# 用于构建所有Docker镜像

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

# 检查Docker是否安装
check_docker() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed or not in PATH"
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        log_error "Docker daemon is not running"
        exit 1
    fi
    
    log_success "Docker is ready"
}

# 清理旧镜像
cleanup_images() {
    log_info "Cleaning up old images..."
    
    # 删除未使用的镜像
    docker image prune -f
    
    # 删除项目相关的旧镜像
    docker images | grep gaokao | awk '{print $3}' | xargs -r docker rmi -f || true
    
    log_success "Image cleanup completed"
}

# 构建镜像
build_image() {
    local service=$1
    local env=$2
    local dockerfile_path=$3
    local context_path=$4
    local image_tag="gaokao/${service}:${env}"
    
    log_info "Building ${image_tag}..."
    
    if docker build -t "$image_tag" -f "$dockerfile_path" "$context_path"; then
        log_success "Successfully built ${image_tag}"
        
        # 运行安全扫描（如果安装了trivy）
        if command -v trivy &> /dev/null; then
            log_info "Running security scan for ${image_tag}..."
            trivy image --exit-code 0 --severity HIGH,CRITICAL "$image_tag"
        fi
        
        return 0
    else
        log_error "Failed to build ${image_tag}"
        return 1
    fi
}

# 显示帮助信息
show_help() {
    cat << EOF
高考志愿填报系统 - Docker 构建脚本

用法: $0 [选项] [环境]

选项:
    -h, --help          显示此帮助信息
    -c, --cleanup       构建前清理旧镜像
    -s, --skip-tests    跳过测试镜像构建
    -v, --verbose       详细输出
    --version VERSION   指定镜像版本标签 (默认: latest)

环境:
    dev                 构建开发环境镜像
    prod                构建生产环境镜像
    test                构建测试环境镜像
    all                 构建所有环境镜像 (默认)

示例:
    $0                  # 构建所有环境的镜像
    $0 dev              # 只构建开发环境镜像
    $0 -c prod          # 清理后构建生产环境镜像
    $0 --version v1.0.0 prod  # 构建带版本标签的生产环境镜像

EOF
}

# 构建开发环境镜像
build_dev() {
    log_info "Building development environment images..."
    
    build_image "api-gateway" "dev" "docker/dev/Dockerfile.api-gateway" "services/api-gateway" || return 1
    build_image "user-service" "dev" "docker/dev/Dockerfile.user-service" "services/user-service" || return 1
    build_image "cpp-modules" "dev" "docker/dev/Dockerfile.cpp-modules" "." || return 1
    
    log_success "Development environment images built successfully"
}

# 构建生产环境镜像
build_prod() {
    log_info "Building production environment images..."
    
    # 设置版本标签
    local version=${VERSION:-latest}
    
    # 构建生产环境镜像
    docker build -t "gaokao/api-gateway:${version}" -f "docker/prod/Dockerfile.api-gateway" "services/api-gateway" || return 1
    docker build -t "gaokao/user-service:${version}" -f "docker/prod/Dockerfile.user-service" "services/user-service" || return 1
    docker build -t "gaokao/cpp-modules:${version}" -f "docker/prod/Dockerfile.cpp-modules" "." || return 1
    
    # 如果版本不是latest，也打上latest标签
    if [ "$version" != "latest" ]; then
        docker tag "gaokao/api-gateway:${version}" "gaokao/api-gateway:latest"
        docker tag "gaokao/user-service:${version}" "gaokao/user-service:latest"
        docker tag "gaokao/cpp-modules:${version}" "gaokao/cpp-modules:latest"
    fi
    
    log_success "Production environment images built successfully (version: ${version})"
}

# 构建测试环境镜像
build_test() {
    if [ "$SKIP_TESTS" = "true" ]; then
        log_warning "Skipping test image build"
        return 0
    fi
    
    log_info "Building test environment images..."
    
    build_image "test-runner" "latest" "docker/test/Dockerfile.test-runner" "." || return 1
    
    log_success "Test environment images built successfully"
}

# 验证构建结果
verify_builds() {
    log_info "Verifying built images..."
    
    local failed_images=()
    
    # 检查镜像是否存在
    for image in $(docker images --format "{{.Repository}}:{{.Tag}}" | grep gaokao); do
        if docker inspect "$image" > /dev/null 2>&1; then
            log_success "✓ $image"
        else
            log_error "✗ $image"
            failed_images+=("$image")
        fi
    done
    
    if [ ${#failed_images[@]} -eq 0 ]; then
        log_success "All images verified successfully"
    else
        log_error "Failed to verify images: ${failed_images[*]}"
        return 1
    fi
}

# 显示构建统计
show_stats() {
    log_info "Build statistics:"
    echo
    docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}" | grep -E "(REPOSITORY|gaokao)"
    echo
    
    local total_size=$(docker images --format "{{.Size}}" | grep -E "MB|GB" | sed 's/MB//' | sed 's/GB//' | awk '{sum += $1} END {print sum "MB"}')
    log_info "Total image size: $total_size"
}

# 主函数
main() {
    local environment="all"
    local cleanup=false
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -c|--cleanup)
                cleanup=true
                shift
                ;;
            -s|--skip-tests)
                SKIP_TESTS=true
                shift
                ;;
            -v|--verbose)
                set -x
                shift
                ;;
            --version)
                VERSION="$2"
                shift 2
                ;;
            dev|prod|test|all)
                environment="$1"
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # 检查Docker环境
    check_docker
    
    # 清理镜像（如果需要）
    if [ "$cleanup" = true ]; then
        cleanup_images
    fi
    
    # 记录开始时间
    local start_time=$(date +%s)
    
    # 根据环境构建镜像
    case $environment in
        dev)
            build_dev || exit 1
            ;;
        prod)
            build_prod || exit 1
            ;;
        test)
            build_test || exit 1
            ;;
        all)
            build_dev || exit 1
            build_prod || exit 1
            build_test || exit 1
            ;;
    esac
    
    # 验证构建结果
    verify_builds || exit 1
    
    # 计算构建时间
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    # 显示统计信息
    show_stats
    
    log_success "Build completed in ${duration} seconds"
}

# 运行主函数
main "$@"