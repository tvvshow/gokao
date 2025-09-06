#!/bin/bash

# 完整构建脚本

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 构建C++库
echo "正在构建C++库..."
cd "$SCRIPT_DIR/../cpp-modules/device-fingerprint"
mkdir -p build && cd build
cmake ..
make

# 复制库文件到服务目录
cp libdevice_fingerprint.so* ../../services/device-auth-service/lib/
cp ../include/*.h ../../services/device-auth-service/include/

# 构建Go服务
echo "正在构建Go服务..."
cd "$SCRIPT_DIR/../services/device-auth-service"
go build -o bin/device-auth-service .

echo "完整构建完成"