#!/bin/bash

# 高考志愿填报系统 - 部署脚本
# 用于部署到不同环境

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

# 检查依赖
check_dependencies() {
    local deps=("docker" "docker-compose")
    
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            log_error "$dep is not installed or not in PATH"
            exit 1
        fi
    done
    
    if ! docker info &> /dev/null; then
        log_error "Docker daemon is not running"
        exit 1
    fi
    
    log_success "All dependencies are ready"
}

# 检查环境文件
check_env_files() {
    local env=$1
    local env_file="docker/${env}/.env"
    
    if [ ! -f "$env_file" ]; then
        log_warning "Environment file $env_file not found, creating from template..."
        
        # 创建默认环境文件
        create_default_env "$env"
    fi
    
    # 验证必要的环境变量
    if [ "$env" = "prod" ]; then
        local required_vars=("POSTGRES_PASSWORD" "REDIS_PASSWORD" "JWT_SECRET")
        for var in "${required_vars[@]}"; do
            if ! grep -q "^${var}=" "$env_file" || grep -q "^${var}=$" "$env_file"; then
                log_error "Required environment variable $var is not set in $env_file"
                log_info "Please set all required variables before deploying to production"
                exit 1
            fi
        done
    fi
    
    log_success "Environment files validated"
}

# 创建默认环境文件
create_default_env() {
    local env=$1
    local env_file="docker/${env}/.env"
    
    mkdir -p "docker/${env}"
    
    case $env in
        dev)
            cat > "$env_file" << EOF
# 开发环境配置
COMPOSE_PROJECT_NAME=gaokao-dev
VERSION=dev

# 数据库配置
POSTGRES_DB=gaokao_dev
POSTGRES_USER=gaokao_user
POSTGRES_PASSWORD=gaokao_pass

# Redis配置
REDIS_PASSWORD=

# 应用配置
JWT_SECRET=dev-jwt-secret-key
DEBUG=true
LOG_LEVEL=debug

# 端口配置
API_GATEWAY_PORT=8080
USER_SERVICE_PORT=8081
CPP_MODULE_PORT=8082
POSTGRES_PORT=5432
REDIS_PORT=6379
PGADMIN_PORT=5050
GRAFANA_PORT=3000
PROMETHEUS_PORT=9090
EOF
            ;;
        prod)
            cat > "$env_file" << EOF
# 生产环境配置
COMPOSE_PROJECT_NAME=gaokao-prod
VERSION=latest

# 数据库配置（请修改密码）
POSTGRES_DB=gaokao_prod
POSTGRES_USER=gaokao_user
POSTGRES_PASSWORD=CHANGE_ME_STRONG_PASSWORD

# Redis配置（请修改密码）
REDIS_PASSWORD=CHANGE_ME_REDIS_PASSWORD

# 应用配置（请修改密钥）
JWT_SECRET=CHANGE_ME_JWT_SECRET_KEY
DEBUG=false
LOG_LEVEL=info

# SSL/TLS配置
ENABLE_HTTPS=true
TLS_CERT_FILE=/run/secrets/tls_cert
TLS_KEY_FILE=/run/secrets/tls_key

# 监控配置
GRAFANA_ADMIN_PASSWORD=CHANGE_ME_GRAFANA_PASSWORD
EOF
            log_warning "Please update the production environment variables in $env_file before deploying!"
            ;;
        test)
            cat > "$env_file" << EOF
# 测试环境配置
COMPOSE_PROJECT_NAME=gaokao-test
VERSION=test

# 数据库配置
POSTGRES_DB=gaokao_test
POSTGRES_USER=gaokao_test_user
POSTGRES_PASSWORD=gaokao_test_pass

# Redis配置
REDIS_PASSWORD=

# 应用配置
JWT_SECRET=test-jwt-secret-key
DEBUG=true
LOG_LEVEL=debug
TEST_MODE=true

# 测试超时设置
TEST_TIMEOUT=300
EOF
            ;;
    esac
    
    log_info "Created default environment file: $env_file"
}

