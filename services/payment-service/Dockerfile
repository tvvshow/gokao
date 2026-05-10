# 多阶段构建
FROM golang:1.25-alpine AS builder

# 设置工作目录
WORKDIR /app/services/payment-service

# 安装必要的系统依赖
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev

# 复制go mod文件
COPY services/payment-service/go.mod services/payment-service/go.sum ./
COPY pkg/auth /app/pkg/auth
COPY pkg/health /app/pkg/health

# 下载依赖
RUN go mod download

# 复制源代码
COPY services/payment-service ./

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o payment-service main.go

# 运行阶段
FROM alpine:3.20

# 安装ca证书和时区数据
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/services/payment-service/payment-service .

# 创建必要的目录
RUN mkdir -p /app/logs /app/data

# 更改文件所有者
RUN chown -R appuser:appgroup /app

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8085

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8085/health || exit 1

# 运行应用
CMD ["./payment-service"]
