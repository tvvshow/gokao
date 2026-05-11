FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app/services/monitoring-service

COPY services/monitoring-service/go.mod services/monitoring-service/go.sum ./
COPY services/monitoring-service/internal/alerts/go.mod services/monitoring-service/internal/alerts/go.sum ./internal/alerts/
COPY services/monitoring-service/internal/metrics/go.mod ./internal/metrics/
COPY pkg/response /app/pkg/response
RUN go mod download

COPY services/monitoring-service ./
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o monitoring-service .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/services/monitoring-service/monitoring-service .

RUN chown -R appuser:appgroup /app
USER appuser

EXPOSE 8086

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -q -O - http://127.0.0.1:8086/metrics || exit 1

CMD ["./monitoring-service"]
