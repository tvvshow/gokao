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

## 日志聚合（L-08，docker/prod 编排）

`docker/prod/docker-compose.prod.yml` 接入 Loki + Promtail，链路：

```
gaokao-* container stdout → docker json-file driver
                          → /var/lib/docker/containers/*/-json.log
                          → Promtail (docker_sd_configs 自动发现)
                          → Loki (filesystem 后端, 14d 保留)
                          → Grafana (provisioned datasource)
```

- `monitoring/loki.yml`：Loki 单节点配置；tsdb v13 schema + filesystem chunks + 14d retention + compactor on
- `monitoring/promtail.yml`：抓 `gaokao-*` 容器，按 `container_name`/`service`/`stream`/`level` 打标
- `monitoring/grafana/provisioning/datasources/datasources.yml`：Grafana 启动自动注册 Prometheus + Loki 两个数据源

LogQL 示例（Grafana → Explore → Loki）：

```
# 单服务最近 ERROR
{container_name="gaokao-api-gateway-prod"} |= "ERROR"

# 全栈 ERROR + WARN
{container_name=~"gaokao-.*"} | level=~"ERROR|WARN"

# 推荐服务慢请求（结合 level/正则）
{container_name="gaokao-recommendation-prod"} |~ "took (\\d{4,})ms"
```

`promtail` 与 `loki` 都接 backend 内网，外部仅经 Grafana 暴露查询面。

## 当前边界

- `user-service` / `data-service` / `payment-service` / `recommendation-service` 目前仍以健康探测为主，尚未统一暴露 Prometheus 原生业务指标
- `docker/prod/docker-compose.prod.yml` 仍是独立的旧生产编排，后续需要继续和当前根编排收敛
