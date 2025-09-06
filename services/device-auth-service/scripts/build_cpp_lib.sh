# C++库构建脚本

# 创建构建目录
mkdir -p build
cd build

# 运行CMake配置
cmake ..

# 构建库
make

# 复制库文件到服务目录
cp libdevice_fingerprint.so* ../../../services/device-auth-service/lib/
cp ../include/*.h ../../../services/device-auth-service/include/

echo "C++库构建完成"