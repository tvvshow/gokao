# 技术债务审计与上线路线

**初版**：2026-05-10（全量审计）
**最近更新**：2026-05-11（清理 33 项 FIXED 详情、新增上线缺口章节、重组后续计划）
**审计范围**：6 微服务 + 14 pkg 模块 + C++ 算法模块 + 前端 + 部署侧

> 本文档定位为**当前活动工作清单**：已完成项目仅保留索引（commit hash + 一句话），便于历史追溯；未完成项目保留完整诊断与计划。上线侧（代码之外）缺口纳入第 3 章。

---

## 0. 状态总览

| 类别 | 总数 | ✓ FIXED | ⏸ DEFERRED | ✗ PENDING |
|------|------|---------|-----------|-----------|
| 严重 (P-01~P-10) | 10 | **10** | 0 | 0 |
| 中等 (P-11~P-25) | 15 | **14** | **1** | 0 |
| 重复代码 (A~I) | 9 | **9** | 0 | 0 |
| 算法 (Phase 4) | 5 | 0 | 0 | **5** |
| **代码层合计** | **39** | **33** | **1** | **5** |

代码层债务清完率 **85%**。剩余 5 项 PENDING 全部集中在算法侧（D.1-D.5），其中 2 项可立即自主推进，3 项需产品/数据团队 sign-off。

代码层完成 ≠ 可上线 — 见第 3 章上线缺口。

---

## 1. 已完成项目索引（33 项）

> 每项保留主要 commit + 核心收益一句话。详细修复记录见 git log / commit message。

### 1.1 严重问题（P-01 ~ P-10，10/10 已修复）

| ID | 项目 | 主要 commit | 核心收益 |
|----|------|-------------|---------|
| P-01 | LIKE 全表扫描 → pg_trgm GIN 索引 + hot_searches.keyword | eef5eb7 + 后续 | seq scan → bitmap index scan，零业务代码改动 |
| P-02 | GetUserPermissions N+1 | c237473 | JOIN 收敛单查询 |
| P-03 | RequireAuth 每请求查库 | c237473 | Redis 缓存 5min |
| P-04 | RequirePermission 空实现 | c237473 | 接入真权限校验 |
| P-05 | payment-service 双 PaymentService | 多次 | 删 dead code，service/ 与 handler/ 唯一化 |
| P-06 | c.JSON 双写 | c237473 | 一次响应一次写入 |
| P-07 | 订单号纳秒冲突 | c237473 | `%d%04d` + rand 防碰撞 |
| P-08 | 支付幂等性 | 多次 | `pkg/middleware/idempotency.go` + Redis SetNX，TTL 24h |
| P-09 | 事务路径 db=nil panic | c237473 | nil 防御 |
| P-10 | callback body 重复读取 | c237473 | 单次 buffer cache |

### 1.2 中等问题（P-11 ~ P-25，14/15 已修复，1 DEFERRED 见第 2 章）

| ID | 项目 | 主要 commit | 核心收益 |
|----|------|-------------|---------|
| P-11 | Prometheus 重复注册 | 多次 | per-router registry |
| P-12 | 缓存中间件失效 | 674ad84 | bodyWriter 包装真实响应 |
| P-13 | logrus 每请求 2x WithFields | 本次 | shared base entry + cache Debug IsLevelEnabled 守卫 |
| P-14 | rateLimiter sync.Map 无 TTL | 本次 | cleanup goroutine + 注入时钟 |
| P-15 | stats 7 查询 → 4 → 并发 errgroup | 44c0d71 + 本次 | 1 聚合 + 3 并发 GROUP BY，Session 隔离防 builder 污染 |
| P-16 | 冒泡排序 | 732b9e4 | 标准 sort.Slice |
| P-17 | search recordSearch ctx 解耦 | 本次 | timeout 父 ctx 用请求 ctx，cancel 穿透 |
| P-18 | ORDER BY SQL 注入 | 732b9e4 | 白名单 |
| P-19 | admission 多查询无事务 + 隐性 builder 污染 | 多次 | Transaction 包裹 + Session 隔离每次 Select/Group |
| P-20 | 无 HTTP Server 超时 | 732b9e4 | ReadHeader/Read/Write/Idle 全套 |
| P-21 | membership 缓存形同虚设 | 7043249 | 修 cache hit 路径 |
| P-22 | debug log Sprintf 求值（**审计 false positive**） | — | 全仓 `Debug(.*Sprintf)` 零处；logrus `Debugf` 本身自带 IsLevelEnabled 守卫 |
| P-24 | 100MB 流式导入 | 本次 | `data_processing_service.go` Process*DataStream + upsert |
| P-25 | algorithm_service saveAnalysisResult ctx 解耦 | 本次 | ctx 派生 + WithContext |

