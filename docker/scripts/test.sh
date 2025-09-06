#!/bin/bash

# 高考志愿填报系统 - 测试脚本
# 用于运行各种测试

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
    local deps=("docker" "docker-compose" "curl")
    
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            log_error "$dep is not installed or not in PATH"
            exit 1
        fi
    done
    
    log_success "All dependencies are ready"
}

# 启动测试环境
start_test_environment() {
    log_info "Starting test environment..."
    
    # 停止并清理现有测试容器
    docker-compose -f docker/test/docker-compose.test.yml down --remove-orphans --volumes || true
    
    # 启动测试环境
    docker-compose -f docker/test/docker-compose.test.yml up -d postgres-test redis-test cpp-modules-test
    
    # 等待基础服务就绪
    log_info "Waiting for database and cache services..."
    sleep 30
    
    # 启动应用服务
    docker-compose -f docker/test/docker-compose.test.yml up -d user-service-test api-gateway-test
    
    # 等待应用服务就绪
    log_info "Waiting for application services..."
    sleep 30
    
    log_success "Test environment started"
}

# 停止测试环境
stop_test_environment() {
    log_info "Stopping test environment..."
    docker-compose -f docker/test/docker-compose.test.yml down --remove-orphans --volumes
    log_success "Test environment stopped"
}

# 运行单元测试
run_unit_tests() {
    log_info "Running unit tests..."
    
    local test_results_dir="test-results/unit"
    mkdir -p "$test_results_dir"
    
    # Go服务单元测试
    log_info "Running Go unit tests..."
    
    # API Gateway单元测试
    cd services/api-gateway
    go test -v -race -coverprofile="../../${test_results_dir}/api-gateway-coverage.out" ./... 2>&1 | tee "../../${test_results_dir}/api-gateway-tests.log"
    go tool cover -html="../../${test_results_dir}/api-gateway-coverage.out" -o "../../${test_results_dir}/api-gateway-coverage.html"
    cd ../..
    
    # User Service单元测试
    cd services/user-service
    go test -v -race -coverprofile="../../${test_results_dir}/user-service-coverage.out" ./... 2>&1 | tee "../../${test_results_dir}/user-service-tests.log"
    go tool cover -html="../../${test_results_dir}/user-service-coverage.out" -o "../../${test_results_dir}/user-service-coverage.html"
    cd ../..
    
    # C++模块单元测试
    log_info "Running C++ unit tests..."
    cd cpp-modules/device-fingerprint
    if [ -d "build" ]; then
        cd build
        if [ -f "test_device_fingerprint" ]; then
            ./test_device_fingerprint 2>&1 | tee "../../../${test_results_dir}/cpp-modules-tests.log"
        else
            log_warning "C++ test executable not found"
        fi
        cd ..
    else
        log_warning "C++ build directory not found"
    fi
    cd ../..
    
    log_success "Unit tests completed"
}

# 运行集成测试
run_integration_tests() {
    log_info "Running integration tests..."
    
    local test_results_dir="test-results/integration"
    mkdir -p "$test_results_dir"
    
    # 确保测试环境运行
    if ! docker-compose -f docker/test/docker-compose.test.yml ps | grep -q "Up"; then
        log_error "Test environment is not running. Please start it first."
        return 1
    fi
    
    # 等待服务就绪
    wait_for_test_services
    
    # 运行集成测试
    docker-compose -f docker/test/docker-compose.test.yml up --exit-code-from test-runner test-runner
    
    # 复制测试结果
    docker cp gaokao-test-runner:/test-results/. "$test_results_dir/"
    
    log_success "Integration tests completed"
}

# 运行API测试
run_api_tests() {
    log_info "Running API tests..."
    
    local test_results_dir="test-results/api"
    mkdir -p "$test_results_dir"
    
    # 运行Newman API测试
    if [ -f "tests/api/gaokao-api-tests.json" ]; then
        docker-compose -f docker/test/docker-compose.test.yml up --exit-code-from newman-test newman-test
        
        # 复制测试结果
        docker cp gaokao-newman-test:/var/newman-results/. "$test_results_dir/"
    else
        log_warning "API test collection not found, skipping API tests"
    fi
    
    log_success "API tests completed"
}

