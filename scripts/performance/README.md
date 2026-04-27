# 性能基线脚本

当前目录提供最小可执行的 k6 基线压测脚本：

- `k6-smoke.js`

## 目标

先验证当前五个核心 HTTP 服务的基础健康探活性能，而不是直接做高强度业务压测。

默认检查：

- `api-gateway:8080/health`
- `user-service:8083/health`
- `data-service:8082/health`
- `recommendation-service:8084/health`
- `payment-service:8085/health`

## 默认阈值

- `http_req_failed < 1%`
- `http_req_duration p(95) < 500ms`

## 使用方式

```bash
k6 run scripts/performance/k6-smoke.js
```

可通过环境变量覆盖目标地址或并发：

```bash
K6_VUS=50 \
K6_DURATION=1m \
API_GATEWAY_BASE_URL=http://localhost:8080 \
USER_SERVICE_BASE_URL=http://localhost:8083 \
DATA_SERVICE_BASE_URL=http://localhost:8082 \
RECOMMENDATION_SERVICE_BASE_URL=http://localhost:8084 \
PAYMENT_SERVICE_BASE_URL=http://localhost:8085 \
k6 run scripts/performance/k6-smoke.js
```

## 说明

这是第一层健康基线，不替代后续针对推荐生成、录取查询、鉴权链路的专项压测。