# 检查secrets文件（生产环境）
check_secrets() {
    if [ "$1" != "prod" ]; then
        return 0
    fi
    
    local secrets_dir="docker/prod/secrets"
    local required_secrets=("postgres_password.txt" "redis_password.txt" "jwt_secret.txt" "grafana_admin_password.txt")
    
    mkdir -p "$secrets_dir"
    
    for secret in "${required_secrets[@]}"; do
        local secret_file="$secrets_dir/$secret"
        if [ ! -f "$secret_file" ]; then
            log_warning "Creating default secret file: $secret_file"
            
            case $secret in
                postgres_password.txt)
                    openssl rand -base64 32 > "$secret_file"
                    ;;
                redis_password.txt)
                    openssl rand -base64 32 > "$secret_file"
                    ;;
                jwt_secret.txt)
                    openssl rand -base64 64 > "$secret_file"
                    ;;
                grafana_admin_password.txt)
                    openssl rand -base64 16 > "$secret_file"
                    ;;
            esac
            
            chmod 600 "$secret_file"
            log_info "Generated random secret for $secret"
        fi
    done
    
    # 检查TLS证书
    if [ ! -f "$secrets_dir/tls.crt" ] || [ ! -f "$secrets_dir/tls.key" ]; then
        log_warning "TLS certificates not found, generating self-signed certificates..."
        
        openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
            -keyout "$secrets_dir/tls.key" \
            -out "$secrets_dir/tls.crt" \
            -subj "/C=CN/ST=Beijing/L=Beijing/O=Gaokao/OU=IT/CN=gaokao.local"
        
        chmod 600 "$secrets_dir/tls.key" "$secrets_dir/tls.crt"
        log_info "Generated self-signed TLS certificates"
    fi
    
    log_success "Secrets validated"
}

# 停止现有服务
stop_services() {
    local env=$1
    local compose_file="docker/${env}/docker-compose.${env}.yml"
    
    if docker-compose -f "$compose_file" ps -q | grep -q .; then
        log_info "Stopping existing services..."
        docker-compose -f "$compose_file" down --remove-orphans
        log_success "Services stopped"
    else
        log_info "No running services found"
    fi
}

# 拉取最新镜像
pull_images() {
    local env=$1
    local compose_file="docker/${env}/docker-compose.${env}.yml"
    
    log_info "Pulling latest images..."
    docker-compose -f "$compose_file" pull --ignore-pull-failures
    log_success "Images pulled"
}

# 启动服务
start_services() {
    local env=$1
    local compose_file="docker/${env}/docker-compose.${env}.yml"
    local env_file="docker/${env}/.env"
    
    log_info "Starting services..."
    
    # 使用环境文件启动服务
    docker-compose -f "$compose_file" --env-file "$env_file" up -d
    
    log_success "Services started"
}

# 等待服务就绪
wait_for_services() {
    local env=$1
    local timeout=300
    local count=0
    
    log_info "Waiting for services to be ready..."
    
    # 定义健康检查URL
    local services=()
    case $env in
        dev)
            services=("http://localhost:8080/health" "http://localhost:8081/health")
            ;;
        prod)
            services=("http://localhost/health")
            ;;
        test)
            services=("http://localhost:8080/health" "http://localhost:8081/health")
            ;;
    esac
    
    # 等待所有服务就绪
    for service_url in "${services[@]}"; do
        count=0
        while [ $count -lt $timeout ]; do
            if curl -f "$service_url" >/dev/null 2>&1; then
                log_success "✓ $service_url is ready"
                break
            fi
            
            if [ $count -eq 0 ]; then
                log_info "Waiting for $service_url..."
            fi
            
            sleep 5
            count=$((count + 5))
        done
        
        if [ $count -ge $timeout ]; then
            log_error "Timeout waiting for $service_url"
            return 1
        fi
    done
    
    log_success "All services are ready"
}

# 运行健康检查
health_check() {
    local env=$1
    local compose_file="docker/${env}/docker-compose.${env}.yml"
    
    log_info "Running health checks..."
    
    # 显示服务状态
    docker-compose -f "$compose_file" ps
    
    # 检查服务健康状态
    local unhealthy_services=$(docker-compose -f "$compose_file" ps | grep -v "Up (healthy)" | grep "Up" | wc -l)
    
    if [ "$unhealthy_services" -gt 0 ]; then
        log_warning "Some services are not fully healthy yet"
    else
        log_success "All services are healthy"
    fi
}