### 1.3 重复代码治理（A ~ I + C.1 ~ C.6，9/9 已修复）

| ID | 项目 | 主要 commit | 核心收益 |
|----|------|-------------|---------|
| A → C.3 | handler 错误响应 267 处 `gin.H{}` | 多次 | 全部收敛到 `pkg/response` 工厂；生产代码残留 0 处 |
| B | Config 辅助函数 | 80094dc | 各服务 `getEnv*` 委托 `pkg/config` |
| C | CORS 中间件 5 处 | 1f12cf3 | 统一到 `pkg/middleware` |
| D | RequestID/TraceID 4 处 | 1f12cf3 | 统一到 `pkg/middleware` |
| F → C.1 | BeforeCreate UUID hooks 16 处 | 多次 | `pkg/models.AssignNewUUIDIfZero` helper |
| G → C.2 | APIResponse 3 处 struct | 多次 | type alias 收敛到 `pkg/response.APIResponse`，零 wire 破坏 |
| H → C.4 | payment-service JSONB 双类型 | 多次 | 收敛到 `pkg/database/types` 单一 JSONB 实现 |
| (C.5) | monitoring + api-gateway Redis 初始化漂移 | 072a34e | `pkg/config.LoadRedis` + `pkg/database.OpenRedis`，含 ping 校验；v8 archived → v9 |
| (C.6) | SecurityHeaders 三处实现 | 多次 | `pkg/middleware.SecurityHeaders()` 包级函数，payment-service 首次真正挂上安全头 |

### 1.4 CI / 工程基础设施

| ID | 项目 | 主要 commit | 核心收益 |
|----|------|-------------|---------|
| (CI) | Node 20 deprecation | 5358b99 + 8989b9d | annotations 13 → 1（剩 golangci-lint-action 等上游升级） |
| (workspace) | go.work 17 模块整理 | 多次 | replace 直链 + transitive 修复 |

---

## 2. 待处理代码层债务（6 项）

### 2.1 ⏸ DEFERRED — B.8 / P-23 缓存 JSON 序列化优化

**为何 DEFERRED**：
1. 无 pprof 证据表明 `encoding/json.*` 在热路径占比 ≥ 5%；
2. 切换 `jsoniter` + `sync.Pool[bytes.Buffer]` 涉及 8-10 个 go.mod 新增依赖 + replace + wire-compat 回归 + streaming encoder 改写 30+ 调用点；
3. 推荐响应预算 < 500ms，标准 json marshal struct 在 µs 量级，远未触瓶颈；
4. 按 §A.2.1 "改代码前必须能用一句话说清问题；证据弱就继续诊断不要凭直觉改代码"。

**何时取消 DEFERRED**：满足以下任一
- 生产 pprof 显示 `encoding/json.*` CPU 占比 ≥ 5%
- benchmark 显示推荐 / data-service 路径 P95 接近 500ms 上限且 json 是次大头
- 引入 ClickHouse / Kafka 等高频序列化场景

**OOB 路径备忘**：
1. 封装 `pkg/jsonx`（`jsoniter.ConfigCompatibleWithStandardLibrary` + `MarshalToBuffer` / `UnmarshalFromBytes`）。
2. 缓存热路径 import 切换：`encoding/json` → `pkg/jsonx`，接口对齐零调用点改动。
3. wire-compat 回归：生产 Redis dump 样本逐条 std-json vs jsonx 字节级比对。
4. pprof 复测确认占比下降。

---

