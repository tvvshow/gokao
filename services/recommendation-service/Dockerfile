# 构建阶段：Go + C++ (CGO)
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache \
    gcc g++ make cmake git \
    openssl-dev jsoncpp-dev \
    linux-headers

# WORKDIR must match the go.mod location so replace directives resolve correctly
WORKDIR /app/services/recommendation-service

COPY services/recommendation-service/go.mod services/recommendation-service/go.sum ./
COPY pkg/config /app/pkg/config
COPY pkg/health /app/pkg/health
COPY pkg/middleware /app/pkg/middleware
RUN go mod download

# 构建 C++ 共享库
COPY cpp-modules/volunteer-matcher /tmp/volunteer-matcher
RUN cd /tmp/volunteer-matcher \
    && mkdir -p build && cd build \
    && cmake .. \
        -DCMAKE_BUILD_TYPE=Release \
        -DBUILD_TESTS=OFF \
        -DBUILD_EXAMPLES=OFF \
        -DBUILD_BENCHMARKS=OFF \
        -DBUILD_GO_BINDINGS=OFF \
    && make -j$(nproc)

COPY services/recommendation-service ./

# CGO: 通过 env vars 注入容器内路径，与源码 #cgo 指令互补
RUN CGO_ENABLED=1 \
    CGO_CFLAGS="-I/tmp/volunteer-matcher/include" \
    CGO_LDFLAGS="-L/tmp/volunteer-matcher/build -lvolunteer_matcher -lstdc++ -Wl,-rpath,/usr/local/lib" \
    go build -tags cppengine -ldflags="-s -w" -o recommendation-service .

# 运行阶段
FROM alpine:3.20

RUN apk add --no-cache \
    libstdc++ libgcc \
    openssl jsoncpp \
    ca-certificates tzdata

ENV TZ=Asia/Shanghai

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /tmp/volunteer-matcher/build/libvolunteer_matcher.so /usr/local/lib/
RUN cd /usr/local/lib \
    && ln -sf libvolunteer_matcher.so libvolunteer_matcher.so.1 \
    && jsoncpp_ver=$(ls /usr/lib/libjsoncpp.so.* 2>/dev/null | head -1 | sed 's|.*/||') \
    && if [ -n "$jsoncpp_ver" ] && [ ! -e /usr/lib/libjsoncpp.so.26 ]; then \
         ln -sf "$jsoncpp_ver" /usr/lib/libjsoncpp.so.26; \
       fi
COPY --from=builder /app/services/recommendation-service/recommendation-service .

RUN mkdir -p logs data && chown -R appuser:appgroup /app
USER appuser

EXPOSE 8084

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -q -O - http://127.0.0.1:8084/health || exit 1

CMD ["./recommendation-service"]
