#!/bin/bash

set -e

# 脚本配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
BUILD_DIR="${PROJECT_ROOT}/build"
LOG_FILE="${BUILD_DIR}/build.log"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1" | tee -a "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] ✅ $1${NC}" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ❌ $1${NC}" | tee -a "$LOG_FILE"
}

log_warning() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] ⚠️  $1${NC}" | tee -a "$LOG_FILE"
}

# 创建构建目录
create_build_dir() {
    log "创建构建目录..."
    mkdir -p "$BUILD_DIR"
    mkdir -p "$BUILD_DIR/bin"
    mkdir -p "$BUILD_DIR/lib"
    mkdir -p "$BUILD_DIR/logs"
    log_success "构建目录创建完成"
}

# 检查依赖
check_dependencies() {
    log "检查系统依赖..."
    
    # 检查Go
    if ! command -v go &> /dev/null; then
        log_error "Go未安装，请先安装Go 1.21+"
        exit 1
    fi
    
    GO_VERSION=$(go version | grep -oP 'go\d+\.\d+' | sed 's/go//')
    log "检测到Go版本: $GO_VERSION"
    
    # 检查C++编译器
    if ! command -v g++ &> /dev/null && ! command -v clang++ &> /dev/null; then
        log_error "C++编译器未安装，请先安装g++或clang++"
        exit 1
    fi
    
    # 检查CMake
    if ! command -v cmake &> /dev/null; then
        log_warning "CMake未安装，C++模块将跳过构建"
        SKIP_CPP=true
    fi
    
    # 检查Docker
    if ! command -v docker &> /dev/null; then
        log_warning "Docker未安装，容器镜像构建将跳过"
        SKIP_DOCKER=true
    fi
    
    log_success "依赖检查完成"
}

# 构建C++模块
build_cpp_modules() {
    if [ "$SKIP_CPP" = true ]; then
        log_warning "跳过C++模块构建"
        return
    fi
    
    log "开始构建C++模块..."
    
    # 构建设备指纹模块
    if [ -d "${PROJECT_ROOT}/cpp-modules/device-fingerprint" ]; then
        log "构建设备指纹模块..."
        cd "${PROJECT_ROOT}/cpp-modules/device-fingerprint"
        
        mkdir -p build
        cd build
        cmake .. -DCMAKE_BUILD_TYPE=Release \
                 -DCMAKE_INSTALL_PREFIX="${BUILD_DIR}" \
                 -DBUILD_TESTING=OFF
        make -j$(nproc)
        make install
        
        log_success "设备指纹模块构建完成"
    fi
    
    # 构建许可证模块
    if [ -d "${PROJECT_ROOT}/cpp-modules/license" ]; then
        log "构建许可证模块..."
        cd "${PROJECT_ROOT}/cpp-modules/license"
        
        mkdir -p build
        cd build
        cmake .. -DCMAKE_BUILD_TYPE=Release \
                 -DCMAKE_INSTALL_PREFIX="${BUILD_DIR}" \
                 -DBUILD_TESTING=OFF
        make -j$(nproc)
        make install
        
        log_success "许可证模块构建完成"
    fi
    
    # 构建推荐引擎模块
    if [ -d "${PROJECT_ROOT}/cpp-modules/volunteer-matcher" ]; then
        log "构建推荐引擎模块..."
        cd "${PROJECT_ROOT}/cpp-modules/volunteer-matcher"
        
        mkdir -p build
        cd build
        cmake .. -DCMAKE_BUILD_TYPE=Release \
                 -DCMAKE_INSTALL_PREFIX="${BUILD_DIR}" \
                 -DBUILD_TESTING=OFF
        make -j$(nproc)
        make install
        
        log_success "推荐引擎模块构建完成"
    fi
    
    log_success "所有C++模块构建完成"
}

# 构建Go服务
build_go_services() {
    log "开始构建Go服务..."
    
    cd "$PROJECT_ROOT"
    
    # 服务列表
    services=(
        "api-gateway"
        "user-service"
        "data-service"
        "recommendation-service"
        "payment-service"
    )
    
    for service in "${services[@]}"; do
        service_dir="${PROJECT_ROOT}/services/${service}"
        if [ -d "$service_dir" ]; then
            log "构建服务: $service"
            cd "$service_dir"
            
            # 下载依赖
            go mod tidy
            go mod download
            
            # 构建二进制文件
            CGO_ENABLED=1 GOOS=linux go build -a -ldflags="-s -w" -o "${BUILD_DIR}/bin/${service}" .
            
            # 复制配置文件
            if [ -f "config.yaml" ]; then
                cp "config.yaml" "${BUILD_DIR}/bin/${service}.yaml"
            fi
            
            log_success "服务 $service 构建完成"
        else
            log_warning "服务目录不存在: $service"
        fi
    done
    
    log_success "所有Go服务构建完成"
}

# 运行测试
run_tests() {
    log "开始运行测试..."
    
    cd "$PROJECT_ROOT"
    
    # Go测试
    log "运行Go测试..."
    for service_dir in services/*/; do
        if [ -f "${service_dir}/go.mod" ]; then
            service_name=$(basename "$service_dir")
            log "测试服务: $service_name"
            cd "$service_dir"
            go test -v ./... >> "$LOG_FILE" 2>&1 || log_warning "服务 $service_name 测试失败"
            cd "$PROJECT_ROOT"
        fi
    done
    
    # C++测试
    if [ "$SKIP_CPP" != true ]; then
        log "运行C++测试..."
        for cpp_dir in cpp-modules/*/; do
            if [ -f "${cpp_dir}/CMakeLists.txt" ]; then
                module_name=$(basename "$cpp_dir")
                log "测试模块: $module_name"
                cd "${cpp_dir}/build"
                if [ -f "test_${module_name}" ]; then
                    ./test_${module_name} >> "$LOG_FILE" 2>&1 || log_warning "模块 $module_name 测试失败"
                fi
                cd "$PROJECT_ROOT"
            fi
        done
    fi
    
    log_success "测试完成"
}

