# 构建阶段
<<<<<<< HEAD
FROM golang:1.23.8-alpine AS builder
=======
FROM golang:1.21-alpine AS builder
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc

# 设置工作目录
WORKDIR /app

# 安装依赖
RUN apk add --no-cache git ca-certificates tzdata

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# 运行阶段
FROM alpine:latest

# 安装必要的包
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

# 创建非root用户
RUN adduser -D -s /bin/sh appuser

# 设置工作目录
WORKDIR /root/

# 从builder阶段复制可执行文件
COPY --from=builder /app/main .

# 复制配置文件（如果有）
COPY --from=builder /app/.env* ./

# 设置文件权限
RUN chown -R appuser:appuser /root/

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8082

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8082/health || exit 1

# 启动应用
CMD ["./main"]