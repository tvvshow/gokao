# 构建阶段
FROM golang:1.25-alpine AS builder

# 安装构建工具（CGO 链接 C++ 推荐引擎所需）
RUN apk add --no-cache \
    gcc \
    g++ \
    make \
    git \
    linux-headers

# 设置工作目录
WORKDIR /app

# 复制Go模块文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建Go应用
# 注意：libvolunteer_matcher.so 必须由宿主 cmake 预构建并通过 docker-compose
# volumes 挂载到 /app/cpp-modules（见 docker-compose.yml）
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o recommendation-service .

# 运行阶段
FROM alpine:3.20

# 安装运行时依赖
RUN apk add --no-cache \
    libstdc++ \
    libgcc \
    ca-certificates \
    tzdata \
    curl

# 设置时区
ENV TZ=Asia/Shanghai

# 创建应用用户
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制和配置
COPY --from=builder /app/recommendation-service .
COPY --from=builder /app/config ./config

# 创建必要的目录
RUN mkdir -p logs && chown appuser:appgroup logs

# 设置权限
RUN chown -R appuser:appgroup /app
USER appuser

# 暴露端口（与 docker-compose.yml 一致）
EXPOSE 8084

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8084/health || exit 1

# 启动命令
CMD ["./recommendation-service"]