# 构建Docker镜像
build_docker_images() {
    if [ "$SKIP_DOCKER" = true ]; then
        log_warning "跳过Docker镜像构建"
        return
    fi
    
    log "开始构建Docker镜像..."
    
    cd "$PROJECT_ROOT"
    
    # 构建各个服务的Docker镜像
    services=(
        "api-gateway"
        "user-service"
        "data-service"
        "recommendation-service"
        "payment-service"
    )
    
    for service in "${services[@]}"; do
        service_dir="${PROJECT_ROOT}/services/${service}"
        if [ -f "${service_dir}/Dockerfile" ]; then
            log "构建Docker镜像: $service"
            cd "$service_dir"
            
            docker build -t "gaokaohub/${service}:latest" . >> "$LOG_FILE" 2>&1
            
            log_success "Docker镜像 $service 构建完成"
        else
            log_warning "Dockerfile不存在: $service"
        fi
    done
    
    log_success "所有Docker镜像构建完成"
}

# 生成部署文件
generate_deployment_files() {
    log "生成部署文件..."
    
    # 创建docker-compose部署文件
    cat > "${BUILD_DIR}/docker-compose.prod.yml" << 'EOF'
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: gaokaohub
      POSTGRES_USER: gaokaohub
      POSTGRES_PASSWORD: gaokaohub123
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - gaokaohub-network

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    networks:
      - gaokaohub-network

  elasticsearch:
    image: elasticsearch:8.11.0
    environment:
      - discovery.type=single-node
      - "ES_JAVA_OPTS=-Xms512m -Xmx512m"
      - xpack.security.enabled=false
    volumes:
      - elasticsearch_data:/usr/share/elasticsearch/data
    networks:
      - gaokaohub-network

  api-gateway:
    image: gaokaohub/api-gateway:latest
    ports:
      - "8080:8080"
    environment:
      - GIN_MODE=release
      - DB_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis
    networks:
      - gaokaohub-network

  user-service:
    image: gaokaohub/user-service:latest
    ports:
      - "8081:8081"
    environment:
      - GIN_MODE=release
      - DB_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis
    networks:
      - gaokaohub-network

  data-service:
    image: gaokaohub/data-service:latest
    ports:
      - "8082:8082"
    environment:
      - GIN_MODE=release
      - DB_HOST=postgres
      - REDIS_HOST=redis
      - ELASTICSEARCH_HOST=elasticsearch
    depends_on:
      - postgres
      - redis
      - elasticsearch
    networks:
      - gaokaohub-network

  recommendation-service:
    image: gaokaohub/recommendation-service:latest
    ports:
      - "8083:8083"
    environment:
      - GIN_MODE=release
      - DB_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis
    networks:
      - gaokaohub-network

  payment-service:
    image: gaokaohub/payment-service:latest
    ports:
      - "8084:8084"
    environment:
      - GIN_MODE=release
      - DB_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis
    networks:
      - gaokaohub-network

volumes:
  postgres_data:
  redis_data:
  elasticsearch_data:

networks:
  gaokaohub-network:
    driver: bridge
EOF

    # 创建Kubernetes部署文件
    mkdir -p "${BUILD_DIR}/k8s"
    
    # 这里可以添加更多K8s配置文件生成
    
    log_success "部署文件生成完成"
}

# 清理函数
cleanup() {
    log "清理临时文件..."
    # 可以在这里添加清理逻辑
    log_success "清理完成"
}

# 主函数
main() {
    echo "=========================================="
    echo "🏗️  高考志愿填报助手 - 构建脚本"
    echo "=========================================="
    
    # 初始化
    create_build_dir
    
    # 开始构建日志
    log "开始构建过程..."
    log "项目根目录: $PROJECT_ROOT"
    log "构建目录: $BUILD_DIR"
    
    # 检查依赖
    check_dependencies
    
    # 构建C++模块
    build_cpp_modules
    
    # 构建Go服务
    build_go_services
    
    # 运行测试
    if [ "$1" != "--skip-tests" ]; then
        run_tests
    else
        log_warning "跳过测试"
    fi
    
    # 构建Docker镜像
    if [ "$1" != "--skip-docker" ]; then
        build_docker_images
    fi
    
    # 生成部署文件
    generate_deployment_files
    
    # 完成
    log_success "构建完成！"
    echo ""
    echo "构建结果:"
    echo "  - 二进制文件: ${BUILD_DIR}/bin/"
    echo "  - C++库文件: ${BUILD_DIR}/lib/"
    echo "  - 部署文件: ${BUILD_DIR}/"
    echo "  - 构建日志: ${LOG_FILE}"
    echo ""
    echo "启动服务:"
    echo "  cd ${BUILD_DIR}"
    echo "  docker-compose -f docker-compose.prod.yml up -d"
    echo ""
}

# 捕获退出信号
trap cleanup EXIT

# 执行主函数
main "$@"