# 显示部署信息
show_deployment_info() {
    local env=$1
    
    log_info "Deployment Information:"
    echo
    
    case $env in
        dev)
            cat << EOF
🌟 开发环境部署完成！

访问地址:
  • API Gateway: http://localhost:8080
  • Swagger UI: http://localhost:8080/swagger/index.html
  • User Service: http://localhost:8081
  • pgAdmin: http://localhost:5050 (admin@gaokao.dev / admin123)
  • Redis Commander: http://localhost:8081
  • Grafana: http://localhost:3000 (admin / admin123)
  • Prometheus: http://localhost:9090

数据库连接:
  • PostgreSQL: localhost:5432 (gaokao_user / gaokao_pass / gaokao_dev)
  • Redis: localhost:6379

开发工具:
  • 热重载已启用
  • 调试端口: 2345 (API Gateway), 2346 (User Service)
EOF
            ;;
        prod)
            cat << EOF
🚀 生产环境部署完成！

访问地址:
  • 主站点: https://your-domain.com
  • 监控面板: https://your-domain.com:3000 (仅内网访问)

安全提醒:
  • 请确保已修改所有默认密码
  • 请配置正确的TLS证书
  • 请配置防火墙规则
  • 请定期备份数据库
EOF
            ;;
        test)
            cat << EOF
🧪 测试环境部署完成！

测试地址:
  • API Gateway: http://localhost:8080
  • User Service: http://localhost:8081

数据库连接:
  • PostgreSQL: localhost:5433 (gaokao_test_user / gaokao_test_pass / gaokao_test)
  • Redis: localhost:6380

运行测试:
  docker-compose -f docker/test/docker-compose.test.yml exec test-runner /usr/local/bin/run-tests.sh
EOF
            ;;
    esac
    
    echo
}

# 显示帮助信息
show_help() {
    cat << EOF
高考志愿填报系统 - 部署脚本

用法: $0 [选项] [环境]

选项:
    -h, --help          显示此帮助信息
    -f, --force         强制重新部署
    -s, --skip-pull     跳过镜像拉取
    -w, --no-wait       不等待服务就绪
    -c, --check-only    仅运行健康检查

环境:
    dev                 部署开发环境 (默认)
    prod                部署生产环境
    test                部署测试环境

示例:
    $0                  # 部署开发环境
    $0 prod             # 部署生产环境
    $0 -f dev           # 强制重新部署开发环境
    $0 -c prod          # 检查生产环境状态

EOF
}

# 主函数
main() {
    local environment="dev"
    local force=false
    local skip_pull=false
    local no_wait=false
    local check_only=false
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -f|--force)
                force=true
                shift
                ;;
            -s|--skip-pull)
                skip_pull=true
                shift
                ;;
            -w|--no-wait)
                no_wait=true
                shift
                ;;
            -c|--check-only)
                check_only=true
                shift
                ;;
            dev|prod|test)
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
    
    # 检查依赖
    check_dependencies
    
    # 仅运行健康检查
    if [ "$check_only" = true ]; then
        health_check "$environment"
        exit 0
    fi
    
    # 检查环境文件和secrets
    check_env_files "$environment"
    check_secrets "$environment"
    
    # 记录开始时间
    local start_time=$(date +%s)
    
    # 强制重新部署或停止现有服务
    if [ "$force" = true ]; then
        stop_services "$environment"
    fi
    
    # 拉取最新镜像
    if [ "$skip_pull" = false ]; then
        pull_images "$environment"
    fi
    
    # 启动服务
    start_services "$environment"
    
    # 等待服务就绪
    if [ "$no_wait" = false ]; then
        wait_for_services "$environment"
    fi
    
    # 运行健康检查
    health_check "$environment"
    
    # 计算部署时间
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    # 显示部署信息
    show_deployment_info "$environment"
    
    log_success "Deployment completed in ${duration} seconds"
}

# 运行主函数
main "$@"