#!/bin/bash

# 高考志愿填报系统 - 清理脚本
# 用于清理Docker资源

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

# 确认操作
confirm_action() {
    local message="$1"
    local default="${2:-n}"
    
    if [ "$FORCE" = true ]; then
        return 0
    fi
    
    echo -e "${YELLOW}[CONFIRM]${NC} $message"
    read -p "Are you sure? [y/N]: " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        return 0
    else
        return 1
    fi
}

# 显示系统资源使用情况
show_resource_usage() {
    log_info "Current Docker resource usage:"
    echo
    
    # 显示镜像使用情况
    echo "📦 Images:"
    docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}" | head -10
    echo
    
    # 显示容器使用情况
    echo "🚀 Containers:"
    docker ps -a --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | head -10
    echo
    
    # 显示卷使用情况
    echo "💾 Volumes:"
    docker volume ls --format "table {{.Driver}}\t{{.Name}}" | head -10
    echo
    
    # 显示网络使用情况
    echo "🌐 Networks:"
    docker network ls --format "table {{.Name}}\t{{.Driver}}\t{{.Scope}}" | head -10
    echo
    
    # 显示总体磁盘使用情况
    echo "💽 Disk Usage:"
    docker system df
    echo
}

# 停止所有项目容器
stop_project_containers() {
    log_info "Stopping all project containers..."
    
    # 停止所有环境的容器
    for env in dev prod test; do
        local compose_file="docker/${env}/docker-compose.${env}.yml"
        if [ -f "$compose_file" ]; then
            log_info "Stopping $env environment..."
            docker-compose -f "$compose_file" down --remove-orphans || true
        fi
    done
    
    # 停止其他可能的项目容器
    docker ps -a --filter "name=gaokao-*" --format "{{.Names}}" | xargs -r docker stop || true
    
    log_success "All project containers stopped"
}

# 删除项目容器
remove_project_containers() {
    if confirm_action "This will remove all project containers"; then
        log_info "Removing project containers..."
        
        # 删除所有项目相关容器
        docker ps -a --filter "name=gaokao-*" --format "{{.Names}}" | xargs -r docker rm -f || true
        
        # 删除其他可能的容器
        docker ps -a --filter "label=com.docker.compose.project=gaokao-dev" --format "{{.Names}}" | xargs -r docker rm -f || true
        docker ps -a --filter "label=com.docker.compose.project=gaokao-prod" --format "{{.Names}}" | xargs -r docker rm -f || true
        docker ps -a --filter "label=com.docker.compose.project=gaokao-test" --format "{{.Names}}" | xargs -r docker rm -f || true
        
        log_success "Project containers removed"
    else
        log_info "Skipping container removal"
    fi
}

# 删除项目镜像
remove_project_images() {
    if confirm_action "This will remove all project images"; then
        log_info "Removing project images..."
        
        # 删除项目相关镜像
        docker images --filter "reference=gaokao/*" --format "{{.ID}}" | xargs -r docker rmi -f || true
        
        # 删除悬挂镜像
        docker images --filter "dangling=true" --format "{{.ID}}" | xargs -r docker rmi -f || true
        
        log_success "Project images removed"
    else
        log_info "Skipping image removal"
    fi
}

# 删除项目数据卷
remove_project_volumes() {
    if confirm_action "This will remove all project volumes and DATA WILL BE LOST"; then
        log_info "Removing project volumes..."
        
        # 删除项目相关数据卷
        docker volume ls --filter "name=gaokao-*" --format "{{.Name}}" | xargs -r docker volume rm -f || true
        
        # 删除孤立卷
        docker volume ls --filter "dangling=true" --format "{{.Name}}" | xargs -r docker volume rm -f || true
        
        log_success "Project volumes removed"
    else
        log_info "Skipping volume removal"
    fi
}

# 删除项目网络
remove_project_networks() {
    if confirm_action "This will remove all project networks"; then
        log_info "Removing project networks..."
        
        # 删除项目相关网络
        docker network ls --filter "name=gaokao-*" --format "{{.Name}}" | xargs -r docker network rm || true
        
        log_success "Project networks removed"
    else
        log_info "Skipping network removal"
    fi
}

