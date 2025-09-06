FROM golang:1.22-alpine AS builder

# 安装GCC和相关工具
RUN apk add --no-cache gcc g++ make cmake

WORKDIR /app

# 复制C++模块源代码
COPY cpp-modules/device-fingerprint /app/cpp-modules/device-fingerprint

# 构建C++库
RUN cd cpp-modules/device-fingerprint && \
    mkdir -p build && cd build && \
    cmake .. && \
    make && \
    mkdir -p /app/lib && \
    cp libdevice_fingerprint.so* /app/lib/ && \
    mkdir -p /app/include && \
    cp ../include/*.h /app/include/

# 复制Go服务源代码
COPY services/device-auth-service /app/services/device-auth-service

# 复制go mod和sum文件
COPY go.mod go.sum ./
RUN go mod download

# 设置CGO环境变量
ENV CGO_ENABLED=1
ENV CC=gcc
ENV CXX=g++

# 构建Go服务
RUN cd services/device-auth-service && \
    mkdir -p lib include && \
    cp /app/lib/* lib/ && \
    cp /app/include/* include/ && \
    go build -o bin/device-auth-service .

# 使用alpine镜像作为运行环境
FROM alpine:latest

# 安装ca证书和glibc兼容层
RUN apk --no-cache add ca-certificates libc6-compat

WORKDIR /app

# 从builder镜像复制编译好的应用
COPY --from=builder /app/services/device-auth-service/bin/device-auth-service ./bin/device-auth-service

# 复制C++库
COPY --from=builder /app/lib/* /usr/lib/

# 复制配置文件
COPY services/device-auth-service/config/ ./config/

# 暴露端口
EXPOSE 8085

# 运行应用
CMD ["./bin/device-auth-service"]