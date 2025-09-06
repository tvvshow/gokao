#!/bin/sh

# Docker entrypoint script for Gaokao system

set -e

# 默认配置
export API_GATEWAY_PORT=${API_GATEWAY_PORT:-8080}
export USER_SERVICE_PORT=${USER_SERVICE_PORT:-8081}
export DATA_SERVICE_PORT=${DATA_SERVICE_PORT:-8082}
export PAYMENT_SERVICE_PORT=${PAYMENT_SERVICE_PORT:-8083}
export RECOMMENDATION_SERVICE_PORT=${RECOMMENDATION_SERVICE_PORT:-8084}
export FRONTEND_PORT=${FRONTEND_PORT:-3000}

# 数据库配置
export DATABASE_URL=${DATABASE_URL:-"postgres://gaokao:password@localhost:5432/gaokao_db?sslmode=disable"}
export REDIS_URL=${REDIS_URL:-"redis://localhost:6379"}

# JWT配置
export JWT_SECRET=${JWT_SECRET:-"your-super-secret-jwt-key-change-in-production"}
export JWT_EXPIRE_HOURS=${JWT_EXPIRE_HOURS:-24}

# 日志级别
export LOG_LEVEL=${LOG_LEVEL:-"info"}

# 函数：启动服务
start_service() {
    local service_name=$1
    local service_port=$2
    local service_binary="./bin/${service_name}"
    
    echo "Starting ${service_name} on port ${service_port}..."
    
    if [ -f "${service_binary}" ]; then
        ${service_binary} &
        echo "${service_name} started with PID $!"
    else
        echo "Error: ${service_binary} not found"
        exit 1
    fi
}

# 函数：等待服务启动
wait_for_service() {
    local service_name=$1
    local service_port=$2
    local max_attempts=30
    local attempt=1
    
    echo "Waiting for ${service_name} to be ready..."
    
    while [ $attempt -le $max_attempts ]; do
        if wget --quiet --tries=1 --timeout=2 --spider "http://localhost:${service_port}/health" 2>/dev/null; then
            echo "${service_name} is ready!"
            return 0
        fi
        
        echo "Attempt ${attempt}/${max_attempts}: ${service_name} not ready yet..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    echo "Error: ${service_name} failed to start within timeout"
    return 1
}

# 函数：启动前端服务
start_frontend() {
    echo "Starting frontend server on port ${FRONTEND_PORT}..."
    
    if [ -d "./frontend/dist" ]; then
        # 使用简单的HTTP服务器提供静态文件
        cd frontend/dist
        python3 -m http.server ${FRONTEND_PORT} &
        echo "Frontend server started with PID $!"
        cd ../..
    else
        echo "Error: Frontend dist directory not found"
        exit 1
    fi
}

# 函数：优雅关闭
graceful_shutdown() {
    echo "Received shutdown signal, stopping services..."
    
    # 获取所有子进程并发送TERM信号
    jobs -p | xargs -r kill -TERM
    
    # 等待进程结束
    wait
    
    echo "All services stopped"
    exit 0
}

# 设置信号处理
trap graceful_shutdown TERM INT

# 主逻辑
case "$1" in
    "api-gateway")
        start_service "api-gateway" $API_GATEWAY_PORT
        wait_for_service "api-gateway" $API_GATEWAY_PORT
        ;;
    "user-service")
        start_service "user-service" $USER_SERVICE_PORT
        wait_for_service "user-service" $USER_SERVICE_PORT
        ;;
    "data-service")
        start_service "data-service" $DATA_SERVICE_PORT
        wait_for_service "data-service" $DATA_SERVICE_PORT
        ;;
    "payment-service")
        start_service "payment-service" $PAYMENT_SERVICE_PORT
        wait_for_service "payment-service" $PAYMENT_SERVICE_PORT
        ;;
    "recommendation-service")
        start_service "recommendation-service" $RECOMMENDATION_SERVICE_PORT
        wait_for_service "recommendation-service" $RECOMMENDATION_SERVICE_PORT
        ;;
    "frontend")
        start_frontend
        ;;
    "all"|*)
        echo "Starting all services..."
        
        # 按依赖顺序启动服务
        echo "=== Starting backend services ==="
        start_service "user-service" $USER_SERVICE_PORT
        start_service "data-service" $DATA_SERVICE_PORT
        start_service "payment-service" $PAYMENT_SERVICE_PORT
        start_service "recommendation-service" $RECOMMENDATION_SERVICE_PORT
        
        # 等待后端服务启动
        sleep 5
        
        # 启动API网关
        start_service "api-gateway" $API_GATEWAY_PORT
        wait_for_service "api-gateway" $API_GATEWAY_PORT
        
        # 启动前端
        echo "=== Starting frontend ==="
        start_frontend
        
        echo "=== All services started successfully ==="
        echo "API Gateway: http://localhost:${API_GATEWAY_PORT}"
        echo "Frontend: http://localhost:${FRONTEND_PORT}"
        echo "Health check: http://localhost:${API_GATEWAY_PORT}/health"
        ;;
esac

# 保持容器运行
echo "Services are running. Press Ctrl+C to stop."
wait