### 2.2 ✗ PENDING — D.1 录取概率改正态分布 CDF

**当前问题**：`services/recommendation-service/pkg/cppbridge/simple_rule_bridge.go:229-260` 用硬编码 7 阶梯 switch（29→30 跳 10%），省份差异忽略。

**期望**：`P = Φ((score - μ) / σ)`，μ/σ 来自该校该专业近 3 年录取数据。

**步骤**：
1. 验证 admission schema 是否已有 `avg_score` / `std_score` 字段；缺则补 SQL 计算或在查询时聚合。
2. 实现 `normalCDF(x float64) float64 { return 0.5 * (1 + math.Erf(x / math.Sqrt2)) }`。
3. 重写 `calculateAdmissionProbability`：取近 3 年记录加权平均求 μ/σ，返回 `clamp(normalCDF((score-μ)/sigma), 0.01, 0.99)`。
4. 同步改 `enhanced_rule_bridge.go:325` 同款方法。

**风险**：
- 旧测试断言概率值（"分数差 30 → 0.95"）会全部 fail，需重写期望值。
- σ 估计噪声（3 年样本太少）→ 加 σ floor（如 σ ≥ 5）。

**验证**：
- 单测：σ=10、scoreDiff=20 → 应 ≈ Φ(2.0) = 0.9772。
- 5 个典型 case 前后概率对照（确认无突变）。
- 让产品/算法同学审一组 case 是否符合直觉。

**工作量**：~6h（含测试改写）。

**前置依赖**：无 — **可立即自主推进**（纯技术修正，不破坏 wire 契约）。

---

### 2.3 ✗ PENDING — D.2 "冲稳保"三分法

**当前问题**：`services/recommendation-service/internal/handlers/simple_recommendation_handler.go:1393-1399` 仅"稳妥/适中/冲刺"且单维度，缺"保底"类。

**期望**：标准三分法 — "冲" / "稳" / "保"，按类别分组返回，至少考虑 (probability, score_gap) 两维。

**步骤**：
1. 与产品/算法同学对齐阈值：
   - 保：probability ≥ 0.85 且 scoreDiff ≥ +20
   - 稳：probability ≥ 0.55 且 scoreDiff ≥ -5
   - 冲：probability ≥ 0.20 且 scoreDiff ≥ -25
2. 改返回结构：`[]Recommendation` → `{ "rush": [...], "stable": [...], "safety": [...] }`。
3. 同步 `analyzer.go:165-174` 话术（line 174 格式化字符串"稳妥%d个，适中%d个，冲刺%d个"也要改）。
4. 前端 ts 类型 + 视图组件配套（`frontend/src/types/recommendation.ts` + `RecommendationList.vue`）。

**风险**：breaking API change，**需协调前端同步上线**。建议双写 1-2 周（同时返回 `rec_type` + `category`）平滑过渡。

**工作量**：~12h（后端 4h + 前端 6h + 测试 2h）。

**前置依赖**：**需产品 sign-off 三分法阈值**。

---

### 2.4 ✗ PENDING — D.3 ML 模块 stub 删除

**当前问题**：`services/recommendation-service/pkg/ml/ml_enhanced_engine.go` 整个文件是"占位实现"：
- `DeepLearningModel.Predict`: `base + (gap/100)*0.2` 线性占位（伪装深度学习）
- `CollaborativeFilter.GetAdjustment`: 硬返回 `0.02`
- `ContentBasedFilter.GetAdjustment`: 硬返回 `0.01`
- `ReinforcementLearner.Optimize`: `base + cf + cb` 简单加法
- `FeatureEngineering`: O(n²) 双循环 dead code

**死代码验证（2026-05-11 grep）**：
```
services/recommendation-service/pkg/ml/ml_enhanced_engine.go: 类型/方法定义
其他位置: 零调用点
```
→ `MLEnhancedRecommendationEngine` **未被 services 路径任何 main / handler 引用**，是纯 dead code。

**推荐路径**：**删除** — 真上 ML 是另一个量级的工作（数据集 + 训练 + 部署），不在本路线图。