# 清理系统资源
cleanup_system_resources() {
    if confirm_action "This will cleanup all unused Docker resources"; then
        log_info "Cleaning up system resources..."
        
        # 清理构建缓存
        docker builder prune -f
        
        # 清理系统资源
        docker system prune -f --volumes
        
        log_success "System resources cleaned up"
    else
        log_info "Skipping system cleanup"
    fi
}

# 备份重要数据
backup_data() {
    local backup_dir="backup/$(date +%Y%m%d_%H%M%S)"
    
    if confirm_action "Do you want to backup important data before cleanup?"; then
        log_info "Creating backup..."
        
        mkdir -p "$backup_dir"
        
        # 备份数据库数据
        if docker ps --filter "name=gaokao-postgres" --format "{{.Names}}" | grep -q postgres; then
            log_info "Backing up PostgreSQL data..."
            docker exec gaokao-postgres-prod pg_dumpall -U gaokao_user > "$backup_dir/postgres_backup.sql" 2>/dev/null || \
            docker exec gaokao-postgres-dev pg_dumpall -U gaokao_user > "$backup_dir/postgres_backup.sql" 2>/dev/null || \
            log_warning "Could not backup PostgreSQL data"
        fi
        
        # 备份Redis数据
        if docker ps --filter "name=gaokao-redis" --format "{{.Names}}" | grep -q redis; then
            log_info "Backing up Redis data..."
            docker exec gaokao-redis-prod redis-cli SAVE > /dev/null 2>&1 || \
            docker exec gaokao-redis-dev redis-cli SAVE > /dev/null 2>&1 || true
            
            docker cp gaokao-redis-prod:/data/dump.rdb "$backup_dir/redis_dump.rdb" 2>/dev/null || \
            docker cp gaokao-redis-dev:/data/dump.rdb "$backup_dir/redis_dump.rdb" 2>/dev/null || \
            log_warning "Could not backup Redis data"
        fi
        
        # 备份配置文件
        log_info "Backing up configuration files..."
        cp -r docker/ "$backup_dir/" 2>/dev/null || true
        
        # 创建备份信息文件
        cat > "$backup_dir/backup_info.txt" << EOF
Backup Information
==================
Date: $(date)
Docker Version: $(docker --version)
Project: 高考志愿填报系统

Contents:
- postgres_backup.sql: PostgreSQL database dump
- redis_dump.rdb: Redis data dump
- docker/: Docker configuration files

Restore Instructions:
1. Start the Docker environment
2. Restore PostgreSQL: docker exec -i gaokao-postgres psql -U gaokao_user < postgres_backup.sql
3. Restore Redis: docker cp redis_dump.rdb gaokao-redis:/data/dump.rdb && docker restart gaokao-redis
EOF
        
        log_success "Backup created in: $backup_dir"
    else
        log_info "Skipping backup"
    fi
}

# 清理日志文件
cleanup_logs() {
    if confirm_action "This will remove all log files"; then
        log_info "Cleaning up log files..."
        
        # 清理Docker容器日志
        docker ps -a --format "{{.Names}}" | grep gaokao | xargs -I {} sh -c 'docker logs {} > /dev/null 2>&1 && echo "" > $(docker inspect --format="{{.LogPath}}" {}) 2>/dev/null' || true
        
        # 清理项目日志文件
        find . -name "*.log" -type f -delete 2>/dev/null || true
        find . -path "*/logs/*" -type f -delete 2>/dev/null || true
        
        log_success "Log files cleaned up"
    else
        log_info "Skipping log cleanup"
    fi
}

# 重置开发环境
reset_dev_environment() {
    if confirm_action "This will reset the development environment to a clean state"; then
        log_info "Resetting development environment..."
        
        # 停止开发环境
        docker-compose -f docker/dev/docker-compose.dev.yml down --remove-orphans --volumes
        
        # 删除开发环境相关资源
        docker volume ls --filter "name=gaokao-*-dev*" --format "{{.Name}}" | xargs -r docker volume rm -f || true
        docker images --filter "reference=gaokao/*:dev" --format "{{.ID}}" | xargs -r docker rmi -f || true
        
        # 重新构建和启动
        ./docker/scripts/build.sh dev
        ./docker/scripts/deploy.sh dev
        
        log_success "Development environment reset completed"
    else
        log_info "Skipping development environment reset"
    fi
}

