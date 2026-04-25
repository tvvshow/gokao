#!/bin/bash

# 高考志愿填报系统部署脚本
# 使用方法: ./deploy.sh [development|production]

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量
ENVIRONMENT=${1:-development}
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_FILE="$PROJECT_ROOT/deploy.log"
COMPOSE_FILE="$PROJECT_ROOT/docker-compose.yml"

# 日志函数
log() {
    echo -e "$1" | tee -a "$LOG_FILE"
}

log_info() {
    log "${BLUE}[INFO]${NC} $1"
}

log_success() {
    log "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    log "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    log "${RED}[ERROR]${NC} $1"
}

# 检查依赖
check_dependencies() {
    log_info "检查部署依赖..."

    # 检查Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker未安装，请先安装Docker"
        exit 1
    fi

    # 检查Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose未安装，请先安装Docker Compose"
        exit 1
    fi

    # 检查Git
    if ! command -v git &> /dev/null; then
        log_error "Git未安装，请先安装Git"
        exit 1
    fi

    log_success "所有依赖检查通过"
}

# 清理旧构建
clean_build() {
    log_info "清理旧构建..."

    # 停止并删除容器
    docker-compose -f "$COMPOSE_FILE" down --remove-orphans 2>/dev/null || true

    # 清理Docker缓存
    docker system prune -f 2>/dev/null || true

    log_success "清理完成"
}

# 拉取最新代码
pull_latest() {
    log_info "拉取最新代码..."

    cd "$PROJECT_ROOT"
    git pull origin master

    log_success "代码更新完成"
}

# 构建镜像
build_images() {
    log_info "构建Docker镜像..."

    # 构建后端服务
    log_info "构建数据服务镜像..."
    docker-compose -f "$COMPOSE_FILE" build data-service

    log_info "构建API网关镜像..."
    docker-compose -f "$COMPOSE_FILE" build api-gateway

    log_info "构建用户服务镜像..."
    docker-compose -f "$COMPOSE_FILE" build user-service

    log_info "构建推荐服务镜像..."
    docker-compose -f "$COMPOSE_FILE" build recommendation-service

    log_info "构建前端镜像..."
    docker-compose -f "$COMPOSE_FILE" build frontend

    log_success "镜像构建完成"
}

# 启动数据库
start_database() {
    log_info "启动数据库服务..."

    docker-compose -f "$COMPOSE_FILE" up -d postgres redis

    # 等待数据库就绪
    log_info "等待数据库就绪..."
    sleep 10

    # 检查数据库状态
    until docker-compose -f "$COMPOSE_FILE" exec postgres pg_isready -U postgres; do
        log_warning "等待数据库启动..."
        sleep 5
    done

    log_success "数据库启动完成"
}

# 初始化数据库
init_database() {
    log_info "初始化数据库..."

    # 执行数据库初始化脚本
    docker-compose -f "$COMPOSE_FILE" exec postgres psql -U postgres -d gaokao_db -f /docker-entrypoint-initdb.d/init.sql

    log_success "数据库初始化完成"
}

# 启动服务
start_services() {
    log_info "启动所有服务..."

    # 启动数据服务
    docker-compose -f "$COMPOSE_FILE" up -d data-service

    # 等待数据服务就绪
    log_info "等待数据服务就绪..."
    sleep 15

    # 启动其他服务
    docker-compose -f "$COMPOSE_FILE" up -d api-gateway user-service recommendation-service frontend

    log_success "所有服务启动完成"
}

# 健康检查
health_check() {
    log_info "执行健康检查..."

    # 检查服务状态
    services=("postgres" "redis" "data-service" "api-gateway" "user-service" "recommendation-service" "frontend")

    for service in "${services[@]}"; do
        if docker-compose -f "$COMPOSE_FILE" ps -q "$service" | xargs -r docker inspect --format='{{.State.Health.Status}}' 2>/dev/null | grep -q "healthy"; then
            log_success "$service 服务健康"
        else
            log_warning "$service 服务可能存在问题"
        fi
    done

    # 检查API连通性
    log_info "检查API连通性..."
    if curl -f http://localhost:8080/health >/dev/null 2>&1; then
        log_success "API网关健康检查通过"
    else
        log_warning "API网关健康检查失败"
    fi

    # 检查前端服务
    if curl -f http://localhost:3000 >/dev/null 2>&1; then
        log_success "前端服务健康检查通过"
    else
        log_warning "前端服务健康检查失败"
    fi
}