**步骤**：
1. （已完成）grep 确认无生产调用。
2. 删除整个 `pkg/ml/` 目录。
3. 检查 `services/recommendation-service/go.mod` 是否需移除相关依赖（如 gonum 等若仅 ML 用）。
4. 确认 build/test 全绿。

**工作量**：~2h。

**前置依赖**：无 — **可立即自主推进**。

---

### 2.5 ✗ PENDING — D.4 Score / Confidence 语义冲突

**当前问题**：`generateEnhancedRecommendations`（推 reco handler 第 849 行附近）用 confidence 覆盖 Score 字段，导致 Score 原本表示"多维匹配分"的语义丢失。

**期望**：保留 Score（多维匹配）+ 新增 Confidence（概率置信度）两个独立字段。

**步骤**：
1. Read 实际 line 849 周围确认覆盖发生位置。
2. response struct 增加 `Confidence float64` 字段。
3. Score 仍由 bridge 层算多维匹配，Confidence 由 D.1 的 normalCDF 输出。
4. 前端列表卡片同时展示两值（设计同学决定 UX）。

**工作量**：~4h（**通常与 D.2 一并做** — 都涉及 wire 契约变更与前端协同）。

**前置依赖**：与 D.2 同（产品/前端 sign-off）。

---

### 2.6 ✗ PENDING — D.5 analytics 硬编码数据 → 真数据接入

**当前问题**：
- `services/recommendation-service/internal/services/analytics_service.go:336-398` `GetRecommendationTrends` 全 mock
- 同文件 `:741-763` `getTopRecommendations` 硬编码"清华大学/北京大学/上海交通大学"

**期望**：从 ClickHouse / data-service / 推荐日志聚合真数据。

**步骤**：
1. **建数据管道**：评估是否已有推荐日志表；没有则建（推荐 `recommendation_log` 表，schema：`user_id + university_id + major_id + score + probability + rec_type + timestamp`）。
2. **埋点**：在 `recommendation-service` 推荐成功路径追加异步写入。
3. **聚合**：改造为 `SELECT university_id, COUNT(*) FROM recommendation_log WHERE created_at > NOW() - INTERVAL '7 days' GROUP BY university_id ORDER BY count DESC LIMIT 10`。
4. **缓存**：Redis 5min TTL 避免每次分析请求都打 DB。

**风险**：若无推荐日志表，需要先建数据管道（应用层埋点 → 写入）—— 是基础设施级工作。

**工作量**：~1d 起（含管道建设）。

**前置依赖**：**需后端 + 数据团队商定 `recommendation_log` schema 与存储后端（PostgreSQL vs ClickHouse）**。

---

## 3. 上线侧缺口（代码层之外，2026-05-11 新增）

代码层 85% 完成，但**上线需要的不仅是代码质量**。下面是真实缺口画像，按风险降序。

### 3.1 🔴 致命 — 阻塞上线

| ID | 缺口 | 现状证据 | 影响 |
|----|------|----------|------|
| L-01 | **真实数据未注入** | `data/tasks.json` = 0 字节；DB 仅有 init.sql schema，无大学/专业/录取分数线数据 | 上线即"空壳"，推荐无法运作 |
| L-02 | **ICP 备案** | 未见任何备案号字段或文档 | 中国境内法律强制，无备案不能开 |
| L-03 | **支付商户号 + 证书** | adapter 代码完整（alipay/wechat/unionpay）；`config/.env.production` 中商户证书路径 `/run/secrets/rsa_*` 是空挂载点 | 无法收款 |
| L-04 | **真实域名 + SSL** | `CORS_ALLOWED_ORIGINS=https://yourdomain.com` 占位；无 nginx / cert-manager / Let's Encrypt 配置 | 无法对外服务 |

### 3.2 🟠 高 — 影响生产稳定性