# 运行性能测试
run_performance_tests() {
    log_info "Running performance tests..."
    
    local test_results_dir="test-results/performance"
    mkdir -p "$test_results_dir"
    
    # 运行K6性能测试
    if [ -f "tests/performance/load-test.js" ]; then
        docker-compose -f docker/test/docker-compose.test.yml up -d influxdb-test
        sleep 10
        
        docker-compose -f docker/test/docker-compose.test.yml up --exit-code-from k6-test k6-test
        
        # 复制测试结果
        docker cp gaokao-k6-test:/results/. "$test_results_dir/"
    else
        log_warning "Performance test scripts not found, skipping performance tests"
    fi
    
    log_success "Performance tests completed"
}

# 运行安全测试
run_security_tests() {
    log_info "Running security tests..."
    
    local test_results_dir="test-results/security"
    mkdir -p "$test_results_dir"
    
    # 运行镜像安全扫描
    if command -v trivy &> /dev/null; then
        log_info "Scanning Docker images for vulnerabilities..."
        
        for image in gaokao/api-gateway:dev gaokao/user-service:dev gaokao/cpp-modules:dev; do
            if docker images -q "$image" > /dev/null; then
                trivy image --format json --output "$test_results_dir/${image//\//-}-scan.json" "$image"
                trivy image --format table "$image" | tee "$test_results_dir/${image//\//-}-scan.txt"
            fi
        done
    else
        log_warning "Trivy not installed, skipping security scans"
    fi
    
    # 运行OWASP ZAP安全测试（如果可用）
    if command -v zap-baseline.py &> /dev/null; then
        log_info "Running OWASP ZAP baseline scan..."
        zap-baseline.py -t http://localhost:8080 -J "$test_results_dir/zap-report.json" || true
    else
        log_warning "OWASP ZAP not available, skipping web security tests"
    fi
    
    log_success "Security tests completed"
}

# 等待测试服务就绪
wait_for_test_services() {
    local timeout=180
    local count=0
    
    log_info "Waiting for test services to be ready..."
    
    local services=("http://localhost:8080/health" "http://localhost:8081/health")
    
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
    
    log_success "All test services are ready"
}

