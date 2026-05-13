# Performance Tests (k6)

L-11 落地的 k6 压测套件。覆盖 auth / data 浏览 / 推荐三条核心路径。

## SLA 目标

| 路径 | 指标 | 目标 |
|---|---|---|
| `POST /api/v1/users/auth/login` | P95 | < 200ms |
| `GET  /api/v1/data/universities` | P95 | < 150ms |
| `GET  /api/v1/data/majors?keyword=...` | P95 | < 150ms |
| `POST /api/v1/recommendations` | P95 | < 500ms |
| 全路径 | 错误率 | < 1% |

阈值写在 `load-test.js` 的 `options.thresholds`；超限 k6 退出码非零，CI 标红。

## 安装 k6

```bash
# macOS
brew install k6

# Linux (Debian)
sudo apt install k6

# Docker（无须本地装）
docker run --rm -i grafana/k6 run - < tests/performance/load-test.js
```

## 本地起跑

需要本地起 docker compose（postgres + redis + 五个服务 + api-gateway）：

```bash
docker compose up -d
sleep 30  # 等服务 ready

# 全套，使用默认 full profile
k6 run --env API_BASE_URL=http://localhost:8080 tests/performance/load-test.js

# CI smoke 档（更少 VU、30s 时长）
k6 run --env API_BASE_URL=http://localhost:8080 --env PROFILE=smoke tests/performance/load-test.js

# 只跑推荐（重点关注 < 500ms SLA）
k6 run --env API_BASE_URL=http://localhost:8080 --env SCENARIO=recommend tests/performance/load-test.js
```

## 推荐场景需要的测试账号

`recommendations` 场景需要 JWT。在环境变量里传：

```bash
export TEST_USER_EMAIL="loadtest@gaokao.dev"
export TEST_USER_PASSWORD="..."
k6 run --env API_BASE_URL=... tests/performance/load-test.js
```

无凭证时 `recommend` 场景静默跳过（不报错），其他场景仍跑。CI staging deploy 后用
`secrets.TEST_USER_EMAIL` / `TEST_USER_PASSWORD` 注入。

## profile / scenario 矩阵

环境变量：

- `PROFILE=full` (默认) — ramping-vus，peak 50/200/50 VU
- `PROFILE=smoke` — constant-vus，5/10/3 VU × 30s
- `SCENARIO=all` (默认) — 跑 auth + browse + recommend 三个
- `SCENARIO=auth | browse | recommend` — 只跑一个

组合示例：

```bash
# 仅 recommend 全量
PROFILE=full SCENARIO=recommend k6 run --env API_BASE_URL=... load-test.js
```

## CI 集成

`.github/workflows/ci-cd.yml` 的 `performance-tests` job：
- 仅在 push 到 `develop` 分支（→ staging 部署）后触发
- 用 `grafana/k6-action@v0.3.1` 直跑 `tests/performance/load-test.js`
- `API_BASE_URL` 取 `secrets.STAGING_API_URL`
- 当前 CI 默认是 full profile；若 staging 资源不够，CI 那行加 `--env PROFILE=smoke`

## 解读输出

k6 关键指标：

| 指标 | 含义 |
|---|---|
| `http_req_duration` | 请求耗时总分布（k6 内置） |
| `recommend_duration` / `browse_duration` / `auth_duration` | 业务标记后的耗时（手打 Trend，便于在 thresholds 里隔离 SLA） |
| `business_errors` | check() 失败 ratio。non-2xx + JSON 解析失败都计错 |
| `vus` / `iterations` / `iteration_duration` | k6 内置：当前 VU 数、累计迭代数、每轮平均耗时 |

通过：所有 thresholds 满足；失败：任意 threshold 超阈值 → 退出码 1。

## 后续扩展点

- `tests/performance/scenarios/` 留作单路径独立脚本（未来若 SLA 调整或需要更复杂 think-time 模型时拆分）
- 想接 Grafana 实时看板：`k6 run --out experimental-prometheus-rw`，prometheus.yml 加 remote_write
