# 多阶段构建 Dockerfile
# 阶段1: 构建Go后端服务
FROM golang:1.21-alpine AS go-builder

WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git ca-certificates tzdata

# 复制go mod文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY services/ ./services/

# 构建所有Go服务
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/api-gateway ./services/api-gateway
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/user-service ./services/user-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/data-service ./services/data-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/payment-service ./services/payment-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/recommendation-service ./services/recommendation-service

# 阶段2: 构建前端
FROM node:18-alpine AS frontend-builder

WORKDIR /app/frontend

# 复制package文件
COPY frontend/package*.json ./
RUN npm ci --only=production

# 复制前端源代码
COPY frontend/ ./

# 构建前端
RUN npm run build

# 阶段3: 最终运行镜像
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=go-builder /app/bin/ ./bin/

# 从前端构建阶段复制静态文件
COPY --from=frontend-builder /app/frontend/dist/ ./frontend/dist/

# 复制配置文件
COPY config/ ./config/

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置权限
RUN chown -R appuser:appgroup /root/

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080 8081 8082 8083 8084 3000

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 启动脚本
COPY docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["all"]
