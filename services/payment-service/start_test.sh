#!/bin/bash

# 支付服务测试启动脚本

echo "=== 启动支付服务测试 ==="

# 设置环境变量
export $(grep -v '^#' .env.test | xargs)

# 检查是否已安装依赖
if ! command -v go &> /dev/null; then
    echo "错误: Go 未安装"
    exit 1
fi

# 构建支付服务
echo "构建支付服务..."
go build -o payment-service-test.exe main.go

if [ $? -ne 0 ]; then
    echo "构建失败"
    exit 1
fi

echo "支付服务构建成功"

# 启动服务 (后台运行)
echo "启动支付服务..."
./payment-service-test.exe &
SERVICE_PID=$!

# 等待服务启动
sleep 3

# 检查服务是否正常运行
if ps -p $SERVICE_PID > /dev/null; then
    echo "支付服务已启动 (PID: $SERVICE_PID)"
    
    # 测试健康检查
    echo "测试健康检查..."
    curl -f http://localhost:8084/health
    
    if [ $? -eq 0 ]; then
        echo "✓ 健康检查通过"
    else
        echo "✗ 健康检查失败"
        kill $SERVICE_PID
        exit 1
    fi
    
    # 测试支付渠道接口
    echo "测试支付渠道接口..."
    curl -f http://localhost:8084/api/v1/payments/channels
    
    if [ $? -eq 0 ]; then
        echo "✓ 支付渠道接口正常"
    else
        echo "✗ 支付渠道接口异常"
    fi
    
    echo "\n支付服务测试完成，服务正在运行 (PID: $SERVICE_PID)"
    echo "按 Ctrl+C 停止服务"
    
    # 等待用户中断
    wait $SERVICE_PID
else
    echo "支付服务启动失败"
    exit 1
fi