# 数据迁移
migrate_data() {
    log_info "执行数据迁移..."

    # 运行数据导入脚本
    docker-compose -f "$COMPOSE_FILE" exec data-service go run scripts/import_moe_universities.go

    log_success "数据迁移完成"
}

# 性能优化
optimize_performance() {
    log_info "执行性能优化..."

    # 数据库优化
    docker-compose -f "$COMPOSE_FILE" exec postgres psql -U postgres -d gaokao_db -c "
        ANALYZE;
        VACUUM ANALYZE;
        CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
    "

    # 清理日志
    docker system prune -f 2>/dev/null || true

    log_success "性能优化完成"
}

# 配置SSL证书
setup_ssl() {
    log_info "配置SSL证书..."

    SSL_DIR="$PROJECT_ROOT/ssl"

    # 检查SSL证书目录
    if [ ! -d "$SSL_DIR" ]; then
        mkdir -p "$SSL_DIR"
        log_warning "SSL证书目录不存在，已创建"
    fi

    # 检查证书文件
    if [ ! -f "$SSL_DIR/gaokao.example.com.crt" ] || [ ! -f "$SSL_DIR/gaokao.example.com.key" ]; then
        log_warning "SSL证书文件不存在，请按ssl/README.md配置证书"

        # 创建临时自签名证书（仅用于测试）
        log_info "创建临时测试证书..."
        openssl req -x509 -newkey rsa:4096 -keyout "$SSL_DIR/gaokao.example.com.key" -out "$SSL_DIR/gaokao.example.com.crt" -days 365 -nodes -subj "/C=CN/ST=Beijing/L=Beijing/O=GaokaoSystem/OU=IT/CN=gaokao.example.com"

        # 设置权限
        chmod 600 "$SSL_DIR/gaokao.example.com.key"
        chmod 644 "$SSL_DIR/gaokao.example.com.crt"

        log_success "临时测试证书已生成，生产环境请替换为正式证书"
    else
        log_success "SSL证书配置正确"
    fi
}

# 显示状态
show_status() {
    log_info "服务状态："
    docker-compose -f "$COMPOSE_FILE" ps

    echo ""
    log_info "服务日志："
    docker-compose -f "$COMPOSE_FILE" logs --tail=20
}

# 回滚函数
rollback() {
    log_warning "开始回滚..."

    docker-compose -f "$COMPOSE_FILE" down

    log_info "回滚完成"
}

# 主函数
main() {
    echo "=========================================="
    log_info "开始部署高考志愿填报系统 - $ENVIRONMENT 环境"
    echo "=========================================="

    # 记录开始时间
    START_TIME=$(date)
    log_info "部署开始时间: $START_TIME"

    # 执行部署步骤
    check_dependencies
    clean_build
    pull_latest
    build_images
    setup_ssl
    start_database
    init_database
    start_services
    health_check
    migrate_data
    optimize_performance

    # 显示最终状态
    show_status

    # 记录完成时间
    END_TIME=$(date)
    log_success "部署完成时间: $END_TIME"
    log_success "系统已成功部署到 $ENVIRONMENT 环境"

    # 显示访问信息
    echo ""
    echo "=========================================="
    log_info "访问信息："
    echo "  前端地址: http://gaokao.example.com"
    echo "  前端地址(HTTPS): https://gaokao.example.com"
    echo "  API文档: http://gaokao.example.com/api/swagger/index.html"
    echo "  数据库端口: 5433"
    echo "  Redis端口: 6380"
    echo ""
    log_info "SSL证书配置："
    echo "  证书目录: $PROJECT_ROOT/ssl/"
    echo "  证书文件: gaokao.example.com.crt, gaokao.example.com.key"
    echo "  如需更新证书，请参考: $PROJECT_ROOT/ssl/README.md"
    echo "=========================================="
}

# 错误处理
trap 'log_error "部署过程中发生错误"; rollback' ERR

# 执行主函数
main "$@"