# 显示帮助信息
show_help() {
    cat << EOF
高考志愿填报系统 - 清理脚本

用法: $0 [选项] [操作]

选项:
    -h, --help          显示此帮助信息
    -f, --force         强制执行，不询问确认
    -b, --backup        执行清理前先备份数据
    -s, --status        只显示资源使用状态

操作:
    containers          删除项目容器
    images              删除项目镜像
    volumes             删除项目数据卷 (数据将丢失!)
    networks            删除项目网络
    logs                清理日志文件
    system              清理系统资源
    dev-reset           重置开发环境
    all                 执行完整清理 (默认)

示例:
    $0                  # 显示状态并提示清理选项
    $0 -f all           # 强制执行完整清理
    $0 -b volumes       # 备份后删除数据卷
    $0 containers       # 只删除容器
    $0 dev-reset        # 重置开发环境

⚠️  警告: 
- volumes 操作会永久删除所有数据
- 生产环境请务必先备份数据
- 建议在清理前运行 -s 查看当前状态

EOF
}

# 交互式清理菜单
interactive_cleanup() {
    show_resource_usage
    
    echo "请选择要执行的清理操作:"
    echo "1) 停止所有容器"
    echo "2) 删除项目容器"
    echo "3) 删除项目镜像"
    echo "4) 删除项目数据卷 (⚠️  数据将丢失)"
    echo "5) 删除项目网络"
    echo "6) 清理日志文件"
    echo "7) 清理系统资源"
    echo "8) 重置开发环境"
    echo "9) 完整清理 (全部操作)"
    echo "0) 退出"
    echo
    
    read -p "请输入选项 [0-9]: " choice
    
    case $choice in
        1)
            stop_project_containers
            ;;
        2)
            remove_project_containers
            ;;
        3)
            remove_project_images
            ;;
        4)
            remove_project_volumes
            ;;
        5)
            remove_project_networks
            ;;
        6)
            cleanup_logs
            ;;
        7)
            cleanup_system_resources
            ;;
        8)
            reset_dev_environment
            ;;
        9)
            if [ "$BACKUP" = true ]; then
                backup_data
            fi
            stop_project_containers
            remove_project_containers
            remove_project_images
            remove_project_volumes
            remove_project_networks
            cleanup_logs
            cleanup_system_resources
            ;;
        0)
            log_info "Exiting..."
            exit 0
            ;;
        *)
            log_error "Invalid option: $choice"
            exit 1
            ;;
    esac
}

# 主函数
main() {
    local operation=""
    local show_status=false
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -f|--force)
                FORCE=true
                shift
                ;;
            -b|--backup)
                BACKUP=true
                shift
                ;;
            -s|--status)
                show_status=true
                shift
                ;;
            containers|images|volumes|networks|logs|system|dev-reset|all)
                operation="$1"
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # 只显示状态
    if [ "$show_status" = true ]; then
        show_resource_usage
        exit 0
    fi
    
    # 检查Docker是否可用
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed or not in PATH"
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        log_error "Docker daemon is not running"
        exit 1
    fi
    
    # 备份数据（如果需要）
    if [ "$BACKUP" = true ]; then
        backup_data
    fi
    
    # 执行指定操作或交互式菜单
    case $operation in
        containers)
            stop_project_containers
            remove_project_containers
            ;;
        images)
            remove_project_images
            ;;
        volumes)
            remove_project_volumes
            ;;
        networks)
            remove_project_networks
            ;;
        logs)
            cleanup_logs
            ;;
        system)
            cleanup_system_resources
            ;;
        dev-reset)
            reset_dev_environment
            ;;
        all)
            stop_project_containers
            remove_project_containers
            remove_project_images
            remove_project_volumes
            remove_project_networks
            cleanup_logs
            cleanup_system_resources
            ;;
        "")
            interactive_cleanup
            ;;
    esac
    
    # 显示清理后的资源使用情况
    echo
    log_info "Cleanup completed. Current resource usage:"
    show_resource_usage
}

# 运行主函数
main "$@"