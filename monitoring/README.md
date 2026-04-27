# Monitoring 配置说明

当前仓库已提供可直接落地的监控基线：

- `monitoring/prometheus.yml`
- `monitoring/alerts/gaokao-alerts.yml`
- `monitoring/alertmanager.yml`

根 `docker-compose.yml` 已接入：

- `prometheus`
- `alertmanager`
- `blackbox-exporter`

## 已对齐端口与探测目标

Prometheus 当前分两类采集：

1. **原生指标抓取**
   - `api-gateway:8080/metrics`
   - `monitoring-service:8086/metrics`

2. **blackbox 健康探测**
   - `api-gateway:8080/health`
   - `user-service:8083/health`
   - `data-service:8082/health`
   - `recommendation-service:8084/health`
   - `payment-service:8085/health`
   - `monitoring-service:8086/metrics`

## 告警规则范围

当前保留的是**可被现有仓库真实支撑**的规则：

- 服务健康检查失败
- API Gateway 5xx 错误率过高
- API Gateway P95 延迟过高
- 监控抓取目标不可达

已移除依赖不存在 exporter / 指标名的旧规则，避免 Prometheus 启动后出现大量无效告警。

## 使用方式

```bash
docker compose up -d
```

访问地址：

- Prometheus: `http://localhost:9090`
- Alertmanager: `http://localhost:9093`
- Blackbox Exporter: `http://localhost:9115`

## 当前边界

- `user-service` / `data-service` / `payment-service` / `recommendation-service` 目前仍以健康探测为主，尚未统一暴露 Prometheus 原生业务指标
- `docker/prod/docker-compose.prod.yml` 仍是独立的旧生产编排，后续需要继续和当前根编排收敛
