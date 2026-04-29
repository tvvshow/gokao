# syntax=docker/dockerfile:1

# --- Builder stage ---
FROM golang:1.25-alpine AS builder
WORKDIR /app/services/api-gateway

# Enable Go modules and download deps first (better layer caching)
COPY services/api-gateway/go.mod services/api-gateway/go.sum ./
COPY pkg/auth /app/pkg/auth
COPY pkg/errors /app/pkg/errors
COPY pkg/middleware /app/pkg/middleware
RUN go mod download

# Copy the rest of the source
COPY services/api-gateway ./

# Build static binary for linux/amd64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o /out/api-gateway

# --- Runtime stage ---
FROM alpine:3.20
RUN adduser -D -H app && apk --no-cache add ca-certificates
WORKDIR /home/app

COPY --from=builder /out/api-gateway /usr/local/bin/api-gateway

# Default runtime envs. In production, set ENABLE_SWAGGER=0 to disable Swagger UI.
ENV PORT=8080 \
    GIN_MODE=release

EXPOSE 8080
USER app
ENTRYPOINT ["/usr/local/bin/api-gateway"]