| ID | 缺口 | 现状证据 |
|----|------|----------|
| L-05 | **K8s / 编排 manifest** | `infrastructure/` 只有 `docker/`，没有 K8s yaml / Helm chart / Argo / Kustomize |
| L-06 | **监控栈** | 代码侧 prometheus 已埋点；**没有** Prometheus + Grafana + AlertManager 部署 yaml |
| L-07 | **告警通道** | `alert_manager.go` 代码完整（dingtalk/wechat/email/webhook），但**接收方未配置** |
| L-08 | **日志聚合** | logrus 写 stdout，无 Loki / ELK / Fluent Bit pipeline |
| L-09 | **versioned DB migration** | 只有 `init.sql` + `data-service/cmd/migrator/main.go`；无 `goose` / `golang-migrate` 之类带 rollback 的迁移工具链 |
| L-10 | **secrets 管理** | `config/.env.production` 本地明文（已 gitignore，未泄漏到 repo）；未走 Vault / k8s sealed-secrets / docker secrets |
| L-11 | **压力测试** | 项目目标"10 万并发 / 推荐 < 500ms / SLA > 99.9%"**全部未实测**；无 k6 / locust / wrk 脚本 |
| L-12 | **灾备** | 无 DB 备份策略 / Redis 持久化策略 / 跨可用区方案 |

### 3.3 🟡 中 — 上线前应补

| ID | 缺口 | 现状证据 |
|----|------|----------|
| L-13 | **运维 runbook** | DEPLOYMENT_GUIDE.md 有，但无故障处置手册 / on-call rotation |
| L-14 | **法务合规** | 涉及未成年学生数据 + 高考成绩，**隐私政策 / 用户协议**未见法律审核痕迹 |
| L-15 | **等保合规** | 教育/支付双重合规要求，未见审计报告 |
| L-16 | **CDN + 前端部署** | `frontend/dist/` 已构建，但 CDN / OSS / 静态资源加速未定 |

### 3.4 🟢 已就绪 — 不构成阻塞

- ✓ 代码质量：33/39 审计项 FIXED，CI 全绿
- ✓ Docker 镜像：6 微服务 Dockerfile 完整 + 多阶段构建
- ✓ 健康检查：`/healthz` `/readyz` `/health` 齐全
- ✓ API 文档：Swagger（docs.go + swagger.json/yaml）完整
- ✓ 部署脚本：`deploy.sh` (317 行) + `deploy-frontend.sh` + `deploy-remote.sh`
- ✓ 多支付通道代码：alipay/wechat/unionpay 真实 adapter（非 stub）
- ✓ `StubPaymentAdapter` 确认**不在 factory 路径**，生产无关

---

## 4. 后续推进计划

按"是否需外部输入"分 Sprint。**Sprint A / B 可立即开干，无前置依赖**。

### Sprint A — 代码层最后一公里（~8h，全部可自主推进）

| 任务 | 工时 | 前置 |
|------|------|------|
| **D.3** ML stub 删除 | ~2h | 无（dead code 已验证） |
| **D.1** 录取概率正态 CDF | ~6h | 无（纯技术修正，不破坏 wire） |

**完成后状态**：代码层债务 33 → 35 FIXED，剩 D.2 / D.4 / D.5 三项卡外部输入。

---

### Sprint B — 生产基础设施（~3-5 工作日，技术自主可推）

| 任务 | 工时 | 产出 |
|------|------|------|
| **L-05** K8s manifest（Deployment + Service + Ingress + HPA） | ~1d | `infrastructure/k8s/*.yaml`（按目标集群定 ACK / TKE / 自建） |
| **L-06** 监控栈部署 yaml（Prometheus + Grafana + AlertManager） | ~1d | `infrastructure/monitoring/*.yaml` + ServiceMonitor + 默认 dashboard |
| **L-08** 日志聚合（Loki + Promtail） | ~0.5d | `infrastructure/logging/*.yaml` |
| **L-09** versioned migration 工具链（建议 `golang-migrate`） | ~0.5d | `services/*/migrations/*.sql` 改造 + Makefile 集成 |
| **L-11** 压测脚本（k6 起步） | ~1d | `tests/load/*.js`，覆盖推荐 / 搜索 / 鉴权三条核心路径 |

**前置选择题（**需要你决定 1 个**）**：目标部署环境是
- 自建 K8s
- 阿里云 ACK
- 腾讯云 TKE
- 其他

不同环境的 ingress / cert-manager / 存储类配置差异较大，定下来后我才能精确写 yaml。

---

