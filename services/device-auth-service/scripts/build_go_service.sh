# Go服务构建脚本

# 设置CGO环境变量
export CGO_ENABLED=1

# 构建服务
go build -o bin/device-auth-service .

echo "Go服务构建完成"