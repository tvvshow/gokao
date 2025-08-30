# 构建阶段
FROM golang:1.25-alpine AS builder

# 安装构建工具
RUN apk add --no-cache \
    gcc \
    g++ \
    make \
    cmake \
    git \
    linux-headers

# 设置工作目录
WORKDIR /app

# 复制Go模块文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建C++模块（如果需要）
RUN cd ../../cpp-modules/volunteer-matcher && \
    mkdir -p build && \
    cd build && \
    cmake .. && \
    make -j$(nproc)

# 构建Go应用
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o recommendation-service .

# 运行阶段
FROM alpine:3.18

# 安装运行时依赖
RUN apk add --no-cache \
    libstdc++ \
    libgcc \
    ca-certificates \
    tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建应用用户
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件和库文件
COPY --from=builder /app/recommendation-service .
COPY --from=builder /app/cpp-modules/volunteer-matcher/build/libvolunteer_matcher.so /usr/local/lib/
COPY --from=builder /app/config ./config

# 创建必要的目录
RUN mkdir -p logs && chown appuser:appgroup logs

# 设置权限
RUN chown -R appuser:appgroup /app
USER appuser

# 暴露端口
EXPOSE 8083

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8083/health || exit 1

# 启动命令
CMD ["./recommendation-service"]