### Sprint C — 等产品 / 数据团队 sign-off

| 任务 | 工时 | 前置 |
|------|------|------|
| **D.2** 冲稳保三分法 | ~12h | 产品确认 (probability, scoreDiff) 阈值 |
| **D.4** Score/Confidence 语义分离 | ~4h | 与 D.2 一起做 |
| **D.5** analytics 真数据 | ~1d+ | 后端 + 数据团队商定 `recommendation_log` schema + 存储后端 |

---

### Sprint D — 业务 / 合规侧（非代码侧，需运营/法务/商务推动）

| 任务 | 责任方 |
|------|--------|
| **L-01** 真实数据采集（大学/专业/录取分数线，2024-2026 三年） | 数据 / 内容运营 |
| **L-02** ICP 备案 | 法务 / 运营 |
| **L-03** 支付商户号申请（微信支付 / 支付宝 / 银联） + 证书入库 | 商务 / 财务 |
| **L-04** 真实域名注册 + SSL 证书签发 | 运维 / 商务 |
| **L-10** secrets 流转规范（Vault / sealed-secrets） | 安全 / 运维 |
| **L-14** 隐私政策 / 用户协议法律审核 | 法务 |
| **L-15** 等保合规审计 | 法务 / 安全 |

**这一栏不是技术能解决的** —— 列出来是为了让上线时间表能真实评估。

---

## 5. 验证基线

每个改动 PR 必须满足：
- `go test ./...` 全绿
- 相关 unit test 覆盖率 ≥ 60%（与 CI 阈值对齐）
- `go vet ./...` 0 warning
- `gosec ./...` 无新增高危
- `golangci-lint run` 通过
- **CI 必须当回合内跟到 success**（feedback_ci_watch.md 已记录）

跨服务改动（如涉及前端 wire 契约的 D.2 / D.4）需要前端 e2e 配合：
- `cd frontend && npm run test:unit && npm run test:e2e`

每次 commit 后立即 `gh run list --branch <branch> --limit 1` → `gh run view <id>` 跟到终态，红立刻当场修。

---

## 附录 A：本次复审用到的 grep 凭证

```bash
# A.1 幂等性
rg "Idempotency-Key|SetNX|SETNX" services/

# B.1 LIKE
rg "LIKE|ILIKE" services/data-service/internal/services/

# B.8 P-22 验证（结果 0 处确认 false positive）
rg "(logger|log|logrus|s\.logger|am\.logger|h\.logger)\.Debug\(.*Sprintf" services/ pkg/
rg "\.Debug\(fmt\.Sprintf" services/ pkg/

# C.1 BeforeCreate 计数
rg -c "func\s+\(\w+\s+\*\w+\)\s+BeforeCreate" services/

# C.2 / G. APIResponse 重复
rg "type\s+APIResponse\s+struct" services/

# C.4 / H. JSONB 重复
rg "type\s+\w*JSON\w*\s+|JSONB" services/payment-service/internal/models/

# C.5 Redis init 漂移
rg "redis\.NewClient\(" services/ pkg/

# D.3 ML stub 死代码确认
rg "MLEnhanced|ml_enhanced|MLEnhancedRecommendationEngine" services/

# 上线缺口扫描（2026-05-11）
rg "TODO|FIXME|XXX|HACK:" services/ pkg/                  # 结果：0 处
rg "占位|stub|placeholder" services/recommendation-service/pkg/ml/
git ls-files | grep -iE "\.env"                            # 确认无 .env.production 泄漏
find infrastructure/ -maxdepth 3 -type f                   # 确认 K8s/monitoring 缺失
```

---

## 附录 B：变更日志

| 日期 | 变更 |
|------|------|
| 2026-05-10 | 初版全量审计（39 项） |
| 2026-05-10 ~ 05-11 | 推进 C.3 / C.5 / B.8 等多项，FIXED 由 24 → 33 |
| **2026-05-11** | **本次重写**：精简已完成项详情、B.8 P-22 标记 false positive、P-23 标记 DEFERRED、新增第 3 章上线侧缺口（L-01 ~ L-16）、新增第 4 章 Sprint 计划 |
