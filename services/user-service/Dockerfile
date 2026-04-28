# 多阶段构建
# 第一阶段：构建阶段
FROM golang:1.25-alpine AS builder

# 设置工作目录
WORKDIR /app/services/user-service

# 安装必要的包
RUN apk add --no-cache git ca-certificates tzdata

# 复制go mod文件
COPY services/user-service/go.mod services/user-service/go.sum ./
COPY pkg/auth /app/pkg/auth
COPY pkg/database /app/pkg/database
COPY pkg/errors /app/pkg/errors
COPY pkg/logger /app/pkg/logger

# 下载依赖
RUN go mod download

# 复制源代码
COPY services/user-service ./

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# 第二阶段：运行阶段
FROM alpine:3.20

# 安装ca-certificates和tzdata
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/services/user-service/main .

# 复制配置文件（如果有的话）
COPY --from=builder /app/services/user-service/.env.example .env.example

# 更改文件所有者
RUN chown -R appuser:appgroup /root

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8083

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8083/health || exit 1

# 运行应用
CMD ["./main"]
