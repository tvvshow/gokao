# Monitoring 配置说明

当前仓库已提供 Prometheus 抓取配置与基础告警规则：

- `monitoring/prometheus.yml`
- `monitoring/alerts/gaokao-alerts.yml`

## 已对齐端口

Prometheus 抓取目标已按当前服务默认端口统一：

- `api-gateway:8080`
- `data-service:8082`
- `user-service:8083`
- `recommendation-service:8084`
- `payment-service:8085`

## 当前边界

仓库根 `docker-compose.yml` 目前会启动业务服务与 `monitoring-service`，但**不会自动拉起 Prometheus / Alertmanager / Grafana**。

因此，`monitoring/` 目录下的配置目前属于：

1. 外部 Prometheus 部署时直接挂载使用的配置
2. 后续补充监控编排时的基线配置

## 使用建议

至少保证以下两项同时成立：

1. Prometheus 挂载 `monitoring/prometheus.yml`
2. 规则目录挂载 `monitoring/alerts/`

这样可以直接复用现有服务抓取与告警规则，避免再次出现端口漂移。
