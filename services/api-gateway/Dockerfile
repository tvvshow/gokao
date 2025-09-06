# syntax=docker/dockerfile:1

# --- Builder stage ---
<<<<<<< HEAD
FROM golang:1.23.8-alpine AS builder
=======
FROM golang:1.21-alpine AS builder
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
WORKDIR /app

# Enable Go modules and download deps first (better layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

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