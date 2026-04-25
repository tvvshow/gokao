#!/bin/bash

# 高考志愿填报系统远端部署脚本
# 使用方法: ./deploy-remote.sh [development|production|staging]

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置变量
ENVIRONMENT=${1:-production}
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_FILE="$PROJECT_ROOT/deploy-remote.log"
COMPOSE_FILE="$PROJECT_ROOT/docker-compose.yml"
ENV_FILE="$PROJECT_ROOT/.env.remote"

# 加载环境变量
if [ -f "$ENV_FILE" ]; then
    export $(grep -v '^#' "$ENV_FILE" | xargs)
fi

# 远端服务器配置
SERVER_USER=${SERVER_USER:-ubuntu}
SERVER_HOST=${SERVER_HOST:-gaokao.example.com}
SERVER_PORT=${SERVER_PORT:-22}
SERVER_PATH=${SERVER_PATH:-/opt/gaokao-system}

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

    # 检查SSH连接
    if ! ssh -p $SERVER_PORT -o ConnectTimeout=10 $SERVER_USER@$SERVER_HOST "echo 'SSH连接正常'" >/dev/null 2>&1; then
        log_error "无法连接到远端服务器 $SERVER_USER@$SERVER_HOST:$SERVER_PORT"
        exit 1
    fi

    # 检查本地依赖
    if ! command -v ssh &> /dev/null; then
        log_error "SSH客户端未安装"
        exit 1
    fi

    if ! command -v scp &> /dev/null; then
        log_error "SCP客户端未安装"
        exit 1
    fi

    log_success "所有依赖检查通过"
}

# 准备部署文件
prepare_deployment() {
    log_info "准备部署文件..."

    # 复制环境配置
    cp "$ENV_FILE" "$PROJECT_ROOT/.env"

    # 生成随机密码（如果未设置）
    if [ -z "$POSTGRES_PASSWORD" ]; then
        export POSTGRES_PASSWORD=$(openssl rand -base64 32)
        echo "POSTGRES_PASSWORD=$POSTGRES_PASSWORD" >> "$PROJECT_ROOT/.env"
    fi

    if [ -z "$REDIS_PASSWORD" ]; then
        export REDIS_PASSWORD=$(openssl rand -base64 32)
        echo "REDIS_PASSWORD=$REDIS_PASSWORD" >> "$PROJECT_ROOT/.env"
    fi

    if [ -z "$JWT_SECRET" ]; then
        export JWT_SECRET=$(openssl rand -base64 64)
        echo "JWT_SECRET=$JWT_SECRET" >> "$PROJECT_ROOT/.env"
    fi

    log_success "部署文件准备完成"
}

# 上传项目文件
upload_project() {
    log_info "上传项目文件到远端服务器..."

    # 创建远端目录
    ssh -p $SERVER_PORT $SERVER_USER@$SERVER_HOST "mkdir -p $SERVER_PATH"

    # 上传项目文件（排除不必要的文件）
    scp -P $SERVER_PORT -r \
        --exclude='.git' \
        --exclude='node_modules' \
        --exclude='vcpkg' \
        --exclude='build' \
        --exclude='dist' \
        --exclude='*.log' \
        --exclude='*.tmp' \
        --exclude='.DS_Store' \
        "$PROJECT_ROOT" $SERVER_USER@$SERVER_HOST:$SERVER_PATH

    log_success "项目文件上传完成"
}

# 远端部署
remote_deploy() {
    log_info "在远端服务器执行部署..."

    # 执行远端部署命令
    ssh -p $SERVER_PORT $SERVER_USER@$SERVER_HOST "cd $SERVER_PATH && chmod +x deploy.sh && ./deploy.sh $ENVIRONMENT"

    log_success "远端部署完成"
}

# 验证部署
verify_deployment() {
    log_info "验证部署结果..."

    # 等待服务启动
    sleep 30

    # 检查服务状态
    if curl -f http://$SERVER_HOST/health >/dev/null 2>&1; then
        log_success "前端服务健康检查通过"
    else
        log_warning "前端服务健康检查失败"
    fi

    # 检查API服务
    if curl -f http://$SERVER_HOST/api/health >/dev/null 2>&1; then
        log_success "API服务健康检查通过"
    else
        log_warning "API服务健康检查失败"
    fi

    log_success "部署验证完成"
}

# 显示最终信息
show_final_info() {
    echo ""
    echo "=========================================="
    log_info "🎉 远端部署完成！"
    echo ""
    log_info "访问地址："
    echo "  前端地址: http://$SERVER_HOST"
    echo "  前端地址(HTTPS): https://$SERVER_HOST"
    echo "  API文档: http://$SERVER_HOST/api/swagger/index.html"
    echo ""
    log_info "服务器信息："
    echo "  服务器: $SERVER_USER@$SERVER_HOST:$SERVER_PORT"
    echo "  项目路径: $SERVER_PATH"
    echo "  配置文件: $SERVER_PATH/.env"
    echo ""
    log_info "管理命令："
    echo "  查看日志: ssh -p $SERVER_PORT $SERVER_USER@$SERVER_HOST \"cd $SERVER_PATH && docker-compose logs -f\""
    echo "  重启服务: ssh -p $SERVER_PORT $SERVER_USER@$SERVER_HOST \"cd $SERVER_PATH && docker-compose restart\""
    echo "  停止服务: ssh -p $SERVER_PORT $SERVER_USER@$SERVER_HOST \"cd $SERVER_PATH && docker-compose down\""
    echo ""
    log_info "SSL证书配置："
    echo "  证书文件路径: $SERVER_PATH/ssl/"
    echo "  如需更新证书，请参考: $SERVER_PATH/ssl/README.md"
    echo "=========================================="
}

# 主函数
main() {
    echo "=========================================="
    log_info "开始远端部署高考志愿填报系统 - $ENVIRONMENT 环境"
    echo "=========================================="

    # 记录开始时间
    START_TIME=$(date)
    log_info "部署开始时间: $START_TIME"

    # 执行部署步骤
    check_dependencies
    prepare_deployment
    upload_project
    remote_deploy
    verify_deployment

    # 显示最终信息
    show_final_info

    # 记录完成时间
    END_TIME=$(date)
    log_success "远端部署完成时间: $END_TIME"
}

# 错误处理
trap 'log_error "部署过程中发生错误，正在清理..."; exit 1' ERR

# 执行主函数
main "$@"