# 生成测试报告
generate_test_report() {
    log_info "Generating test report..."
    
    local report_file="test-results/test-report.html"
    
    cat > "$report_file" << EOF
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>高考志愿填报系统 - 测试报告</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f4f4f4; padding: 20px; border-radius: 5px; }
        .section { margin: 20px 0; padding: 15px; border-left: 4px solid #007cba; }
        .success { border-left-color: #28a745; }
        .warning { border-left-color: #ffc107; }
        .error { border-left-color: #dc3545; }
        table { width: 100%; border-collapse: collapse; margin: 10px 0; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        .timestamp { color: #666; font-size: 0.9em; }
    </style>
</head>
<body>
    <div class="header">
        <h1>高考志愿填报系统 - 测试报告</h1>
        <p class="timestamp">生成时间: $(date)</p>
    </div>
    
    <div class="section success">
        <h2>测试概览</h2>
        <table>
            <tr><th>测试类型</th><th>状态</th><th>结果文件</th></tr>
            <tr><td>单元测试</td><td>完成</td><td><a href="unit/">查看详情</a></td></tr>
            <tr><td>集成测试</td><td>完成</td><td><a href="integration/">查看详情</a></td></tr>
            <tr><td>API测试</td><td>完成</td><td><a href="api/">查看详情</a></td></tr>
            <tr><td>性能测试</td><td>完成</td><td><a href="performance/">查看详情</a></td></tr>
            <tr><td>安全测试</td><td>完成</td><td><a href="security/">查看详情</a></td></tr>
        </table>
    </div>
    
    <div class="section">
        <h2>覆盖率报告</h2>
        <ul>
            <li><a href="unit/api-gateway-coverage.html">API Gateway 覆盖率</a></li>
            <li><a href="unit/user-service-coverage.html">User Service 覆盖率</a></li>
        </ul>
    </div>
    
    <div class="section">
        <h2>详细日志</h2>
        <ul>
            <li><a href="unit/api-gateway-tests.log">API Gateway 测试日志</a></li>
            <li><a href="unit/user-service-tests.log">User Service 测试日志</a></li>
            <li><a href="unit/cpp-modules-tests.log">C++ Modules 测试日志</a></li>
        </ul>
    </div>
</body>
</html>
EOF
    
    log_success "Test report generated: $report_file"
}

# 清理测试资源
cleanup_test_resources() {
    log_info "Cleaning up test resources..."
    
    # 停止并删除测试容器
    docker-compose -f docker/test/docker-compose.test.yml down --remove-orphans --volumes
    
    # 删除测试镜像（可选）
    if [ "$CLEANUP_IMAGES" = "true" ]; then
        docker images | grep test | awk '{print $3}' | xargs -r docker rmi -f
    fi
    
    log_success "Test resources cleaned up"
}

# 显示帮助信息
show_help() {
    cat << EOF
高考志愿填报系统 - 测试脚本

用法: $0 [选项] [测试类型]

选项:
    -h, --help          显示此帮助信息
    -s, --start-env     启动测试环境
    -e, --stop-env      停止测试环境
    -c, --cleanup       测试后清理资源
    -r, --report        生成测试报告
    --cleanup-images    清理测试镜像

测试类型:
    unit                运行单元测试
    integration         运行集成测试
    api                 运行API测试
    performance         运行性能测试
    security            运行安全测试
    all                 运行所有测试 (默认)

示例:
    $0                  # 运行所有测试
    $0 unit             # 只运行单元测试
    $0 -s               # 只启动测试环境
    $0 -e               # 只停止测试环境
    $0 -c all           # 运行所有测试并清理

EOF
}

# 主函数
main() {
    local test_type="all"
    local start_env=false
    local stop_env=false
    local cleanup=false
    local generate_report=false
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -s|--start-env)
                start_env=true
                shift
                ;;
            -e|--stop-env)
                stop_env=true
                shift
                ;;
            -c|--cleanup)
                cleanup=true
                shift
                ;;
            -r|--report)
                generate_report=true
                shift
                ;;
            --cleanup-images)
                CLEANUP_IMAGES=true
                shift
                ;;
            unit|integration|api|performance|security|all)
                test_type="$1"
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
    
    # 创建测试结果目录
    mkdir -p test-results
    
    # 只启动测试环境
    if [ "$start_env" = true ] && [ "$stop_env" = false ]; then
        start_test_environment
        exit 0
    fi
    
    # 只停止测试环境
    if [ "$stop_env" = true ] && [ "$start_env" = false ]; then
        stop_test_environment
        exit 0
    fi
    
    # 记录开始时间
    local start_time=$(date +%s)
    
    # 启动测试环境
    start_test_environment
    
    # 运行测试
    case $test_type in
        unit)
            run_unit_tests
            ;;
        integration)
            run_integration_tests
            ;;
        api)
            run_api_tests
            ;;
        performance)
            run_performance_tests
            ;;
        security)
            run_security_tests
            ;;
        all)
            run_unit_tests
            run_integration_tests
            run_api_tests
            run_performance_tests
            run_security_tests
            ;;
    esac
    
    # 生成测试报告
    if [ "$generate_report" = true ] || [ "$test_type" = "all" ]; then
        generate_test_report
    fi
    
    # 清理资源
    if [ "$cleanup" = true ]; then
        cleanup_test_resources
    else
        stop_test_environment
    fi
    
    # 计算测试时间
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    log_success "Testing completed in ${duration} seconds"
    log_info "Test results available in: test-results/"
}

# 运行主函数
main "$@"